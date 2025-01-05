package repository

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"

	"github.com/maheshrc27/postflow/internal/models"
)

type PostMediaRepository interface {
	Create(ctx context.Context, tx *sql.Tx, pm *models.PostMedia) error
	GetByPostID(ctx context.Context, postID int64) (*models.PostMedia, error)
	ListByPostID(ctx context.Context, postID int64) ([]*models.PostMedia, error)
	Update(ctx context.Context, pm *models.PostMedia) error
	Remove(ctx context.Context, postID int64) error
}

type postMediaRepository struct {
	db *sql.DB
}

func NewPostMediaRepository(db *sql.DB) PostMediaRepository {
	return &postMediaRepository{db: db}
}

func (r *postMediaRepository) Create(ctx context.Context, tx *sql.Tx, pm *models.PostMedia) error {
	var err error

	query := `
		INSERT INTO post_media (post_id, asset_id, display_order)
		VALUES ($1, $2, $3)
	`
	if tx != nil {
		_, err = tx.ExecContext(ctx, query, pm.PostID, pm.AssetID, pm.DisplayOrder)
	} else {
		_, err = r.db.ExecContext(ctx, query, pm.PostID, pm.AssetID, pm.DisplayOrder)
	}

	if err != nil {
		slog.Info(err.Error())
		return err
	}

	return nil
}

func (r *postMediaRepository) GetByPostID(ctx context.Context, postID int64) (*models.PostMedia, error) {
	query := `
		SELECT post_id, asset_id, display_order
		FROM post_media
		WHERE post_id = $1 AND display_order = 0
	`

	var pm models.PostMedia
	err := r.db.QueryRowContext(ctx, query, postID).Scan(&pm.PostID, &pm.AssetID, &pm.DisplayOrder)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		slog.Info(err.Error())
		return nil, err
	}

	return &pm, nil
}

func (r *postMediaRepository) ListByPostID(ctx context.Context, postID int64) ([]*models.PostMedia, error) {
	query := `
		SELECT post_id, asset_id, display_order
		FROM post_media
		WHERE post_id = $1
		ORDER BY display_order
	`

	rows, err := r.db.QueryContext(ctx, query, postID)
	if err != nil {
		slog.Info(err.Error())
		return nil, err
	}
	defer rows.Close()

	var postMedias []*models.PostMedia
	for rows.Next() {
		var pm models.PostMedia
		if err := rows.Scan(&pm.PostID, &pm.AssetID, &pm.DisplayOrder); err != nil {
			slog.Info(err.Error())
			return nil, err
		}
		postMedias = append(postMedias, &pm)
	}

	if err = rows.Err(); err != nil {
		slog.Info(err.Error())
		return nil, err
	}

	return postMedias, nil
}

func (r *postMediaRepository) Remove(ctx context.Context, id int64) error {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		slog.Info(err.Error())
		return err
	}
	defer tx.Rollback()

	query := `
		DELETE FROM post_media
		WHERE id = $1
	`

	_, err = tx.ExecContext(ctx, query, id)
	if err != nil {
		slog.Info(err.Error())
		return err
	}

	if err = tx.Commit(); err != nil {
		slog.Info(err.Error())
		return err
	}

	return nil
}

func (r *postMediaRepository) Update(ctx context.Context, pm *models.PostMedia) error {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		slog.Info(err.Error())
		return err
	}
	defer tx.Rollback()

	query := `
		UPDATE post_media
		SET display_order = $1
		WHERE post_id = $2 AND asset_id = $3
	`

	result, err := tx.ExecContext(ctx, query, pm.DisplayOrder, pm.PostID, pm.AssetID)
	if err != nil {
		slog.Info(err.Error())
		return err
	}

	affectedRows, err := result.RowsAffected()
	if err != nil {
		slog.Info(err.Error())
		return err
	}
	if affectedRows == 0 {
		return errors.New("no rows affected")
	}

	if err = tx.Commit(); err != nil {
		slog.Info(err.Error())
		return err
	}

	return nil
}
