// internal/handler/dto/create_shorturl_response.go
package dto

type CreateShortURLResponse struct {
	ShortCode string `json:"short_code"`
	ShortURL  string `json:"short_url"`
}
