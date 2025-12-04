# Сводка: Исправление валидации телефона при миграции

## Что было сделано

Создан отдельный gRPC метод `AddAdministratorForMigration` для добавления администраторов без валидации телефона через maxbot-service.

## Ключевые изменения

### chat-service

1. **Protobuf** (`api/proto/chat.proto`):
   - Добавлен метод `AddAdministratorForMigration`
   - Добавлены сообщения `AddAdministratorForMigrationRequest`, `AddAdministratorForMigrationResponse`, `Administrator`

2. **Use Case** (`internal/usecase/chat_service.go`):
   - Добавлен параметр `skipPhoneValidation` в `AddAdministratorWithFlags`
   - Валидация телефона пропускается, если `skipPhoneValidation = true`

3. **gRPC Handler** (`internal/infrastructure/grpc/chat_handler.go`):
   - Реализован метод `AddAdministratorForMigration` с `skipPhoneValidation = true`

4. **HTTP Handler** (`internal/infrastructure/http/handler.go`):
   - Добавлен опциональный флаг `skip_phone_validation` в `AddAdministratorRequest`

### migration-service

1. **gRPC Client** (`internal/infrastructure/grpc/chat_client.go`):
   - Реализован метод `AddAdministrator` через gRPC вызов `AddAdministratorForMigration`

2. **Composite Client** (`internal/infrastructure/chat/composite_client.go`):
   - Создан композитный клиент, использующий gRPC для администраторов и HTTP для остального

3. **Server** (`internal/app/server.go`):
   - Инициализация gRPC клиента для chat-service
   - Использование композитного клиента в use case

## Результат

✅ Миграция больше не зависит от доступности maxbot-service  
✅ Администраторы добавляются без валидации телефона  
✅ Обычные операции продолжают использовать валидацию  
✅ Сервисы пересобраны и работают  

## Команды для проверки

```bash
# Проверить, что gRPC сервер запущен
docker-compose exec chat-service netstat -tuln | grep 9092

# Проверить логи
docker-compose logs chat-service --tail=20
docker-compose logs migration-service --tail=20

# Запустить миграцию
curl -X POST http://localhost:8084/api/v1/migrate/excel \
  -F "file=@data.xlsx"
```

## Файлы документации

- `docs/MIGRATION_PHONE_VALIDATION_FIX.md` - подробное описание изменений
- `docs/MIGRATION_PHONE_VALIDATION_SUMMARY.md` - краткая сводка (этот файл)
