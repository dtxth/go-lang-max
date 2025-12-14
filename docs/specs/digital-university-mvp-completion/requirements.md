# Requirements Document

## Introduction

Данная спецификация описывает доработку микросервисной архитектуры мини-приложения "Цифровой вуз" для соответствия техническому заданию MVP. Система должна обеспечить централизованное управление вузами, сотрудниками, чатами и структурными подразделениями с поддержкой ролевой модели ABAC и миграцией более 150,000 существующих чатов из различных источников.

## Glossary

- **System** - Микросервисная система "Цифровой вуз"
- **Auth Service** - Сервис аутентификации и авторизации
- **Employee Service** - Сервис управления сотрудниками вузов
- **Chat Service** - Сервис управления групповыми чатами
- **Structure Service** - Сервис управления иерархической структурой вузов
- **MaxBot Service** - Сервис интеграции с MAX Messenger Bot API
- **ABAC** - Attribute-Based Access Control (управление доступом на основе атрибутов)
- **Superadmin** - Представитель VK с полными правами
- **Curator** - Ответственный представитель от вуза
- **Operator** - Представитель подразделения вуза
- **MAX_id** - Уникальный идентификатор пользователя в MAX Messenger
- **University** - Вуз (высшее учебное заведение)
- **Branch** - Филиал или головное подразделение вуза
- **Faculty** - Факультет или институт
- **Group** - Академическая группа студентов
- **Chat** - Групповой чат в MAX Messenger
- **Administrator** - Администратор чата
- **INN** - Идентификационный номер налогоплательщика
- **KPP** - Код причины постановки на учет
- **FOIV** - Федеральный орган исполнительной власти

## Requirements

### Requirement 1: Ролевая модель ABAC

**User Story:** Как администратор системы, я хочу иметь гибкую ролевую модель на основе атрибутов, чтобы разграничивать права доступа между суперадминами, кураторами и операторами.

#### Acceptance Criteria

1. WHEN the System initializes THEN the Auth Service SHALL create tables for roles and user-role assignments
2. WHEN a user authenticates THEN the Auth Service SHALL include role information in the JWT token payload
3. WHEN a service receives a request THEN the System SHALL validate user permissions based on role attributes and resource context
4. WHEN a Curator attempts to access resources THEN the System SHALL restrict access to only their assigned University
5. WHERE role hierarchy exists THEN the System SHALL grant higher roles all permissions of lower roles

### Requirement 2: Управление ролями сотрудников

**User Story:** Как куратор вуза, я хочу назначать операторов для подразделений, чтобы делегировать управление структурными единицами.

#### Acceptance Criteria

1. WHEN a Curator creates an employee record THEN the Employee Service SHALL allow assignment of Operator role
2. WHEN a Superadmin creates an employee record THEN the Employee Service SHALL allow assignment of any role including Curator
3. WHEN an employee is assigned a role THEN the System SHALL create corresponding user account in Auth Service
4. WHEN an employee role is updated THEN the System SHALL synchronize the change with Auth Service
5. WHEN an employee is deleted THEN the System SHALL revoke all associated permissions and invalidate tokens

### Requirement 3: Интеграция MAX_id для сотрудников

**User Story:** Как система, я хочу автоматически получать MAX_id по номеру телефона сотрудника, чтобы обеспечить интеграцию с MAX Messenger.

#### Acceptance Criteria

1. WHEN an employee is created with phone number THEN the Employee Service SHALL request MAX_id from MaxBot Service via gRPC
2. WHEN MaxBot Service receives phone number THEN the MaxBot Service SHALL normalize and validate the phone format
3. WHEN MaxBot Service finds user in MAX THEN the MaxBot Service SHALL return MAX_id to Employee Service
4. WHEN MAX_id is received THEN the Employee Service SHALL store MAX_id in employee record
5. WHEN MAX_id lookup fails THEN the System SHALL store employee record without MAX_id and log the failure

### Requirement 4: Пакетное обновление MAX_id

**User Story:** Как администратор системы, я хочу обновлять MAX_id для существующих сотрудников пакетно, чтобы заполнить недостающие данные.

#### Acceptance Criteria

1. WHEN a batch update is initiated THEN the Employee Service SHALL retrieve all employees without MAX_id
2. WHEN employees are retrieved THEN the Employee Service SHALL send phone numbers to MaxBot Service in batches
3. WHEN MaxBot Service receives batch request THEN the MaxBot Service SHALL process up to 100 phone numbers per request
4. WHEN MAX_ids are returned THEN the Employee Service SHALL update employee records with received MAX_ids
5. WHEN batch processing completes THEN the System SHALL generate report with success and failure counts

### Requirement 5: Фильтрация чатов по ролям

**User Story:** Как пользователь с определенной ролью, я хочу видеть только те чаты, к которым у меня есть доступ, чтобы не перегружать интерфейс лишней информацией.

#### Acceptance Criteria

1. WHEN a Superadmin requests chat list THEN the Chat Service SHALL return all chats from all universities
2. WHEN a Curator requests chat list THEN the Chat Service SHALL return only chats from their assigned University
3. WHEN an Operator requests chat list THEN the Chat Service SHALL return only chats from their assigned Branch or Faculty
4. WHEN filtering by role THEN the Chat Service SHALL validate role from JWT token via Auth Service gRPC
5. WHEN invalid role is detected THEN the System SHALL return 403 Forbidden error

### Requirement 6: Управление администраторами чатов с проверкой прав

**User Story:** Как куратор или оператор, я хочу добавлять и удалять администраторов чатов в рамках моих полномочий, чтобы управлять доступом к чатам.

#### Acceptance Criteria

1. WHEN adding administrator to chat THEN the Chat Service SHALL verify user has permission for that chat's University or Branch
2. WHEN adding administrator by phone THEN the Chat Service SHALL request MAX_id from MaxBot Service
3. WHEN removing administrator THEN the Chat Service SHALL verify at least one other administrator remains for the chat
4. WHEN attempting to remove last administrator THEN the System SHALL reject the request with error message
5. WHEN administrator is added or removed THEN the Chat Service SHALL update the administrators table and return success

### Requirement 7: Миграция чатов из админки

**User Story:** Как администратор системы, я хочу импортировать 6,000 чатов из существующей админки групповых чатов, чтобы сохранить существующие данные.

#### Acceptance Criteria

1. WHEN migration is initiated THEN the System SHALL read chat data from existing database with INN, chat name, and URL
2. WHEN processing each chat record THEN the System SHALL lookup or create University by INN
3. WHEN University is identified THEN the Chat Service SHALL create chat record with source 'admin_panel'
4. WHEN chat has administrator phone THEN the System SHALL create administrator record with phone number
5. WHEN migration completes THEN the System SHALL generate report with imported chat count and any errors

### Requirement 8: Миграция чатов из Google-таблицы

**User Story:** Как администратор системы, я хочу импортировать чаты, добавленные через бота-регистратора из Google-таблицы, чтобы включить их в систему.

#### Acceptance Criteria

1. WHEN migration is initiated THEN the System SHALL authenticate with Google Sheets API using service account
2. WHEN reading Google Sheet THEN the System SHALL parse columns: INN, KPP, URL, administrator phone
3. WHEN processing each row THEN the System SHALL lookup or create University by INN and KPP
4. WHEN University is identified THEN the Chat Service SHALL create chat record with source 'bot_registrar'
5. WHEN migration completes THEN the System SHALL log total processed rows and any validation errors

### Requirement 9: Миграция чатов академических групп из Excel

**User Story:** Как администратор системы, я хочу импортировать более 155,000 чатов академических групп из Excel файлов, чтобы связать их со структурой вузов.

#### Acceptance Criteria

1. WHEN Excel file is uploaded THEN the System SHALL validate file format and required columns
2. WHEN processing Excel row THEN the System SHALL parse: administrator phone, INN, FOIV, organization name, branch name, KPP, faculty, course, group number, chat name, chat URL
3. WHEN processing structure data THEN the Structure Service SHALL create or update University, Branch, Faculty, and Group records
4. WHEN structure is created THEN the Chat Service SHALL create chat record with source 'academic_group'
5. WHEN chat is created THEN the Structure Service SHALL link Group to Chat by storing chat_id reference

### Requirement 10: Связь структуры с чатами

**User Story:** Как пользователь системы, я хочу видеть связь между структурными подразделениями и чатами, чтобы понимать организационную иерархию.

#### Acceptance Criteria

1. WHEN Group is created with chat information THEN the Structure Service SHALL store chat_id reference
2. WHEN retrieving University structure THEN the Structure Service SHALL include chat information for each Group
3. WHEN chat_id is present THEN the Structure Service SHALL request chat details from Chat Service via gRPC
4. WHEN chat is deleted THEN the System SHALL set chat_id to NULL in Group record without deleting Group
5. WHEN displaying structure THEN the System SHALL show hierarchy: University → Branch → Faculty → Group → Chat

### Requirement 11: Связь подразделений с операторами

**User Story:** Как куратор вуза, я хочу назначать операторов для управления конкретными подразделениями, чтобы делегировать административные функции.

#### Acceptance Criteria

1. WHEN Curator assigns Operator to Branch THEN the Structure Service SHALL create department_managers record
2. WHEN Curator assigns Operator to Faculty THEN the Structure Service SHALL create department_managers record with faculty_id
3. WHEN Operator is assigned THEN the System SHALL verify Operator employee exists in Employee Service
4. WHEN retrieving Operator permissions THEN the System SHALL return list of assigned Branches and Faculties
5. WHEN Operator is removed from department THEN the System SHALL delete department_managers record

### Requirement 12: API для импорта структуры из Excel

**User Story:** Как куратор вуза, я хочу загружать структуру вуза из Excel файла через API, чтобы быстро заполнить систему данными.

#### Acceptance Criteria

1. WHEN Excel file is uploaded to import endpoint THEN the Structure Service SHALL validate file format and size limit
2. WHEN file is valid THEN the Structure Service SHALL parse Excel rows into structure entities
3. WHEN parsing completes THEN the Structure Service SHALL create University, Branch, Faculty, and Group records in transaction
4. WHEN duplicate records are detected THEN the System SHALL update existing records instead of creating duplicates
5. WHEN import completes THEN the System SHALL return summary with created, updated, and failed record counts

### Requirement 13: Отображение иерархической структуры вуза

**User Story:** Как пользователь системы, я хочу просматривать структуру вуза в иерархическом виде, чтобы понимать организационное устройство.

#### Acceptance Criteria

1. WHEN requesting University structure THEN the Structure Service SHALL return nested JSON with all levels
2. WHEN structure includes Branches THEN the System SHALL display: University → Branch → Faculty → Group → Chat
3. WHEN structure has no Branches THEN the System SHALL display: University → Faculty → Group → Chat
4. WHEN Group has associated chat THEN the System SHALL include chat name and URL in response
5. WHEN retrieving structure THEN the System SHALL order entities alphabetically within each level

### Requirement 14: Поиск сотрудников с фильтрацией по вузу

**User Story:** Как пользователь системы, я хочу искать сотрудников по имени, фамилии и названию вуза, чтобы быстро находить нужных людей.

#### Acceptance Criteria

1. WHEN search query is provided THEN the Employee Service SHALL search by first_name, last_name, and university name
2. WHEN Superadmin searches THEN the Employee Service SHALL return employees from all universities
3. WHEN Curator searches THEN the Employee Service SHALL return only employees from their University
4. WHEN search results are returned THEN the System SHALL include employee full name, phone, role, and university name
5. WHEN no matches found THEN the System SHALL return empty array with 200 status code

### Requirement 15: Добавление сотрудника с автоматическим созданием вуза

**User Story:** Как куратор или суперадмин, я хочу добавлять сотрудников с автоматическим созданием вуза при необходимости, чтобы упростить процесс регистрации.

#### Acceptance Criteria

1. WHEN creating employee with new University INN THEN the Employee Service SHALL create University record first
2. WHEN creating employee with existing University INN THEN the Employee Service SHALL reuse existing University
3. WHEN University is created THEN the System SHALL store name, INN, and KPP
4. WHEN employee is created THEN the System SHALL request MAX_id from MaxBot Service
5. WHEN all data is saved THEN the Employee Service SHALL return complete employee record with University details

### Requirement 16: Пагинация для списков чатов

**User Story:** Как пользователь системы, я хочу просматривать чаты с пагинацией, чтобы эффективно работать с большими объемами данных.

#### Acceptance Criteria

1. WHEN requesting chat list THEN the Chat Service SHALL accept limit and offset parameters
2. WHEN limit is not provided THEN the System SHALL use default limit of 50
3. WHEN limit exceeds 100 THEN the System SHALL cap limit at 100
4. WHEN returning results THEN the Chat Service SHALL include total count in response metadata
5. WHEN offset exceeds total count THEN the System SHALL return empty array

### Requirement 17: Поиск чатов по названию

**User Story:** Как пользователь системы, я хочу искать чаты по названию, чтобы быстро находить нужные чаты.

#### Acceptance Criteria

1. WHEN search query is provided THEN the Chat Service SHALL use full-text search on chat name
2. WHEN searching THEN the System SHALL apply role-based filtering before returning results
3. WHEN search uses Russian text THEN the Chat Service SHALL use Russian language text search configuration
4. WHEN multiple words are provided THEN the System SHALL match chats containing all words
5. WHEN no matches found THEN the System SHALL return empty array with 200 status code

### Requirement 18: gRPC интеграция между сервисами

**User Story:** Как система, я хочу использовать gRPC для внутренней коммуникации между сервисами, чтобы обеспечить высокую производительность и типобезопасность.

#### Acceptance Criteria

1. WHEN service needs to validate token THEN the System SHALL call Auth Service ValidateToken gRPC method
2. WHEN service needs MAX_id THEN the System SHALL call MaxBot Service GetUserByPhone gRPC method
3. WHEN service needs chat details THEN the System SHALL call Chat Service GetChat gRPC method
4. WHEN gRPC call fails THEN the System SHALL retry up to 3 times with exponential backoff
5. WHEN all retries fail THEN the System SHALL log error and return appropriate HTTP error to client

### Requirement 19: Валидация номеров телефонов

**User Story:** Как система, я хочу валидировать и нормализовать номера телефонов, чтобы обеспечить единообразие данных.

#### Acceptance Criteria

1. WHEN phone number is provided THEN the MaxBot Service SHALL normalize to E.164 format
2. WHEN phone starts with 8 THEN the System SHALL replace with +7
3. WHEN phone starts with 9 THEN the System SHALL prepend +7
4. WHEN phone contains non-digit characters THEN the System SHALL remove them before validation
5. WHEN phone format is invalid THEN the System SHALL return validation error with clear message

### Requirement 20: Логирование и мониторинг миграции

**User Story:** Как администратор системы, я хочу видеть детальные логи процесса миграции, чтобы отслеживать прогресс и выявлять проблемы.

#### Acceptance Criteria

1. WHEN migration starts THEN the System SHALL log start time and source type
2. WHEN processing each record THEN the System SHALL log record identifier and processing status
3. WHEN error occurs THEN the System SHALL log error details with record context
4. WHEN migration completes THEN the System SHALL log summary with total, success, and failure counts
5. WHEN migration runs THEN the System SHALL expose progress metrics via HTTP endpoint
