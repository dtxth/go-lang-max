# MaxBot Service Integration Summary

## Обзор изменений

MaxBot Service был расширен дополнительными методами для полноценной интеграции с Max Messenger Bot API. Теперь сервис поддерживает не только поиск пользователей по номеру телефона, но и отправку сообщений, получение информации о чатах и многое другое.

## Новые возможности MaxBot Service

### 1. Отправка сообщений
- **SendMessage** - отправка текстовых сообщений пользователям или в чаты
- **SendNotification** - отправка VIP-уведомлений по номеру телефона

### 2. Управление чатами
- **GetChatInfo** - получение информации о чате (название, тип, количество участников)
- **GetChatMembers** - получение списка участников чата с пагинацией
- **GetChatAdmins** - получение списка администраторов чата

### 3. Пакетные операции
- **CheckPhoneNumbers** - проверка существования множества номеров телефонов

## Обновленные файлы

### MaxBot Service

1. **api/proto/maxbot.proto**
   - Добавлены новые RPC методы и message типы
   - Расширен сервис MaxBotService

2. **internal/domain/max_api_client.go**
   - Расширен интерфейс MaxAPIClient
   - Добавлены типы ChatInfo, ChatMember, ChatMembersList

3. **internal/infrastructure/maxapi/client.go**
   - Реализованы все новые методы
   - Добавлена обработка ошибок и логирование

4. **internal/usecase/maxbot_service.go**
   - Добавлены методы usecase слоя

5. **internal/infrastructure/grpc/handler.go**
   - Реализованы gRPC handlers для новых методов
   - Добавлен маппинг между proto и domain типами

6. **README.md**
   - Обновлена документация с примерами использования
   - Добавлены описания всех новых методов

7. **INTEGRATION_GUIDE.md** (новый файл)
   - Подробное руководство по интеграции для других сервисов
   - Примеры кода и конфигурации
   - Best practices и рекомендации


### Документация для других сервисов

1. **chat-service/MAXBOT_INTEGRATION_EXAMPLES.md** (новый файл)
   - Примеры расширения MaxClient
   - Use cases для синхронизации чатов
   - Отправка уведомлений администраторам
   - HTTP endpoints для новых функций
   - Автоматическая синхронизация с Max Messenger

2. **employee-service/MAXBOT_INTEGRATION_EXAMPLES.md** (новый файл)
   - Примеры расширения MaxClient
   - Use cases для уведомлений сотрудников
   - Пакетная проверка номеров телефонов
   - HTTP endpoints для новых функций
   - Валидация номеров телефонов

3. **README.md** (обновлен)
   - Добавлена информация о новых возможностях MaxBot Service
   - Ссылка на руководство по интеграции

## Архитектура интеграции

```
┌─────────────────────────────────────────────────────────────┐
│                    Other Services                            │
│              (Employee, Chat, Auth, etc.)                    │
└────────────────────────┬────────────────────────────────────┘
                         │ gRPC
                         ▼
┌─────────────────────────────────────────────────────────────┐
│                   MaxBot Service                             │
│                    (gRPC Server)                             │
└────────────────────────┬────────────────────────────────────┘
                         │ HTTP/REST
                         ▼
┌─────────────────────────────────────────────────────────────┐
│              Max Messenger Bot API                           │
│           (max-bot-api-client-go library)                    │
└─────────────────────────────────────────────────────────────┘
```

## Примеры использования

### Employee Service

```go
// Отправка уведомления сотруднику
err := employeeService.NotifyEmployee(employeeID, "Важное сообщение")

// Проверка номеров телефонов всех сотрудников вуза
result, err := employeeService.ValidateEmployeePhones(universityID)
fmt.Printf("Найдено %d из %d номеров\n", result.Existing, result.Total)
```

### Chat Service

```go
// Синхронизация информации о чате из Max Messenger
err := chatService.SyncChatFromMax(chatID)

// Отправка уведомления всем администраторам чата
err := chatService.NotifyAdministrators(chatID, "Важное уведомление")

// Получение администраторов из Max Messenger
admins, err := maxClient.GetChatAdmins(maxChatID)
```

### Auth Service (потенциальное использование)

```go
// Отправка уведомления о входе в систему
if user.Phone != "" {
    go maxClient.SendNotification(user.Phone, "Выполнен вход в систему")
}
```

## Конфигурация

### Переменные окружения для MaxBot Service

```bash
MAX_BOT_TOKEN=your-bot-token          # Обязательно
MAX_API_URL=https://api.max.ru        # По умолчанию
MAX_API_TIMEOUT=5s                    # По умолчанию
GRPC_PORT=9095                        # По умолчанию
```

### Переменные окружения для других сервисов

```bash
MAXBOT_GRPC_ADDR=localhost:9095       # Адрес MaxBot Service
MAXBOT_TIMEOUT=5s                     # Таймаут для запросов
```


## Доступные gRPC методы

### Базовые методы (уже реализованы)

1. **GetMaxIDByPhone** - получение Max ID по номеру телефона
2. **ValidatePhone** - валидация и нормализация номера телефона

### Новые методы

3. **SendMessage** - отправка сообщения пользователю или в чат
   ```protobuf
   rpc SendMessage(SendMessageRequest) returns (SendMessageResponse);
   ```

4. **SendNotification** - отправка VIP-уведомления по номеру телефона
   ```protobuf
   rpc SendNotification(SendNotificationRequest) returns (SendNotificationResponse);
   ```

5. **GetChatInfo** - получение информации о чате
   ```protobuf
   rpc GetChatInfo(GetChatInfoRequest) returns (GetChatInfoResponse);
   ```

6. **GetChatMembers** - получение списка участников чата
   ```protobuf
   rpc GetChatMembers(GetChatMembersRequest) returns (GetChatMembersResponse);
   ```

7. **GetChatAdmins** - получение списка администраторов чата
   ```protobuf
   rpc GetChatAdmins(GetChatAdminsRequest) returns (GetChatAdminsResponse);
   ```

8. **CheckPhoneNumbers** - пакетная проверка номеров телефонов
   ```protobuf
   rpc CheckPhoneNumbers(CheckPhoneNumbersRequest) returns (CheckPhoneNumbersResponse);
   ```

## Следующие шаги для интеграции

### Для Employee Service

1. Добавить методы `SendNotification` и `CheckPhoneNumbers` в MaxClient
2. Реализовать use cases для уведомлений сотрудников
3. Добавить HTTP endpoints для новых функций
4. Обновить domain интерфейс MaxService
5. (Опционально) Добавить фоновую задачу для валидации номеров

### Для Chat Service

1. Добавить методы для работы с чатами в MaxClient
2. Реализовать синхронизацию информации о чатах
3. Добавить уведомления администраторов
4. Добавить HTTP endpoints для новых функций
5. (Опционально) Добавить автоматическую синхронизацию

### Для Auth Service

1. (Опционально) Добавить уведомления о входе в систему
2. (Опционально) Добавить двухфакторную аутентификацию через Max Messenger

### Для Structure Service

1. Изучить потребности сервиса в Max API
2. Добавить необходимые интеграции

## Генерация proto файлов

После обновления proto файлов необходимо сгенерировать Go код:

```bash
# Из корневой директории проекта
./generate_proto.sh
```

Или для конкретного сервиса:

```bash
protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    maxbot-service/api/proto/maxbot.proto
```

## Тестирование

### Запуск MaxBot Service

```bash
cd maxbot-service
export MAX_BOT_TOKEN="your-bot-token"
go run cmd/maxbot/main.go
```

### Тестирование через grpcurl

```bash
# GetMaxIDByPhone
grpcurl -plaintext -d '{"phone": "+79991234567"}' \
    localhost:9095 maxbot.MaxBotService/GetMaxIDByPhone

# SendNotification
grpcurl -plaintext -d '{"phone": "+79991234567", "text": "Test notification"}' \
    localhost:9095 maxbot.MaxBotService/SendNotification

# GetChatInfo
grpcurl -plaintext -d '{"chat_id": 12345}' \
    localhost:9095 maxbot.MaxBotService/GetChatInfo
```

## Лучшие практики

1. **Используйте таймауты** - всегда устанавливайте разумные таймауты для gRPC запросов
2. **Обрабатывайте ошибки gracefully** - не блокируйте основную функциональность при недоступности Max API
3. **Кэшируйте результаты** - кэшируйте Max ID для уменьшения нагрузки
4. **Используйте пакетные операции** - CheckPhoneNumbers вместо множества GetMaxIDByPhone
5. **Логируйте с маскировкой** - маскируйте номера телефонов в логах
6. **Асинхронная отправка** - отправляйте уведомления асинхронно

## Документация

- **MaxBot Service README**: [maxbot-service/README.md](./maxbot-service/README.md)
- **Integration Guide**: [maxbot-service/INTEGRATION_GUIDE.md](./maxbot-service/INTEGRATION_GUIDE.md)
- **Chat Service Examples**: [chat-service/MAXBOT_INTEGRATION_EXAMPLES.md](./chat-service/MAXBOT_INTEGRATION_EXAMPLES.md)
- **Employee Service Examples**: [employee-service/MAXBOT_INTEGRATION_EXAMPLES.md](./employee-service/MAXBOT_INTEGRATION_EXAMPLES.md)
- **Max Messenger Bot API**: https://dev.max.ru/

## Поддержка

Для вопросов и проблем создайте issue в репозитории проекта.

