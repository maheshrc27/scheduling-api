package repository

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/maheshrc27/scheduling-api/internal/models"
)

type PostingHistoryRepository interface {
	GetByID(ctx context.Context, id int64) (*models.PostingHistory, error)
	Create(ctx context.Context, ph *models.PostingHistory) (int64, error)
	GetByUserID(ctx context.Context, userID int64) ([]*models.PostingHistory, error)
}

type postingHistoryRepository struct {
	db *sql.DB
}

func NewPostingHistoryRepository(db *sql.DB) PostingHistoryRepository {
	return &postingHistoryRepository{db: db}
}

func (r *postingHistoryRepository) Create(ctx context.Context, ph *models.PostingHistory) (int64, error) {
	query := `
		INSERT INTO posting_history (user_id, post_id, account_id, error_message)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`

	var id int64
	err := r.db.QueryRowContext(ctx, query, ph.UserID, ph.PostID, ph.AccountID, ph.ErrorMessage).Scan(&id)
	if err != nil {
		slog.Info(err.Error())
		return 0, err
	}

	return id, nil
}

func (r *postingHistoryRepository) GetByID(ctx context.Context, id int64) (*models.PostingHistory, error) {
	query := `SELECT * FROM posting_history WHERE id = ?`
	row := r.db.QueryRowContext(ctx, query, id)

	var ph models.PostingHistory
	err := row.Scan(&ph.ID, &ph.UserID, &ph.PostID, &ph.AccountID, &ph.ErrorMessage, &ph.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		slog.Info(err.Error())
		return nil, err
	}

	return &ph, nil
}

func (r *postingHistoryRepository) GetByUserID(ctx context.Context, userID int64) ([]*models.PostingHistory, error) {
	query := `SELECT * FROM posting_history WHERE user_id = $1`
	var rows *sql.Rows
	var err error

	rows, err = r.db.QueryContext(ctx, query, userID)
	if err != nil {
		slog.Info(err.Error())
		return nil, err
	}
	defer rows.Close()

	var phs []*models.PostingHistory
	for rows.Next() {
		var ph models.PostingHistory
		err := rows.Scan(&ph.ID, &ph.UserID, &ph.PostID, &ph.AccountID, &ph.ErrorMessage, &ph.CreatedAt)
		if err != nil {
			slog.Info(err.Error())
			return nil, err
		}
		phs = append(phs, &ph)
	}
	return phs, nil
}
