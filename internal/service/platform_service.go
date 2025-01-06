package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/url"

	config "github.com/maheshrc27/scheduling-api/configs"
	"github.com/maheshrc27/scheduling-api/internal/models"
	"github.com/maheshrc27/scheduling-api/internal/repository"
	"github.com/maheshrc27/scheduling-api/pkg/utils"
)

const (
	TIKTOK_AUTH_URL    = "https://www.tiktok.com/v2/auth/authorize"
	GOOGLE_AUTH_URL    = "https://accounts.google.com/o/oauth2/v2/auth"
	INSTAGRAM_AUTH_URL = "https://www.instagram.com/oauth/authorize"
)

type PlatformService interface {
	GetAuthURL(ctx context.Context, platform, tokenString string) string
	List(ctx context.Context, userID int64) ([]*models.SocialAccount, error)
	Delete(ctx context.Context, userID, accountID int64) error
}

type platformService struct {
	cfg config.Config
	sa  repository.SocialAccountRepository
}

func NewPlatformService(cfg config.Config, sa repository.SocialAccountRepository) PlatformService {
	return &platformService{
		cfg: cfg,
		sa:  sa,
	}
}

func (s *platformService) GetAuthURL(ctx context.Context, platform, tokenString string) string {
	switch platform {
	case "instagram":
		authURL := INSTAGRAM_AUTH_URL
		params := url.Values{}
		params.Add("client_id", s.cfg.InstagramClientID)
		params.Add("scope", "instagram_business_basic,instagram_business_content_publish")
		params.Add("response_type", "code")
		params.Add("redirect_uri", s.cfg.InstagramRedirectURI)
		params.Add("state", tokenString)

		fullURL := fmt.Sprintf("%s?%s", authURL, params.Encode())
		return fullURL

	case "tiktok":
		authURL := TIKTOK_AUTH_URL
		params := url.Values{}
		params.Add("client_key", s.cfg.TiktokClientKey)
		params.Add("scope", "user.info.basic,user.info.profile,video.publish,video.upload")
		params.Add("response_type", "code")
		params.Add("redirect_uri", s.cfg.TiktokRedirectURI)
		params.Add("state", tokenString)

		fullURL := fmt.Sprintf("%s?%s", authURL, params.Encode())
		return fullURL

	case "youtube":
		authURL := GOOGLE_AUTH_URL
		params := url.Values{}
		params.Add("client_id", s.cfg.GoogleClientID)
		params.Add("redirect_uri", s.cfg.GoogleRedirectURI)
		params.Add("response_type", "code")
		params.Add("scope", "https://www.googleapis.com/auth/userinfo.profile https://www.googleapis.com/auth/userinfo.email https://www.googleapis.com/auth/youtube.upload")
		params.Add("state", tokenString)
		params.Add("access_type", "offline")

		fullURL := fmt.Sprintf("%s?%s", authURL, params.Encode())
		return fullURL

	default:
		return ""
	}
}

func (s *platformService) List(ctx context.Context, userID int64) ([]*models.SocialAccount, error) {
	var err error

	if userID == 0 {
		err = errors.New("UserID is not valid")
		slog.Info(err.Error())
		return nil, err
	}

	accounts, err := s.sa.ListInfoByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("Error getting social accounts")
	}

	return accounts, nil
}

func (s *platformService) Delete(ctx context.Context, userID, accountID int64) error {
	var err error

	if userID == 0 {
		err = errors.New("UserID is not valid")
		slog.Info(err.Error())
		return err
	}

	if accountID == 0 {
		err = errors.New("AccountID is not valid")
		slog.Info(err.Error())
		return err
	}

	isValid, err := s.sa.CheckByUserID(ctx, accountID, userID)
	if err != nil {
		return err
	}

	if !isValid {
		err = errors.New("Social account doesn't exist")
		slog.Info(err.Error())
		return err
	}

	accountInfo, err := s.sa.GetByID(ctx, accountID)
	if err != nil {
		return fmt.Errorf("Unable to get social account info")
	}

	decryptedAccessToken, err := utils.Decrypt(accountInfo.AccessToken, []byte(s.cfg.SecretKey))
	if err != nil {
		return err
	}

	switch accountInfo.Platform {

	case "tiktok":
		err = RevokeTiktokAccess(accountInfo.AccountID, decryptedAccessToken)
		if err != nil {
			slog.Info(err.Error())
			return fmt.Errorf("Unable to revoke access")
		}
	case "youtube":
		err = RevokeGoogleAccess(decryptedAccessToken)
		if err != nil {
			slog.Info(err.Error())
			return fmt.Errorf("Unable to revoke access")
		}
	}

	err = s.sa.Remove(ctx, accountID)
	if err != nil {
		return fmt.Errorf("Error removing account Info")
	}

	return nil
}
