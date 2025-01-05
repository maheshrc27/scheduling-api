package models

import (
	"time"
)

type Subscription struct {
	ID                  int64     `db:"id" json:"id"`
	UserID              int64     `db:"user_id" json:"user_id"`
	SubscriptionID      string    `db:"subscription_id" json:"subscription_id"`
	SubscriptionEndDate time.Time `db:"subscription_end_date" json:"subscription_end_date"`
	Status              string    `db:"status" json:"status"`
	CreatedAt           time.Time `db:"created_at" json:"created_at"`
	UpdatedAt           time.Time `db:"updated_at" json:"updated_at"`
}
