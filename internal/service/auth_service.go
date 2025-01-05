package service

import (
	"context"
	"errors"
	"log/slog"

	config "github.com/maheshrc27/postflow/configs"
	"github.com/maheshrc27/postflow/internal/models"
	"github.com/maheshrc27/postflow/internal/repository"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type AuthService interface {
	LoginCallback(ctx context.Context, code string) (int64, error)
}

type authService struct {
	cfg config.Config
	u   repository.UserRepository
}

func NewAuthService(cfg config.Config, u repository.UserRepository) AuthService {
	return &authService{
		cfg: cfg,
		u:   u,
	}
}

func (s *authService) LoginCallback(ctx context.Context, code string) (int64, error) {
	var err error
	if code == "" {
		err = errors.New("code or state is empty")
		slog.Info(err.Error())
		return 0, err
	}

	oauth2Config := &oauth2.Config{
		ClientID:     s.cfg.GoogleClientID,
		ClientSecret: s.cfg.GoogleClientSecret,
		RedirectURL:  "http://localhost:3000/login/callback",
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
		Endpoint:     google.Endpoint,
	}

	if oauth2Config.ClientID == "" || oauth2Config.ClientSecret == "" || oauth2Config.RedirectURL == "" {
		err = errors.New("OAuth2 configuration is incomplete")
		slog.Info(err.Error())
		return 0, err
	}

	token, err := oauth2Config.Exchange(context.Background(), code)
	if err != nil {
		slog.Info(err.Error())
		return 0, err
	}

	client := oauth2Config.Client(context.Background(), token)

	userInfo, err := GetUserInfo(client)
	if err != nil {
		return 0, err
	}

	user, isExist, err := s.u.GetByEmail(ctx, userInfo.Email)
	if err != nil {
		return 0, err
	}

	var userID int64

	if !isExist {
		userID, err = s.u.Create(ctx, nil, &models.User{
			GoogleID:       userInfo.ID,
			Email:          userInfo.Email,
			Name:           userInfo.Name,
			ProfilePicture: userInfo.Picture,
		})

		if err != nil {
			slog.Info(err.Error())
			return 0, err
		}
	} else {
		userID = user.ID
	}

	if user.GoogleID == "" {
		updateUser := &models.User{
			ID:             userID,
			GoogleID:       userInfo.ID,
			Email:          userInfo.Email,
			Name:           userInfo.Name,
			ProfilePicture: userInfo.Picture,
		}

		err = s.u.Update(ctx, updateUser)
		if err != nil {
			return 0, err
		}

	}

	return userID, nil
}
