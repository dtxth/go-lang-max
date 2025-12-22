# Gateway Service API Documentation

Этот каталог содержит документацию API для Gateway Service.

## Файлы

- `swagger.yaml` - OpenAPI 3.0 спецификация для всех эндпоинтов Gateway Service

## Просмотр документации

### Swagger UI

Для просмотра интерактивной документации можно использовать Swagger UI:

1. **Онлайн редактор Swagger**:
   - Перейдите на https://editor.swagger.io/
   - Скопируйте содержимое файла `swagger.yaml`
   - Вставьте в редактор

2. **Локальный Swagger UI**:
   ```bash
   # Установите swagger-ui-serve
   npm install -g swagger-ui-serve
   
   # Запустите сервер документации
   swagger-ui-serve swagger.yaml
   ```

3. **Docker**:
   ```bash
   # Запустите Swagger UI в Docker
   docker run -p 8081:8080 -e SWAGGER_JSON=/swagger.yaml -v $(pwd)/swagger.yaml:/swagger.yaml swaggerapi/swagger-ui
   ```

### Redoc

Альтернативно можно использовать Redoc:

```bash
# Установите redoc-cli
npm install -g redoc-cli

# Сгенерируйте HTML документацию
redoc-cli build swagger.yaml --output docs.html
```

## Структура API

Gateway Service предоставляет единую точку входа для следующих сервисов:

### Auth Service
- Регистрация и аутентификация пользователей
- Управление токенами (access/refresh)
- Сброс и изменение паролей
- Аутентификация через MAX (Telegram Mini App)

### Chat Service
- Управление чатами
- Управление администраторами чатов
- Поиск чатов
- Обновление количества участников

### Employee Service
- CRUD операции с сотрудниками
- Поиск сотрудников
- Массовые операции (batch update MAX ID)
- Создание простых сотрудников

### Structure Service
- Управление университетами
- Управление структурой университетов
- Импорт данных из Excel
- Управление менеджерами департаментов

## Аутентификация

Большинство эндпоинтов требуют аутентификации через Bearer токен:

```
Authorization: Bearer <access_token>
```

Получить токен можно через эндпоинты:
- `POST /login` - вход по email
- `POST /login-phone` - вход по телефону
- `POST /register` - регистрация
- `POST /auth/max` - аутентификация через MAX

## Пагинация

Эндпоинты, возвращающие списки, поддерживают пагинацию:

- `page` - номер страницы (начиная с 1)
- `limit` - количество элементов на странице (по умолчанию 10)
- `sort_by` - поле для сортировки (по умолчанию "created_at")
- `sort_order` - порядок сортировки: "asc" или "desc" (по умолчанию "desc")

Пример:
```
GET /employees/all?page=2&limit=20&sort_by=first_name&sort_order=asc
```

## Обработка ошибок

API возвращает стандартные HTTP коды состояния:

- `200` - Успешный запрос
- `201` - Ресурс создан
- `400` - Ошибка валидации данных
- `401` - Неавторизован
- `404` - Ресурс не найден
- `405` - Метод не разрешен
- `500` - Внутренняя ошибка сервера
- `503` - Сервис недоступен

Формат ошибки:
```json
{
  "error": "error_type",
  "message": "Human readable error message",
  "request_id": "req_1234567890"
}
```

## Мониторинг

- `GET /health` - проверка состояния Gateway Service и всех подключенных сервисов
- `GET /metrics` - метрики системы

## Примеры использования

### Регистрация пользователя
```bash
curl -X POST http://localhost:8080/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123",
    "phone": "+1234567890"
  }'
```

### Получение списка сотрудников
```bash
curl -X GET "http://localhost:8080/employees/all?page=1&limit=10" \
  -H "Authorization: Bearer <access_token>"
```

### Создание чата
```bash
curl -X POST http://localhost:8080/chats \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <access_token>" \
  -d '{
    "name": "General Chat",
    "url": "https://t.me/generalchat",
    "max_chat_id": -1001234567890,
    "university_id": 1,
    "department": "Computer Science"
  }'
```

## Обновление документации

При изменении API необходимо обновить файл `swagger.yaml`:

1. Добавить новые эндпоинты в секцию `paths`
2. Добавить новые схемы данных в секцию `components/schemas`
3. Обновить описания и примеры
4. Проверить корректность спецификации в Swagger Editor

## Валидация спецификации

Для проверки корректности OpenAPI спецификации:

```bash
# Установите swagger-codegen
npm install -g swagger-codegen-cli

# Проверьте спецификацию
swagger-codegen validate -i swagger.yaml
```

Или используйте онлайн валидатор: https://validator.swagger.io/