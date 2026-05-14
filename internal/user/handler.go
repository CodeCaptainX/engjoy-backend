package user

import (
	"errors"
	"sentenceminer/pkg/http/response"

	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Login(c *fiber.Ctx) error {
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request")
	}

	loginResponse, err := h.service.Login(c.Context(), req)
	if err != nil {
		if errors.Is(err, ErrInvalidLogin) {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid email or password")
		}
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.Status(fiber.StatusOK).JSON(response.NewResponse("login successful", fiber.StatusOK, loginResponse))
}
