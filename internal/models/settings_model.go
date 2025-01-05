package models

import "time"

type Settings struct {
	ID          int64     `db:"id" json:"id"`
	UserID      string    `db:"user_id" json:"user_id"`
	PostingTime time.Time `db:"posting_time" json:"posting_time"`
	Category    string    `db:"category" json:"category"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
}
