package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/maheshrc27/postflow/internal/service"
)

type UserHandler struct {
	s service.UserService
}

func NewUserHandler(service service.UserService) *UserHandler {
	return &UserHandler{s: service}
}

func (h *UserHandler) GetUserInfo(c *fiber.Ctx) error {
	userId := GetUserID(c)

	userInfo, err := h.s.GetUserInfo(c.Context(), userId)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(userInfo)
}
