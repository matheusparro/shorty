package handler

import (
	"errors"

	"github.com/gofiber/fiber/v2"

	"github.com/matheusparro/shorty/internal/domain"
	"github.com/matheusparro/shorty/internal/service"
)

type AuthHandler struct {
	auth *service.AuthService
}

func NewAuthHandler(auth *service.AuthService) *AuthHandler {
	return &AuthHandler{auth: auth}
}

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req registerRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid payload"})
	}

	res, err := h.auth.Register(c.Context(), req.Email, req.Password)
	if err != nil {
		return h.mapError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"userId": res.UserID,
		"email":  res.Email,
	})
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req loginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid payload"})
	}

	res, err := h.auth.Login(c.Context(), req.Email, req.Password)
	if err != nil {
		return h.mapError(c, err)
	}

	return c.JSON(fiber.Map{
		"userId": res.UserID,
		"email":  res.Email,
	})
}

func (h *AuthHandler) mapError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, domain.ErrInvalidEmail),
		errors.Is(err, domain.ErrInvalidPassword):
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})

	case errors.Is(err, domain.ErrEmailAlreadyUsed):
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": err.Error()})

	case errors.Is(err, domain.ErrInvalidCredentials):
		// não vaza detalhe
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid credentials"})

	default:
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal error"})
	}
}
