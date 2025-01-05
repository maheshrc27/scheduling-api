package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/maheshrc27/postflow/internal/service"
	"github.com/maheshrc27/postflow/internal/transfer"
)

type SettingsHandler struct {
	s service.SettingsService
}

func NewSettingsHandler(service service.SettingsService) *SettingsHandler {
	return &SettingsHandler{s: service}
}

func (h *SettingsHandler) GetSettingsInfo(c *fiber.Ctx) error {
	userId := GetUserID(c)

	settingsInfo, err := h.s.GetSettingsInfo(c.Context(), userId)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Unable to find settings for given user",
		})
	}

	return c.JSON(settingsInfo)
}

func (h *SettingsHandler) UpdateSettings(c *fiber.Ctx) error {
	userId := GetUserID(c)

	var settings transfer.SettingsUpdate
	err := c.BodyParser(&settings)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Unable to parse json",
		})
	}

	err = h.s.UpdateSettings(c.Context(), userId, settings.PostingTime, settings.Category)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Unable to update settings",
		})
	}

	return c.SendStatus(fiber.StatusOK)
}
