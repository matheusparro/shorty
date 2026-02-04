// internal/handler/dto/create_shorturl_request.go
package dto

type CreateShortURLRequest struct {
	URL       string `json:"url"`
	ExpiresAt *string `json:"expires_at,omitempty"` // RFC3339 (ex: 2026-02-10T12:00:00Z)
}
