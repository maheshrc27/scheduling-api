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
	"net/url"
	"strings"
	"time"

	config "github.com/maheshrc27/postflow/configs"
	"github.com/maheshrc27/postflow/internal/models"
	"github.com/maheshrc27/postflow/internal/repository"
	"github.com/maheshrc27/postflow/internal/transfer"
	"github.com/maheshrc27/postflow/pkg/utils"
)

const tiktokTokenURL = "https://open.tiktokapis.com/v2/oauth/token/"

type TiktokService interface {
	TiktokCallback(ctx context.Context, code string, userID int64) (err error)
	RefreshTiktokToken(ctx context.Context, userID int64, accessToken, refreshToken string) error
	HandleTiktokPost(ctx context.Context, post *models.Post, acc *models.SocialAccount) error
}

type tiktokService struct {
	cfg config.Config
	p   repository.PostRepository
	sa  repository.SocialAccountRepository
	pm  repository.PostMediaRepository
	ma  repository.MediaAssetRepository
}

func NewTiktokService(
	cfg config.Config,
	p repository.PostRepository,
	sa repository.SocialAccountRepository,
	pm repository.PostMediaRepository,
	ma repository.MediaAssetRepository) TiktokService {
	return &tiktokService{
		cfg: cfg,
		p:   p,
		sa:  sa,
		pm:  pm,
		ma:  ma,
	}
}

func (s *tiktokService) TiktokCallback(ctx context.Context, code string, userID int64) (err error) {

	if code == "" {
		err = errors.New("code or state is empty")
		slog.Info(err.Error())
		return err
	}

	tokenResponse, err := s.exchangeCodeForToken(code)
	if err != nil {
		return err
	}

	userInfo, err := TiktokUserInfo(tokenResponse.AccessToken)
	if err != nil {
		return err
	}

	log.Printf("TikTok user info fetched: %v", userInfo)

	encryptedAccessToken, err := utils.Encrypt([]byte(tokenResponse.AccessToken), []byte(s.cfg.SecretKey))
	if err != nil {
		return err
	}

	encryptedRefreshToken, err := utils.Encrypt([]byte(tokenResponse.RefreshToken), []byte(s.cfg.SecretKey))
	if err != nil {
		return err
	}

	accountInfo := &models.SocialAccount{
		UserID:          int64(userID),
		Platform:        "tiktok",
		AccountID:       userInfo.Data.User.OpenID,
		AccountName:     userInfo.Data.User.DisplayName,
		AccountUsername: userInfo.Data.User.Username,
		ProfilePicture:  userInfo.Data.User.AvatarURL,
		AccessToken:     encryptedAccessToken,
		RefreshToken:    encryptedRefreshToken,
		TokenExpiresAt:  GetExpiresAt(tokenResponse.ExpiresIn),
	}

	_, err = s.sa.Create(ctx, nil, accountInfo)
	if err != nil {
		return err
	}

	return nil
}

func (s *tiktokService) exchangeCodeForToken(code string) (*transfer.TiktokTokenResponse, error) {
	data := url.Values{}
	data.Add("client_key", s.cfg.TiktokClientKey)
	data.Add("client_secret", s.cfg.TiktokClientSecret)
	data.Add("scopes", "user.info.basic,user.info.profile,video.publish,video.upload")
	data.Add("code", code)
	data.Add("grant_type", "authorization_code")
	data.Add("redirect_uri", s.cfg.TiktokRedirectURI)

	resp, err := http.Post(
		tiktokTokenURL,
		"application/x-www-form-urlencoded",
		strings.NewReader(data.Encode()),
	)
	if err != nil {
		slog.Info(err.Error())
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		slog.Info("TikTok token endpoint returned non-200 status")
		return nil, errors.New("TikTok token endpoint returned non-200 status")
	}

	var tokenResponse transfer.TiktokTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		slog.Info(err.Error())
		return nil, fmt.Errorf("failed to decode token response: %w", err)
	}

	return &tokenResponse, nil
}

func TiktokUserInfo(accessToken string) (*transfer.TikTokResponse, error) {
	url := "https://open.tiktokapis.com/v2/user/info/?fields=open_id,avatar_url,display_name,username"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Println("Error creating request:", err)
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		slog.Info(err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	var result transfer.TikTokResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		slog.Info(err.Error())
		return nil, err
	}

	return &result, nil
}

func (s *tiktokService) RefreshTiktokToken(ctx context.Context, userID int64, accessToken, refreshToken string) error {

	decryptedRefreshToken, err := utils.Decrypt(refreshToken, []byte(s.cfg.SecretKey))
	if err != nil {
		return err
	}
	apiURL := "https://open.tiktokapis.com/v2/oauth/token/"

	data := url.Values{}
	data.Set("client_key", s.cfg.TiktokClientKey)
	data.Set("client_secret", s.cfg.TiktokClientSecret)
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", decryptedRefreshToken)

	// Create a POST request
	req, err := http.NewRequest("POST", apiURL, bytes.NewBufferString(data.Encode()))
	if err != nil {
		return err
	}

	// Set headers
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Execute the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check for non-200 status code
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		log.Printf("%+v", bodyBytes)
		return err
	}

	// Parse response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Info(err.Error())
		return err
	}

	var tokenResponse transfer.TiktokTokenResponse
	err = json.Unmarshal(bodyBytes, &tokenResponse)
	if err != nil {
		return err
	}

	ExpiresAt := time.Now().Add(time.Second * time.Duration(tokenResponse.ExpiresIn))

	encryptedAccessToken, err := utils.Encrypt([]byte(tokenResponse.AccessToken), []byte(s.cfg.SecretKey))
	if err != nil {
		return err
	}

	encryptedRefreshToken, err := utils.Encrypt([]byte(tokenResponse.RefreshToken), []byte(s.cfg.SecretKey))
	if err != nil {
		return err
	}

	socialAccount := models.SocialAccount{
		AccessToken:    encryptedAccessToken,
		RefreshToken:   encryptedRefreshToken,
		TokenExpiresAt: ExpiresAt,
	}

	err = s.sa.SetToken(ctx, userID, accessToken, &socialAccount)
	if err != nil {
		return err
	}

	return nil
}

func (s *tiktokService) HandleTiktokPost(ctx context.Context, post *models.Post, acc *models.SocialAccount) error {
	var err error
	switch post.PostType {

	case "multiple":
		err = s.PostTiktokPhotos(ctx, post, acc)
		if err != nil {
			return err
		}
	default:
		err = s.PostTiktokVideo(ctx, post, acc)
		if err != nil {
			return err
		}
	}

	if err := s.p.UpdatePostStatus(ctx, models.PostStatusPosted, post.ID); err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}

	return nil
}

func (s *tiktokService) PostTiktokVideo(ctx context.Context, post *models.Post, acc *models.SocialAccount) error {

	decryptedAccessToken, err := utils.Decrypt(acc.AccessToken, []byte(s.cfg.SecretKey))
	if err != nil {
		return err
	}

	postMedia, err := s.pm.GetByPostID(ctx, post.ID)
	if err != nil {
		log.Printf("Error getting post media: %v", err)
		return err
	}

	videoInfo, err := s.ma.GetByID(ctx, postMedia.AssetID)
	if err != nil {
		log.Printf("Error getting asset info: %v", err)
		return err
	}

	// Set post_info
	postInfo := transfer.VideoPostInfo{
		Title:                 post.Caption,
		PrivacyLevel:          "PUBLIC_TO_EVERYONE",
		DisableDuet:           false,
		DisableComment:        false,
		DisableStitch:         false,
		VideoCoverTimestampMs: 1000,
	}

	// Set video source_info
	sourceInfo := transfer.VideoSourceInfo{
		Source:   "PULL_FROM_URL",
		VideoURL: videoInfo.FileURL,
	}

	// Prepare the request payload
	videoUploadRequest := transfer.VideoUploadRequest{
		PostInfo:   postInfo,
		SourceInfo: sourceInfo,
	}

	// Marshal the request data into JSON
	jsonData, err := json.Marshal(videoUploadRequest)
	if err != nil {
		log.Println("Error marshalling data:", err)
		return err
	}

	err = QueryCreatorInfoRequest(decryptedAccessToken)
	if err != nil {
		log.Println("Error querying creator info: ", err.Error())
		return err
	}

	// Send the request to TikTok API
	uploadURL := "https://open.tiktokapis.com/v2/post/publish/video/init/"
	req, err := http.NewRequest("POST", uploadURL, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Println("Error creating request:", err)
		return err
	}
	req.Header.Set("Authorization", "Bearer "+decryptedAccessToken)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")

	// Execute the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Error uploading video:", err)
		return err
	}
	defer resp.Body.Close()

	var result transfer.TikTokUploadResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Println(err.Error())
		return err
	}

	// Check the response status
	if resp.StatusCode != http.StatusOK {
		log.Printf("Error posting video on tiktok: %s", result.Error.Message)
		return err
	}

	log.Printf("Tiktok Publish Data: %v", result)

	return nil
}

func (s *tiktokService) PostTiktokPhotos(ctx context.Context, post *models.Post, acc *models.SocialAccount) error {
	decryptedAccessToken, err := utils.Decrypt(acc.AccessToken, []byte(s.cfg.SecretKey))
	if err != nil {
		return err
	}

	postMedias, err := s.pm.ListByPostID(ctx, post.ID)
	if err != nil {
		return err
	}

	photos := make([]string, len(postMedias))

	for _, postMedia := range postMedias {
		assetInfo, err := s.ma.GetByID(ctx, postMedia.AssetID)
		if err != nil {
			return err
		}
		photos = append(photos, assetInfo.FileURL)
	}

	postInfo := transfer.PhotoPostInfo{
		Title:                post.Caption,
		PrivacyLevel:         "PUBLIC_TO_EVERYONE",
		AutoAddMusic:         true,
		DisableComment:       false,
		BrandContentToggle:   false,
		Brand_Organic_Toggle: false,
	}

	sourceInfo := transfer.PhotoSourceInfo{
		Source:          "PULL_FROM_URL",
		PhotoCoverIndex: 1,
		PhotoImages:     photos,
	}

	photoUploadRequest := transfer.PhotUploadRequest{
		PostInfo:   postInfo,
		SourceInfo: sourceInfo,
		PostMode:   "DIRECT_POST",
		MediaType:  "PHOTO",
	}

	jsonData, err := json.Marshal(photoUploadRequest)
	if err != nil {
		log.Println("Error marshalling data:", err)
		return err
	}

	err = QueryCreatorInfoRequest(decryptedAccessToken)
	if err != nil {
		log.Println("Error querying creator info: ", err.Error())
		return err
	}

	uploadURL := "https://open.tiktokapis.com/v2/post/publish/content/init/"
	req, err := http.NewRequest("POST", uploadURL, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Println("Error creating request:", err)
		return err
	}
	req.Header.Set("Authorization", "Bearer "+decryptedAccessToken)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Error uploading video:", err)
		return err
	}
	defer resp.Body.Close()

	var result transfer.TikTokUploadResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Println(err.Error())
		return err
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("Error posting phostos on tiktok: %s", result.Error.Message)
		return err
	}

	log.Printf("Tiktok Publish Data: %v", result)

	return nil
}

func QueryCreatorInfoRequest(accessToken string) error {
	requestURL := "https://open.tiktokapis.com/v2/post/publish/creator_info/query/"
	req, err := http.NewRequest("POST", requestURL, nil)
	if err != nil {
		log.Println("Error creating request:", err)
		return err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")

	return nil
}

func RevokeTiktokAccess(openID, accessToken string) error {
	urlRevoke := "https://open-api.tiktok.com/oauth/revoke/"
	params := url.Values{}
	params.Add("open_id", openID)
	params.Add("access_token", accessToken)

	req, err := http.NewRequest("POST", urlRevoke, strings.NewReader(params.Encode()))
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

	var result transfer.TiktokRevokeData
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		slog.Info(err.Error())
		return err
	}

	log.Println("desc", result.Description)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to revoke token, status code: %s", result.Description)
	}
	return nil
}
