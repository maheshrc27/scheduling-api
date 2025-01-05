package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/maheshrc27/postflow/internal/models"
	"github.com/maheshrc27/postflow/internal/repository"
)

type UserService interface {
	GetUserInfo(ctx context.Context, id int64) (*models.User, error)
}

type userService struct {
	u repository.UserRepository
}

func NewUserService(u repository.UserRepository) UserService {
	return &userService{
		u: u,
	}
}

func (s *userService) GetUserInfo(ctx context.Context, id int64) (*models.User, error) {
	user, isExist, err := s.u.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("Error getting user info")
	}

	if !isExist {
		err = errors.New("User not found")
		slog.Info(err.Error())
		return nil, fmt.Errorf("User doesn't exist")
	}

	return user, nil
}
