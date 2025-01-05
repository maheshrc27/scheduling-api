package models

import "time"

type ApiKey struct {
	ID        int64     `db:"id" json:"id"`
	UserID    int64     `db:"user_id" json:"user_id"`
	ApiKey    string    `db:"api_key" json:"api_key"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}
