package handlers

import (
	"fmt"
	"log"
	"log/slog"
	"strconv"

	"github.com/gofiber/fiber/v2"
	config "github.com/maheshrc27/postflow/configs"
	"github.com/maheshrc27/postflow/internal/service"
	"github.com/maheshrc27/postflow/pkg/utils"
)

type PlatformHandler struct {
	ps  service.PlatformService
	ig  service.InstagramService
	tt  service.TiktokService
	yt  service.YoutubeService
	cfg config.Config
}

func NewPlatformHandler(ps service.PlatformService, ig service.InstagramService, tt service.TiktokService, yt service.YoutubeService, cfg config.Config) *PlatformHandler {
	return &PlatformHandler{
		ps:  ps,
		ig:  ig,
		tt:  tt,
		yt:  yt,
		cfg: cfg,
	}
}

func (h *PlatformHandler) AddSocialAccount(c *fiber.Ctx) error {
	authURL := h.ps.GetAuthURL(c.Context(), c.Params("platform"), c.Query("state"))
	return c.Redirect(authURL)
}

func (h *PlatformHandler) CallbackHandler(c *fiber.Ctx) error {
	code := c.Query("code")
	state := c.Query("state")
	platform := c.Params("platform")

	claims, err := utils.ValidateToken(h.cfg.SecretKey, state)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Unable to validate user",
		})
	}

	userID, err := strconv.ParseInt(claims.UserID, 10, 64)
	if err != nil {
		slog.Info(err.Error())
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Unable to validate user",
		})
	}

	switch platform {
	case "instagram":
		err = h.ig.InstagramCallback(c.Context(), code, userID)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "something went wrong",
			})
		}
	case "tiktok":
		err = h.tt.TiktokCallback(c.Context(), code, userID)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "something went wrong",
			})
		}
	case "youtube":
		err = h.yt.YoutubeCallback(c.Context(), code, userID)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Something went wrong",
			})
		}
	}

	redirectURL := fmt.Sprintf("%s/dashboard/accounts", h.cfg.FrontendURL)
	return c.Redirect(redirectURL, fiber.StatusTemporaryRedirect)
}

func (h *PlatformHandler) ListSocialAccounts(c *fiber.Ctx) error {
	userID := GetUserID(c)

	accountList, err := h.ps.List(c.Context(), userID)
	if err != nil {
		log.Println(err.Error())
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch social accounts",
		})
	}

	return c.Status(fiber.StatusOK).JSON(accountList)
}

func (h *PlatformHandler) DeleteSocialAccount(c *fiber.Ctx) error {
	userID := GetUserID(c)
	accountId := c.QueryInt("id", 0)

	err := h.ps.Delete(c.Context(), userID, int64(accountId))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Unable to delete social account",
		})
	}

	return c.SendStatus(fiber.StatusOK)
}
