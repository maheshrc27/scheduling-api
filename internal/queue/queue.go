package queue

import (
	"encoding/json"
	"log"
	"time"

	"github.com/hibiken/asynq"
)

func EnqueuePost(asynqClient *asynq.Client, payload SchedulePostPayload, delay time.Duration) error {
	taskPayload, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	task := asynq.NewTask(TaskTypeSchedulePost, taskPayload)

	_, err = asynqClient.Enqueue(task, asynq.ProcessIn(delay))
	if err != nil {
		return err
	}

	log.Printf("Task scheduled: %+v", payload)
	return nil
}
