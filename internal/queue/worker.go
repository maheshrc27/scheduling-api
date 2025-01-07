package queue

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"sync"

	"github.com/hibiken/asynq"
	"github.com/maheshrc27/scheduling-api/internal/models"
)

func (j *Queue) HandleSchedulePostTask(ctx context.Context, task *asynq.Task) error {
	var payload SchedulePostPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return err
	}

	j.PublishPost(payload.PostID)

	return nil
}

func (j *Queue) PublishPost(postID int64) error {
	ctx := context.Background()

	// Fetch the post
	post, err := j.pr.GetByID(ctx, postID)
	if err != nil {
		return err
	}

	// Fetch accounts associated with the post
	accountsSelected, err := j.sa.ListByPostID(ctx, postID)
	if err != nil {
		return err
	}
	if accountsSelected == nil {
		return errors.New("no accounts selected for publishing")
	}

	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 10) // Concurrency limit

	// Helper to handle posting
	postToPlatform := func(post *models.Post, socialAcc *models.SocialAccount) {
		defer wg.Done()
		defer func() { <-semaphore }()

		var err error
		switch socialAcc.Platform {
		case "tiktok":
			err = j.tt.HandleTiktokPost(ctx, post, socialAcc)
		case "instagram":
			err = j.ig.HandleInstagramPost(ctx, post, socialAcc)
		case "youtube":
			err = j.yt.PostYoutubeVideo(ctx, post, socialAcc)
		}

		// Log posting history
		postingHistory := models.PostingHistory{
			UserID:       socialAcc.UserID,
			PostID:       postID,
			AccountID:    socialAcc.ID,
			ErrorMessage: "",
		}
		if err != nil {
			postingHistory.ErrorMessage = err.Error()
			log.Printf("Error posting to %s for PostID %d: %v", socialAcc.Platform, post.ID, err)
		}
		if _, err := j.ph.Create(ctx, &postingHistory); err != nil {
			log.Printf("Error saving posting history for PostID %d: %v", post.ID, err)
		}
	}

	// Process each account
	for _, acc := range accountsSelected {
		socialAcc, err := j.ac.GetByID(ctx, acc.AccountID)
		if err != nil {
			log.Printf("Error retrieving social account for AccountID %d: %v", acc.AccountID, err)
			continue
		}
		if socialAcc == nil {
			log.Printf("Social account for AccountID %d is nil", acc.AccountID)
			continue
		}

		wg.Add(1)
		semaphore <- struct{}{}
		go postToPlatform(post, socialAcc)
	}

	wg.Wait() // Wait for all goroutines to finish
	return nil
}
