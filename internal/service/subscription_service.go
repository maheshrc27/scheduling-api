package service

import (
	"context"
	"fmt"

	config "github.com/maheshrc27/scheduling-api/configs"
	"github.com/maheshrc27/scheduling-api/internal/models"
	"github.com/maheshrc27/scheduling-api/internal/repository"
	"github.com/maheshrc27/scheduling-api/internal/transfer"
)

const (
	productId      = "ehajql"
	price1         = 500
	price2         = 900
	price3         = 1500
	InitialCredits = 1
	CreditsPrice1  = 10
	CreditsPrice2  = 30
	CreditsPrice3  = 50
)

type SubscriptionService interface {
	HandleSubscription(ctx context.Context, payload *transfer.SubscriptionEvent) error
}

type subscriptionService struct {
	cfg config.Config
	u   repository.UserRepository
	s   repository.SubscriptionRepository
}

func NewSubscriptionService(cfg config.Config, u repository.UserRepository, s repository.SubscriptionRepository) SubscriptionService {
	return &subscriptionService{
		cfg: cfg,
		u:   u,
		s:   s,
	}
}

func (s *subscriptionService) HandleSubscription(ctx context.Context, payload *transfer.SubscriptionEvent) error {

	switch payload.EventType {
	case "subscription.paid":
		customerEmail := payload.Object.Customer.Email

		user, isExist, err := s.u.GetByEmail(ctx, customerEmail)
		if err != nil {
			return fmt.Errorf("fetching user by email failed: %w", err)
		}

		var userID int64
		if !isExist {
			newUser := &models.User{
				Email: customerEmail,
			}
			userID, err = s.u.Create(ctx, nil, newUser)
			if err != nil {
				return err
			}

			subscriptionInfo := &models.Subscription{
				UserID:              userID,
				SubscriptionID:      payload.Object.ID,
				SubscriptionEndDate: payload.Object.CurrentPeriodEndDate,
				Status:              payload.Object.Status,
			}

			_, err = s.s.Create(ctx, subscriptionInfo)
			if err != nil {
				return err
			}
		} else {
			userID = user.ID

			subscriptionInfo := &models.Subscription{
				UserID:              userID,
				SubscriptionID:      payload.Object.ID,
				SubscriptionEndDate: payload.Object.CurrentPeriodEndDate,
				Status:              payload.Object.Status,
			}

			err := s.s.UpdateSubscription(ctx, subscriptionInfo)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
