package domain

import (
	"context"
	"time"
)

// MaxWebhookEvent представляет входящее webhook событие от MAX Messenger
type MaxWebhookEvent struct {
	Type     string         `json:"type"`
	Message  *MessageEvent  `json:"message,omitempty"`
	Callback *CallbackEvent `json:"callback_query,omitempty"`
}

// MessageEvent представляет событие нового сообщения
type MessageEvent struct {
	From UserInfo        `json:"from"`
	Text string          `json:"text"`
	Chat WebhookChatInfo `json:"chat"`
}

// CallbackEvent представляет событие callback query
type CallbackEvent struct {
	User UserInfo        `json:"user"`
	Data string          `json:"data"`
	Chat WebhookChatInfo `json:"chat"`
}

// UserInfo содержит информацию о пользователе из webhook события
type UserInfo struct {
	UserID    string `json:"user_id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

// WebhookChatInfo содержит информацию о чате из webhook события
type WebhookChatInfo struct {
	ChatID int64  `json:"chat_id"`
	Type   string `json:"type"`
}

// WebhookHandler определяет интерфейс для обработки webhook событий
type WebhookHandler interface {
	// HandleMaxWebhook обрабатывает входящее webhook событие от MAX
	HandleMaxWebhook(ctx context.Context, event MaxWebhookEvent) error
}

// WebhookProcessingResult содержит результат обработки webhook события
type WebhookProcessingResult struct {
	UserID          string    `json:"user_id"`
	ProfileExtracted bool      `json:"profile_extracted"`
	ProfileUpdated   bool      `json:"profile_updated"`
	ProcessedAt      time.Time `json:"processed_at"`
	EventType        string    `json:"event_type"`
}