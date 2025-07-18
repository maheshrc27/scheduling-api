package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/maheshrc27/scheduling-api/internal/service"
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

func (h *UserHandler) DeleteAccount(c *fiber.Ctx) error {
	userId := GetUserID(c)
	confirmation := c.FormValue("confirmation")

	if confirmation != "confirm" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "verify before deleting account",
		})
	}
	err := h.s.RemoveUser(c.Context(), userId)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Unable to delete user",
		})
	}

	return c.SendStatus(fiber.StatusOK)

}
