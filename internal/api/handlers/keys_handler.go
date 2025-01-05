package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/maheshrc27/postflow/internal/service"
)

type ApiKeyHandler struct {
	s service.ApiKeyService
}

func NewApiKeyHandler(service service.ApiKeyService) *ApiKeyHandler {
	return &ApiKeyHandler{s: service}
}

func (h *ApiKeyHandler) CreateApiKey(c *fiber.Ctx) error {
	userId := GetUserID(c)

	err := h.s.Create(c.Context(), userId)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Unable to crete API Key",
		})
	}

	return c.SendStatus(fiber.StatusOK)
}

func (h *ApiKeyHandler) ListKeys(c *fiber.Ctx) error {
	userId := GetUserID(c)

	keys, err := h.s.List(c.Context(), userId)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Unable to list api keys",
		})
	}

	return c.Status(fiber.StatusOK).JSON(keys)
}

func (h *ApiKeyHandler) RemoveAPIKey(c *fiber.Ctx) error {
	userId := GetUserID(c)
	keyId := c.QueryInt("id", 0)

	err := h.s.RemoveAPIKey(c.Context(), userId, int64(keyId))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Unable to delete API Key",
		})
	}

	return c.SendStatus(fiber.StatusOK)
}
