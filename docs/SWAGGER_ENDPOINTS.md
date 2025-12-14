# Swagger UI - Все сервисы

Полный список Swagger UI для всех микросервисов системы "Цифровой Вуз".

## Доступ к Swagger UI

После запуска сервисов через `docker-compose up`, Swagger UI доступен по следующим адресам:

### 1. Auth Service
**URL:** http://localhost:8080/swagger/index.html

**Endpoints:**
- `POST /register` - Регистрация нового пользователя
- `POST /login` - Вход в систему (получение JWT токенов)
- `POST /refresh` - Обновление access токена
- `POST /logout` - Выход из системы (инвалидация refresh токена)
- `GET /health` - Health check

**Описание:** Сервис аутентификации и авторизации с поддержкой JWT токенов и ролевой модели ABAC.

---

### 2. Employee Service
**URL:** http://localhost:8081/swagger/index.html

**Endpoints:**
- `POST /employees` - Создание сотрудника (с автоматическим получением MAX_id)
- `GET /employees` - Поиск сотрудников (с фильтрацией по ролям)
- `GET /employees/all` - Список всех сотрудников
- `GET /employees/{id}` - Получение сотрудника по ID
- `PUT /employees/{id}` - Обновление сотрудника
- `DELETE /employees/{id}` - Удаление сотрудника
- `POST /employees/batch-update-maxid` - Пакетное обновление MAX_id
- `GET /employees/batch-status` - Статус пакетного обновления
- `GET /employees/batch-status/{id}` - Статус конкретного batch job

**Описание:** Управление сотрудниками вузов с интеграцией ролей и автоматическим получением MAX_id.

---

### 3. Chat Service
**URL:** http://localhost:8082/swagger/index.html

**Endpoints:**
- `POST /chats` - Создание чата
- `GET /chats` - Список чатов (с фильтрацией по ролям и пагинацией)
- `GET /chats/all` - Все чаты
- `GET /chats/{id}` - Получение чата по ID
- `PUT /chats/{id}` - Обновление чата
- `DELETE /chats/{id}` - Удаление чата
- `POST /chats/{chat_id}/administrators` - Добавление администратора чата
- `DELETE /administrators/{admin_id}` - Удаление администратора

**Описание:** Управление групповыми чатами с ролевой фильтрацией и поддержкой администраторов.

---

### 4. Structure Service
**URL:** http://localhost:8083/swagger/index.html

**Endpoints:**
- `POST /import/excel` - Импорт структуры из Excel файла
- `GET /universities` - Список университетов
- `GET /universities/{id}` - Получение университета по ID
- `GET /universities/{university_id}/structure` - Полная иерархия университета
- `POST /departments/managers` - Назначение оператора на подразделение
- `GET /departments/managers` - Список операторов подразделений
- `DELETE /departments/managers/{id}` - Удаление назначения оператора

**Описание:** Управление иерархической структурой вузов (University → Branch → Faculty → Group → Chat).

---

### 5. Migration Service
**URL:** http://localhost:8084/swagger/index.html

**Endpoints:**
- `POST /migration/database` - Миграция из базы данных (6,000 чатов)
- `POST /migration/google-sheets` - Миграция из Google Sheets
- `POST /migration/excel` - Миграция из Excel (155,000+ чатов)
- `GET /migration/jobs` - Список всех миграций
- `GET /migration/jobs/{id}` - Статус конкретной миграции

**Описание:** Миграция данных из трех различных источников с отслеживанием прогресса.

---

## Быстрый доступ

Если все сервисы запущены локально, используйте эти ссылки:

| Сервис | Swagger UI | Порт |
|--------|-----------|------|
| Auth Service | [http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html) | 8080 |
| Employee Service | [http://localhost:8081/swagger/index.html](http://localhost:8081/swagger/index.html) | 8081 |
| Chat Service | [http://localhost:8082/swagger/index.html](http://localhost:8082/swagger/index.html) | 8082 |
| Structure Service | [http://localhost:8083/swagger/index.html](http://localhost:8083/swagger/index.html) | 8083 |
| Migration Service | [http://localhost:8084/swagger/index.html](http://localhost:8084/swagger/index.html) | 8084 |

## Обновление Swagger документации

### Обновить все сервисы:

```bash
./update_swagger.sh
```

### Обновить конкретный сервис:

```bash
cd <service-name>
make swagger

# Или напрямую
swag init -g cmd/<service>/main.go -o internal/infrastructure/http/docs
```

### Примеры:

```bash
# Auth Service
cd auth-service && make swagger

# Employee Service
cd employee-service && make swagger

# Chat Service
cd chat-service && make swagger

# Structure Service
cd structure-service && make swagger

# Migration Service
cd migration-service && make swagger
```

## Использование Swagger UI

### 1. Аутентификация

Для защищенных endpoints (помеченных 🔒):

1. Получите JWT токен через `/login` в Auth Service
2. Нажмите кнопку **"Authorize"** в правом верхнем углу Swagger UI
3. Введите: `Bearer <ваш_access_token>`
4. Нажмите **"Authorize"**

### 2. Тестирование endpoints

1. Выберите endpoint
2. Нажмите **"Try it out"**
3. Заполните параметры
4. Нажмите **"Execute"**
5. Просмотрите ответ

### 3. Примеры запросов

Swagger UI автоматически генерирует примеры запросов для каждого endpoint с правильной структурой данных.

## Swagger JSON/YAML

Если нужен raw Swagger spec:

- **JSON**: `http://localhost:<port>/swagger/doc.json`
- **YAML**: Доступен в `internal/infrastructure/http/docs/swagger.yaml`

### Примеры:

```bash
# Auth Service Swagger JSON
curl http://localhost:8080/swagger/doc.json

# Employee Service Swagger JSON
curl http://localhost:8081/swagger/doc.json
```

## Интеграция с внешними инструментами

### Postman

1. Откройте Postman
2. File → Import
3. Вставьте URL: `http://localhost:8080/swagger/doc.json`
4. Импортируйте коллекцию

### Insomnia

1. Откройте Insomnia
2. Create → Import from URL
3. Вставьте URL Swagger JSON
4. Импортируйте

### OpenAPI Generator

Генерация клиентов для различных языков:

```bash
# Установка
npm install -g @openapitools/openapi-generator-cli

# Генерация TypeScript клиента
openapi-generator-cli generate \
  -i http://localhost:8080/swagger/doc.json \
  -g typescript-axios \
  -o ./generated-client

# Генерация Python клиента
openapi-generator-cli generate \
  -i http://localhost:8080/swagger/doc.json \
  -g python \
  -o ./generated-client
```

## Troubleshooting

### Swagger UI не открывается

1. Проверьте, что сервис запущен:
```bash
curl http://localhost:8080/health
```

2. Проверьте логи сервиса:
```bash
docker-compose logs auth-service
```

3. Убедитесь, что порт не занят:
```bash
lsof -i :8080
```

### Endpoints не отображаются

1. Проверьте наличие Swagger аннотаций в handler'ах
2. Регенерируйте Swagger:
```bash
cd <service> && make swagger
```
3. Перезапустите сервис

### Изменения не применяются

1. Регенерируйте Swagger документацию
2. Перезапустите сервис
3. Очистите кэш браузера (Ctrl+Shift+R)

## Дополнительная документация

- [API Reference](./API_REFERENCE.md) - Подробная документация API
- [Migration Service Swagger Guide](./migration-service/SWAGGER_GUIDE.md) - Детальное руководство
- [README](./README.md) - Общая документация проекта
- [Deployment Guide](./DEPLOYMENT_GUIDE.md) - Руководство по развертыванию

## Контакты

Для вопросов и предложений по API документации создайте issue в репозитории.
