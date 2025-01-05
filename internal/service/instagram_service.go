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

type InstagramService interface {
	InstagramCallback(ctx context.Context, code string, userID int64) (err error)
	RefreshInstagramToken(ctx context.Context, userID int64, refreshToken string) error
	HandleInstagramPost(ctx context.Context, post *models.Post, socialAcc *models.SocialAccount) error
}

type instagramService struct {
	cfg config.Config
	sa  repository.SocialAccountRepository
	p   repository.PostRepository
	pm  repository.PostMediaRepository
	ma  repository.MediaAssetRepository
}

func NewInstagramService(
	cfg config.Config,
	sa repository.SocialAccountRepository,
	p repository.PostRepository,
	pm repository.PostMediaRepository,
	ma repository.MediaAssetRepository) InstagramService {
	return &instagramService{
		cfg: cfg,
		sa:  sa,
		p:   p,
		pm:  pm,
		ma:  ma,
	}
}

func (ig *instagramService) InstagramCallback(ctx context.Context, code string, userID int64) (err error) {

	if code == "" {
		err = errors.New("code or state is empty")
		slog.Info(err.Error())
		return err
	}

	if userID == 0 {
		err = errors.New("User not found")
		slog.Info(err.Error())
		return err
	}

	token, err := ig.ExchangeCodeForToken(ctx, code)
	if err != nil {
		return err
	}

	userInfo, err := ig.GetInstagramUserInfo(token.LongLivedToken)
	if err != nil {
		return err
	}

	encryptedAccessToken, err := utils.Encrypt([]byte(token.AccessToken), []byte(ig.cfg.SecretKey))
	if err != nil {
		return err
	}

	accountInfo := &models.SocialAccount{
		UserID:          userID,
		Platform:        "instagram",
		AccountID:       userInfo.UserID,
		AccountName:     userInfo.Name,
		AccountUsername: userInfo.Username,
		ProfilePicture:  userInfo.ProfilePicture,
		AccessToken:     encryptedAccessToken,
		RefreshToken:    encryptedAccessToken,
		TokenExpiresAt:  token.ExpiresAt,
	}

	_, err = ig.sa.Create(ctx, nil, accountInfo)
	if err != nil {
		return err
	}

	return nil
}

func (ig *instagramService) getShortLivedToken(code string) (*transfer.InstagramToken, error) {
	// Prepare the request body
	data := url.Values{}
	data.Set("client_id", ig.cfg.InstagramClientID)
	data.Set("client_secret", ig.cfg.InstagramClientSecret)
	data.Set("grant_type", "authorization_code")
	data.Set("redirect_uri", ig.cfg.InstagramRedirectURI)
	data.Set("code", code)

	// Make the request to Instagram
	resp, err := http.Post(
		"https://api.instagram.com/oauth/access_token",
		"application/x-www-form-urlencoded",
		strings.NewReader(data.Encode()),
	)
	if err != nil {
		slog.Info(err.Error())
		return nil, fmt.Errorf("failed to get short-lived token: %v", err)
	}
	defer resp.Body.Close()

	// Parse the response
	var result struct {
		AccessToken string `json:"access_token"`
		UserID      int    `json:"user_id"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		slog.Info(err.Error())
		return nil, fmt.Errorf("failed to decode token response: %v", err)
	}

	// Create token object
	token := &transfer.InstagramToken{
		UserID:      result.UserID,
		AccessToken: result.AccessToken,
		ExpiresAt:   time.Now().Add(time.Hour),
	}

	return token, nil
}

func (ig *instagramService) getLongLivedToken(shortLivedToken string) (*struct {
	AccessToken string    `json:"access_token"`
	ExpiresAt   time.Time `json:"expires_at"`
}, error) {
	url := fmt.Sprintf(
		"https://graph.instagram.com/access_token?grant_type=ig_exchange_token&client_secret=%s&access_token=%s",
		ig.cfg.InstagramClientSecret,
		shortLivedToken,
	)

	resp, err := http.Get(url)
	if err != nil {
		slog.Info(err.Error())
		return nil, fmt.Errorf("failed to get long-lived token: %v", err)
	}
	defer resp.Body.Close()

	// Check for HTTP errors
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body) // Read the body for debugging
		return nil, fmt.Errorf("error response from Instagram: %s (status code: %d)", body, resp.StatusCode)
	}

	var result struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		ExpiresIn   int64  `json:"expires_in"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		slog.Info(err.Error())
		return nil, fmt.Errorf("failed to decode long-lived token response: %v", err)
	}

	return &struct {
		AccessToken string    `json:"access_token"`
		ExpiresAt   time.Time `json:"expires_at"`
	}{
		AccessToken: result.AccessToken,
		ExpiresAt:   time.Now().Add(time.Second * time.Duration(result.ExpiresIn)),
	}, nil
}

func (ig *instagramService) ExchangeCodeForToken(ctx context.Context, code string) (*transfer.InstagramToken, error) {

	shortLivedToken, err := ig.getShortLivedToken(code)
	if err != nil {
		return nil, fmt.Errorf("failed to get short-lived token: %v", err)
	}

	// Exchange for long-lived token
	longLivedToken, err := ig.getLongLivedToken(shortLivedToken.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get long-lived token: %v", err)
	}

	token := &transfer.InstagramToken{
		AccessToken:    longLivedToken.AccessToken,
		LongLivedToken: longLivedToken.AccessToken,
		ExpiresAt:      longLivedToken.ExpiresAt,
	}

	return token, nil
}

func (ig *instagramService) GetInstagramUserInfo(accessToken string) (*transfer.InstagramUserInfo, error) {
	var userInfo transfer.InstagramUserInfo

	reqUrl := fmt.Sprintf(
		"https://graph.instagram.com/me?fields=id,username,name,account_type,profile_picture_url&access_token=%s",
		accessToken,
	)

	resp, err := http.Get(reqUrl)
	if err != nil {
		slog.Info(err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		slog.Info(err.Error())
		return nil, err
	}
	return &userInfo, nil
}

func (s *instagramService) RefreshInstagramToken(ctx context.Context, userID int64, refreshToken string) error {

	decryptedRefreshToken, err := utils.Decrypt(refreshToken, []byte(s.cfg.SecretKey))
	if err != nil {
		return err
	}

	// Refresh long-lived token
	url := fmt.Sprintf(
		"https://graph.instagram.com/refresh_access_token?grant_type=ig_refresh_token&access_token=%s",
		decryptedRefreshToken,
	)

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var result struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int64  `json:"expires_in"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	ExpiresAt := time.Now().Add(time.Second * time.Duration(result.ExpiresIn))

	encryptedAccessToken, err := utils.Encrypt([]byte(result.AccessToken), []byte(s.cfg.SecretKey))
	if err != nil {
		return err
	}

	socialAccount := models.SocialAccount{
		AccessToken:    encryptedAccessToken,
		RefreshToken:   encryptedAccessToken,
		TokenExpiresAt: ExpiresAt,
	}

	err = s.sa.SetToken(ctx, userID, refreshToken, &socialAccount)
	if err != nil {
		return err
	}

	return nil
}

func (s *instagramService) HandleInstagramPost(ctx context.Context, post *models.Post, socialAcc *models.SocialAccount) error {
	var err error

	decryptedAccessToken, err := utils.Decrypt(socialAcc.AccessToken, []byte(s.cfg.SecretKey))
	if err != nil {
		return err
	}

	switch post.PostType {
	case "single":
		err = s.InstagramSinglePost(ctx, post.ID, socialAcc.AccountID, post.Caption, decryptedAccessToken)
		if err != nil {
			return fmt.Errorf("failed to schedule single post on Instagram: %w", err)
		}
	case "multiple":
		err := s.InstagramCarouselPost(ctx, post.ID, socialAcc.AccountID, post.Caption, decryptedAccessToken)
		if err != nil {
			return fmt.Errorf("failed to schedule carousel post on Instagram: %w", err)
		}
	}

	if err := s.p.UpdatePostStatus(ctx, models.PostStatusPosted, post.ID); err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}
	return nil
}

func (s *instagramService) InstagramSinglePost(ctx context.Context, postID int64, accountID, caption, accessToken string) error {
	url := fmt.Sprintf("https://graph.instagram.com/v21.0/%s/media", accountID)

	postMedia, err := s.pm.GetByPostID(ctx, postID)
	if err != nil {
		return fmt.Errorf("error fetching post media for PostID %d: %w", postID, err)
	}

	if postMedia == nil {
		return fmt.Errorf("no media found for PostID %d", postID)
	}

	mediaAsset, err := s.ma.GetByID(ctx, postMedia.AssetID)
	if err != nil {
		return fmt.Errorf("error retrieving media asset for AssetID %d: %w", postMedia.AssetID, err)
	}

	if mediaAsset == nil || mediaAsset.FileURL == "" {
		return fmt.Errorf("media asset is missing or incomplete for AssetID %d", postMedia.AssetID)
	}

	payload := map[string]interface{}{
		"image_url":    mediaAsset.FileURL,
		"caption":      caption,
		"access_token": accessToken,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("error marshalling payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("HTTP request error: %w", err)
	}
	if resp.Body != nil {
		defer resp.Body.Close()
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code from Instagram: %d", resp.StatusCode)
	}

	var result struct {
		ID string `json:"id"`
	}
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response body: %w", err)
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return fmt.Errorf("error parsing response: %w", err)
	}

	if result.ID == "" {
		return fmt.Errorf("no media ID returned from Instagram")
	}

	err = InstagramPublishPost(accountID, result.ID, accessToken)
	if err != nil {
		return err
	}

	return nil
}

func (s *instagramService) InstagramCarouselPost(ctx context.Context, postID int64, accountID, caption, accessToken string) error {
	url := fmt.Sprintf("https://graph.instagram.com/v21.0/%s/media", accountID)
	postMedias, err := s.pm.ListByPostID(ctx, postID)
	if err != nil {
		return fmt.Errorf("error fetching post media for PostID %d: %w", postID, err)
	}

	if postMedias == nil {
		return fmt.Errorf("no media found for PostID %d", postID)
	}

	postMediasLength := len(postMedias)

	containerIDs := make([]string, postMediasLength)

	for _, postMedia := range postMedias {
		mediaAsset, err := s.ma.GetByID(ctx, postMedia.AssetID)
		if err != nil {
			return fmt.Errorf("error retrieving media asset for AssetID %d: %w", postMedia.AssetID, err)
		}

		if mediaAsset == nil || mediaAsset.FileURL == "" {
			return fmt.Errorf("media asset is missing or incomplete for AssetID %d", postMedia.AssetID)
		}

		payload := map[string]interface{}{
			"image_url":        mediaAsset.FileURL,
			"is_carousel_item": true,
			"access_token":     accessToken,
		}
		body, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("error marshalling payload: %w", err)
		}

		req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
		if err != nil {
			return fmt.Errorf("error creating request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return fmt.Errorf("HTTP request error: %w", err)
		}
		if resp.Body != nil {
			defer resp.Body.Close()
		}

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("unexpected status code from Instagram: %d", resp.StatusCode)
		}

		var result struct {
			ID string `json:"id"`
		}
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("error reading response body: %w", err)
		}
		if err := json.Unmarshal(respBody, &result); err != nil {
			return fmt.Errorf("error parsing response: %w", err)
		}

		if result.ID == "" {
			return fmt.Errorf("no media ID returned from Instagram")
		}

		containerIDs = append(containerIDs, result.ID)
	}

	payload := map[string]interface{}{
		"media_type":   "CAROUSEL",
		"caption":      caption,
		"children":     containerIDs,
		"access_token": accessToken,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("error marshalling payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("HTTP request error: %w", err)
	}
	if resp.Body != nil {
		defer resp.Body.Close()
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code from Instagram: %d", resp.StatusCode)
	}

	var result struct {
		ID string `json:"id"`
	}
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response body: %w", err)
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return fmt.Errorf("error parsing response: %w", err)
	}

	if result.ID == "" {
		return fmt.Errorf("no media ID returned from Instagram")
	}

	err = InstagramPublishPost(accountID, result.ID, accessToken)
	if err != nil {
		return err
	}

	return nil
}

func InstagramPublishPost(accountID, mediaID, accessToken string) error {
	url := fmt.Sprintf("https://graph.instagram.com/v21.0/%s/media_publish", accountID)
	payload := map[string]string{
		"creation_id":  mediaID,
		"access_token": accessToken,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("error marshalling payload: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("HTTP request error: %w", err)
	}
	if resp.Body != nil {
		defer resp.Body.Close()
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code from Instagram: %d", resp.StatusCode)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response body: %w", err)
	}
	log.Printf("Publish response from Instagram: %s\n", string(respBody))
	return nil
}
