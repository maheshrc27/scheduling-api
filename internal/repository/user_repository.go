// repository/user_repository.go
package repository

import (
	"context"
	"database/sql"
	"log/slog"
	"time"

	"github.com/maheshrc27/scheduling-api/internal/models"
)

type UserRepository interface {
	GetByID(ctx context.Context, id int64) (*models.User, bool, error)
	GetByEmail(ctx context.Context, email string) (*models.User, bool, error)
	Create(ctx context.Context, tx *sql.Tx, user *models.User) (int64, error)
	Update(ctx context.Context, user *models.User) error
	Remove(ctx context.Context, id int64) error
}

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) GetByID(ctx context.Context, id int64) (*models.User, bool, error) {
	var user models.User
	query := "SELECT id, name, email, profile_picture FROM users WHERE id = $1"
	err := r.db.QueryRowContext(ctx, query, id).Scan(&user.ID, &user.Name, &user.Email, &user.ProfilePicture)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, false, nil
		}
		slog.Info(err.Error())
		return nil, false, err
	}
	return &user, true, nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*models.User, bool, error) {
	var user models.User
	query := "SELECT id, google_id, email, name FROM users WHERE email = $1"
	err := r.db.QueryRowContext(ctx, query, email).Scan(&user.ID, &user.GoogleID, &user.Email, &user.Name)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, false, nil
		}
		slog.Info(err.Error())
		return nil, false, err
	}
	return &user, true, nil
}

func (r *userRepository) Create(ctx context.Context, tx *sql.Tx, user *models.User) (int64, error) {
	query := "INSERT INTO users (google_id, email, name, profile_picture) VALUES ($1, $2, $3, $4) RETURNING id"

	var err error
	var id int64

	if tx != nil {
		err = tx.QueryRowContext(ctx, query, user.GoogleID, user.Email, user.Name, user.ProfilePicture).Scan(&id)
	} else {
		err = r.db.QueryRowContext(ctx, query, user.GoogleID, user.Email, user.Name, user.ProfilePicture).Scan(&id)
	}
	if err != nil {
		slog.Info(err.Error())
		return 0, err
	}
	return id, nil
}

func (r *userRepository) Update(ctx context.Context, user *models.User) error {
	query := `
		UPDATE users
		SET google_id = $1,
			name = $2,
			profile_picture = $3,
			updated_at = $4
		WHERE id = $5
	`
	_, err := r.db.ExecContext(ctx, query, user.GoogleID, user.Name, user.ProfilePicture, time.Now(), user.ID)
	if err != nil {
		slog.Info(err.Error())
		return err
	}

	return nil
}

func (r *userRepository) Remove(ctx context.Context, id int64) error {
	query := `DELETE FROM users WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)

	if err != nil {
		slog.Info(err.Error())
		return err
	}
	return nil
}
