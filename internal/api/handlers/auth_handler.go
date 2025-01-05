package handlers

import (
	"fmt"
	"net/url"
	"time"

	"github.com/gofiber/fiber/v2"
	config "github.com/maheshrc27/postflow/configs"
	"github.com/maheshrc27/postflow/internal/service"
	"github.com/maheshrc27/postflow/pkg/utils"
)

type AuthHandler struct {
	s   service.AuthService
	cfg config.Config
}

func NewAuthHandler(cfg config.Config, service service.AuthService) *AuthHandler {
	return &AuthHandler{s: service, cfg: cfg}
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	authURL := "https://accounts.google.com/o/oauth2/v2/auth"
	params := url.Values{}
	params.Add("client_id", h.cfg.GoogleClientID)
	params.Add("redirect_uri", "http://localhost:3000/login/callback")
	params.Add("response_type", "code")
	params.Add("scope", "https://www.googleapis.com/auth/userinfo.profile https://www.googleapis.com/auth/userinfo.email")
	params.Add("state", "secureRandomState")
	params.Add("access_type", "offline")

	fullURL := fmt.Sprintf("%s?%s", authURL, params.Encode())
	return c.Redirect(fullURL)
}

func (h *AuthHandler) LoginCallbackHandler(c *fiber.Ctx) error {
	code := c.Query("code")

	userID, err := h.s.LoginCallback(c.Context(), code)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "something went wrong",
		})
	}

	token, err := utils.GenerateToken(h.cfg.SecretKey, fmt.Sprintf("%d", userID), 24*time.Hour)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "something went wrong",
		})
	}

	c.Cookie(&fiber.Cookie{
		Name:     h.cfg.CookieName,
		Value:    token,
		HTTPOnly: true,
		Secure:   false,
		SameSite: fiber.CookieSameSiteNoneMode,
		Path:     "/",
		Expires:  time.Now().Add(24 * time.Hour),
	})

	return c.Redirect(h.cfg.FrontendURL, fiber.StatusTemporaryRedirect)
}
