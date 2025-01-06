package handlers

import (
	"log/slog"

	"github.com/gofiber/fiber/v2"
	"github.com/hibiken/asynq"
	"github.com/maheshrc27/scheduling-api/internal/queue"
	"github.com/maheshrc27/scheduling-api/internal/service"
	"github.com/maheshrc27/scheduling-api/internal/transfer"
)

type PostHandler struct {
	s           service.PostService
	AsynqClient *asynq.Client
}

func NewPostHandler(service service.PostService, asynqClient *asynq.Client) *PostHandler {
	return &PostHandler{s: service, AsynqClient: asynqClient}
}

func (h *PostHandler) CreatePost(c *fiber.Ctx) error {
	userID := GetUserID(c)
	form, err := c.MultipartForm()
	if err != nil {
		slog.Error(err.Error())
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Unable to parse form",
		})
	}

	caption := c.FormValue("caption")
	title := c.FormValue("title")
	scheduledTime := c.FormValue("scheduling_time")
	selectedAccountsStr := c.FormValue("selected_accounts")

	files := form.File["files"]
	if len(files) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "No files selected",
		})
	}

	postID, delay, err := h.s.CreatePost(c.Context(), userID, &transfer.PostCreation{
		Caption:          caption,
		Title:            title,
		ScheduledTime:    scheduledTime,
		SelectedAccounts: selectedAccountsStr},
		files)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	err = queue.EnqueuePost(h.AsynqClient, queue.SchedulePostPayload{PostID: postID}, delay)
	if err != nil {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"error": "Error scheduling post",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Post scheduled successfully",
	})
}

func (h *PostHandler) ListPosts(c *fiber.Ctx) error {
	userId := GetUserID(c)
	postId := c.QueryInt("id", 0)

	if postId != 0 {
		post, err := h.s.PostInfo(c.Context(), int64(postId), userId)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Unable to list api posts",
			})
		}

		return c.Status(fiber.StatusOK).JSON(post)

	}

	posts, err := h.s.List(c.Context(), userId)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Unable to list api posts",
		})
	}

	return c.Status(fiber.StatusOK).JSON(posts)
}

func (h *PostHandler) RemovePost(c *fiber.Ctx) error {
	userID := GetUserID(c)
	postId := c.QueryInt("id", 0)

	err := h.s.Remove(c.Context(), userID, int64(postId))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Unable to remove post",
		})
	}

	return c.SendStatus(fiber.StatusOK)
}
