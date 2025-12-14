# Swagger Types Summary

Этот документ содержит сводку всех типов данных, определенных в Swagger документации каждого сервиса.

## Auth Service

**Базовый путь:** `/`  
**Swagger файлы:** `auth-service/internal/infrastructure/http/docs/swagger.{json,yaml}`

### Определенные типы (definitions):

1. **domain.TokenPair**
   - `access_token` (string)
   - `refresh_token` (string)

2. **domain.User**
   - `id` (integer)
   - `email` (string)
   - `role` (string)

### Эндпоинты:
- `POST /login` - Вход пользователя
- `POST /register` - Регистрация пользователя
- `POST /refresh` - Обновление токена
- `POST /logout` - Выход пользователя
- `GET /health` - Проверка здоровья сервиса

---

## Employee Service

**Базовый путь:** `/`  
**Swagger файлы:** `employee-service/internal/infrastructure/http/docs/swagger.{json,yaml}`

### Определенные типы (definitions):

1. **domain.BatchUpdateJob**
   - `id` (integer)
   - `job_type` (string) - 'max_id_update'
   - `status` (string) - 'running', 'completed', 'failed'
   - `total` (integer) - Всего записей для обработки
   - `processed` (integer) - Успешно обработано
   - `failed` (integer) - Неудачных записей
   - `started_at` (string)
   - `completed_at` (string)

2. **domain.BatchUpdateResult**
   - `job_id` (integer)
   - `total` (integer)
   - `success` (integer)
   - `failed` (integer)
   - `errors` (array of strings)

3. **domain.University**
   - `id` (integer)
   - `name` (string)
   - `inn` (string)
   - `kpp` (string)
   - `created_at` (string)
   - `updated_at` (string)

4. **http.AddEmployeeRequest**
   - `first_name` (string, required)
   - `last_name` (string, required)
   - `middle_name` (string)
   - `phone` (string, required)
   - `university_name` (string)
   - `inn` (string)
   - `kpp` (string)

5. **http.UpdateEmployeeRequest**
   - `first_name` (string)
   - `last_name` (string)
   - `middle_name` (string)
   - `phone` (string)
   - `university_id` (integer)
   - `inn` (string)
   - `kpp` (string)

6. **http.Employee**
   - `id` (integer)
   - `first_name` (string)
   - `last_name` (string)
   - `middle_name` (string)
   - `phone` (string) - Номер телефона
   - `max_id` (string) - MAX_id (заменяет телефон)
   - `max_id_updated_at` (string) - Время последнего обновления MAX_id
   - `role` (string) - Роль: curator, operator, или пусто
   - `user_id` (integer) - ID пользователя в auth-service
   - `university_id` (integer)
   - `university` (domain.University)
   - `inn` (string) - ИНН
   - `kpp` (string) - КПП
   - `created_at` (string)
   - `updated_at` (string)

7. **http.DeleteResponse**
   - `status` (string) - "deleted"

8. **usecase.SearchEmployeeResult**
   - `id` (integer)
   - `full_name` (string)
   - `phone` (string)
   - `role` (string)
   - `university_name` (string)

### Эндпоинты:
- `POST /employees` - Добавить сотрудника
- `GET /employees` - Поиск сотрудников
- `GET /employees/all` - Получить всех сотрудников
- `GET /employees/{id}` - Получить сотрудника по ID
- `PUT /employees/{id}` - Обновить сотрудника
- `DELETE /employees/{id}` - Удалить сотрудника
- `POST /employees/batch-update-maxid` - Запустить пакетное обновление MAX_id
- `GET /employees/batch-status/{id}` - Получить статус пакетного обновления
- `GET /employees/batch-status` - Список всех пакетных заданий

---

## Chat Service

**Базовый путь:** `/`  
**Swagger файлы:** `chat-service/internal/infrastructure/http/docs/swagger.{json,yaml}`

### Определенные типы (definitions):

1. **domain.Administrator**
   - `id` (integer)
   - `chat_id` (integer)
   - `phone` (string) - Номер телефона администратора
   - `max_id` (string) - MAX_id администратора
   - `created_at` (string)
   - `updated_at` (string)

2. **domain.University**
   - `id` (integer)
   - `name` (string)
   - `inn` (string)
   - `kpp` (string)
   - `created_at` (string)
   - `updated_at` (string)

3. **http.AddAdministratorRequest**
   - `phone` (string, required)

4. **http.Administrator**
   - `id` (integer)
   - `chat_id` (integer)
   - `phone` (string) - Номер телефона администратора
   - `max_id` (string) - MAX_id администратора
   - `created_at` (string)
   - `updated_at` (string)

5. **http.Chat**
   - `id` (integer)
   - `name` (string) - Название чата
   - `department` (string) - Подразделение вуза
   - `max_chat_id` (string) - ID чата в MAX
   - `url` (string) - Ссылка на чат
   - `participants_count` (integer) - Количество участников
   - `source` (string) - Источник: "admin_panel", "bot_registrar", "academic_group"
   - `university_id` (integer) - ID вуза (опционально)
   - `university` (domain.University) - Вуз (для суперадмина)
   - `administrators` (array of domain.Administrator) - Администраторы чата
   - `created_at` (string)
   - `updated_at` (string)

6. **http.ChatListResponse**
   - `chats` (array of http.Chat)
   - `total_count` (integer)
   - `limit` (integer)
   - `offset` (integer)

7. **http.DeleteResponse**
   - `status` (string) - "deleted"

### Эндпоинты:
- `GET /chats` - Поиск чатов
- `GET /chats/all` - Получить все чаты
- `GET /chats/{id}` - Получить чат по ID
- `POST /chats/{chat_id}/administrators` - Добавить администратора к чату
- `DELETE /administrators/{admin_id}` - Удалить администратора из чата

---

## Structure Service

**Базовый путь:** `/`  
**Swagger файлы:** `structure-service/internal/infrastructure/http/docs/swagger.{json,yaml}`

### Определенные типы (definitions):

1. **domain.Chat**
   - `id` (integer)
   - `name` (string)
   - `max_id` (string)
   - `url` (string)

2. **domain.DepartmentManager**
   - `id` (integer)
   - `employee_id` (integer)
   - `branch_id` (integer)
   - `faculty_id` (integer)
   - `assigned_by` (integer) - User ID куратора
   - `assigned_at` (string)

3. **domain.ImportResult**
   - `created` (integer) - Количество созданных записей
   - `updated` (integer) - Количество обновленных записей
   - `failed` (integer) - Количество неудачных записей
   - `errors` (array of strings) - Список ошибок

4. **domain.StructureNode**
   - `id` (integer)
   - `type` (string) - "university", "branch", "faculty", "group"
   - `name` (string)
   - `course` (integer)
   - `group_num` (string)
   - `chat` (domain.Chat)
   - `children` (array of domain.StructureNode)

5. **domain.University**
   - `id` (integer)
   - `name` (string)
   - `inn` (string)
   - `kpp` (string)
   - `foiv` (string) - ФОИВ
   - `created_at` (string)
   - `updated_at` (string)

6. **http.AssignOperatorRequest**
   - `employee_id` (integer)
   - `branch_id` (integer)
   - `faculty_id` (integer)
   - `assigned_by` (integer)

### Эндпоинты:
- `GET /universities` - Получить все вузы
- `POST /universities` - Создать вуз
- `GET /universities/{id}` - Получить вуз по ID
- `GET /universities/{university_id}/structure` - Получить структуру вуза
- `POST /import/excel` - Импортировать структуру из Excel
- `GET /departments/managers` - Получить все назначения операторов
- `POST /departments/managers` - Назначить оператора на подразделение
- `DELETE /departments/managers/{id}` - Удалить назначение оператора

---

## Migration Service

**Базовый путь:** `/`  
**Swagger файлы:** `migration-service/internal/infrastructure/http/docs/swagger.{json,yaml}`

### Определенные типы (definitions):

1. **http.ErrorResponse**
   - `error` (string)

2. **http.MigrationJobResponse**
   - `id` (integer)
   - `source_type` (string)
   - `source_identifier` (string)
   - `status` (string)
   - `total` (integer)
   - `processed` (integer)
   - `failed` (integer)
   - `started_at` (string)
   - `completed_at` (string)

3. **http.StartDatabaseMigrationRequest**
   - `source_identifier` (string)

4. **http.StartGoogleSheetsMigrationRequest**
   - `spreadsheet_id` (string)

### Эндпоинты:
- `POST /migration/database` - Запустить миграцию из базы данных
- `POST /migration/google-sheets` - Запустить миграцию из Google Sheets
- `POST /migration/excel` - Запустить миграцию из Excel
- `GET /migration/jobs` - Список всех заданий миграции
- `GET /migration/jobs/{id}` - Получить статус задания миграции

---

## Maxbot Service

**Тип:** gRPC сервис (без HTTP/Swagger документации)

Maxbot Service использует только gRPC протокол для коммуникации. Swagger документация не применима.

### gRPC методы:
- `GetUserByPhone` - Получить пользователя по номеру телефона
- `BatchGetUsersByPhone` - Пакетное получение пользователей по номерам телефонов

---

## Общие замечания

### Безопасность
Все HTTP эндпоинты (кроме health checks) требуют Bearer токен в заголовке Authorization.

### Форматы данных
- Даты возвращаются в формате ISO 8601: `2006-01-02T15:04:05Z07:00`
- Все JSON поля используют snake_case

### Пагинация
Сервисы поддерживают пагинацию через параметры:
- `limit` - максимум записей (по умолчанию 50, максимум 100)
- `offset` - смещение для пагинации

### Коды ответов
- `200` - Успешный запрос
- `201` - Ресурс создан
- `202` - Запрос принят (асинхронная обработка)
- `204` - Успешно, без содержимого
- `400` - Неверный запрос
- `401` - Не авторизован
- `403` - Доступ запрещен
- `404` - Не найдено
- `409` - Конфликт
- `500` - Внутренняя ошибка сервера
