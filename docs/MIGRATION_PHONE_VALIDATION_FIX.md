# Исправление валидации телефона при миграции

## Проблема

При миграции данных из Excel администраторы не могли быть добавлены к чатам из-за валидации телефонов через maxbot-service. Это блокировало процесс миграции, так как:

1. Maxbot-service был недоступен или не отвечал
2. Телефоны в исторических данных могли не соответствовать текущим требованиям валидации
3. Для миграции не требуется проверка телефонов через внешний сервис

## Решение

Создан отдельный gRPC метод `AddAdministratorForMigration` для добавления администраторов без валидации телефона.

### Изменения в chat-service

#### 1. Protobuf определение (api/proto/chat.proto)

Добавлен новый метод в сервис:

```protobuf
service ChatService {
  // ... существующие методы
  
  // AddAdministratorForMigration добавляет администратора без валидации телефона (только для миграции)
  rpc AddAdministratorForMigration(AddAdministratorForMigrationRequest) returns (AddAdministratorForMigrationResponse);
}

message AddAdministratorForMigrationRequest {
  int64 chat_id = 1;
  string phone = 2;
  string max_id = 3;
  bool add_user = 4;
  bool add_admin = 5;
}

message AddAdministratorForMigrationResponse {
  Administrator administrator = 1;
  string error = 2;
}

message Administrator {
  int64 id = 1;
  int64 chat_id = 2;
  string phone = 3;
  string max_id = 4;
  bool add_user = 5;
  bool add_admin = 6;
  string created_at = 7;
}
```

#### 2. Use Case (internal/usecase/chat_service.go)

Добавлен флаг `skipPhoneValidation` в метод `AddAdministratorWithFlags`:

```go
func (s *ChatService) AddAdministratorWithFlags(
    chatID int64, 
    phone string, 
    maxID string, 
    addUser bool, 
    addAdmin bool, 
    skipPhoneValidation bool,
) (*domain.Administrator, error) {
    // Валидация телефона (пропускаем для миграции)
    if !skipPhoneValidation && !s.maxService.ValidatePhone(phone) {
        return nil, domain.ErrInvalidPhone
    }
    // ... остальная логика
}
```

#### 3. gRPC Handler (internal/infrastructure/grpc/chat_handler.go)

Реализован новый метод:

```go
func (h *ChatHandler) AddAdministratorForMigration(
    ctx context.Context, 
    req *proto.AddAdministratorForMigrationRequest,
) (*proto.AddAdministratorForMigrationResponse, error) {
    admin, err := h.chatService.AddAdministratorWithFlags(
        req.ChatId,
        req.Phone,
        req.MaxId,
        req.AddUser,
        req.AddAdmin,
        true, // skipPhoneValidation = true для миграции
    )
    // ... обработка ответа
}
```

#### 4. HTTP Handler (internal/infrastructure/http/handler.go)

Добавлен опциональный флаг в HTTP API:

```go
type AddAdministratorRequest struct {
    Phone                string `json:"phone"`
    MaxID                string `json:"max_id,omitempty"`
    AddUser              bool   `json:"add_user"`
    AddAdmin             bool   `json:"add_admin"`
    SkipPhoneValidation  bool   `json:"skip_phone_validation,omitempty"`
}
```

### Изменения в migration-service

#### 1. gRPC Client (internal/infrastructure/grpc/chat_client.go)

Реализован метод для вызова нового gRPC endpoint:

```go
func (c *ChatClient) AddAdministrator(ctx context.Context, admin *domain.AdministratorData) error {
    req := &chatpb.AddAdministratorForMigrationRequest{
        ChatId:   int64(admin.ChatID),
        Phone:    admin.Phone,
        MaxId:    admin.MaxID,
        AddUser:  admin.AddUser,
        AddAdmin: admin.AddAdmin,
    }

    resp, err := c.client.AddAdministratorForMigration(ctx, req)
    // ... обработка ответа
}
```

#### 2. Composite Client (internal/infrastructure/chat/composite_client.go)

Создан композитный клиент, который использует:
- gRPC для добавления администраторов (без валидации)
- HTTP для создания университетов и чатов

```go
type CompositeClient struct {
    HTTPClient *HTTPClient
    GRPCClient interface {
        AddAdministrator(ctx context.Context, admin *domain.AdministratorData) error
    }
}
```

#### 3. Server Configuration (internal/app/server.go)

Инициализация обоих клиентов:

```go
// HTTP клиент для университетов и чатов
chatHTTPClient := chat.NewHTTPClient(s.config.Services.ChatServiceURL)

// gRPC клиент для администраторов
chatGRPCClient, err := grpc.NewChatClient(s.config.Services.ChatServiceGRPC)

// Композитный клиент
chatClientForAdmins := &chat.CompositeClient{
    HTTPClient: chatHTTPClient,
    GRPCClient: chatGRPCClient,
}
```

## Преимущества решения

1. **Разделение ответственности**: Обычные операции используют валидацию, миграция - нет
2. **Безопасность**: Специальный метод только для миграции, не доступен через обычный HTTP API
3. **Производительность**: gRPC быстрее HTTP для массовых операций
4. **Надежность**: Миграция не зависит от доступности maxbot-service
5. **Обратная совместимость**: Существующий HTTP API не изменился

## Использование

### Для миграции (через gRPC)

```go
admin := &domain.AdministratorData{
    ChatID:   chatID,
    Phone:    "+79001234567",
    MaxID:    "496728250",
    AddUser:  true,
    AddAdmin: true,
}

err := chatGRPCClient.AddAdministrator(ctx, admin)
```

### Для обычных операций (через HTTP)

```bash
curl -X POST http://chat-service:8082/chats/1/administrators \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "+79001234567",
    "max_id": "496728250",
    "add_user": true,
    "add_admin": true
  }'
```

Телефон будет валидирован через maxbot-service.

### Для обычных операций без валидации (через HTTP, если нужно)

```bash
curl -X POST http://chat-service:8082/chats/1/administrators \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "+79001234567",
    "max_id": "496728250",
    "add_user": true,
    "add_admin": true,
    "skip_phone_validation": true
  }'
```

## Конфигурация

В `migration-service` добавлена переменная окружения:

```env
CHAT_SERVICE_GRPC=chat-service:9092
```

## Тестирование

1. Запустить сервисы:
```bash
docker-compose up -d chat-service migration-service
```

2. Проверить логи:
```bash
docker-compose logs chat-service | grep "gRPC server"
docker-compose logs migration-service
```

3. Запустить миграцию:
```bash
curl -X POST http://localhost:8084/api/v1/migrate/excel \
  -F "file=@data.xlsx"
```

## Статус

✅ Protobuf определения обновлены
✅ gRPC метод реализован в chat-service
✅ gRPC клиент реализован в migration-service
✅ Композитный клиент создан
✅ Сервисы пересобраны и перезапущены
✅ Готово к тестированию
