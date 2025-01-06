package repository

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/maheshrc27/scheduling-api/internal/models"
)

type MediaAssetRepository interface {
	Create(ctx context.Context, tx *sql.Tx, ma *models.MediaAsset) (int64, error)
	GetByID(ctx context.Context, id int64) (*models.MediaAsset, error)
	Remove(ctx context.Context, id int64) error
}

type mediaAssetRepository struct {
	db *sql.DB
}

func NewMediaAssetRepository(db *sql.DB) MediaAssetRepository {
	return &mediaAssetRepository{db: db}
}

func (r *mediaAssetRepository) Create(ctx context.Context, tx *sql.Tx, ma *models.MediaAsset) (int64, error) {
	var id int64
	var err error

	query := `
		INSERT INTO media_assets (user_id, file_name, file_type, file_size, file_url)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`
	if tx != nil {
		err = tx.QueryRowContext(ctx, query, ma.UserID, ma.FileName, ma.FileType, ma.FileSize, ma.FileURL).Scan(&id)
	} else {
		err = r.db.QueryRowContext(ctx, query, ma.UserID, ma.FileName, ma.FileType, ma.FileSize, ma.FileURL).Scan(&id)
	}

	if err != nil {
		slog.Info(err.Error())
		return 0, err
	}

	return id, nil
}

func (r *mediaAssetRepository) GetByID(ctx context.Context, id int64) (*models.MediaAsset, error) {
	query := `
		SELECT id, user_id, file_name, file_type, file_url, created_at
		FROM media_assets
		WHERE id = $1
	`

	var ma models.MediaAsset
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&ma.ID,
		&ma.UserID,
		&ma.FileName,
		&ma.FileType,
		&ma.FileURL,
		&ma.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		slog.Info(err.Error())
		return nil, err
	}

	return &ma, nil
}

func (r *mediaAssetRepository) Remove(ctx context.Context, id int64) error {
	query := `
		DELETE FROM post_media
		WHERE id = $1
	`
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		slog.Info(err.Error())
		return err
	}
	return nil
}
