package user

import (
	response "sentenceminer/pkg/http/response"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) AddFavorite(c *fiber.Ctx) error {

	var req struct {
		UserID     int64 `json:"userId"`
		SentenceID int64 `json:"sentenceId"`
	}
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request")
	}

	if err := h.service.AddFavorite(c.Context(), req.UserID, req.SentenceID); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.Status(fiber.StatusOK).JSON(response.NewResponse("favorite added", fiber.StatusOK, nil))
}

func (h *Handler) RemoveFavorite(c *fiber.Ctx) error {
	var req struct {
		UserID     int64 `json:"userId"`
		SentenceID int64 `json:"sentenceId"`
	}
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request")
	}

	if err := h.service.RemoveFavorite(c.Context(), req.UserID, req.SentenceID); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.Status(fiber.StatusOK).JSON(response.NewResponse("favorite removed", fiber.StatusOK, nil))
}

func (h *Handler) GetFavorites(c *fiber.Ctx) error {
	userID, err := strconv.ParseInt(c.Params("userId"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid user id")
	}

	favorites, err := h.service.GetFavorites(c.Context(), userID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.Status(fiber.StatusOK).JSON(response.NewResponse("favorites fetched", fiber.StatusOK, fiber.Map{"favorites": favorites}))
}
