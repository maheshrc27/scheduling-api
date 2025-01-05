package models

import "time"

type Post struct {
	ID            int64     `db:"id" json:"id"`
	UserID        int64     `db:"user_id" json:"user_id"`
	PostType      string    `db:"post_type" json:"post_type"`
	Caption       string    `db:"caption" json:"caption"`
	Title         string    `db:"title" json:"title"`
	ScheduledTime time.Time `db:"scheduled_time" json:"scheduled_time"`
	Status        string    `db:"status" json:"status"` // posted, scheduled, failed, draft
	CreatedAt     time.Time `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time `db:"updated_at" json:"updated_at"`
}

type MediaAsset struct {
	ID           int64     `db:"id"`
	UserID       int64     `db:"user_id"`
	FileName     string    `db:"file_name"`
	FileType     string    `db:"file_type"`
	FileSize     int64     `db:"file_size"`
	FileURL      string    `db:"file_url"`
	ThumbnailURL string    `db:"thumbnail_url"`
	CreatedAt    time.Time `db:"created_at"`
}

type PostMedia struct {
	PostID       int64     `db:"post_id"`
	AssetID      int64     `db:"asset_id"`
	DisplayOrder int       `db:"display_order"`
	CreatedAt    time.Time `db:"created_at"`
}

const (
	PostStatusScheduled = "scheduled"
	PostStatusPosted    = "posted"
	PostStatusFailed    = "failed"
	PostStatusDraft     = "draft"
)
