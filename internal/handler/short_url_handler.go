// internal/handler/shorturl_handler.go
package handler

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/matheusparro/shorty/internal/handler/dto"
	"github.com/matheusparro/shorty/internal/service"
	"github.com/matheusparro/shorty/internal/service/input"
	sinput "github.com/matheusparro/shorty/internal/service/input"
)

type ShortURLHandler struct {
	service *service.ShortURLService
	baseURL string
}

// baseURL é fixo (config) e fica guardado no handler
func NewShortURLHandler(svc *service.ShortURLService, baseURL string) *ShortURLHandler {
	return &ShortURLHandler{
		service: svc,
		baseURL: strings.TrimRight(baseURL, "/"),
	}
}

func (h *ShortURLHandler) Create(c *fiber.Ctx) error {
	var req dto.CreateShortURLRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid payload"})
	}

	userID, ok := c.Locals("userID").(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	// expires_at: string -> *time.Time
	var expiresAt *time.Time
	if req.ExpiresAt != nil {
		t, err := time.Parse(time.RFC3339, *req.ExpiresAt)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "expires_at must be RFC3339"})
		}
		expiresAt = &t
	}

	in := input.CreateShortURLInput{
		OriginalURL: req.URL,
		UserID:      userID,
		ExpiresAt:   expiresAt,
	}

	entity, err := h.service.CreateShortURL(c.Context(), in)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	// Se tua rota de redirect é /api/v1/r/:code, monta certinho:
	shortURL := h.baseURL + "/" + entity.ShortCode

	return c.Status(fiber.StatusCreated).JSON(dto.CreateShortURLResponse{
		ShortCode: entity.ShortCode,
		ShortURL:  shortURL,
	})
}

func (h *ShortURLHandler) Redirect(c *fiber.Ctx) error {
	code := c.Params("code")

	url, err := h.service.FindActiveForRedirect(c.Context(), sinput.RedirectInput{
		Code:      code,
		IP:        c.IP(),
		UserAgent: c.Get("User-Agent"),
		Referer:   c.Get("Referer"),
	})
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "short URL not found"})
	}

	return c.Redirect(url, fiber.StatusTemporaryRedirect)
}

