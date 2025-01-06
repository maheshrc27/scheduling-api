package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/maheshrc27/scheduling-api/internal/models"
)

type SelectedAccountRepository interface {
	Create(ctx context.Context, tx *sql.Tx, sa *models.SelectedAccount) error
	GetByID(ctx context.Context, postID, accountID int64) (*models.SelectedAccount, error)
	ListByPostID(ctx context.Context, postID int64) ([]*models.SelectedAccount, error)
	ListByAccountID(ctx context.Context, userID int64) ([]*models.SelectedAccount, error)
	Remove(ctx context.Context, postID, accountID int64) error
}

type selectedAccountRepository struct {
	db *sql.DB
}

func NewSelectedAccountRepository(db *sql.DB) SelectedAccountRepository {
	return &selectedAccountRepository{db: db}
}

func (r *selectedAccountRepository) Create(ctx context.Context, tx *sql.Tx, sa *models.SelectedAccount) error {
	var err error

	query := `
		INSERT INTO selected_accounts (post_id, account_id)
		VALUES ($1, $2)
	`
	if tx != nil {
		_, err = tx.ExecContext(ctx, query, sa.PostID, sa.AccountID)
	} else {
		_, err = r.db.ExecContext(ctx, query, sa.PostID, sa.AccountID)
	}

	if err != nil {
		slog.Info(err.Error())
		return err
	}
	return nil
}

func (r *selectedAccountRepository) GetByID(ctx context.Context, postID, accountID int64) (*models.SelectedAccount, error) {
	query := "SELECT post_id, account_id FROM selected_accounts WHERE post_id = $1 AND account_id = $2"

	var sa models.SelectedAccount
	err := r.db.QueryRowContext(ctx, query, postID, accountID).Scan(&sa.PostID, &sa.AccountID, &sa.CreatedAt, &sa.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		slog.Info(err.Error())
		return nil, fmt.Errorf("query row: %w", err)
	}

	return &sa, nil
}

func (r *selectedAccountRepository) ListByPostID(ctx context.Context, postID int64) ([]*models.SelectedAccount, error) {
	query := "SELECT post_id, account_id FROM selected_accounts WHERE post_id = $1"

	rows, err := r.db.QueryContext(ctx, query, postID)
	if err != nil {
		slog.Info(err.Error())
		return nil, fmt.Errorf("query rows: %w", err)
	}
	defer rows.Close()

	var accounts []*models.SelectedAccount
	for rows.Next() {
		var sa models.SelectedAccount
		if err := rows.Scan(&sa.PostID, &sa.AccountID); err != nil {
			slog.Info(err.Error())
			return nil, fmt.Errorf("scan row: %w", err)
		}
		accounts = append(accounts, &sa)
	}

	if err := rows.Err(); err != nil {
		slog.Info(err.Error())
		return nil, fmt.Errorf("rows iteration: %w", err)
	}

	return accounts, nil
}

func (r *selectedAccountRepository) ListByAccountID(ctx context.Context, accountID int64) ([]*models.SelectedAccount, error) {
	query := "SELECT post_id, account_id FROM selected_accounts WHERE account_id = $1"

	rows, err := r.db.QueryContext(ctx, query, accountID)
	if err != nil {
		slog.Info(err.Error())
		return nil, fmt.Errorf("query rows: %w", err)
	}
	defer rows.Close()

	var accounts []*models.SelectedAccount
	for rows.Next() {
		var sa models.SelectedAccount
		if err := rows.Scan(&sa.PostID, &sa.AccountID); err != nil {
			slog.Info(err.Error())
			return nil, fmt.Errorf("scan row: %w", err)
		}
		accounts = append(accounts, &sa)
	}

	if err := rows.Err(); err != nil {
		slog.Info(err.Error())
		return nil, fmt.Errorf("rows iteration: %w", err)
	}

	return accounts, nil
}

func (r *selectedAccountRepository) Remove(ctx context.Context, postID, accountID int64) error {
	query := `DELETE FROM social_accounts WHERE post_id = $1 AND account_id = $2`
	_, err := r.db.ExecContext(ctx, query, postID, accountID)
	if err != nil {
		slog.Info(err.Error())
		return err
	}
	return nil
}
