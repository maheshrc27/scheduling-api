package repository

import (
	"context"
	"database/sql"
	"log/slog"
	"time"

	"github.com/maheshrc27/scheduling-api/internal/models"
)

type SettingsRepository interface {
	GetByID(ctx context.Context, id int64) (*models.Settings, error)
	Create(ctx context.Context, post *models.Settings) (int64, error)
	GetByUserID(ctx context.Context, userID int64) (*models.Settings, bool, error)
	GetByPostingTime(ctx context.Context, postingTime time.Time) ([]*models.Settings, error)
	UpdateSettings(ctx context.Context, s *models.Settings, userID int64) error
}

type settingsRepository struct {
	db *sql.DB
}

func NewSettingsRepository(db *sql.DB) SettingsRepository {
	return &settingsRepository{db: db}
}

func (r *settingsRepository) Create(ctx context.Context, settings *models.Settings) (int64, error) {
	query := `
		INSERT INTO settings (user_id, posting_time, category)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`

	var id int64
	err := r.db.QueryRowContext(ctx, query, settings.UserID, settings.PostingTime, settings.Category).Scan(&id)
	if err != nil {
		slog.Info(err.Error())
		return 0, err
	}

	return id, nil
}

func (r *settingsRepository) GetByID(ctx context.Context, id int64) (*models.Settings, error) {
	query := `SELECT * FROM settings WHERE id = $1`
	row := r.db.QueryRowContext(ctx, query, id)

	var settings models.Settings
	err := row.Scan(&settings.ID, &settings.UserID, &settings.PostingTime, &settings.Category, &settings.CreatedAt, &settings.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		slog.Info(err.Error())
		return nil, err
	}

	return &settings, nil
}

func (r *settingsRepository) GetByUserID(ctx context.Context, userID int64) (*models.Settings, bool, error) {
	query := `SELECT * FROM settings WHERE user_id = $1`
	row := r.db.QueryRowContext(ctx, query, userID)

	var settings models.Settings
	err := row.Scan(&settings.ID, &settings.UserID, &settings.PostingTime, &settings.Category, &settings.CreatedAt, &settings.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, false, nil
		}
		slog.Info(err.Error())
		return nil, false, err
	}

	return &settings, true, nil
}

func (r *settingsRepository) GetByPostingTime(ctx context.Context, postingTime time.Time) ([]*models.Settings, error) {
	query := `SELECT * FROM settings WHERE posting_time = $1`
	var rows *sql.Rows
	var err error

	rows, err = r.db.QueryContext(ctx, query, postingTime)
	if err != nil {
		slog.Info(err.Error())
		return nil, err
	}
	defer rows.Close()

	var settings []*models.Settings
	for rows.Next() {
		var s models.Settings
		err := rows.Scan(&s.ID, &s.UserID, &s.PostingTime, &s.Category, &s.CreatedAt, &s.UpdatedAt)
		if err != nil {
			slog.Info(err.Error())
			return nil, err
		}
		settings = append(settings, &s)
	}
	return settings, nil
}

func (r *settingsRepository) UpdateSettings(ctx context.Context, s *models.Settings, userID int64) error {
	query := `
		UPDATE settings 
		SET posting_time = $1,
			category = $2,  
			updated_at = $3
		WHERE user_id = $4
	`
	_, err := r.db.ExecContext(ctx, query, s.PostingTime, s.Category, time.Now(), userID)
	if err != nil {
		slog.Info(err.Error())
		return err
	}

	return nil
}
