package domain

import "time"

type EventMeta struct {
	EventID        string    `json:"event_id"`
	Version        int       `json:"version"`
	CorrelationID  string    `json:"correlation_id,omitempty"`
	OccurredAt     time.Time `json:"occurred_at"`
	Source         string    `json:"source,omitempty"`
}

type ClickEvent struct {
	EventMeta
	ShortCode string `json:"short_code"`
	IP        string `json:"ip"`
	UserAgent string `json:"user_agent"`
	Referer   string `json:"referer,omitempty"`
	Country   string `json:"country,omitempty"`
}

type URLCreatedEvent struct {
	EventMeta
	ID          string     `json:"id"`
	ShortCode   string     `json:"short_code"`
	OriginalURL string     `json:"original_url"`
	UserID      string     `json:"user_id"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
}

type URLExpiredEvent struct {
	EventMeta
	ShortCode string    `json:"short_code"`
	UserID    string    `json:"user_id"`
	ExpiredAt time.Time `json:"expired_at"`
}
