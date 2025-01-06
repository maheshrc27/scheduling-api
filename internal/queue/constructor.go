package queue

import (
	"github.com/maheshrc27/scheduling-api/internal/repository"
	"github.com/maheshrc27/scheduling-api/internal/service"
)

type Queue struct {
	pr repository.PostRepository
	sa repository.SelectedAccountRepository
	ac repository.SocialAccountRepository
	ma repository.MediaAssetRepository
	pm repository.PostMediaRepository
	yt service.YoutubeService
	tt service.TiktokService
	ig service.InstagramService
}

func NewQueue(
	pr repository.PostRepository,
	sa repository.SelectedAccountRepository,
	ma repository.MediaAssetRepository,
	ac repository.SocialAccountRepository,
	pm repository.PostMediaRepository,
	yt service.YoutubeService,
	tt service.TiktokService,
	ig service.InstagramService) *Queue {
	return &Queue{
		pr: pr,
		sa: sa,
		ac: ac,
		ma: ma,
		pm: pm,
		yt: yt,
		tt: tt,
		ig: ig,
	}
}

const TaskTypeSchedulePost = "schedule:post"

type SchedulePostPayload struct {
	PostID int64 `json:"post_id"`
}
