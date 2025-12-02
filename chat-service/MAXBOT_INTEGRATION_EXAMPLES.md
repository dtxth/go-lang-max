# MaxBot Integration Examples for Chat Service

Примеры расширенного использования MaxBot Service в Chat Service.

## Дополнительные методы для MaxClient

Добавьте следующие методы в `internal/infrastructure/max/max_client.go`:

```go
// SendNotification отправляет уведомление администратору
func (c *MaxClient) SendNotification(phone, text string) error {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	resp, err := c.client.SendNotification(ctx, &maxbotproto.SendNotificationRequest{
		Phone: phone,
		Text:  text,
	})
	if err != nil {
		return err
	}

	if resp.Error != "" {
		return mapError(resp.ErrorCode, resp.Error)
	}

	if !resp.Success {
		return errors.New("failed to send notification")
	}

	return nil
}

// GetChatInfo получает информацию о чате из Max Messenger
func (c *MaxClient) GetChatInfo(maxChatID int64) (*ChatInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	resp, err := c.client.GetChatInfo(ctx, &maxbotproto.GetChatInfoRequest{
		ChatId: maxChatID,
	})
	if err != nil {
		return nil, err
	}

	if resp.Error != "" {
		return nil, mapError(resp.ErrorCode, resp.Error)
	}

	return &ChatInfo{
		ChatID:            resp.Chat.ChatId,
		Title:             resp.Chat.Title,
		Type:              resp.Chat.Type,
		ParticipantsCount: int(resp.Chat.ParticipantsCount),
		Description:       resp.Chat.Description,
	}, nil
}

// GetChatAdmins получает список администраторов чата
func (c *MaxClient) GetChatAdmins(maxChatID int64) ([]*ChatMember, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	resp, err := c.client.GetChatAdmins(ctx, &maxbotproto.GetChatAdminsRequest{
		ChatId: maxChatID,
	})
	if err != nil {
		return nil, err
	}

	if resp.Error != "" {
		return nil, mapError(resp.ErrorCode, resp.Error)
	}

	admins := make([]*ChatMember, 0, len(resp.Admins))
	for _, admin := range resp.Admins {
		admins = append(admins, &ChatMember{
			UserID:  admin.UserId,
			Name:    admin.Name,
			IsAdmin: admin.IsAdmin,
			IsOwner: admin.IsOwner,
		})
	}

	return admins, nil
}
```


## Расширение Domain интерфейса

Обновите `internal/domain/max_service.go`:

```go
package domain

// MaxService определяет интерфейс для работы с MAX API
type MaxService interface {
	// Существующие методы
	GetMaxIDByPhone(phone string) (string, error)
	ValidatePhone(phone string) bool
	
	// Новые методы
	SendNotification(phone, text string) error
	GetChatInfo(maxChatID int64) (*ChatInfo, error)
	GetChatAdmins(maxChatID int64) ([]*ChatMember, error)
}

// ChatInfo содержит информацию о чате из Max Messenger
type ChatInfo struct {
	ChatID            int64
	Title             string
	Type              string
	ParticipantsCount int
	Description       string
}

// ChatMember представляет участника чата
type ChatMember struct {
	UserID  int64
	Name    string
	IsAdmin bool
	IsOwner bool
}
```

## Use Case: Синхронизация информации о чате

Добавьте в `internal/usecase/chat_service.go`:

```go
// SyncChatFromMax синхронизирует информацию о чате из Max Messenger
func (s *ChatService) SyncChatFromMax(chatID int64) error {
	chat, err := s.chatRepo.GetByID(chatID)
	if err != nil {
		return domain.ErrChatNotFound
	}

	// Получаем актуальную информацию из Max Messenger
	maxChatInfo, err := s.maxService.GetChatInfo(chat.MaxChatID)
	if err != nil {
		return fmt.Errorf("failed to get chat info from Max: %w", err)
	}

	// Обновляем локальную информацию
	chat.Name = maxChatInfo.Title
	chat.ParticipantsCount = maxChatInfo.ParticipantsCount

	if err := s.chatRepo.Update(chat); err != nil {
		return fmt.Errorf("failed to update chat: %w", err)
	}

	return nil
}

// SyncChatAdmins синхронизирует администраторов чата из Max Messenger
func (s *ChatService) SyncChatAdmins(chatID int64) error {
	chat, err := s.chatRepo.GetByID(chatID)
	if err != nil {
		return domain.ErrChatNotFound
	}

	// Получаем администраторов из Max Messenger
	maxAdmins, err := s.maxService.GetChatAdmins(chat.MaxChatID)
	if err != nil {
		return fmt.Errorf("failed to get chat admins from Max: %w", err)
	}

	// Получаем текущих администраторов из БД
	currentAdmins, err := s.administratorRepo.GetByChatID(chatID)
	if err != nil {
		return fmt.Errorf("failed to get current admins: %w", err)
	}

	// Создаем map для быстрого поиска
	currentAdminsMap := make(map[int64]*domain.Administrator)
	for _, admin := range currentAdmins {
		currentAdminsMap[admin.MaxID] = admin
	}

	// Добавляем новых администраторов
	for _, maxAdmin := range maxAdmins {
		if _, exists := currentAdminsMap[maxAdmin.UserID]; !exists {
			// Администратор есть в Max, но нет в БД - добавляем
			admin := &domain.Administrator{
				ChatID: chatID,
				MaxID:  fmt.Sprintf("%d", maxAdmin.UserID),
				// Phone будет пустым, так как Max API не возвращает телефоны
			}
			if err := s.administratorRepo.Create(admin); err != nil {
				log.Printf("Failed to add admin %d: %v", maxAdmin.UserID, err)
			}
		}
	}

	return nil
}
```


## Use Case: Уведомления администраторов

```go
// NotifyAdministrators отправляет уведомление всем администраторам чата
func (s *ChatService) NotifyAdministrators(chatID int64, message string) error {
	admins, err := s.administratorRepo.GetByChatID(chatID)
	if err != nil {
		return fmt.Errorf("failed to get administrators: %w", err)
	}

	var errors []error
	successCount := 0

	for _, admin := range admins {
		if admin.Phone == "" {
			continue
		}

		// Отправляем уведомление асинхронно
		go func(phone, msg string) {
			if err := s.maxService.SendNotification(phone, msg); err != nil {
				log.Printf("Failed to notify admin %s: %v", maskPhone(phone), err)
			}
		}(admin.Phone, message)

		successCount++
	}

	if successCount == 0 {
		return fmt.Errorf("no administrators with phone numbers found")
	}

	return nil
}

// NotifyAdministratorSync отправляет уведомление администратору синхронно
func (s *ChatService) NotifyAdministratorSync(adminID int64, message string) error {
	admin, err := s.administratorRepo.GetByID(adminID)
	if err != nil {
		return domain.ErrAdministratorNotFound
	}

	if admin.Phone == "" {
		return fmt.Errorf("administrator has no phone number")
	}

	if err := s.maxService.SendNotification(admin.Phone, message); err != nil {
		return fmt.Errorf("failed to send notification: %w", err)
	}

	return nil
}

func maskPhone(phone string) string {
	if len(phone) <= 4 {
		return "****"
	}
	return "****" + phone[len(phone)-4:]
}
```

## HTTP Handler: Новые эндпоинты

Добавьте в `internal/infrastructure/http/handler.go`:

```go
// SyncChatInfo синхронизирует информацию о чате из Max Messenger
// @Summary Sync chat info from Max Messenger
// @Tags chats
// @Param id path int true "Chat ID"
// @Success 200 {object} map[string]string
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /chats/{id}/sync [post]
func (h *Handler) SyncChatInfo(w http.ResponseWriter, r *http.Request) {
	chatID, err := getChatIDFromPath(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid chat ID")
		return
	}

	if err := h.service.SyncChatFromMax(chatID); err != nil {
		if errors.Is(err, domain.ErrChatNotFound) {
			respondError(w, http.StatusNotFound, "Chat not found")
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"message": "Chat info synchronized successfully",
	})
}

// NotifyAdmins отправляет уведомление всем администраторам чата
// @Summary Notify chat administrators
// @Tags chats
// @Param id path int true "Chat ID"
// @Param request body NotifyRequest true "Notification message"
// @Success 200 {object} map[string]string
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /chats/{id}/notify [post]
func (h *Handler) NotifyAdmins(w http.ResponseWriter, r *http.Request) {
	chatID, err := getChatIDFromPath(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid chat ID")
		return
	}

	var req NotifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Message == "" {
		respondError(w, http.StatusBadRequest, "Message is required")
		return
	}

	if err := h.service.NotifyAdministrators(chatID, req.Message); err != nil {
		if errors.Is(err, domain.ErrChatNotFound) {
			respondError(w, http.StatusNotFound, "Chat not found")
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"message": "Notifications sent successfully",
	})
}

type NotifyRequest struct {
	Message string `json:"message"`
}
```


## Обновление роутера

Добавьте новые маршруты в `internal/infrastructure/http/router.go`:

```go
// Синхронизация с Max Messenger
r.Post("/chats/{id}/sync", h.SyncChatInfo)
r.Post("/chats/{id}/notify", h.NotifyAdmins)
```

## Примеры использования API

### Синхронизация информации о чате

```bash
curl -X POST http://localhost:8082/chats/123/sync
```

Ответ:
```json
{
  "message": "Chat info synchronized successfully"
}
```

### Отправка уведомления администраторам

```bash
curl -X POST http://localhost:8082/chats/123/notify \
  -H "Content-Type: application/json" \
  -d '{
    "message": "Важное уведомление для администраторов чата"
  }'
```

Ответ:
```json
{
  "message": "Notifications sent successfully"
}
```

## Автоматическая синхронизация

Добавьте фоновую задачу для периодической синхронизации:

```go
// internal/app/sync_worker.go
package app

import (
	"context"
	"log"
	"time"

	"chat-service/internal/usecase"
)

type SyncWorker struct {
	service  *usecase.ChatService
	interval time.Duration
	stop     chan struct{}
}

func NewSyncWorker(service *usecase.ChatService, interval time.Duration) *SyncWorker {
	return &SyncWorker{
		service:  service,
		interval: interval,
		stop:     make(chan struct{}),
	}
}

func (w *SyncWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	log.Printf("Starting sync worker with interval: %v", w.interval)

	for {
		select {
		case <-ticker.C:
			w.syncAllChats()
		case <-w.stop:
			log.Println("Stopping sync worker")
			return
		case <-ctx.Done():
			log.Println("Context cancelled, stopping sync worker")
			return
		}
	}
}

func (w *SyncWorker) Stop() {
	close(w.stop)
}

func (w *SyncWorker) syncAllChats() {
	log.Println("Starting chat synchronization...")

	// Получаем все чаты
	chats, _, err := w.service.GetAllChats(100, 0, "superadmin", nil)
	if err != nil {
		log.Printf("Failed to get chats: %v", err)
		return
	}

	successCount := 0
	errorCount := 0

	for _, chat := range chats {
		if err := w.service.SyncChatFromMax(chat.ID); err != nil {
			log.Printf("Failed to sync chat %d: %v", chat.ID, err)
			errorCount++
		} else {
			successCount++
		}
	}

	log.Printf("Chat synchronization completed: %d success, %d errors", successCount, errorCount)
}
```

Запуск worker в main.go:

```go
// Создаем и запускаем sync worker
syncWorker := app.NewSyncWorker(chatService, 1*time.Hour)
go syncWorker.Start(context.Background())

// Graceful shutdown
defer syncWorker.Stop()
```

## Конфигурация

Добавьте в `internal/config/config.go`:

```go
type Config struct {
	// ... существующие поля
	
	// MaxBot Service
	MaxBotGRPCAddr string
	MaxBotTimeout  time.Duration
	
	// Sync Worker
	SyncInterval time.Duration
}

func Load() (*Config, error) {
	// ... существующий код
	
	maxBotTimeout := 5 * time.Second
	if timeout := os.Getenv("MAXBOT_TIMEOUT"); timeout != "" {
		if d, err := time.ParseDuration(timeout); err == nil {
			maxBotTimeout = d
		}
	}
	
	syncInterval := 1 * time.Hour
	if interval := os.Getenv("SYNC_INTERVAL"); interval != "" {
		if d, err := time.ParseDuration(interval); err == nil {
			syncInterval = d
		}
	}
	
	return &Config{
		// ... существующие поля
		
		MaxBotGRPCAddr: getEnv("MAXBOT_GRPC_ADDR", "localhost:9095"),
		MaxBotTimeout:  maxBotTimeout,
		SyncInterval:   syncInterval,
	}, nil
}
```

## Docker Compose

Обновите `docker-compose.yml`:

```yaml
services:
  chat-service:
    environment:
      - MAXBOT_GRPC_ADDR=maxbot-service:9095
      - MAXBOT_TIMEOUT=5s
      - SYNC_INTERVAL=1h
    depends_on:
      - maxbot-service
```

