package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
)

func GetUserID(c *fiber.Ctx) int64 {
	userID, _ := strconv.Atoi(c.Locals("user_id").(string))
	return int64(userID)
}
