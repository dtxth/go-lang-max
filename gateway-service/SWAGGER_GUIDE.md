# Руководство по Swagger документации Gateway Service

## Обзор

Swagger документация предоставляет полное описание всех API эндпоинтов Gateway Service, включая:
- Описание всех HTTP методов и путей
- Схемы запросов и ответов
- Примеры использования
- Коды ошибок и их описания
- Интерактивное тестирование API

## Быстрый старт

### 1. Просмотр документации локально

```bash
# Запустите сервер документации
make docs-serve

# Откройте в браузере
# http://localhost:8082
```

### 2. Просмотр онлайн

Перейдите на https://editor.swagger.io/ и загрузите файл `docs/swagger.yaml`

### 3. Использование Docker

```bash
# Запустите Swagger UI в Docker
docker run -p 8082:8080 \
  -e SWAGGER_JSON=/swagger.yaml \
  -v $(pwd)/docs/swagger.yaml:/swagger.yaml \
  swaggerapi/swagger-ui
```

## Структура документации

### Основные разделы

1. **Auth** - Аутентификация и авторизация
   - Регистрация пользователей
   - Вход по email/телефону
   - Управление токенами
   - Сброс пароля
   - MAX аутентификация

2. **Chat** - Управление чатами
   - CRUD операции с чатами
   - Поиск чатов
   - Обновление участников

3. **Administrator** - Администраторы чатов
   - Добавление/удаление администраторов
   - Управление правами

4. **Employee** - Управление сотрудниками
   - CRUD операции
   - Поиск сотрудников
   - Массовые операции

5. **University** - Университеты
   - Управление университетами
   - Структура университетов

6. **Structure** - Структура
   - Создание структуры
   - Импорт из Excel

7. **Department Manager** - Менеджеры департаментов
   - Назначение менеджеров
   - Управление правами

8. **Health** - Мониторинг
   - Проверка состояния сервисов

## Использование Swagger UI

### Тестирование эндпоинтов

1. Откройте Swagger UI (http://localhost:8082)
2. Выберите нужный эндпоинт
3. Нажмите "Try it out"
4. Заполните параметры
5. Нажмите "Execute"
6. Просмотрите результат

### Аутентификация

Для эндпоинтов, требующих аутентификации:

1. Получите токен через `/login` или `/register`
2. Нажмите кнопку "Authorize" в верхней части страницы
3. Введите токен в формате: `Bearer <ваш_токен>`
4. Нажмите "Authorize"

Теперь все запросы будут включать токен автоматически.

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
    "university_id": 1
  }'
```

## Пагинация

Все эндпоинты, возвращающие списки, поддерживают пагинацию:

```
GET /employees/all?page=2&limit=20&sort_by=first_name&sort_order=asc
```

Параметры:
- `page` - номер страницы (начиная с 1)
- `limit` - количество элементов (по умолчанию 10)
- `sort_by` - поле для сортировки (по умолчанию "created_at")
- `sort_order` - порядок: "asc" или "desc" (по умолчанию "desc")

## Коды ответов

- `200 OK` - Успешный запрос
- `201 Created` - Ресурс создан
- `400 Bad Request` - Ошибка валидации
- `401 Unauthorized` - Требуется аутентификация
- `404 Not Found` - Ресурс не найден
- `405 Method Not Allowed` - Метод не разрешен
- `500 Internal Server Error` - Внутренняя ошибка
- `503 Service Unavailable` - Сервис недоступен

## Формат ошибок

```json
{
  "error": "error_type",
  "message": "Описание ошибки",
  "request_id": "req_1234567890"
}
```

## Валидация спецификации

Проверьте корректность OpenAPI спецификации:

```bash
# Используя Makefile
make docs-validate

# Или напрямую
npx swagger-parser validate docs/swagger.yaml
```

## Генерация клиентов

Swagger позволяет генерировать клиенты для различных языков:

```bash
# Установите swagger-codegen
npm install -g swagger-codegen-cli

# Генерация JavaScript клиента
swagger-codegen generate -i docs/swagger.yaml -l javascript -o clients/js

# Генерация Python клиента
swagger-codegen generate -i docs/swagger.yaml -l python -o clients/python

# Генерация Go клиента
swagger-codegen generate -i docs/swagger.yaml -l go -o clients/go
```

## Обновление документации

При добавлении новых эндпоинтов:

1. Откройте `docs/swagger.yaml`
2. Добавьте новый путь в секцию `paths`
3. Добавьте схемы данных в `components/schemas`
4. Обновите описания и примеры
5. Проверьте валидацию: `make docs-validate`
6. Проверьте в Swagger UI: `make docs-serve`

## Полезные ссылки

- [OpenAPI Specification](https://swagger.io/specification/)
- [Swagger Editor](https://editor.swagger.io/)
- [Swagger UI](https://swagger.io/tools/swagger-ui/)
- [Swagger Codegen](https://swagger.io/tools/swagger-codegen/)

## Поддержка

При возникновении проблем с документацией:

1. Проверьте валидацию спецификации
2. Убедитесь, что сервер документации запущен
3. Проверьте логи Gateway Service
4. Обратитесь к команде разработки

## Дополнительные инструменты

### Redoc

Альтернативный просмотр документации:

```bash
npm install -g redoc-cli
redoc-cli serve docs/swagger.yaml
```

### Postman

Импорт в Postman:
1. Откройте Postman
2. File → Import
3. Выберите `docs/swagger.yaml`
4. Используйте сгенерированную коллекцию

### VS Code

Расширения для работы со Swagger:
- Swagger Viewer
- OpenAPI (Swagger) Editor
- REST Client