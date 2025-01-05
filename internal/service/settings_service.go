package service

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/maheshrc27/postflow/internal/models"
	"github.com/maheshrc27/postflow/internal/repository"
)

type SettingsService interface {
	GetSettingsInfo(ctx context.Context, id int64) (*models.Settings, error)
	UpdateSettings(ctx context.Context, userID int64, postingTime string, category string) error
}

type settingsService struct {
	sr repository.SettingsRepository
}

func NewSettingsService(sr repository.SettingsRepository) SettingsService {
	return &settingsService{
		sr: sr,
	}
}

func (s *settingsService) GetSettingsInfo(ctx context.Context, id int64) (*models.Settings, error) {
	settings, isExist, err := s.sr.GetByUserID(ctx, id)
	if err != nil {
		return nil, err
	}

	if !isExist {
		err = errors.New("setting for given user doesn't exist")
		slog.Info(err.Error())
		return nil, err
	}

	return settings, nil
}

func (s *settingsService) UpdateSettings(ctx context.Context, userID int64, postingTime string, category string) error {

	const layout = "15:04:05"

	parsedTime, err := time.Parse(layout, postingTime)
	if err != nil {
		slog.Info(err.Error())
		return err
	}

	settings := models.Settings{
		PostingTime: parsedTime,
		Category:    category,
	}
	err = s.sr.UpdateSettings(ctx, &settings, userID)
	if err != nil {
		return err
	}
	return nil
}
