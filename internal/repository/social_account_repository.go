package repository

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"time"

	"github.com/maheshrc27/postflow/internal/models"
)

type SocialAccountRepository interface {
	Create(ctx context.Context, tx *sql.Tx, sa *models.SocialAccount) (int64, error)
	GetByID(ctx context.Context, id int64) (*models.SocialAccount, error)
	ListByUserID(ctx context.Context, userID int64) ([]*models.SocialAccount, error)
	ListInfoByUserID(ctx context.Context, userID int64) ([]*models.SocialAccount, error)
	ListByTimeInterval(ctx context.Context, initialTime, finalTime time.Time) ([]*models.SocialAccount, error)
	CheckByUserID(ctx context.Context, accountID, userID int64) (bool, error)
	SetToken(ctx context.Context, userID int64, oldAccessToken string, sa *models.SocialAccount) error
	Remove(ctx context.Context, id int64) error
}

type socialAccountRepository struct {
	db *sql.DB
}

func NewSocialAccountRepository(db *sql.DB) SocialAccountRepository {
	return &socialAccountRepository{db: db}
}

func (r *socialAccountRepository) Create(ctx context.Context, tx *sql.Tx, sa *models.SocialAccount) (int64, error) {
	var err error
	var id int64

	var insertQuery = `
			INSERT INTO social_accounts(
				user_id, 
				platform, 
				account_id, 
				account_name, 
				account_username, 
				profile_picture_url, 
				access_token, 
				refresh_token, 
				token_expires_at
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			RETURNING id
		`

	if tx != nil {
		err = tx.QueryRowContext(ctx, insertQuery,
			sa.UserID,
			sa.Platform,
			sa.AccountID,
			sa.AccountName,
			sa.AccountUsername,
			sa.ProfilePicture,
			sa.AccessToken,
			sa.RefreshToken,
			sa.TokenExpiresAt,
		).Scan(&id)
	} else {
		err = r.db.QueryRowContext(ctx, insertQuery,
			sa.UserID,
			sa.Platform,
			sa.AccountID,
			sa.AccountName,
			sa.AccountUsername,
			sa.ProfilePicture,
			sa.AccessToken,
			sa.RefreshToken,
			sa.TokenExpiresAt,
		).Scan(&id)
	}

	if err != nil {
		slog.Info(err.Error())
		return 0, err
	}

	return id, nil

}

func (r *socialAccountRepository) GetByID(ctx context.Context, id int64) (*models.SocialAccount, error) {
	query := `SELECT * FROM social_accounts WHERE id = $1`
	row := r.db.QueryRowContext(ctx, query, id)

	var sa models.SocialAccount
	err := row.Scan(&sa.ID, &sa.UserID, &sa.Platform, &sa.AccountID, &sa.AccountName,
		&sa.AccountUsername, &sa.ProfilePicture, &sa.AccessToken, &sa.RefreshToken,
		&sa.TokenExpiresAt, &sa.AccountStatus, &sa.CreatedAt, &sa.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		slog.Info(err.Error())
		return nil, err
	}

	return &sa, nil
}

func (r *socialAccountRepository) ListByUserID(ctx context.Context, userID int64) ([]*models.SocialAccount, error) {
	query := `SELECT * FROM social_accounts`
	args := []interface{}{}

	if userID != 0 {
		query += ` WHERE user_id = ?`
		args = append(args, userID)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		slog.Info(err.Error())
		return nil, err
	}
	defer rows.Close()

	var accounts []*models.SocialAccount
	for rows.Next() {
		var sa models.SocialAccount
		err := rows.Scan(&sa.ID, &sa.UserID, &sa.Platform, &sa.AccountID, &sa.AccountName,
			&sa.AccountUsername, &sa.ProfilePicture, &sa.AccessToken, &sa.RefreshToken,
			&sa.TokenExpiresAt, &sa.CreatedAt, &sa.UpdatedAt)
		if err != nil {
			slog.Info(err.Error())
			return nil, err
		}
		accounts = append(accounts, &sa)
	}
	return accounts, nil
}

func (r *socialAccountRepository) ListByTimeInterval(ctx context.Context, initialTime, finalTime time.Time) ([]*models.SocialAccount, error) {
	query := `SELECT
			user_id,
			platform,
			access_token, 
			refresh_token, 
			token_expires_at
			FROM social_accounts 
			WHERE (token_expires_at BETWEEN $1 AND $2)
			OR (token_expires_at < $3)`
	rows, err := r.db.QueryContext(ctx, query, initialTime, finalTime, initialTime)
	if err != nil {
		slog.Info(err.Error())
		return nil, err
	}
	defer rows.Close()

	var socialAccounts []*models.SocialAccount
	for rows.Next() {
		var sa models.SocialAccount
		err := rows.Scan(&sa.UserID, &sa.Platform, &sa.AccessToken, &sa.RefreshToken, &sa.TokenExpiresAt)
		if err != nil {
			slog.Info(err.Error())
			return nil, err
		}
		socialAccounts = append(socialAccounts, &sa)
	}

	if err := rows.Err(); err != nil {
		slog.Info(err.Error())
		return nil, err
	}

	return socialAccounts, nil
}

func (r *socialAccountRepository) ListInfoByUserID(ctx context.Context, userID int64) ([]*models.SocialAccount, error) {
	query := `SELECT id, account_name, profile_picture_url, platform FROM social_accounts WHERE user_id = $1`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		slog.Info(err.Error())
		return nil, err
	}
	defer rows.Close()

	var socialAccounts []*models.SocialAccount
	for rows.Next() {
		var sa models.SocialAccount
		err := rows.Scan(&sa.ID, &sa.AccountName, &sa.ProfilePicture, &sa.Platform)
		if err != nil {
			slog.Info(err.Error())
			return nil, err
		}
		socialAccounts = append(socialAccounts, &sa)
	}
	return socialAccounts, nil
}

func (r *socialAccountRepository) CheckByUserID(ctx context.Context, accountID, userID int64) (bool, error) {
	query := "SELECT 1 FROM social_accounts WHERE id = $1 AND user_id = $2"

	var result int
	err := r.db.QueryRowContext(ctx, query, accountID, userID).Scan(&result)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		slog.Info(err.Error())
		return false, err
	}

	return result == 1, nil
}

func (r *socialAccountRepository) SetToken(ctx context.Context, userID int64, oldAccessToken string, sa *models.SocialAccount) error {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		slog.Info(err.Error())
		return err
	}
	defer tx.Rollback()

	updateTokenQuery := `
		UPDATE social_accounts
		SET 
			access_token = COALESCE(NULLIF($3, ''), access_token),
			refresh_token = COALESCE(NULLIF($4, ''), refresh_token),
			token_expires_at = COALESCE($5, token_expires_at),
			updated_at = CURRENT_TIMESTAMP
		WHERE user_id = $1 AND access_token = $2;
	`
	result, err := tx.ExecContext(ctx, updateTokenQuery, userID, oldAccessToken, sa.AccessToken, sa.RefreshToken, sa.TokenExpiresAt)
	if err != nil {
		slog.Info(err.Error())
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		slog.Info(err.Error())
		return err
	}
	if affected != 1 {
		slog.Info("no rows affected; user_id may not exist")
		return errors.New("no rows affected; user_id may not exist")
	}

	if err = tx.Commit(); err != nil {
		slog.Info(err.Error())
		return err
	}
	return nil
}

func (r *socialAccountRepository) Remove(ctx context.Context, id int64) error {
	query := `DELETE FROM social_accounts WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		slog.Info(err.Error())
		return err
	}
	return nil
}
