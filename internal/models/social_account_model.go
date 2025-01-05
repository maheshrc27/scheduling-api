package models

import (
	"time"
)

type SocialAccount struct {
	ID              int64     `db:"id" json:"id"`
	UserID          int64     `db:"user_id" json:"user_id"`
	Platform        string    `db:"platform" json:"platform"`
	AccountID       string    `db:"account_id" json:"account_id"`
	AccountName     string    `db:"account_name" json:"account_name"`
	AccountUsername string    `db:"account_username" json:"account_username"`
	ProfilePicture  string    `db:"profile_picture_url" json:"profile_picture"`
	AccessToken     string    `db:"access_token" json:"access_token"`
	RefreshToken    string    `db:"refresh_token" json:"refresh_token"`
	TokenExpiresAt  time.Time `db:"token_expires_at" json:"token_expires_at"`
	AccountStatus   string    `db:"account_status" json:"account_status"`
	CreatedAt       time.Time `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time `db:"updated_at" json:"updated_at"`
}

type SelectedAccount struct {
	PostID    int64     `db:"post_id" json:"post_id"`
	AccountID int64     `db:"account_id" json:"account_id"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}
