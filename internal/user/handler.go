package user

import (
	response "sentenceminer/pkg/http/response"

	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) AddFavorite(c *fiber.Ctx) error {
	userID := c.Locals("userID").(int64)

	var req struct {
		SentenceUUID string `json:"sentenceUuid"`
	}
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request")
	}

	if err := h.service.AddFavorite(c.Context(), userID, req.SentenceUUID); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.Status(fiber.StatusOK).JSON(response.NewResponse(c, "favorite added", fiber.StatusOK, nil))
}

func (h *Handler) RemoveFavorite(c *fiber.Ctx) error {
	userID := c.Locals("userID").(int64)
	sentenceUUID := c.Params("sentenceUuid")

	if err := h.service.RemoveFavorite(c.Context(), userID, sentenceUUID); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.Status(fiber.StatusOK).JSON(response.NewResponse(c, "favorite removed", fiber.StatusOK, nil))
}

func (h *Handler) GetFavorites(c *fiber.Ctx) error {
	userID := c.Locals("userID").(int64)

	favorites, err := h.service.GetFavorites(c.Context(), userID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.Status(fiber.StatusOK).JSON(response.NewResponse(c, "favorites fetched", fiber.StatusOK, fiber.Map{"favorites": favorites}))
}
