# Цифровой Вуз - Микросервисная Архитектура

Микросервисная система управления университетами, сотрудниками, чатами и структурными подразделениями для мини-приложения "Цифровой вуз" на базе Go и MAX Messenger.

## Описание

Проект представляет собой полнофункциональную систему микросервисов с поддержкой:
- **Ролевой модели ABAC** (Attribute-Based Access Control) с иерархией прав
- **Автоматической интеграции с MAX Messenger** через получение MAX_id по номеру телефона
- **Миграции 150,000+ чатов** из трех различных источников данных
- **Иерархической структуры вузов** с поддержкой связей между подразделениями, сотрудниками и чатами
- **Чистой архитектуры** с разделением на domain, usecase, infrastructure слои

## Ключевые возможности

✅ **Ролевая модель ABAC** - Superadmin, Curator, Operator с контекстными правами доступа  
✅ **Автоматическое получение MAX_id** - Интеграция с MAX Messenger Bot API  
✅ **Пакетные операции** - Обновление MAX_id для тысяч сотрудников  
✅ **Миграция данных** - Импорт из базы данных, Google Sheets и Excel  
✅ **Иерархическая структура** - University → Branch → Faculty → Group → Chat  
✅ **Фильтрация по ролям** - Автоматическое ограничение доступа к данным  
✅ **Поиск и пагинация** - Полнотекстовый поиск с поддержкой русского языка  
✅ **gRPC коммуникация** - Высокопроизводительное взаимодействие между сервисами  
✅ **Retry логика** - Автоматические повторы при сбоях с экспоненциальной задержкой  
✅ **Структурированное логирование** - JSON логи с request_id для трассировки  

## Архитектура

Проект состоит из 6 микросервисов:

### 1. Auth Service (порт 8080, gRPC: 9090)
**Сервис аутентификации и авторизации с поддержкой ABAC**

Основные возможности:
- Управление пользователями и JWT токенами (access и refresh)
- **Ролевая модель**: Superadmin, Curator, Operator
- **ABAC валидация**: Проверка прав доступа на основе атрибутов (роль, университет, филиал, факультет)
- **gRPC API**: ValidateToken, GetUserPermissions, AssignRole
- Назначение ролей с контекстом (university_id, branch_id, faculty_id)
- Иерархия прав: Superadmin > Curator > Operator

**Документация:**
- [Валидация прав доступа](./auth-service/internal/usecase/validate_permission.go)
- [Тесты ABAC](./auth-service/internal/usecase/validate_permission_test.go)

### 2. Employee Service (порт 8081, gRPC: 9091)
**Сервис управления сотрудниками с интеграцией ролей и MAX_id**

Основные возможности:
- CRUD операции для сотрудников вузов
- **Автоматическое получение MAX_id** при создании сотрудника
- **Пакетное обновление MAX_id** для существующих сотрудников (до 100 за раз)
- **Синхронизация ролей** с Auth Service при создании/обновлении/удалении
- **Поиск с фильтрацией по ролям** (Curator видит только свой вуз)
- Автоматическое создание университетов по INN
- Управление университетами (name, INN, KPP)

**Документация:**
- [Создание сотрудника с ролью](./employee-service/internal/usecase/create_employee_with_role.go)
- [Пакетное обновление MAX_id](./employee-service/internal/usecase/batch_update_max_id.go)
- [Примеры интеграции MaxBot](./employee-service/MAXBOT_INTEGRATION_EXAMPLES.md)
- [Реализация пакетных операций](./employee-service/BATCH_UPDATE_IMPLEMENTATION.md)

### 3. Chat Service (порт 8082, gRPC: 9092)
**Сервис управления чатами с ролевой фильтрацией**

Основные возможности:
- Создание и управление групповыми чатами
- **Фильтрация по ролям**: Superadmin (все), Curator (свой вуз), Operator (свой филиал/факультет)
- **Управление администраторами** с проверкой прав и получением MAX_id
- **Полнотекстовый поиск** по названию чата (русский язык)
- **Пагинация** с лимитом до 100 записей
- Защита от удаления последнего администратора
- Поддержка источников: admin_panel, bot_registrar, academic_group

**Документация:**
- [Фильтрация по ролям](./chat-service/internal/usecase/list_chats_with_role_filter.go)
- [Реализация фильтрации](./chat-service/ROLE_BASED_FILTERING_IMPLEMENTATION.md)
- [Реализация пагинации](./chat-service/PAGINATION_IMPLEMENTATION.md)
- [Реализация поиска](./chat-service/CHAT_SEARCH_IMPLEMENTATION.md)

### 4. Structure Service (порт 8083, gRPC: 9093)
**Сервис управления иерархической структурой вузов**

Основные возможности:
- **Иерархическая структура**: University → Branch → Faculty → Group → Chat
- **Импорт из Excel** с полной структурой и чатами
- **Связь групп с чатами** через chat_id
- **Управление операторами подразделений** (department_managers)
- Получение полной структуры с деталями чатов через gRPC
- Алфавитная сортировка на каждом уровне иерархии
- Транзакционный импорт с обработкой дубликатов

**Документация:**
- [Импорт из Excel](./structure-service/internal/usecase/import_structure_from_excel.go)
- [Получение иерархии](./structure-service/internal/usecase/get_university_structure.go)
- [Реализация иерархии](./structure-service/STRUCTURE_HIERARCHY_IMPLEMENTATION.md)
- [Реализация импорта Excel](./structure-service/EXCEL_IMPORT_IMPLEMENTATION.md)
- [Управление операторами](./structure-service/DEPARTMENT_MANAGERS_IMPLEMENTATION.md)

### 5. MaxBot Service (gRPC: 9095)
**Сервис интеграции с MAX Messenger Bot API**

Основные возможности:
- Поиск пользователей по номеру телефона (GetUserByPhone)
- **Пакетная обработка** до 100 номеров за запрос (BatchGetUsersByPhone)
- **Нормализация телефонов** в формат E.164 (+7XXXXXXXXXX)
- Поддержка российских форматов (8, 9, +7)
- Удаление нечисловых символов (пробелы, дефисы, скобки)
- Валидация формата телефона
- Интеграция с официальной библиотекой max-bot-api-client-go

**Документация:**
- [README MaxBot Service](./maxbot-service/README.md)
- [Руководство по интеграции](./maxbot-service/INTEGRATION_GUIDE.md)
- [Реализация пакетных операций](./maxbot-service/BATCH_OPERATIONS_IMPLEMENTATION.md)
- [Нормализация телефонов](./maxbot-service/internal/usecase/normalize_phone.go)

### 6. Migration Service (порт 8084)
**Сервис миграции данных из различных источников**

Основные возможности:
- **Миграция из базы данных** (6,000 чатов из админки)
- **Миграция из Google Sheets** (чаты от бота-регистратора)
- **Миграция из Excel** (155,000+ чатов академических групп)
- Отслеживание прогресса миграции (migration_jobs)
- Логирование ошибок (migration_errors)
- Создание/переиспользование университетов по INN+KPP
- Автоматическое создание структуры и связывание с чатами

**Документация:**
- [README Migration Service](./migration-service/README.md)
- [Реализация сервиса миграции](./migration-service/MIGRATION_SERVICE_IMPLEMENTATION.md)
- [Миграция из базы данных](./migration-service/internal/usecase/migrate_from_database.go)
- [Миграция из Google Sheets](./migration-service/internal/usecase/migrate_from_google_sheets.go)
- [Миграция из Excel](./migration-service/internal/usecase/migrate_from_excel.go)

## Технологический стек

- **Язык**: Go 1.21+
- **База данных**: PostgreSQL 15
- **API**: REST (HTTP) + gRPC
- **Аутентификация**: JWT (RS256)
- **Контейнеризация**: Docker, Docker Compose
- **Документация API**: Swagger/OpenAPI
- **Миграции БД**: golang-migrate
- **Логирование**: Structured JSON logs
- **Тестирование**: Go testing, testify, gopter (property-based testing)
- **Внешние API**: MAX Messenger Bot API

## Быстрый старт

### Требования

- Go 1.21+
- Docker и Docker Compose
- protoc (для генерации gRPC кода)
- PostgreSQL 15 (или используйте Docker Compose)

### Установка зависимостей

```bash
# Установка protoc (macOS)
brew install protobuf

# Установка Go плагинов для gRPC
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

### Генерация gRPC кода

```bash
# Из корневой директории проекта
chmod +x generate_proto.sh
./generate_proto.sh
```

### Запуск через Docker Compose

```bash
# Запуск всех сервисов и баз данных
docker-compose up -d

# Просмотр логов
docker-compose logs -f

# Просмотр логов конкретного сервиса
docker-compose logs -f auth-service

# Остановка всех сервисов
docker-compose down

# Остановка с удалением volumes (очистка БД)
docker-compose down -v
```

Все сервисы будут доступны по следующим адресам:
- **Auth Service**: http://localhost:8080 (gRPC: 9090)
- **Employee Service**: http://localhost:8081 (gRPC: 9091)
- **Chat Service**: http://localhost:8082 (gRPC: 9092)
- **Structure Service**: http://localhost:8083 (gRPC: 9093)
- **Migration Service**: http://localhost:8084
- **MaxBot Service**: gRPC: 9095

### Проверка работоспособности

```bash
# Проверка health endpoints
curl http://localhost:8080/health  # Auth Service
curl http://localhost:8081/health  # Employee Service
curl http://localhost:8082/health  # Chat Service
curl http://localhost:8083/health  # Structure Service
curl http://localhost:8084/health  # Migration Service
```

### Запуск локально (без Docker)

Для каждого сервиса:

```bash
# Перейти в директорию сервиса
cd auth-service  # или chat-service, employee-service, structure-service

# Установить зависимости
go mod download

# Запустить миграции (если необходимо)
# Убедитесь, что PostgreSQL запущен и настроен

# Запустить сервис
go run cmd/auth/main.go  # или соответствующий путь для других сервисов
```

## Структура проекта

```
go-lang-max/
├── auth-service/              # Сервис аутентификации и ABAC
│   ├── api/proto/             # gRPC proto файлы
│   ├── cmd/auth/              # Точка входа
│   ├── internal/
│   │   ├── domain/            # Доменные модели (User, Role, Permission)
│   │   ├── usecase/           # Бизнес-логика (ValidatePermission, AssignRole)
│   │   ├── infrastructure/
│   │   │   ├── grpc/          # gRPC сервер и handlers
│   │   │   ├── http/          # HTTP REST API
│   │   │   ├── repository/    # PostgreSQL репозитории
│   │   │   ├── jwt/           # JWT token manager
│   │   │   └── errors/        # Обработка ошибок
│   │   └── config/            # Конфигурация
│   ├── migrations/            # SQL миграции (roles, user_roles)
│   └── Dockerfile
│
├── employee-service/          # Сервис управления сотрудниками
│   ├── internal/
│   │   ├── domain/            # Employee, University, BatchUpdateJob
│   │   ├── usecase/           # CreateEmployeeWithRole, BatchUpdateMaxId
│   │   └── infrastructure/
│   │       ├── auth/          # Auth Service gRPC client
│   │       └── max/           # MaxBot Service gRPC client
│   └── migrations/            # SQL миграции (role, max_id, batch_update_jobs)
│
├── chat-service/              # Сервис управления чатами
│   ├── internal/
│   │   ├── domain/            # Chat, Administrator, ChatFilter
│   │   ├── usecase/           # ListChatsWithRoleFilter, AddAdministrator
│   │   └── infrastructure/
│   │       ├── auth/          # Auth Service gRPC client
│   │       └── max/           # MaxBot Service gRPC client
│   └── migrations/            # SQL миграции
│
├── structure-service/         # Сервис структуры вузов
│   ├── internal/
│   │   ├── domain/            # Structure, DepartmentManager
│   │   ├── usecase/           # ImportFromExcel, GetUniversityStructure
│   │   └── infrastructure/
│   │       ├── excel/         # Excel parser
│   │       └── grpc/          # Chat Service gRPC client
│   └── migrations/            # SQL миграции (department_managers, chat_url)
│
├── maxbot-service/            # Сервис интеграции с MAX Messenger
│   ├── internal/
│   │   ├── domain/            # MAX API client interface
│   │   ├── usecase/           # BatchGetUsersByPhone, NormalizePhone
│   │   └── infrastructure/
│   │       ├── maxapi/        # MAX API HTTP client
│   │       └── grpc/          # gRPC сервер
│   └── api/proto/             # gRPC proto файлы
│
├── migration-service/         # Сервис миграции данных
│   ├── internal/
│   │   ├── domain/            # MigrationJob, MigrationError
│   │   ├── usecase/           # MigrateFromDatabase, MigrateFromExcel
│   │   └── infrastructure/
│   │       ├── chat/          # Chat Service gRPC client
│   │       └── structure/     # Structure Service gRPC client
│   └── migrations/            # SQL миграции (migration_jobs, migration_errors)
│
├── integration-tests/         # End-to-end интеграционные тесты
│   ├── employee_integration_test.go
│   ├── chat_integration_test.go
│   ├── structure_integration_test.go
│   ├── migration_integration_test.go
│   ├── grpc_integration_test.go
│   └── run_tests.sh
│
├── .kiro/specs/               # Спецификации и документация
│   └── digital-university-mvp-completion/
│       ├── requirements.md    # Требования (EARS формат)
│       ├── design.md          # Дизайн и архитектура
│       └── tasks.md           # План реализации
│
├── docker-compose.yml         # Оркестрация всех сервисов
├── generate_proto.sh          # Скрипт генерации gRPC кода
├── GRPC_SETUP.md              # Документация по настройке gRPC
├── MIGRATIONS.md              # Руководство по миграциям БД
└── README.md                  # Этот файл
```

## gRPC взаимодействие

Сервисы взаимодействуют через gRPC для внутренней коммуникации с автоматическими повторами при сбоях:

### Схема взаимодействия

```
Employee Service ──gRPC──> Auth Service (ValidateToken, AssignRole)
                 └──gRPC──> MaxBot Service (GetUserByPhone, BatchGetUsersByPhone)

Chat Service ──gRPC──> Auth Service (ValidateToken, GetUserPermissions)
             └──gRPC──> MaxBot Service (GetUserByPhone)

Structure Service ──gRPC──> Chat Service (GetChat)
                  └──gRPC──> Employee Service (GetEmployee)

Migration Service ──gRPC──> Chat Service (CreateChat)
                  └──gRPC──> Structure Service (ImportStructure)
```

### Основные gRPC методы

**Auth Service (порт 9090):**
- `ValidateToken(token)` → user_id, role, university_id, branch_id, faculty_id
- `GetUserPermissions(user_id)` → список прав доступа
- `AssignRole(user_id, role, context)` → успех/ошибка

**MaxBot Service (порт 9095):**
- `GetUserByPhone(phone)` → MAX_id или ошибка
- `BatchGetUsersByPhone(phones[])` → массив phone→MAX_id mappings
- `NormalizePhone(phone)` → нормализованный номер в E.164

**Chat Service (порт 9092):**
- `GetChat(chat_id)` → детали чата
- `CreateChat(chat_data)` → созданный чат

**Employee Service (порт 9091):**
- `GetEmployee(employee_id)` → детали сотрудника

### Retry логика

Все gRPC вызовы используют автоматические повторы:
- **Количество попыток**: 3
- **Задержки**: 1s, 2s, 4s (экспоненциальная)
- **Логирование**: каждая попытка логируется
- **Ошибка**: возвращается после финальной неудачи

Подробная информация: [GRPC_RETRY_IMPLEMENTATION.md](./GRPC_RETRY_IMPLEMENTATION.md)

**Документация:**
- [Настройка gRPC](./GRPC_SETUP.md)
- [Реализация retry](./GRPC_RETRY_IMPLEMENTATION.md)
- [Интеграционные тесты gRPC](./integration-tests/grpc_integration_test.go)

## Базы данных

Каждый сервис использует свою собственную базу данных PostgreSQL (Database per Service pattern):

| Сервис | База данных | Порт | Основные таблицы |
|--------|-------------|------|------------------|
| Auth Service | auth-db | 5432 | users, roles, user_roles, refresh_tokens |
| Employee Service | employee-db | 5433 | employees, universities, batch_update_jobs |
| Chat Service | chat-db | 5434 | chats, administrators |
| Structure Service | structure-db | 5435 | universities, branches, faculties, groups, department_managers |
| Migration Service | migration-db | 5436 | migration_jobs, migration_errors |

### Миграции

Миграции автоматически применяются при первом запуске через Docker Compose.

**Ручное управление миграциями:**

```bash
# Применить миграции
cd auth-service
make migrate-up

# Откатить последнюю миграцию
make migrate-down

# Проверить статус миграций
make migrate-status
```

**Документация:**
- [Руководство по миграциям](./MIGRATIONS.md)
- [Скрипт проверки миграций](./verify_migrations.sh)
- [Скрипт тестирования миграций](./test_migrations.sh)

## API Документация

### Swagger UI

После запуска сервисов, Swagger документация доступна по адресам:

- **Auth Service**: http://localhost:8080/swagger/index.html
- **Employee Service**: http://localhost:8081/swagger/index.html
- **Chat Service**: http://localhost:8082/swagger/index.html
- **Structure Service**: http://localhost:8083/swagger/index.html
- **Migration Service**: http://localhost:8084/swagger/index.html

📖 **Полная документация Swagger:** [SWAGGER_ENDPOINTS.md](./SWAGGER_ENDPOINTS.md)

### Основные API endpoints

#### Auth Service (порт 8080)

```
POST   /auth/register          - Регистрация пользователя
POST   /auth/login             - Вход (получение JWT токенов)
POST   /auth/refresh           - Обновление access токена
POST   /auth/logout            - Выход (инвалидация refresh токена)
POST   /roles/assign           - Назначение роли пользователю
GET    /users/{id}/permissions - Получение прав доступа пользователя
GET    /health                 - Health check
```

#### Employee Service (порт 8081)

```
POST   /employees                    - Создание сотрудника (с ролью и MAX_id)
GET    /employees                    - Поиск сотрудников (с фильтрацией по ролям)
GET    /employees/{id}               - Получение сотрудника
PUT    /employees/{id}               - Обновление сотрудника (синхронизация роли)
DELETE /employees/{id}               - Удаление сотрудника (отзыв прав)
POST   /employees/batch-update-maxid - Пакетное обновление MAX_id
GET    /employees/batch-status       - Статус пакетного обновления
GET    /universities                 - Список университетов
GET    /health                       - Health check
```

#### Chat Service (порт 8082)

```
POST   /chats                  - Создание чата
GET    /chats                  - Список чатов (с фильтрацией по ролям и пагинацией)
GET    /chats/search           - Поиск чатов по названию
GET    /chats/{id}             - Получение чата
PUT    /chats/{id}             - Обновление чата
DELETE /chats/{id}             - Удаление чата
POST   /chats/{id}/admins      - Добавление администратора (с проверкой прав)
DELETE /chats/{id}/admins/{aid} - Удаление администратора (защита последнего)
GET    /health                 - Health check
```

#### Structure Service (порт 8083)

```
POST   /import/excel                  - Импорт структуры из Excel
GET    /universities/{id}/structure   - Получение полной иерархии вуза
POST   /departments/managers          - Назначение оператора на подразделение
DELETE /departments/managers/{id}     - Удаление назначения оператора
GET    /departments/managers          - Список операторов подразделений
GET    /universities                  - Список университетов
GET    /branches                      - Список филиалов
GET    /faculties                     - Список факультетов
GET    /groups                        - Список групп
GET    /health                        - Health check
```

#### Migration Service (порт 8084)

```
POST   /migration/database      - Миграция из базы данных (6,000 чатов)
POST   /migration/google-sheets - Миграция из Google Sheets
POST   /migration/excel         - Миграция из Excel (155,000+ чатов)
GET    /migration/jobs/{id}     - Статус миграции
GET    /migration/jobs          - Список всех миграций
GET    /health                  - Health check
```

### Примеры использования

**Создание сотрудника с автоматическим получением MAX_id:**

```bash
curl -X POST http://localhost:8081/employees \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <JWT_TOKEN>" \
  -d '{
    "first_name": "Иван",
    "last_name": "Иванов",
    "phone": "+79991234567",
    "role": "operator",
    "university": {
      "name": "МГУ",
      "inn": "7714357576",
      "kpp": "771401001"
    }
  }'
```

**Получение чатов с фильтрацией по роли:**

```bash
# Curator увидит только чаты своего вуза
curl -X GET "http://localhost:8082/chats?limit=50&offset=0" \
  -H "Authorization: Bearer <CURATOR_JWT_TOKEN>"
```

**Поиск чатов:**

```bash
curl -X GET "http://localhost:8082/chats/search?q=математика" \
  -H "Authorization: Bearer <JWT_TOKEN>"
```

**Импорт структуры из Excel:**

```bash
curl -X POST http://localhost:8083/import/excel \
  -H "Authorization: Bearer <JWT_TOKEN>" \
  -F "file=@structure.xlsx"
```

## Переменные окружения

### Auth Service
```bash
DATABASE_URL=postgres://user:password@localhost:5432/auth_db?sslmode=disable
HTTP_ADDR=:8080                    # Адрес HTTP сервера
GRPC_PORT=9090                     # Порт gRPC сервера
JWT_ACCESS_SECRET=your-secret-key  # Секретный ключ для access токенов
JWT_REFRESH_SECRET=your-secret-key # Секретный ключ для refresh токенов
ACCESS_MINUTES=15                  # Время жизни access токена (минуты)
REFRESH_HOURS=168                  # Время жизни refresh токена (часы)
LOG_LEVEL=info                     # Уровень логирования (debug, info, warn, error)
```

### Employee Service
```bash
DATABASE_URL=postgres://user:password@localhost:5433/employee_db?sslmode=disable
PORT=8081                          # Порт HTTP сервера
GRPC_PORT=9091                     # Порт gRPC сервера
AUTH_SERVICE_GRPC=localhost:9090   # Адрес Auth Service gRPC
MAXBOT_SERVICE_GRPC=localhost:9095 # Адрес MaxBot Service gRPC
LOG_LEVEL=info
```

### Chat Service
```bash
DATABASE_URL=postgres://user:password@localhost:5434/chat_db?sslmode=disable
PORT=8082                          # Порт HTTP сервера
GRPC_PORT=9092                     # Порт gRPC сервера
AUTH_SERVICE_GRPC=localhost:9090   # Адрес Auth Service gRPC
MAXBOT_SERVICE_GRPC=localhost:9095 # Адрес MaxBot Service gRPC
LOG_LEVEL=info
```

### Structure Service
```bash
DATABASE_URL=postgres://user:password@localhost:5435/structure_db?sslmode=disable
PORT=8083                          # Порт HTTP сервера
GRPC_PORT=9093                     # Порт gRPC сервера
CHAT_SERVICE_GRPC=localhost:9092   # Адрес Chat Service gRPC
EMPLOYEE_SERVICE_GRPC=localhost:9091 # Адрес Employee Service gRPC
LOG_LEVEL=info
```

### MaxBot Service
```bash
MAX_API_TOKEN=your-max-api-token   # Токен для MAX Messenger Bot API (обязательно!)
MAX_API_URL=https://api.max.ru     # URL MAX Messenger API
MAX_API_TIMEOUT=5s                 # Таймаут для API запросов
GRPC_PORT=9095                     # Порт gRPC сервера
LOG_LEVEL=info
```

### Migration Service
```bash
DATABASE_URL=postgres://user:password@localhost:5436/migration_db?sslmode=disable
PORT=8084                          # Порт HTTP сервера
CHAT_SERVICE_GRPC=localhost:9092   # Адрес Chat Service gRPC
STRUCTURE_SERVICE_GRPC=localhost:9093 # Адрес Structure Service gRPC
GOOGLE_SHEETS_CREDENTIALS_PATH=/path/to/credentials.json # Путь к credentials для Google Sheets API
LOG_LEVEL=info
```

**Примечание:** Все переменные окружения настроены в `docker-compose.yml` для запуска через Docker.

## Ролевая модель и управление доступом (ABAC)

Система использует Attribute-Based Access Control (ABAC) с тремя ролями и контекстными атрибутами.

### Роли

| Роль | Описание | Права доступа |
|------|----------|---------------|
| **Superadmin** | Представитель VK | Полный доступ ко всем данным всех университетов |
| **Curator** | Представитель вуза | Доступ только к данным своего университета |
| **Operator** | Представитель подразделения | Доступ только к данным своего филиала/факультета |

### Иерархия прав

```
Superadmin (все университеты)
    ↓
Curator (один университет)
    ↓
Operator (один филиал или факультет)
```

Высшие роли имеют все права низших ролей (кумулятивные права).

### Контекстные атрибуты

Каждая роль (кроме Superadmin) привязана к контексту:

```go
type UserRole struct {
    UserID       int
    Role         string  // "superadmin", "curator", "operator"
    UniversityID *int    // NULL для superadmin
    BranchID     *int    // NULL для curator и superadmin
    FacultyID    *int    // NULL для curator и superadmin
}
```

### Примеры фильтрации

**Список чатов:**
- Superadmin: `SELECT * FROM chats`
- Curator: `SELECT * FROM chats WHERE university_id = ?`
- Operator: `SELECT * FROM chats WHERE branch_id = ? OR faculty_id = ?`

**Поиск сотрудников:**
- Superadmin: все сотрудники
- Curator: только сотрудники своего университета
- Operator: только сотрудники своего подразделения

### Назначение ролей

```bash
# Назначение роли Curator для университета
curl -X POST http://localhost:8080/roles/assign \
  -H "Authorization: Bearer <SUPERADMIN_TOKEN>" \
  -d '{
    "user_id": 123,
    "role": "curator",
    "university_id": 1
  }'

# Назначение роли Operator для филиала
curl -X POST http://localhost:8080/roles/assign \
  -H "Authorization: Bearer <CURATOR_TOKEN>" \
  -d '{
    "user_id": 456,
    "role": "operator",
    "branch_id": 10
  }'
```

**Документация:**
- [Валидация прав доступа](./auth-service/internal/usecase/validate_permission.go)
- [Тесты ABAC](./auth-service/internal/usecase/validate_permission_test.go)
- [Фильтрация чатов по ролям](./chat-service/ROLE_BASED_FILTERING_IMPLEMENTATION.md)

## Процесс миграции данных

Система поддерживает миграцию более 150,000 чатов из трех различных источников.

### Источники данных

| Источник | Количество | Тип | Source метка |
|----------|------------|-----|--------------|
| База данных админки | ~6,000 | Чаты из существующей админки | `admin_panel` |
| Google Sheets | Переменное | Чаты от бота-регистратора | `bot_registrar` |
| Excel файлы | 155,000+ | Чаты академических групп | `academic_group` |

### Этапы миграции

#### 1. Миграция из базы данных (admin_panel)

```bash
# Запуск миграции
curl -X POST http://localhost:8084/migration/database \
  -H "Authorization: Bearer <TOKEN>" \
  -d '{
    "source_db_url": "postgres://user:pass@host:5432/old_db"
  }'

# Проверка статуса
curl http://localhost:8084/migration/jobs/1
```

**Что происходит:**
1. Чтение данных: INN, название чата, URL, телефон администратора
2. Поиск/создание университета по INN
3. Создание чата с source='admin_panel'
4. Создание записи администратора
5. Получение MAX_id для администратора

**Время выполнения:** ~10 минут

#### 2. Миграция из Google Sheets (bot_registrar)

```bash
curl -X POST http://localhost:8084/migration/google-sheets \
  -H "Authorization: Bearer <TOKEN>" \
  -d '{
    "sheet_id": "1abc...xyz",
    "credentials_path": "/path/to/credentials.json"
  }'
```

**Что происходит:**
1. Аутентификация с Google Sheets API
2. Чтение колонок: INN, KPP, URL, телефон администратора
3. Поиск/создание университета по INN+KPP
4. Создание чата с source='bot_registrar'
5. Логирование обработанных строк и ошибок

#### 3. Миграция из Excel (academic_group)

```bash
curl -X POST http://localhost:8084/migration/excel \
  -H "Authorization: Bearer <TOKEN>" \
  -F "file=@academic_groups.xlsx"
```

**Что происходит:**
1. Валидация формата Excel и обязательных колонок
2. Парсинг: телефон, INN, FOIV, название организации, филиал, KPP, факультет, курс, номер группы, название чата, URL
3. Вызов Structure Service для создания иерархии (University → Branch → Faculty → Group)
4. Вызов Chat Service для создания чата с source='academic_group'
5. Связывание Group с Chat через chat_id

**Время выполнения:** ~2-3 часа для 155,000 записей

### Мониторинг миграции

```bash
# Список всех миграций
curl http://localhost:8084/migration/jobs

# Детали конкретной миграции
curl http://localhost:8084/migration/jobs/1

# Пример ответа:
{
  "id": 1,
  "source_type": "excel",
  "status": "running",
  "total": 155000,
  "processed": 50000,
  "failed": 123,
  "started_at": "2024-01-15T10:00:00Z",
  "estimated_completion": "2024-01-15T13:00:00Z"
}
```

### Обработка ошибок

Все ошибки миграции логируются в таблицу `migration_errors`:

```bash
# Получение ошибок миграции
curl http://localhost:8084/migration/jobs/1/errors

# Пример ответа:
[
  {
    "record_identifier": "row_1234",
    "error_message": "Invalid INN format: 123",
    "created_at": "2024-01-15T10:05:00Z"
  }
]
```

**Документация:**
- [Руководство по миграции](./QUICK_MIGRATION_GUIDE.md)
- [Реализация Migration Service](./migration-service/MIGRATION_SERVICE_IMPLEMENTATION.md)
- [Интеграционные тесты миграции](./integration-tests/migration_integration_test.go)

## Разработка

### Добавление нового сервиса

1. Создайте директорию для сервиса
2. Инициализируйте Go модуль: `go mod init <service-name>`
3. Добавьте структуру проекта (cmd, internal, migrations)
4. Добавьте Dockerfile и docker-compose.yml для сервиса
5. Обновите корневой docker-compose.yml
6. Добавьте proto файлы в `api/proto/`
7. Обновите `generate_proto.sh` для генерации кода

### Генерация Swagger документации

**Обновление всех сервисов одной командой:**

```bash
./update_swagger.sh
```

**Обновление конкретного сервиса:**

```bash
cd <service-name>
make swagger

# Или напрямую
swag init -g cmd/<service>/main.go -o internal/infrastructure/http/docs
```

**Примеры:**

```bash
# Auth Service
cd auth-service && make swagger

# Migration Service  
cd migration-service && make swagger
```

## Развертывание (Deployment)

### Требования для продакшена

- Docker 20.10+
- Docker Compose 2.0+
- PostgreSQL 15
- Минимум 4GB RAM
- 50GB дискового пространства
- SSL сертификаты для HTTPS

### Подготовка к развертыванию

#### 1. Настройка переменных окружения

Создайте `.env` файл для каждого сервиса:

```bash
# Пример для production
cp auth-service/.env.example auth-service/.env
# Отредактируйте .env файлы с production значениями
```

**Критически важные настройки:**
- Используйте сильные JWT секреты (минимум 32 символа)
- Настройте production DATABASE_URL с SSL
- Укажите реальный MAX_API_TOKEN
- Установите LOG_LEVEL=info (не debug)

#### 2. Настройка базы данных

```bash
# Создайте базы данных
psql -U postgres -c "CREATE DATABASE auth_db;"
psql -U postgres -c "CREATE DATABASE employee_db;"
psql -U postgres -c "CREATE DATABASE chat_db;"
psql -U postgres -c "CREATE DATABASE structure_db;"
psql -U postgres -c "CREATE DATABASE migration_db;"

# Примените миграции
./verify_migrations.sh
```

#### 3. Настройка SSL/TLS

Для production обязательно используйте HTTPS и TLS для gRPC:

```yaml
# docker-compose.prod.yml
services:
  auth-service:
    environment:
      - TLS_CERT_FILE=/certs/server.crt
      - TLS_KEY_FILE=/certs/server.key
    volumes:
      - ./certs:/certs:ro
```

### Стратегия развертывания

#### Blue-Green Deployment

1. **Развертывание новой версии (Green)**
   ```bash
   docker-compose -f docker-compose.prod.yml up -d --scale auth-service=2
   ```

2. **Health checks**
   ```bash
   # Проверка всех сервисов
   for port in 8080 8081 8082 8083 8084; do
     curl -f http://localhost:$port/health || exit 1
   done
   ```

3. **Переключение трафика**
   ```bash
   # Обновите load balancer или API gateway
   # для направления трафика на новые инстансы
   ```

4. **Откат при необходимости**
   ```bash
   docker-compose -f docker-compose.prod.yml down
   docker-compose -f docker-compose.prod.yml.backup up -d
   ```

### Миграция базы данных в продакшене

```bash
# 1. Создайте backup
pg_dump -U postgres auth_db > auth_db_backup.sql

# 2. Примените миграции
cd auth-service
make migrate-up

# 3. Проверьте статус
make migrate-status

# 4. При необходимости откатите
make migrate-down
```

### Мониторинг и логирование

#### Структурированные логи

Все сервисы используют JSON логирование:

```json
{
  "level": "info",
  "timestamp": "2024-01-15T10:00:00Z",
  "request_id": "abc-123",
  "service": "auth-service",
  "message": "User authenticated",
  "user_id": 123,
  "duration_ms": 45
}
```

#### Метрики для мониторинга

Рекомендуется отслеживать:
- Request rate (запросов в секунду)
- Response time (p50, p95, p99)
- Error rate (% ошибок)
- gRPC call success/failure rate
- Database connection pool usage
- Migration progress

#### Health checks

Все сервисы предоставляют `/health` endpoint:

```bash
# Пример health check response
{
  "status": "healthy",
  "database": "connected",
  "grpc_services": {
    "auth_service": "available",
    "maxbot_service": "available"
  },
  "uptime_seconds": 3600
}
```

### Масштабирование

#### Горизонтальное масштабирование

```bash
# Увеличение количества инстансов
docker-compose -f docker-compose.prod.yml up -d --scale employee-service=3
docker-compose -f docker-compose.prod.yml up -d --scale chat-service=3
```

#### Рекомендации по масштабированию

- **Auth Service**: 2-3 инстанса (stateless, легко масштабируется)
- **Employee Service**: 2-4 инстанса (зависит от нагрузки)
- **Chat Service**: 3-5 инстансов (высокая нагрузка на чтение)
- **Structure Service**: 1-2 инстанса (низкая нагрузка)
- **MaxBot Service**: 2-3 инстанса (внешние API вызовы)
- **Migration Service**: 1 инстанс (фоновые задачи)

#### Database Connection Pooling

```go
// Рекомендуемые настройки для production
db.SetMaxOpenConns(25)
db.SetMaxIdleConns(5)
db.SetConnMaxLifetime(5 * time.Minute)
```

### Безопасность

#### Checklist для продакшена

- [ ] Используются сильные JWT секреты
- [ ] Включен SSL/TLS для всех соединений
- [ ] Настроен rate limiting
- [ ] Включена валидация всех входных данных
- [ ] Используются parameterized SQL queries
- [ ] Настроены firewall правила
- [ ] Ограничен доступ к базам данных
- [ ] Включено логирование всех операций
- [ ] Настроены alerts для критических ошибок
- [ ] Регулярные backup баз данных

#### Rate Limiting

Рекомендуемые лимиты:
- 100 запросов/минуту на пользователя
- 1000 запросов/минуту на IP
- 10 запросов/час для migration endpoints

### Backup и восстановление

```bash
# Ежедневный backup всех баз данных
#!/bin/bash
DATE=$(date +%Y%m%d)
for db in auth_db employee_db chat_db structure_db migration_db; do
  pg_dump -U postgres $db | gzip > /backups/${db}_${DATE}.sql.gz
done

# Восстановление из backup
gunzip < /backups/auth_db_20240115.sql.gz | psql -U postgres auth_db
```

**Документация:**
- [Руководство по миграциям](./MIGRATIONS.md)
- [Мониторинг и логирование](./MONITORING_AND_LOGGING_IMPLEMENTATION.md)
- [Обработка ошибок](./ERROR_HANDLING_IMPLEMENTATION.md)

## Тестирование

### Unit Tests

```bash
# Запуск тестов для конкретного сервиса
cd auth-service
go test ./...

# Запуск всех unit тестов
go test ./...
```

### Integration Tests

Проект включает комплексные end-to-end интеграционные тесты, которые проверяют взаимодействие всех сервисов.

```bash
# Запуск всех интеграционных тестов
cd integration-tests
./run_tests.sh

# Или используя Make
make test

# Запуск конкретных тестовых наборов
make test-employee    # Тесты Employee Service
make test-chat        # Тесты Chat Service
make test-structure   # Тесты Structure Service
make test-migration   # Тесты Migration Service
make test-grpc        # Тесты gRPC коммуникации
```

**Покрытие интеграционных тестов:**
- ✅ Создание сотрудников с ролями и MAX_id (Employee → Auth → MaxBot)
- ✅ Фильтрация чатов по ролям (Superadmin, Curator, Operator)
- ✅ Импорт структуры из Excel с полной иерархией
- ✅ Миграция из базы данных, Google Sheets и Excel
- ✅ gRPC коммуникация между всеми сервисами
- ✅ Пагинация и поиск
- ✅ Управление администраторами чатов
- ✅ Пакетные операции

**Документация:**
- Быстрый старт: [integration-tests/QUICK_START.md](./integration-tests/QUICK_START.md)
- Полное руководство: [integration-tests/INTEGRATION_TEST_GUIDE.md](./integration-tests/INTEGRATION_TEST_GUIDE.md)
- Требования: [.kiro/specs/digital-university-mvp-completion/requirements.md](./.kiro/specs/digital-university-mvp-completion/requirements.md)

### Property-Based Testing

Система использует property-based testing (gopter) для проверки универсальных свойств:

```bash
# Запуск property-based тестов
cd employee-service
go test -v ./internal/usecase -run TestProperty

cd maxbot-service
go test -v ./internal/usecase -run TestPhoneNormalization
```

**Примеры свойств:**
- Нормализация телефонов всегда возвращает E.164 формат
- ABAC валидация соблюдает иерархию ролей
- Пагинация никогда не превышает лимит 100
- Последний администратор не может быть удален

**Конфигурация:**
- Минимум 100 итераций на свойство
- Генераторы для реалистичных тестовых данных
- Автоматическое shrinking при нахождении контрпримера

## Устранение неполадок (Troubleshooting)

### Проблемы с запуском

**Проблема: Сервис не запускается**

```bash
# Проверьте логи
docker-compose logs auth-service

# Проверьте подключение к БД
docker-compose exec auth-db psql -U postgres -d auth_db -c "SELECT 1;"

# Проверьте порты
netstat -an | grep LISTEN | grep -E "8080|8081|8082|8083|8084"
```

**Проблема: Миграции не применяются**

```bash
# Проверьте статус миграций
cd auth-service
make migrate-status

# Принудительно примените миграции
make migrate-force VERSION=1

# Проверьте таблицу миграций
psql -U postgres -d auth_db -c "SELECT * FROM schema_migrations;"
```

### Проблемы с gRPC

**Проблема: gRPC connection refused**

```bash
# Проверьте, что gRPC сервер запущен
netstat -an | grep 9090

# Проверьте логи gRPC сервера
docker-compose logs auth-service | grep gRPC

# Тест gRPC соединения
grpcurl -plaintext localhost:9090 list
```

**Проблема: gRPC timeout**

Увеличьте таймауты в конфигурации:

```go
// internal/infrastructure/grpc/client.go
conn, err := grpc.Dial(
    address,
    grpc.WithTimeout(10*time.Second),
    grpc.WithBlock(),
)
```

### Проблемы с MAX API

**Проблема: MAX_id не получается**

```bash
# Проверьте токен MAX API
curl -H "Authorization: Bearer $MAX_API_TOKEN" \
     https://api.max.ru/v1/users/search?phone=+79991234567

# Проверьте логи MaxBot Service
docker-compose logs maxbot-service | grep "MAX API"

# Проверьте нормализацию телефона
curl -X POST http://localhost:9095/normalize-phone \
     -d '{"phone": "89991234567"}'
```

**Проблема: Batch операция зависла**

```bash
# Проверьте статус batch job
curl http://localhost:8081/employees/batch-status

# Проверьте логи
docker-compose logs employee-service | grep "batch"

# Перезапустите batch операцию
curl -X POST http://localhost:8081/employees/batch-update-maxid
```

### Проблемы с миграцией данных

**Проблема: Excel импорт не работает**

```bash
# Проверьте формат файла
file structure.xlsx

# Проверьте размер файла (лимит 50MB)
ls -lh structure.xlsx

# Проверьте логи
docker-compose logs migration-service | grep "excel"

# Проверьте обязательные колонки
# Требуются: phone, INN, org_name, chat_name, chat_url
```

**Проблема: Миграция зависла**

```bash
# Проверьте статус
curl http://localhost:8084/migration/jobs/1

# Проверьте ошибки
curl http://localhost:8084/migration/jobs/1/errors

# Перезапустите миграцию
docker-compose restart migration-service
```

### Проблемы с производительностью

**Проблема: Медленные запросы**

```bash
# Включите логирование медленных запросов в PostgreSQL
# postgresql.conf:
log_min_duration_statement = 1000  # логировать запросы > 1s

# Проверьте индексы
psql -U postgres -d chat_db -c "\d+ chats"

# Анализ запроса
psql -U postgres -d chat_db -c "EXPLAIN ANALYZE SELECT * FROM chats WHERE university_id = 1;"
```

**Проблема: Высокое использование памяти**

```bash
# Проверьте использование памяти
docker stats

# Ограничьте память для контейнера
docker-compose.yml:
  auth-service:
    mem_limit: 512m
    mem_reservation: 256m
```

### Проблемы с правами доступа

**Проблема: 403 Forbidden при доступе к ресурсам**

```bash
# Проверьте JWT токен
curl -X POST http://localhost:8080/auth/validate \
     -H "Authorization: Bearer $TOKEN"

# Проверьте роли пользователя
curl http://localhost:8080/users/123/permissions \
     -H "Authorization: Bearer $ADMIN_TOKEN"

# Проверьте логи Auth Service
docker-compose logs auth-service | grep "permission denied"
```

### Полезные команды

```bash
# Полная перезагрузка системы
docker-compose down -v
docker-compose up -d

# Очистка Docker
docker system prune -a --volumes

# Проверка всех health endpoints
for port in 8080 8081 8082 8083 8084; do
  echo "Checking port $port..."
  curl -s http://localhost:$port/health | jq .
done

# Экспорт логов
docker-compose logs > system_logs.txt

# Мониторинг в реальном времени
watch -n 1 'docker-compose ps'
```

## Дополнительная документация

### Спецификации и дизайн
- [Требования (Requirements)](./. kiro/specs/digital-university-mvp-completion/requirements.md)
- [Дизайн системы (Design)](./. kiro/specs/digital-university-mvp-completion/design.md)
- [План реализации (Tasks)](./. kiro/specs/digital-university-mvp-completion/tasks.md)

### Реализация функций
- [Ролевая фильтрация чатов](./chat-service/ROLE_BASED_FILTERING_IMPLEMENTATION.md)
- [Пагинация](./chat-service/PAGINATION_IMPLEMENTATION.md)
- [Поиск чатов](./chat-service/CHAT_SEARCH_IMPLEMENTATION.md)
- [Пакетное обновление MAX_id](./employee-service/BATCH_UPDATE_IMPLEMENTATION.md)
- [Интеграция с MAX_id](./employee-service/MAX_ID_INTEGRATION_SUMMARY.md)
- [Управление операторами подразделений](./structure-service/DEPARTMENT_MANAGERS_IMPLEMENTATION.md)
- [Иерархия структуры](./structure-service/STRUCTURE_HIERARCHY_IMPLEMENTATION.md)
- [Импорт из Excel](./structure-service/EXCEL_IMPORT_IMPLEMENTATION.md)
- [Пакетные операции MaxBot](./maxbot-service/BATCH_OPERATIONS_IMPLEMENTATION.md)

### Инфраструктура
- [Настройка gRPC](./GRPC_SETUP.md)
- [Реализация retry логики](./GRPC_RETRY_IMPLEMENTATION.md)
- [Обработка ошибок](./ERROR_HANDLING_IMPLEMENTATION.md)
- [Мониторинг и логирование](./MONITORING_AND_LOGGING_IMPLEMENTATION.md)
- [Миграции баз данных](./MIGRATIONS.md)

### Миграция данных
- [Быстрое руководство по миграции](./QUICK_MIGRATION_GUIDE.md)
- [Сводка по миграции](./MIGRATION_SUMMARY.md)
- [Реализация Migration Service](./migration-service/MIGRATION_SERVICE_IMPLEMENTATION.md)

### Интеграция с MAX Messenger
- [MaxBot Service README](./maxbot-service/README.md)
- [Руководство по интеграции MaxBot](./maxbot-service/INTEGRATION_GUIDE.md)
- [Примеры интеграции (Employee)](./employee-service/MAXBOT_INTEGRATION_EXAMPLES.md)
- [Примеры интеграции (Chat)](./chat-service/MAXBOT_INTEGRATION_EXAMPLES.md)
- [Быстрый старт MaxBot](./QUICK_START_MAXBOT_INTEGRATION.md)
- [Сводка интеграции MaxBot](./MAXBOT_INTEGRATION_SUMMARY.md)

### Тестирование
- [Быстрый старт интеграционных тестов](./integration-tests/QUICK_START.md)
- [Руководство по интеграционным тестам](./integration-tests/INTEGRATION_TEST_GUIDE.md)
- [Сводка по интеграционным тестам](./integration-tests/IMPLEMENTATION_SUMMARY.md)
- [Завершенные интеграционные тесты](./INTEGRATION_TESTS_COMPLETED.md)

## Контакты и поддержка

Для вопросов и предложений:
- Создайте issue в репозитории
- Обратитесь к документации в директории `.kiro/specs/`
- Проверьте существующие implementation guides

## Лицензия

[Укажите лицензию проекта]
