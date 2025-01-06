// repository/user_repository.go
package repository

import (
	"context"
	"database/sql"
	"log/slog"
	"time"

	"github.com/maheshrc27/scheduling-api/internal/models"
)

type SubscriptionRepository interface {
	GetByUserID(ctx context.Context, id int64) (*models.Subscription, bool, error)
	Create(ctx context.Context, subscription *models.Subscription) (int64, error)
	UpdateSubscription(ctx context.Context, subscription *models.Subscription) error
}

type subscriptionRepository struct {
	db *sql.DB
}

func NewSubscriptionRepository(db *sql.DB) SubscriptionRepository {
	return &subscriptionRepository{db: db}
}

func (r *subscriptionRepository) GetByUserID(ctx context.Context, id int64) (*models.Subscription, bool, error) {
	var subscription models.Subscription
	query := "SELECT subscription_id, subscrition_end_date, status FROM subscriptions WHERE user_id = $1"
	err := r.db.QueryRowContext(ctx, query, id).Scan(&subscription.SubscriptionID, &subscription.SubscriptionEndDate, &subscription.Status)
	if err != nil {
		slog.Info(err.Error())
		return nil, false, err
	}
	return &subscription, true, nil
}

func (r *subscriptionRepository) Create(ctx context.Context, subscription *models.Subscription) (int64, error) {
	query := "INSERT INTO subscriptions (user_id, subscription_id, subscription_end_date, status) VALUES ($1, $2, $3, $4) RETURNING id"
	var id int64
	err := r.db.QueryRowContext(ctx, query, subscription.UserID, subscription.SubscriptionID, subscription.SubscriptionEndDate, subscription.Status).Scan(&id)
	if err != nil {
		slog.Info(err.Error())
		return 0, err
	}
	return id, nil
}

func (r *subscriptionRepository) UpdateSubscription(ctx context.Context, subscription *models.Subscription) error {
	query := `
		UPDATE subscriptions
		SET subscription_end_date = $1,
			status = $2
			updated_at = $3
		WHERE user_id = $4
	`
	_, err := r.db.ExecContext(ctx, query, subscription.SubscriptionEndDate, subscription.Status, time.Now(), subscription.UserID)
	if err != nil {
		slog.Info(err.Error())
		return err
	}

	return nil
}
