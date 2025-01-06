package middleware

import (
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
	config "github.com/maheshrc27/scheduling-api/configs"
	"github.com/maheshrc27/scheduling-api/internal/service"
	"github.com/maheshrc27/scheduling-api/pkg/utils"
)

type AuthMiddleware struct {
	s   service.ApiKeyService
	cfg config.Config
}

func NewAuthMiddleware(cfg config.Config, service service.ApiKeyService) *AuthMiddleware {
	return &AuthMiddleware{s: service, cfg: cfg}
}

func (m *AuthMiddleware) AuthMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		tokenString := c.Cookies(m.cfg.CookieName)
		apiKey := c.Query("api_key")

		if tokenString == "" && apiKey == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Missing Keys or cookies",
			})
		}

		if apiKey != "" {
			userID, err := m.s.GetUserID(c.Context(), apiKey)
			if err != nil {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"error": err.Error(),
				})
			}
			c.Locals("user_id", fmt.Sprintf("%d", userID))
		} else if tokenString != "" {

			claims, err := utils.ValidateToken(m.cfg.SecretKey, tokenString)
			if err != nil {
				c.Cookie(&fiber.Cookie{
					Name:   m.cfg.CookieName,
					Value:  "",
					Path:   "/",
					MaxAge: -1, // Delete cookie
				})

				log.Printf("Token validation failed: %v", err)
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"error": "Invalid or expired token",
				})
			}

			c.Locals("user_id", claims.UserID)
		}
		return c.Next()
	}
}
