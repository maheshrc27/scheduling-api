package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/maheshrc27/scheduling-api/internal/models"
	"github.com/maheshrc27/scheduling-api/internal/repository"
	"github.com/maheshrc27/scheduling-api/pkg/utils"
)

type ApiKeyService interface {
	Create(ctx context.Context, userID int64) error
	List(ctx context.Context, userID int64) ([]*models.ApiKey, error)
	GetUserID(ctx context.Context, apiKey string) (int64, error)
	RemoveAPIKey(ctx context.Context, userID, keyID int64) error
}

type apiKeyService struct {
	k repository.ApiKeyRepository
}

func NewApiKeyService(k repository.ApiKeyRepository) ApiKeyService {
	return &apiKeyService{
		k: k,
	}
}

func (s *apiKeyService) Create(ctx context.Context, userID int64) error {

	keys, err := s.k.GetByUserID(ctx, userID)
	if err != nil {
		return err
	}

	if len(keys) > 4 {
		err = errors.New("Only 5 API Keys can be created.")
		slog.Info(err.Error())
		return err
	}

	key, err := utils.GenerateRandomKey(16)
	if err != nil {
		slog.Info(err.Error())
		return fmt.Errorf("Error generating API key")
	}

	apiKey := &models.ApiKey{
		UserID: userID,
		ApiKey: key,
	}

	_, err = s.k.Create(ctx, apiKey)
	if err != nil {
		return fmt.Errorf("Error saving API key")
	}
	return nil
}

func (s *apiKeyService) GetUserID(ctx context.Context, apiKey string) (int64, error) {
	userID, isExist, err := s.k.GetByKey(ctx, apiKey)
	if err != nil {
		return 0, err
	}

	if !isExist {
		err = errors.New("Key doesn't exist")
		return 0, err
	}

	return *userID, nil
}

func (s *apiKeyService) List(ctx context.Context, userID int64) ([]*models.ApiKey, error) {
	apiKeys, err := s.k.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("Error getting API keys")
	}
	return apiKeys, nil
}

func (s *apiKeyService) RemoveAPIKey(ctx context.Context, userID, keyID int64) error {
	var err error

	if userID == 0 {
		err = errors.New("UserID is not valid")
		slog.Info(err.Error())
		return err
	}

	if keyID == 0 {
		err = errors.New("KeyID is not valid")
		slog.Info(err.Error())
		return err
	}

	isValid, err := s.k.CheckByUserID(ctx, keyID, userID)
	if err != nil {
		return err
	}

	if !isValid {
		err = errors.New("Key doesn't exist")
		slog.Info(err.Error())
		return err
	}

	err = s.k.Remove(ctx, keyID)
	if err != nil {
		return err
	}
	return nil
}
