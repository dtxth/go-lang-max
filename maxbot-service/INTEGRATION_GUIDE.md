# MaxBot Service Integration Guide

Это руководство описывает, как интегрировать MaxBot Service в другие микросервисы проекта.

## Обзор

MaxBot Service предоставляет gRPC API для взаимодействия с Max Messenger Bot API. Сервис поддерживает:

- Поиск пользователей по номеру телефона
- Отправку сообщений и уведомлений
- Получение информации о чатах и участниках
- Валидацию номеров телефонов

## Быстрый старт

### 1. Добавление зависимости

Добавьте в `go.mod` вашего сервиса:

```go
require (
    maxbot-service v0.0.0
    google.golang.org/grpc v1.66.0
)
```

### 2. Создание gRPC клиента

Создайте клиент для взаимодействия с MaxBot Service:

```go
package max

import (
    "context"
    "time"
    
    maxbotproto "maxbot-service/api/proto"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
)

type MaxClient struct {
    conn    *grpc.ClientConn
    client  maxbotproto.MaxBotServiceClient
    timeout time.Duration
}

func NewMaxClient(address string, timeout time.Duration) (*MaxClient, error) {
    conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
    if err != nil {
        return nil, err
    }

    return &MaxClient{
        conn:    conn,
        client:  maxbotproto.NewMaxBotServiceClient(conn),
        timeout: timeout,
    }, nil
}

func (c *MaxClient) Close() error {
    if c.conn == nil {
        return nil
    }
    return c.conn.Close()
}
```


### 3. Реализация методов

#### Получение Max ID по номеру телефона

```go
func (c *MaxClient) GetMaxIDByPhone(phone string) (string, error) {
    ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
    defer cancel()

    resp, err := c.client.GetMaxIDByPhone(ctx, &maxbotproto.GetMaxIDByPhoneRequest{
        Phone: phone,
    })
    if err != nil {
        return "", err
    }

    if resp.Error != "" {
        return "", mapError(resp.ErrorCode, resp.Error)
    }

    return resp.MaxId, nil
}
```

#### Валидация номера телефона

```go
func (c *MaxClient) ValidatePhone(phone string) bool {
    ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
    defer cancel()

    resp, err := c.client.ValidatePhone(ctx, &maxbotproto.ValidatePhoneRequest{
        Phone: phone,
    })
    if err != nil {
        return false
    }

    if resp.Error != "" {
        return false
    }

    return resp.Valid
}
```

#### Отправка сообщения в чат

```go
func (c *MaxClient) SendMessageToChat(chatID int64, text string) (string, error) {
    ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
    defer cancel()

    resp, err := c.client.SendMessage(ctx, &maxbotproto.SendMessageRequest{
        Recipient: &maxbotproto.SendMessageRequest_ChatId{ChatId: chatID},
        Text:      text,
    })
    if err != nil {
        return "", err
    }

    if resp.Error != "" {
        return "", mapError(resp.ErrorCode, resp.Error)
    }

    return resp.MessageId, nil
}
```


#### Отправка уведомления пользователю

```go
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
```

#### Получение информации о чате

```go
func (c *MaxClient) GetChatInfo(chatID int64) (*ChatInfo, error) {
    ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
    defer cancel()

    resp, err := c.client.GetChatInfo(ctx, &maxbotproto.GetChatInfoRequest{
        ChatId: chatID,
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
```

#### Получение администраторов чата

```go
func (c *MaxClient) GetChatAdmins(chatID int64) ([]*ChatMember, error) {
    ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
    defer cancel()

    resp, err := c.client.GetChatAdmins(ctx, &maxbotproto.GetChatAdminsRequest{
        ChatId: chatID,
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


#### Проверка существования номеров телефонов

```go
func (c *MaxClient) CheckPhoneNumbers(phones []string) ([]string, error) {
    ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
    defer cancel()

    resp, err := c.client.CheckPhoneNumbers(ctx, &maxbotproto.CheckPhoneNumbersRequest{
        Phones: phones,
    })
    if err != nil {
        return nil, err
    }

    if resp.Error != "" {
        return nil, mapError(resp.ErrorCode, resp.Error)
    }

    return resp.ExistingPhones, nil
}
```

### 4. Маппинг ошибок

```go
func mapError(code maxbotproto.ErrorCode, message string) error {
    switch code {
    case maxbotproto.ErrorCode_ERROR_CODE_INVALID_PHONE:
        return domain.ErrInvalidPhone
    case maxbotproto.ErrorCode_ERROR_CODE_MAX_ID_NOT_FOUND:
        return domain.ErrMaxIDNotFound
    default:
        return errors.New(message)
    }
}
```

## Примеры использования в сервисах

### Employee Service

Employee Service использует MaxBot для получения Max ID при добавлении сотрудника:

```go
// В usecase/employee_service.go
func (s *EmployeeService) AddEmployeeByPhone(phone, firstName, lastName string) (*domain.Employee, error) {
    // Валидация телефона
    if !s.maxService.ValidatePhone(phone) {
        return nil, domain.ErrInvalidPhone
    }
    
    // Получение Max ID
    maxID, err := s.maxService.GetMaxIDByPhone(phone)
    if err != nil {
        return nil, err
    }
    
    // Создание сотрудника
    employee := &domain.Employee{
        FirstName: firstName,
        LastName:  lastName,
        Phone:     phone,
        MaxID:     maxID,
    }
    
    return s.employeeRepo.Create(employee)
}
```


### Chat Service

Chat Service может использовать MaxBot для:
- Добавления администраторов по номеру телефона
- Получения информации о чатах из Max Messenger
- Отправки уведомлений администраторам

```go
// Добавление администратора
func (s *ChatService) AddAdministrator(chatID int64, phone string) (*domain.Administrator, error) {
    // Валидация телефона
    if !s.maxService.ValidatePhone(phone) {
        return nil, domain.ErrInvalidPhone
    }
    
    // Получение Max ID
    maxID, err := s.maxService.GetMaxIDByPhone(phone)
    if err != nil {
        return nil, err
    }
    
    // Создание администратора
    admin := &domain.Administrator{
        ChatID: chatID,
        Phone:  phone,
        MaxID:  maxID,
    }
    
    return s.administratorRepo.Create(admin)
}

// Синхронизация информации о чате
func (s *ChatService) SyncChatInfo(chatID int64) error {
    // Получение информации из Max Messenger
    chatInfo, err := s.maxClient.GetChatInfo(chatID)
    if err != nil {
        return err
    }
    
    // Обновление локальной информации
    chat, err := s.chatRepo.GetByMaxChatID(chatID)
    if err != nil {
        return err
    }
    
    chat.Name = chatInfo.Title
    chat.ParticipantsCount = chatInfo.ParticipantsCount
    
    return s.chatRepo.Update(chat)
}

// Отправка уведомления администраторам чата
func (s *ChatService) NotifyAdmins(chatID int64, message string) error {
    admins, err := s.administratorRepo.GetByChatID(chatID)
    if err != nil {
        return err
    }
    
    for _, admin := range admins {
        if err := s.maxClient.SendNotification(admin.Phone, message); err != nil {
            log.Printf("Failed to notify admin %s: %v", admin.Phone, err)
        }
    }
    
    return nil
}
```

### Auth Service

Auth Service может использовать MaxBot для отправки уведомлений о входе:

```go
func (s *AuthService) Login(email, password string) (*TokenPair, error) {
    user, err := s.repo.GetByEmail(email)
    if err != nil {
        return nil, domain.ErrInvalidCreds
    }
    
    if !s.hasher.Compare(password, user.Password) {
        return nil, domain.ErrInvalidCreds
    }
    
    tokens, err := s.jwtManager.GenerateTokens(user.ID, user.Email, user.Role)
    if err != nil {
        return nil, err
    }
    
    // Отправка уведомления о входе (опционально)
    if user.Phone != "" {
        go s.maxClient.SendNotification(user.Phone, "Выполнен вход в систему")
    }
    
    return tokens, nil
}
```


## Конфигурация

### Переменные окружения

Добавьте в конфигурацию вашего сервиса:

```bash
# MaxBot Service gRPC адрес
MAXBOT_GRPC_ADDR=localhost:9095

# Таймаут для запросов к MaxBot Service
MAXBOT_TIMEOUT=5s
```

### Пример конфигурации (config.go)

```go
package config

import (
    "os"
    "time"
)

type Config struct {
    // ... другие поля
    
    MaxBotGRPCAddr string
    MaxBotTimeout  time.Duration
}

func Load() (*Config, error) {
    maxBotTimeout := 5 * time.Second
    if timeout := os.Getenv("MAXBOT_TIMEOUT"); timeout != "" {
        if d, err := time.ParseDuration(timeout); err == nil {
            maxBotTimeout = d
        }
    }
    
    return &Config{
        // ... другие поля
        
        MaxBotGRPCAddr: getEnv("MAXBOT_GRPC_ADDR", "localhost:9095"),
        MaxBotTimeout:  maxBotTimeout,
    }, nil
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}
```

### Инициализация в main.go

```go
package main

import (
    "log"
    
    "your-service/internal/config"
    "your-service/internal/infrastructure/max"
)

func main() {
    cfg, err := config.Load()
    if err != nil {
        log.Fatal(err)
    }
    
    // Создание MaxBot клиента
    maxClient, err := max.NewMaxClient(cfg.MaxBotGRPCAddr, cfg.MaxBotTimeout)
    if err != nil {
        log.Fatalf("Failed to create MaxBot client: %v", err)
    }
    defer maxClient.Close()
    
    // Передача клиента в usecase
    employeeService := usecase.NewEmployeeService(
        employeeRepo,
        universityRepo,
        maxClient, // MaxService interface
    )
    
    // ... остальная инициализация
}
```


## Docker Compose

Добавьте MaxBot Service в ваш `docker-compose.yml`:

```yaml
services:
  your-service:
    build: .
    environment:
      - MAXBOT_GRPC_ADDR=maxbot-service:9095
      - MAXBOT_TIMEOUT=5s
    depends_on:
      - maxbot-service
    networks:
      - app-network

  maxbot-service:
    build: ./maxbot-service
    environment:
      - MAX_API_TOKEN=${MAX_API_TOKEN}
      - MAX_API_URL=https://api.max.ru
      - MAX_API_TIMEOUT=5s
      - GRPC_PORT=9095
    ports:
      - "9095:9095"
    networks:
      - app-network

networks:
  app-network:
    driver: bridge
```

## Обработка ошибок

### Типы ошибок

MaxBot Service возвращает следующие типы ошибок:

1. **ERROR_CODE_INVALID_PHONE** - Неверный формат номера телефона
2. **ERROR_CODE_MAX_ID_NOT_FOUND** - Пользователь не найден в Max Messenger
3. **ERROR_CODE_INTERNAL** - Внутренняя ошибка (аутентификация, таймаут и т.д.)

### Рекомендации по обработке

```go
func (s *Service) ProcessPhone(phone string) error {
    maxID, err := s.maxClient.GetMaxIDByPhone(phone)
    if err != nil {
        switch {
        case errors.Is(err, domain.ErrInvalidPhone):
            // Неверный формат - вернуть ошибку пользователю
            return fmt.Errorf("invalid phone format: %w", err)
            
        case errors.Is(err, domain.ErrMaxIDNotFound):
            // Пользователь не найден - можно продолжить без Max ID
            log.Printf("User not found in Max Messenger: %s", phone)
            return nil
            
        default:
            // Внутренняя ошибка - повторить позже
            log.Printf("MaxBot service error: %v", err)
            return fmt.Errorf("temporary error, please try again")
        }
    }
    
    // Успешно получен Max ID
    return s.saveMaxID(phone, maxID)
}
```

## Тестирование

### Мок для тестирования

```go
package mocks

import "context"

type MockMaxClient struct {
    GetMaxIDByPhoneFunc func(phone string) (string, error)
    ValidatePhoneFunc   func(phone string) bool
}

func (m *MockMaxClient) GetMaxIDByPhone(phone string) (string, error) {
    if m.GetMaxIDByPhoneFunc != nil {
        return m.GetMaxIDByPhoneFunc(phone)
    }
    return "mock_max_id", nil
}

func (m *MockMaxClient) ValidatePhone(phone string) bool {
    if m.ValidatePhoneFunc != nil {
        return m.ValidatePhoneFunc(phone)
    }
    return true
}
```

### Пример теста

```go
func TestAddEmployee(t *testing.T) {
    mockMax := &mocks.MockMaxClient{
        GetMaxIDByPhoneFunc: func(phone string) (string, error) {
            if phone == "+79991234567" {
                return "max_id_123", nil
            }
            return "", domain.ErrMaxIDNotFound
        },
    }
    
    service := NewEmployeeService(repo, universityRepo, mockMax)
    
    employee, err := service.AddEmployeeByPhone("+79991234567", "John", "Doe")
    assert.NoError(t, err)
    assert.Equal(t, "max_id_123", employee.MaxID)
}
```


## Лучшие практики

### 1. Используйте таймауты

Всегда устанавливайте разумные таймауты для запросов к MaxBot Service:

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
```

### 2. Обрабатывайте ошибки gracefully

Не блокируйте основную функциональность при недоступности Max API:

```go
maxID, err := s.maxClient.GetMaxIDByPhone(phone)
if err != nil {
    log.Printf("Failed to get Max ID: %v", err)
    // Продолжить без Max ID
    maxID = ""
}
```

### 3. Кэшируйте результаты

Кэшируйте Max ID для уменьшения нагрузки на API:

```go
// Проверяем кэш перед запросом
if cachedMaxID := s.cache.Get(phone); cachedMaxID != "" {
    return cachedMaxID, nil
}

maxID, err := s.maxClient.GetMaxIDByPhone(phone)
if err != nil {
    return "", err
}

// Сохраняем в кэш
s.cache.Set(phone, maxID, 24*time.Hour)
return maxID, nil
```

### 4. Используйте пакетные операции

Для проверки множества номеров используйте `CheckPhoneNumbers`:

```go
// Вместо множества отдельных запросов
phones := []string{"+79991234567", "+79997654321", "+79995555555"}
existingPhones, err := s.maxClient.CheckPhoneNumbers(phones)
```

### 5. Логируйте с маскировкой

Маскируйте номера телефонов в логах:

```go
func maskPhone(phone string) string {
    if len(phone) <= 4 {
        return "****"
    }
    return "****" + phone[len(phone)-4:]
}

log.Printf("Processing phone: %s", maskPhone(phone))
```

### 6. Асинхронная отправка уведомлений

Отправляйте уведомления асинхронно, чтобы не блокировать основной поток:

```go
go func() {
    if err := s.maxClient.SendNotification(phone, message); err != nil {
        log.Printf("Failed to send notification: %v", err)
    }
}()
```

## Миграция существующего кода

Если у вас уже есть интеграция с Max API, следуйте этим шагам:

### Шаг 1: Обновите domain интерфейс

```go
// Старый интерфейс
type MaxService interface {
    GetMaxIDByPhone(phone string) (string, error)
    ValidatePhone(phone string) bool
}

// Новый интерфейс (добавьте новые методы)
type MaxService interface {
    GetMaxIDByPhone(phone string) (string, error)
    ValidatePhone(phone string) bool
    SendNotification(phone, text string) error
    GetChatInfo(chatID int64) (*ChatInfo, error)
    // ... другие методы
}
```

### Шаг 2: Замените реализацию

Замените прямые вызовы Max API на вызовы через MaxBot Service gRPC.

### Шаг 3: Обновите конфигурацию

Добавьте переменные окружения для MaxBot Service.

### Шаг 4: Протестируйте

Убедитесь, что все существующие функции работают корректно.

## Поддержка

Для вопросов и проблем:
- Документация MaxBot Service: [README.md](./README.md)
- Max Messenger Bot API: https://dev.max.ru/
- GitHub Issues: [создайте issue в репозитории]

