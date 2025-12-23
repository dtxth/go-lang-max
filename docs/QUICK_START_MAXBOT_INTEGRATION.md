# Quick Start: MaxBot Service Integration

## 🎯 Что было сделано

MaxBot Service был расширен **8 новыми методами** для полноценной работы с Max Messenger Bot API:

### Новые возможности

✅ **Отправка сообщений**
- SendMessage - отправка в чаты и пользователям
- SendNotification - VIP-уведомления по номеру телефона

✅ **Управление чатами**
- GetChatInfo - информация о чате
- GetChatMembers - список участников с пагинацией
- GetChatAdmins - список администраторов

✅ **Пакетные операции**
- CheckPhoneNumbers - проверка множества номеров

## 📁 Созданные файлы

### Документация
- `MAXBOT_INTEGRATION_SUMMARY.md` - обзор всех изменений
- `INTEGRATION_CHECKLIST.md` - чеклист для интеграции
- `QUICK_START_MAXBOT_INTEGRATION.md` - этот файл

### MaxBot Service
- `maxbot-service/INTEGRATION_GUIDE.md` - подробное руководство по интеграции
- Обновлены: proto, domain, infrastructure, usecase, handler, README

### Примеры для других сервисов
- `chat-service/MAXBOT_INTEGRATION_EXAMPLES.md` - примеры для Chat Service
- `employee-service/MAXBOT_INTEGRATION_EXAMPLES.md` - примеры для Employee Service

## 🚀 Быстрый старт

### Шаг 1: Сгенерировать proto файлы

```bash
# Установить protoc (если не установлен)
brew install protobuf  # macOS

# Сгенерировать код
./generate_proto.sh
```

### Шаг 2: Запустить MaxBot Service

```bash
cd maxbot-service
export MAX_BOT_TOKEN="your-bot-token"
go run cmd/maxbot/main.go
```

### Шаг 3: Протестировать

```bash
# Установить grpcurl
brew install grpcurl

# Тест GetMaxIDByPhone
grpcurl -plaintext -d '{"phone": "+79991234567"}' \
    localhost:9095 maxbot.MaxBotService/GetMaxIDByPhone

# Тест SendNotification
grpcurl -plaintext -d '{"phone": "+79991234567", "text": "Test"}' \
    localhost:9095 maxbot.MaxBotService/SendNotification
```

## 📖 Примеры использования

### Employee Service - Уведомление сотрудника

```go
// Отправить уведомление
err := maxClient.SendNotification(employee.Phone, "Важное сообщение")

// Проверить номера всех сотрудников вуза
result, err := employeeService.ValidateEmployeePhones(universityID)
fmt.Printf("Найдено %d из %d\n", result.Existing, result.Total)
```

### Chat Service - Синхронизация чата

```go
// Получить информацию о чате из Max Messenger
chatInfo, err := maxClient.GetChatInfo(maxChatID)

// Получить администраторов
admins, err := maxClient.GetChatAdmins(maxChatID)

// Уведомить всех администраторов
err := chatService.NotifyAdministrators(chatID, "Важное уведомление")
```


## 🔧 Интеграция в другие сервисы

### Employee Service

1. **Добавить методы в MaxClient** (`internal/infrastructure/max/max_client.go`):
   - `SendNotification(phone, text string) error`
   - `CheckPhoneNumbers(phones []string) ([]string, error)`

2. **Обновить domain интерфейс** (`internal/domain/max_service.go`)

3. **Добавить use cases** (`internal/usecase/employee_service.go`):
   - `NotifyEmployee(employeeID int64, message string) error`
   - `NotifyUniversityEmployees(universityID int64, message string) error`
   - `ValidateEmployeePhones(universityID int64) (*PhoneValidationResult, error)`

4. **Добавить HTTP endpoints**:
   - `POST /employees/{id}/notify`
   - `POST /universities/{id}/notify`
   - `GET /universities/{id}/validate-phones`

📄 **Полные примеры**: `employee-service/MAXBOT_INTEGRATION_EXAMPLES.md`

### Chat Service

1. **Добавить методы в MaxClient** (`internal/infrastructure/max/max_client.go`):
   - `SendNotification(phone, text string) error`
   - `GetChatInfo(maxChatID int64) (*ChatInfo, error)`
   - `GetChatAdmins(maxChatID int64) ([]*ChatMember, error)`

2. **Обновить domain интерфейс** (`internal/domain/max_service.go`)

3. **Добавить use cases** (`internal/usecase/chat_service.go`):
   - `SyncChatFromMax(chatID int64) error`
   - `SyncChatAdmins(chatID int64) error`
   - `NotifyAdministrators(chatID int64, message string) error`

4. **Добавить HTTP endpoints**:
   - `POST /chats/{id}/sync`
   - `POST /chats/{id}/notify`

5. **(Опционально) Добавить автоматическую синхронизацию**

📄 **Полные примеры**: `chat-service/MAXBOT_INTEGRATION_EXAMPLES.md`

## 📚 Документация

### Основная документация
- **MaxBot Service README**: `maxbot-service/README.md`
  - Описание всех методов
  - Примеры использования gRPC API
  - Конфигурация и переменные окружения

- **Integration Guide**: `maxbot-service/INTEGRATION_GUIDE.md`
  - Подробное руководство по интеграции
  - Примеры кода для всех методов
  - Best practices и рекомендации
  - Конфигурация Docker Compose
  - Примеры тестирования

### Примеры для сервисов
- **Chat Service**: `chat-service/MAXBOT_INTEGRATION_EXAMPLES.md`
  - Расширение MaxClient
  - Синхронизация чатов
  - Уведомления администраторов
  - HTTP endpoints
  - Автоматическая синхронизация

- **Employee Service**: `employee-service/MAXBOT_INTEGRATION_EXAMPLES.md`
  - Расширение MaxClient
  - Уведомления сотрудников
  - Валидация номеров
  - HTTP endpoints
  - Пакетная проверка

### Сводная информация
- **Integration Summary**: `MAXBOT_INTEGRATION_SUMMARY.md`
  - Обзор всех изменений
  - Архитектура интеграции
  - Список обновленных файлов

- **Integration Checklist**: `INTEGRATION_CHECKLIST.md`
  - Пошаговый чеклист
  - Задачи для каждого сервиса
  - Тестирование и деплой

## 🎓 Доступные методы MaxBot Service

| Метод | Описание | Использование |
|-------|----------|---------------|
| GetMaxIDByPhone | Получить Max ID по телефону | Employee, Chat |
| ValidatePhone | Валидация номера | Employee, Chat |
| SendMessage | Отправить сообщение | Chat, Employee |
| SendNotification | VIP-уведомление | Employee, Chat, Auth |
| GetChatInfo | Информация о чате | Chat |
| GetChatMembers | Участники чата | Chat |
| GetChatAdmins | Администраторы чата | Chat |
| CheckPhoneNumbers | Пакетная проверка | Employee |

## ⚙️ Конфигурация

### MaxBot Service

```bash
MAX_BOT_TOKEN=your-bot-token          # Обязательно!
MAX_API_URL=https://api.max.ru        # По умолчанию
MAX_API_TIMEOUT=5s                    # По умолчанию
GRPC_PORT=9095                        # По умолчанию
```

### Другие сервисы

```bash
MAXBOT_GRPC_ADDR=localhost:9095       # Адрес MaxBot Service
MAXBOT_TIMEOUT=5s                     # Таймаут запросов
```

### Docker Compose

```yaml
services:
  maxbot-service:
    environment:
      - MAX_BOT_TOKEN=${MAX_BOT_TOKEN}
      - GRPC_PORT=9095
    ports:
      - "9095:9095"

  employee-service:
    environment:
      - MAXBOT_GRPC_ADDR=maxbot-service:9095
    depends_on:
      - maxbot-service

  chat-service:
    environment:
      - MAXBOT_GRPC_ADDR=maxbot-service:9095
    depends_on:
      - maxbot-service
```

## 🧪 Тестирование

### Через grpcurl

```bash
# GetMaxIDByPhone
grpcurl -plaintext -d '{"phone": "+79991234567"}' \
    localhost:9095 maxbot.MaxBotService/GetMaxIDByPhone

# SendNotification
grpcurl -plaintext -d '{"phone": "+79991234567", "text": "Test"}' \
    localhost:9095 maxbot.MaxBotService/SendNotification

# GetChatInfo
grpcurl -plaintext -d '{"chat_id": 12345}' \
    localhost:9095 maxbot.MaxBotService/GetChatInfo

# CheckPhoneNumbers
grpcurl -plaintext -d '{"phones": ["+79991234567", "+79997654321"]}' \
    localhost:9095 maxbot.MaxBotService/CheckPhoneNumbers
```

### Через HTTP (после интеграции)

```bash
# Employee Service - уведомление
curl -X POST http://localhost:8081/employees/1/notify \
  -H "Content-Type: application/json" \
  -d '{"message": "Test notification"}'

# Chat Service - синхронизация
curl -X POST http://localhost:8082/chats/1/sync
```

## 💡 Best Practices

1. **Таймауты**: Используйте 5-10 секунд для gRPC запросов
2. **Graceful degradation**: Сервисы должны работать при недоступности Max API
3. **Кэширование**: Кэшируйте Max ID для уменьшения нагрузки
4. **Пакетные операции**: Используйте CheckPhoneNumbers вместо множества GetMaxIDByPhone
5. **Асинхронность**: Отправляйте уведомления асинхронно (goroutines)
6. **Логирование**: Маскируйте номера телефонов в логах

## 🐛 Troubleshooting

### Proto файлы не генерируются

```bash
# Установить protoc
brew install protobuf

# Установить Go плагины
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Добавить в PATH
export PATH="$PATH:$(go env GOPATH)/bin"
```

### MaxBot Service не запускается

```bash
# Проверить токен
echo $MAX_BOT_TOKEN

# Проверить порт
lsof -i :9095

# Проверить логи
docker-compose logs maxbot-service
```

### Ошибки при вызове методов

- Проверьте, что MaxBot Service запущен
- Проверьте правильность адреса MAXBOT_GRPC_ADDR
- Проверьте таймауты
- Проверьте логи обоих сервисов

## 📞 Поддержка

- **Документация Max API**: https://dev.max.ru/
- **GitHub Issues**: Создайте issue в репозитории
- **Вопросы**: Обратитесь к команде разработки

## ✅ Следующие шаги

1. ✅ Сгенерировать proto файлы: `./generate_proto.sh`
2. ✅ Запустить MaxBot Service с токеном
3. ✅ Протестировать базовые методы через grpcurl
4. 📝 Интегрировать в Employee Service (см. `employee-service/MAXBOT_INTEGRATION_EXAMPLES.md`)
5. 📝 Интегрировать в Chat Service (см. `chat-service/MAXBOT_INTEGRATION_EXAMPLES.md`)
6. 📝 Обновить Swagger документацию
7. 📝 Протестировать все endpoints
8. 🚀 Деплой

---

**Готово к использованию!** Все необходимые методы реализованы и задокументированы.

