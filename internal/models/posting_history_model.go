package models

import "time"

type PostingHistory struct {
	ID           int64     `db:"id" json:"id"`
	UserID       int64     `db:"user_id" json:"user_id"`
	PostID       int64     `db:"post_id" json:"post_id"`
	AccountID    int64     `db:"account_id" json:"account_id"`
	ErrorMessage string    `db:"error_message" json:"error_message"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
}
