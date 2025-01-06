package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"

	config "github.com/maheshrc27/scheduling-api/configs"
	"github.com/maheshrc27/scheduling-api/internal/models"
	"github.com/maheshrc27/scheduling-api/internal/repository"
	"github.com/maheshrc27/scheduling-api/internal/transfer"
	"github.com/maheshrc27/scheduling-api/pkg/utils"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

type YoutubeService interface {
	YoutubeCallback(ctx context.Context, code string, userID int64) (err error)
	RefreshYoutubeToken(ctx context.Context, userID int64, accessToken, refreshToken string) error
	PostYoutubeVideo(ctx context.Context, post *models.Post, socialAcc *models.SocialAccount) error
}

type youtubeService struct {
	cfg config.Config
	p   repository.PostRepository
	sa  repository.SocialAccountRepository
	pm  repository.PostMediaRepository
	ma  repository.MediaAssetRepository
}

func NewYoutubeService(
	cfg config.Config,
	p repository.PostRepository,
	sa repository.SocialAccountRepository,
	pm repository.PostMediaRepository,
	ma repository.MediaAssetRepository) YoutubeService {
	return &youtubeService{
		cfg: cfg,
		p:   p,
		sa:  sa,
		pm:  pm,
		ma:  ma,
	}
}

func (s *youtubeService) YoutubeCallback(ctx context.Context, code string, userID int64) (err error) {

	if code == "" {
		err = errors.New("code or state is empty")
		slog.Info(err.Error())
		return err
	}

	oauth2Config := &oauth2.Config{
		ClientID:     s.cfg.GoogleClientID,
		ClientSecret: s.cfg.GoogleClientSecret,
		RedirectURL:  "http://localhost:3000/auth/youtube/callback",
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile", "https://www.googleapis.com/auth/youtube.upload"},
		Endpoint:     google.Endpoint,
	}

	if oauth2Config.ClientID == "" || oauth2Config.ClientSecret == "" || oauth2Config.RedirectURL == "" {
		err = errors.New("OAuth2 configration is incomplete")
		slog.Info(err.Error())
		return err
	}

	token, err := oauth2Config.Exchange(context.Background(), code)
	if err != nil {
		slog.Info(err.Error())
		return err
	}

	if token.RefreshToken == "" {
		err = errors.New("refresh token is empty")
		slog.Info(err.Error())
		return err
	}

	client := oauth2Config.Client(context.Background(), token)
	userInfo, err := GetUserInfo(client)
	if err != nil {
		return err
	}

	encryptedAccessToken, err := utils.Encrypt([]byte(token.AccessToken), []byte(s.cfg.SecretKey))
	if err != nil {
		return err
	}

	encryptedRefreshToken, err := utils.Encrypt([]byte(token.RefreshToken), []byte(s.cfg.SecretKey))
	if err != nil {
		return err
	}

	accountInfo := &models.SocialAccount{
		UserID:          int64(userID),
		Platform:        "youtube",
		AccountID:       userInfo.ID,
		AccountName:     userInfo.Name,
		AccountUsername: userInfo.Email,
		ProfilePicture:  userInfo.Picture,
		AccessToken:     encryptedAccessToken,
		RefreshToken:    encryptedRefreshToken,
		TokenExpiresAt:  token.Expiry,
	}

	_, err = s.sa.Create(ctx, nil, accountInfo)
	if err != nil {
		return err
	}

	return nil
}

func (s *youtubeService) RefreshYoutubeToken(ctx context.Context, userID int64, accessToken, refreshToken string) error {
	conf := &oauth2.Config{
		ClientID:     s.cfg.GoogleClientID,
		ClientSecret: s.cfg.GoogleClientSecret,
		Scopes:       []string{"https://www.googleapis.com/auth/youtube.upload"},
		Endpoint:     google.Endpoint,
	}

	decryptedRefreshToken, err := utils.Decrypt(refreshToken, []byte(s.cfg.SecretKey))
	if err != nil {
		return err
	}

	tokenSource := conf.TokenSource(context.Background(), &oauth2.Token{RefreshToken: decryptedRefreshToken})

	token, err := tokenSource.Token()
	if err != nil {
		slog.Info(err.Error())
		return err
	}

	encryptedAccessToken, err := utils.Encrypt([]byte(token.AccessToken), []byte(s.cfg.SecretKey))
	if err != nil {
		return err
	}

	socialAccount := models.SocialAccount{
		AccessToken:    encryptedAccessToken,
		TokenExpiresAt: token.Expiry,
	}

	err = s.sa.SetToken(ctx, userID, accessToken, &socialAccount)
	if err != nil {
		return err
	}

	return nil
}

func (s *youtubeService) PostYoutubeVideo(ctx context.Context, post *models.Post, socialAcc *models.SocialAccount) error {

	decryptedAccessToken, err := utils.Decrypt(socialAcc.AccessToken, []byte(s.cfg.SecretKey))
	if err != nil {
		return err
	}

	token := &oauth2.Token{
		AccessToken: decryptedAccessToken,
	}
	client := oauth2.NewClient(ctx, oauth2.StaticTokenSource(token))
	service, err := youtube.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Printf("Error creating YouTube service: %v", err)
		return err
	}

	postMedia, err := s.pm.GetByPostID(ctx, post.ID)
	if err != nil {
		return err
	}

	videoInfo, err := s.ma.GetByID(ctx, postMedia.AssetID)
	if err != nil {
		return err
	}

	uploadVideoFromS3(service, post.Caption, post.Title, videoInfo.FileURL)

	if err := s.p.UpdatePostStatus(ctx, models.PostStatusPosted, post.ID); err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}

	return nil
}

func uploadVideoFromS3(service *youtube.Service, caption, title, s3URL string) error {
	// Step 1: Download video from S3
	tempFile, err := downloadVideoFromS3(s3URL)
	if err != nil {
		log.Fatalf("Error downloading video from S3: %v", err)
	}
	defer os.Remove(tempFile) // Ensure the temporary file is deleted after use

	// Step 2: Open the downloaded video file
	file, err := os.Open(tempFile)
	if err != nil {
		log.Printf("Error opening video file: %v", err)
		return err
	}
	defer file.Close()

	// Step 3: Prepare video metadata
	video := &youtube.Video{
		Snippet: &youtube.VideoSnippet{
			Description: caption,
			Title:       title,
			CategoryId:  "22",
		},
		Status: &youtube.VideoStatus{
			PrivacyStatus: "public",
		},
	}

	// Step 4: Upload the video to YouTube
	call := service.Videos.Insert([]string{"snippet", "status"}, video)
	response, err := call.Media(file).Do()
	if err != nil {
		log.Printf("Error uploading video: %v", err)
		return err
	}

	// Step 5: Log success
	fmt.Printf("Video uploaded successfully: https://youtu.be/%s\n", response.Id)
	return nil
}

func downloadVideoFromS3(s3URL string) (string, error) {
	// Create a temporary file
	tempFile, err := os.CreateTemp("", "video-*.mp4")
	if err != nil {
		return "", fmt.Errorf("error creating temporary file: %w", err)
	}
	defer tempFile.Close()

	// Download the video
	response, err := http.Get(s3URL)
	if err != nil {
		return "", fmt.Errorf("error downloading video from S3: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected response status: %d", response.StatusCode)
	}

	// Save the video to the temporary file
	_, err = io.Copy(tempFile, response.Body)
	if err != nil {
		return "", fmt.Errorf("error saving video to temporary file: %w", err)
	}

	// Return the temporary file path
	return tempFile.Name(), nil
}

func GetUserInfo(client *http.Client) (*transfer.GoogleUserInfo, error) {
	userInfoURL := "https://www.googleapis.com/oauth2/v1/userinfo"

	response, err := client.Get(userInfoURL)
	if err != nil {
		slog.Info(err.Error())
		return nil, fmt.Errorf("error fetching user info: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		slog.Info("Unexpected response status")
		return nil, fmt.Errorf("unexpected response status: %d", response.StatusCode)
	}

	var userInfo transfer.GoogleUserInfo
	if err := json.NewDecoder(response.Body).Decode(&userInfo); err != nil {
		slog.Info(err.Error())
		return nil, fmt.Errorf("error decoding user info: %w", err)
	}

	return &userInfo, nil
}

func RevokeGoogleAccess(accessToken string) error {
	url := "https://oauth2.googleapis.com/revoke"
	payload := []byte("token=" + accessToken)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to revoke token, status code: %d", resp.StatusCode)
	}
	return nil
}
