package feed

import (
	"sentenceminer/internal/feed/service"
	"sentenceminer/pkg/http/response"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
)

type FeedHandler struct {
	service *service.FeedService
}

func NewFeedHandler(db *sqlx.DB) *FeedHandler {
	return &FeedHandler{service: service.NewFeedService(db)}
}

func (h *FeedHandler) getLearningFeed(c *fiber.Ctx) error {
	var userID int64
	if c.Locals("userID") != nil {
		userID = c.Locals("userID").(int64)
	} else {
		userID = 0 // Anonymous
	}

	page, _ := strconv.Atoi(c.Query("page", "1"))
	perPage, _ := strconv.Atoi(c.Query("perPage", "20"))

	items, err := h.service.GetLearningFeed(userID, page, perPage)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.Status(fiber.StatusOK).JSON(response.NewResponse(c, "learning feed fetched", fiber.StatusOK, fiber.Map{
		"items": items,
	}))
}
