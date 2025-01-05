package handlers

import (
	"log/slog"

	"github.com/gofiber/fiber/v2"
	"github.com/maheshrc27/postflow/internal/service"
	"github.com/maheshrc27/postflow/internal/transfer"
)

type PaymentHandler struct {
	s service.SubscriptionService
}

func NewPaymentHandler(service service.SubscriptionService) *PaymentHandler {
	return &PaymentHandler{s: service}
}

func (h *PaymentHandler) PaymentWebhook(c *fiber.Ctx) error {

	var requestData transfer.SubscriptionEvent

	if err := c.BodyParser(&requestData); err != nil {
		slog.Info(err.Error())
		return err
	}

	err := h.s.HandleSubscription(c.Context(), &requestData)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	return c.SendStatus(fiber.StatusOK)
}
