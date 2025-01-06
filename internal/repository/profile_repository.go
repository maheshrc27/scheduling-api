package repository

import (
	"context"
	"database/sql"
	"log/slog"
	"time"

	"github.com/maheshrc27/scheduling-api/internal/models"
)

type ProfileRepository interface {
	GetByID(ctx context.Context, id int64) (*models.Post, error)
	Create(ctx context.Context, post *models.Post) (int64, error)
	GetByUserID(ctx context.Context, userID int64) ([]*models.Post, error)
}

type profileRepository struct {
	db *sql.DB
}

func NewProfileRepository(db *sql.DB) ProfileRepository {
	return &profileRepository{db: db}
}

func (r *profileRepository) Create(ctx context.Context, post *models.Post) (int64, error) {
	query := `
		INSERT INTO posts (user_id, post_type, caption, scheduled_time)
		VALUES (?, ?, ?, ?)
		RETURNING id
	`

	var id int64
	err := r.db.QueryRowContext(ctx, query, post.UserID, post.PostType, post.Caption, post.ScheduledTime).Scan(&id)
	if err != nil {
		slog.Info(err.Error())
		return 0, err
	}

	return id, nil
}

func (r *profileRepository) GetByID(ctx context.Context, id int64) (*models.Post, error) {
	query := `SELECT * FROM profile WHERE id = ?`
	row := r.db.QueryRowContext(ctx, query, id)

	var post models.Post
	err := row.Scan(&post.ID, &post.UserID, &post.PostType, &post.Caption, &post.ScheduledTime, &post.Status, &post.CreatedAt, &post.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		slog.Info(err.Error())
		return nil, err
	}

	return &post, nil
}

func (r *profileRepository) GetByUserID(ctx context.Context, userID int64) ([]*models.Post, error) {
	query := `SELECT * FROM profile WHERE user_id = $1`
	var rows *sql.Rows
	var err error

	rows, err = r.db.QueryContext(ctx, query, userID)
	if err != nil {
		slog.Info(err.Error())
		return nil, err
	}
	defer rows.Close()

	var posts []*models.Post
	for rows.Next() {
		var post models.Post
		err := rows.Scan(&post.ID, &post.UserID, &post.PostType, &post.Caption, &post.ScheduledTime, &post.Status, &post.CreatedAt, &post.UpdatedAt)
		if err != nil {
			slog.Info(err.Error())
			return nil, err
		}
		posts = append(posts, &post)
	}
	return posts, nil
}

func (r *profileRepository) GetScheduled(ctx context.Context, userID int64) ([]*models.Post, error) {
	query := `SELECT * FROM profile WHERE user_id=$1 AND status = $2`
	var rows *sql.Rows
	var err error

	rows, err = r.db.QueryContext(ctx, query, userID, "scheduled")
	if err != nil {
		slog.Info(err.Error())
		return nil, err
	}

	var posts []*models.Post
	for rows.Next() {
		var post models.Post
		err := rows.Scan(&post.ID, &post.UserID, &post.PostType, &post.Caption, &post.ScheduledTime, &post.Status, &post.CreatedAt, &post.UpdatedAt)
		if err != nil {
			slog.Info(err.Error())
			return nil, err
		}
		posts = append(posts, &post)
	}
	return posts, nil

}

func (r *profileRepository) UpdatePostStatus(ctx context.Context, status string, postID int64) error {
	query := `
		UPDATE posts 
		SET status = ?,  
			updated_at = ?
		WHERE id = ?
	`
	_, err := r.db.ExecContext(ctx, query, status, time.Now(), postID)
	return err
}

func (r *profileRepository) Remove(ctx context.Context, id int64) error {
	query := `DELETE FROM posts WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// to edit queries
