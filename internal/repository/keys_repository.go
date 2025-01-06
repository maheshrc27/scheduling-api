// repository/user_repository.go
package repository

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/maheshrc27/scheduling-api/internal/models"
)

type ApiKeyRepository interface {
	GetByKey(ctx context.Context, apiKey string) (*int64, bool, error)
	GetByUserID(ctx context.Context, userID int64) ([]*models.ApiKey, error)
	Create(ctx context.Context, apiKey *models.ApiKey) (int64, error)
	CheckByUserID(ctx context.Context, keyID, userID int64) (bool, error)
	Remove(ctx context.Context, id int64) error
}

type apiKeyRepository struct {
	db *sql.DB
}

func NewApiKeyRepository(db *sql.DB) ApiKeyRepository {
	return &apiKeyRepository{db: db}
}

func (r *apiKeyRepository) GetByKey(ctx context.Context, apiKey string) (*int64, bool, error) {
	var userId int64
	query := "SELECT user_id FROM api_keys WHERE api_key = $1"
	err := r.db.QueryRowContext(ctx, query, apiKey).Scan(&userId)
	if err != nil {
		slog.Info(err.Error())
		return nil, false, err
	}
	return &userId, true, nil
}

func (r *apiKeyRepository) GetByUserID(ctx context.Context, userID int64) ([]*models.ApiKey, error) {
	query := `SELECT * FROM api_keys WHERE user_id = $1`
	var rows *sql.Rows
	var err error

	rows, err = r.db.QueryContext(ctx, query, userID)
	if err != nil {
		slog.Info(err.Error())
		return nil, err
	}
	defer rows.Close()

	var apiKeys []*models.ApiKey
	for rows.Next() {
		var apiKey models.ApiKey
		err := rows.Scan(&apiKey.ID, &apiKey.UserID, &apiKey.ApiKey, &apiKey.CreatedAt)
		if err != nil {
			slog.Info(err.Error())
			return nil, err
		}
		apiKeys = append(apiKeys, &apiKey)
	}
	return apiKeys, nil
}

func (r *apiKeyRepository) Create(ctx context.Context, apiKey *models.ApiKey) (int64, error) {
	query := "INSERT INTO api_keys (user_id, api_key) VALUES ($1, $2) RETURNING id"
	var id int64
	err := r.db.QueryRowContext(ctx, query, apiKey.UserID, apiKey.ApiKey).Scan(&id)
	if err != nil {
		slog.Info(err.Error())
		return 0, err
	}
	return id, nil
}

func (r *apiKeyRepository) CheckByUserID(ctx context.Context, keyID, userID int64) (bool, error) {
	query := "SELECT 1 FROM api_keys WHERE id = $1 AND user_id = $2"

	var result int
	err := r.db.QueryRowContext(ctx, query, keyID, userID).Scan(&result)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		slog.Info(err.Error())
		return false, err
	}

	return result == 1, nil
}

func (r *apiKeyRepository) Remove(ctx context.Context, id int64) error {
	query := `DELETE FROM api_keys WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		slog.Info(err.Error())
		return err
	}
	return nil
}
