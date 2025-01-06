package job

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/maheshrc27/scheduling-api/internal/models"
	"github.com/maheshrc27/scheduling-api/internal/repository"
	"github.com/maheshrc27/scheduling-api/internal/service"
)

type TokenRefreshJob struct {
	sr repository.SocialAccountRepository
	yt service.YoutubeService
	tt service.TiktokService
	ig service.InstagramService
}

func NewtokenRefreshJob(
	sr repository.SocialAccountRepository,
	yt service.YoutubeService,
	tt service.TiktokService,
	ig service.InstagramService) *TokenRefreshJob {
	return &TokenRefreshJob{
		sr: sr,
		yt: yt,
		tt: tt,
		ig: ig,
	}
}

func (c *TokenRefreshJob) RefreshTokens() {
	ctx := context.Background()

	currentTime := time.Now()
	timeIn30Minutes := currentTime.Add(30 * time.Minute)

	accounts, err := c.sr.ListByTimeInterval(ctx, currentTime, timeIn30Minutes)
	if err != nil {
		slog.Info(err.Error())
		return
	}

	var wg sync.WaitGroup

	concurrencyLimit := 10
	semaphore := make(chan struct{}, concurrencyLimit)

	for _, acc := range accounts {

		wg.Add(1)
		semaphore <- struct{}{}

		go func(acc *models.SocialAccount) {
			defer wg.Done()
			defer func() { <-semaphore }()

			switch acc.Platform {
			case "youtube":
				err = c.yt.RefreshYoutubeToken(ctx, acc.UserID, acc.AccessToken, acc.RefreshToken)
				if err != nil {
					slog.Info("Unable to refresh tokens for YouTbe")
				}

			case "instagram":
				err = c.ig.RefreshInstagramToken(ctx, acc.UserID, acc.RefreshToken)
				if err != nil {
					slog.Info("Unable to refresh tokens for Instagram")
				}

			case "tiktok":
				err = c.tt.RefreshTiktokToken(ctx, acc.UserID, acc.AccessToken, acc.RefreshToken)
				if err != nil {
					slog.Info("Unable to refresh tokens for TikTok")
				}
			}
		}(acc)
	}
}
