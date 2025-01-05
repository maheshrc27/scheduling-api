package queue

import (
	"context"
	"encoding/json"
	"errors"
	"log"

	"github.com/hibiken/asynq"
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

	post, err := j.pr.GetByID(ctx, postID)
	if err != nil {
		return err
	}

	accountsSelected, err := j.sa.ListByPostID(ctx, postID)
	if err != nil {
		return err
	}

	if accountsSelected == nil {
		err = errors.New("No accounts selected for publishing.")
		return err
	}

	for _, acc := range accountsSelected {
		socialAcc, err := j.ac.GetByID(ctx, acc.AccountID)
		if err != nil {
			log.Printf("Error retrieving social account for AccountID %d: %v\n", acc.AccountID, err)
			continue
		}

		if socialAcc == nil {
			log.Printf("Social account for AccountID %d is nil\n", acc.AccountID)
			continue
		}

		switch socialAcc.Platform {
		case "tiktok":
			err = j.tt.HandleTiktokPost(ctx, post, socialAcc)
			if err != nil {
				log.Printf("Error posting to Tiktok for PostID %d: %v\n", post.ID, err)
			}
		case "instagram":
			if err := j.ig.HandleInstagramPost(ctx, post, socialAcc); err != nil {
				log.Printf("Error posting to Instagram for PostID %d: %v\n", post.ID, err)
			}

		case "youtube":
			if err := j.yt.PostYoutubeVideo(ctx, post, socialAcc); err != nil {
				log.Printf("Error posting to YouTube for PostID %d: %v\n", post.ID, err)
			}
		}
	}

	return nil
}
