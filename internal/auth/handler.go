package auth

import (
	"errors"
	response "sentenceminer/pkg/http/response"

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

func (h *Handler) GoogleLogin(c *fiber.Ctx) error {
	// In a real app, use a random state and store it in session/cookie
	state := "random-state"
	url := h.service.GetGoogleLoginURL(state)
	return c.Redirect(url)
}

func (h *Handler) GoogleCallback(c *fiber.Ctx) error {
	code := c.Query("code")
	if code == "" {
		return fiber.NewError(fiber.StatusBadRequest, "missing code")
	}

	// Verify state here if we stored it

	loginResp, err := h.service.HandleGoogleCallback(c.Context(), code)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.Status(fiber.StatusOK).JSON(response.NewResponse("login successful", fiber.StatusOK, loginResp))
}
