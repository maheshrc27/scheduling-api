package repository

import (
	"context"
	"database/sql"
	"log/slog"
	"time"

	"github.com/maheshrc27/postflow/internal/models"
)

type PostRepository interface {
	GetByID(ctx context.Context, id int64) (*models.Post, error)
	Create(ctx context.Context, tx *sql.Tx, post *models.Post) (int64, error)
	GetByUserID(ctx context.Context, userID int64) ([]*models.Post, error)
	UpdatePostStatus(ctx context.Context, status string, postID int64) error
	CheckByUserID(ctx context.Context, accountID, userID int64) (bool, error)
	Remove(ctx context.Context, id int64) error
}

type postRepository struct {
	db *sql.DB
}

func NewPostRepository(db *sql.DB) PostRepository {
	return &postRepository{db: db}
}

func (r *postRepository) Create(ctx context.Context, tx *sql.Tx, post *models.Post) (int64, error) {
	query := `
		INSERT INTO posts (user_id, post_type, caption, title, scheduled_time)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`

	var id int64
	var err error

	if tx != nil {
		err = tx.QueryRowContext(ctx, query, post.UserID, post.PostType, post.Caption, post.Title, post.ScheduledTime).Scan(&id)
	} else {
		err = r.db.QueryRowContext(ctx, query, post.UserID, post.PostType, post.Caption, post.Title, post.ScheduledTime).Scan(&id)
	}
	if err != nil {
		slog.Info(err.Error())
		return 0, err
	}

	return id, nil
}

func (r *postRepository) GetByID(ctx context.Context, id int64) (*models.Post, error) {
	query := `SELECT id, user_id, post_type, caption, title, scheduled_time, status, created_at, updated_at FROM posts WHERE id = $1`
	row := r.db.QueryRowContext(ctx, query, id)

	var post models.Post
	err := row.Scan(&post.ID, &post.UserID, &post.PostType, &post.Caption, &post.Title, &post.ScheduledTime, &post.Status, &post.CreatedAt, &post.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		slog.Info(err.Error())
		return nil, err
	}

	return &post, nil
}

func (r *postRepository) GetByUserID(ctx context.Context, userID int64) ([]*models.Post, error) {
	query := `SELECT id, user_id, post_type, caption, title, scheduled_time, status, created_at, updated_at FROM posts WHERE user_id = $1`
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
		err := rows.Scan(&post.ID, &post.UserID, &post.PostType, &post.Caption, &post.Title, &post.ScheduledTime, &post.Status, &post.CreatedAt, &post.UpdatedAt)
		if err != nil {
			slog.Info(err.Error())
			return nil, err
		}
		posts = append(posts, &post)
	}
	return posts, nil
}

func (r *postRepository) CheckByUserID(ctx context.Context, accountID, userID int64) (bool, error) {
	query := "SELECT 1 FROM posts WHERE id = $1 AND user_id = $2"

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

func (r *postRepository) GetScheduled(ctx context.Context, userID int64) ([]*models.Post, error) {
	query := `SELECT id, user_id, post_type, caption, title, scheduled_time, status, created_at, updated_at FROM posts WHERE user_id=$1 AND status = $2`
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
		err := rows.Scan(&post.ID, &post.UserID, &post.PostType, &post.Caption, &post.Title, &post.ScheduledTime, &post.Status, &post.CreatedAt, &post.UpdatedAt)
		if err != nil {
			slog.Info(err.Error())
			return nil, err
		}
		posts = append(posts, &post)
	}
	return posts, nil

}

func (r *postRepository) UpdatePostStatus(ctx context.Context, status string, postID int64) error {
	query := `
		UPDATE posts 
		SET status = $1,  
			updated_at = $2
		WHERE id = $3
	`
	_, err := r.db.ExecContext(ctx, query, status, time.Now(), postID)
	if err != nil {
		slog.Info(err.Error())
		return err
	}
	return nil
}

func (r *postRepository) Remove(ctx context.Context, id int64) error {
	query := `DELETE FROM posts WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)

	if err != nil {
		slog.Info(err.Error())
		return err
	}
	return nil
}
