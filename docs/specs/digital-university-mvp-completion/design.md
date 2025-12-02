# Design Document

## Overview

Данный дизайн описывает доработку микросервисной архитектуры "Цифровой вуз" для реализации MVP согласно техническому заданию. Система состоит из 5 микросервисов (Auth, Employee, Chat, Structure, MaxBot), взаимодействующих через HTTP REST API и gRPC.

Ключевые аспекты дизайна:
- **Ролевая модель ABAC** с поддержкой иерархии прав (Superadmin > Curator > Operator)
- **Автоматическая интеграция с MAX Messenger** через получение MAX_id по номеру телефона
- **Миграция 150,000+ чатов** из трех различных источников данных
- **Иерархическая структура вузов** с поддержкой связей между подразделениями, сотрудниками и чатами
- **Чистая архитектура** с разделением на domain, usecase, infrastructure слои

## Architecture

### High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                         Client Layer                             │
│                    (React WebApp - Future)                       │
└────────────────────────────┬────────────────────────────────────┘
                             │ HTTP REST
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│                      API Gateway Layer                           │
│                    (Future: Kong/Nginx)                          │
└─────┬──────────┬──────────┬──────────┬──────────┬──────────────┘
      │          │          │          │          │
      │ HTTP     │ HTTP     │ HTTP     │ HTTP     │ gRPC
      ▼          ▼          ▼          ▼          ▼
┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐
│   Auth   │ │ Employee │ │   Chat   │ │Structure │ │  MaxBot  │
│ Service  │ │ Service  │ │ Service  │ │ Service  │ │ Service  │
│          │ │          │ │          │ │          │ │          │
│ :8080    │ │ :8081    │ │ :8082    │ │ :8083    │ │ :9095    │
│ gRPC:9090│ │ gRPC:9091│ │ gRPC:9092│ │ gRPC:9093│ │ (gRPC)   │
└────┬─────┘ └────┬─────┘ └────┬─────┘ └────┬─────┘ └────┬─────┘
     │            │            │            │            │
     │ PostgreSQL │ PostgreSQL │ PostgreSQL │ PostgreSQL │ MAX API
     ▼            ▼            ▼            ▼            ▼
┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐
│ auth-db  │ │employee  │ │ chat-db  │ │structure │ │   MAX    │
│  :5432   │ │  -db     │ │  :5434   │ │  -db     │ │Messenger │
│          │ │  :5433   │ │          │ │  :5435   │ │   API    │
└──────────┘ └──────────┘ └──────────┘ └──────────┘ └──────────┘
```

### Service Interactions

```
┌─────────────┐
│   Client    │
└──────┬──────┘
       │ 1. POST /employees (with phone)
       ▼
┌─────────────────┐
│ Employee Service│
└──────┬──────────┘
       │ 2. gRPC: GetUserByPhone(phone)
       ▼
┌─────────────────┐
│  MaxBot Service │
└──────┬──────────┘
       │ 3. HTTP: MAX API
       ▼
┌─────────────────┐
│   MAX API       │
└──────┬──────────┘
       │ 4. Return MAX_id
       ▼
┌─────────────────┐
│ Employee Service│ 5. Store employee with MAX_id
└─────────────────┘
```

### Role-Based Access Control Flow

```
┌─────────────┐
│   Client    │
└──────┬──────┘
       │ 1. Request with JWT
       ▼
┌─────────────────┐
│  Chat Service   │
└──────┬──────────┘
       │ 2. gRPC: ValidateToken(jwt)
       ▼
┌─────────────────┐
│  Auth Service   │
└──────┬──────────┘
       │ 3. Return user_id, role, university_id
       ▼
┌─────────────────┐
│  Chat Service   │ 4. Apply role-based filtering
└──────┬──────────┘
       │ 5. Query chats with filters
       ▼
┌─────────────────┐
│    chat-db      │
└─────────────────┘
```

## Components and Interfaces

### Auth Service

**Responsibilities:**
- User authentication and JWT token generation
- Role management (Superadmin, Curator, Operator)
- Token validation via gRPC for other services
- User-role assignment and permission checking

**New Components:**

1. **Domain Layer:**
   - `Role` entity: id, name, description
   - `UserRole` entity: user_id, role_id, university_id, branch_id, faculty_id
   - `Permission` value object: resource, action, context

2. **Use Cases:**
   - `AssignRoleUseCase`: Assign role to user with context (university/branch/faculty)
   - `ValidatePermissionUseCase`: Check if user has permission for specific action
   - `GetUserRolesUseCase`: Retrieve all roles and contexts for user

3. **gRPC Interface:**
```protobuf
service AuthService {
  rpc ValidateToken(ValidateTokenRequest) returns (ValidateTokenResponse);
  rpc GetUserPermissions(GetUserPermissionsRequest) returns (GetUserPermissionsResponse);
  rpc AssignRole(AssignRoleRequest) returns (AssignRoleResponse);
}

message ValidateTokenResponse {
  int32 user_id = 1;
  string role = 2;
  int32 university_id = 3;
  int32 branch_id = 4;
  int32 faculty_id = 5;
}
```

### Employee Service

**Responsibilities:**
- CRUD operations for employees
- Integration with Auth Service for role assignment
- Integration with MaxBot Service for MAX_id retrieval
- University management
- Batch MAX_id updates

**New Components:**

1. **Domain Layer:**
   - Enhanced `Employee` entity: add `role`, `max_id`, `user_id` fields
   - `BatchUpdateResult` value object: total, success, failed, errors

2. **Use Cases:**
   - `CreateEmployeeWithRoleUseCase`: Create employee and assign role via Auth Service
   - `BatchUpdateMaxIdUseCase`: Update MAX_id for employees without it
   - `SearchEmployeesWithRoleFilterUseCase`: Search with role-based filtering

3. **HTTP Endpoints:**
```
POST   /employees                    - Create employee with role
PUT    /employees/{id}                - Update employee
DELETE /employees/{id}                - Delete employee and revoke role
GET    /employees                     - Search with role filtering
POST   /employees/batch-update-maxid  - Trigger batch MAX_id update
GET    /employees/batch-status        - Get batch update status
```

### Chat Service

**Responsibilities:**
- CRUD operations for chats
- Administrator management with permission checks
- Role-based chat filtering
- Integration with Auth Service for permission validation
- Integration with MaxBot Service for administrator MAX_id

**New Components:**

1. **Domain Layer:**
   - Enhanced `Chat` entity: add `source` field (admin_panel, bot_registrar, academic_group)
   - `ChatFilter` value object: role, university_id, branch_id, faculty_id

2. **Use Cases:**
   - `ListChatsWithRoleFilterUseCase`: Filter chats based on user role and context
   - `AddAdministratorWithPermissionCheckUseCase`: Verify permissions before adding admin
   - `RemoveAdministratorWithValidationUseCase`: Ensure at least one admin remains

3. **Middleware:**
   - `AuthMiddleware`: Validate JWT and extract role information
   - `PermissionMiddleware`: Check resource-level permissions

### Structure Service

**Responsibilities:**
- Hierarchical structure management (University → Branch → Faculty → Group)
- Excel import for structure and chats
- Link groups to chats
- Department manager assignments (Operators to Branches/Faculties)

**New Components:**

1. **Domain Layer:**
   - `DepartmentManager` entity: id, employee_id, branch_id, faculty_id, assigned_at
   - `ImportResult` value object: created, updated, failed, errors
   - `StructureHierarchy` aggregate: University with nested Branches, Faculties, Groups, Chats

2. **Use Cases:**
   - `ImportStructureFromExcelUseCase`: Parse Excel and create structure entities
   - `AssignOperatorToDepartmentUseCase`: Link operator to branch/faculty
   - `GetUniversityStructureUseCase`: Retrieve full hierarchy with chat details

3. **HTTP Endpoints:**
```
POST   /import/excel                      - Import structure from Excel
GET    /universities/{id}/structure       - Get full hierarchy
POST   /departments/managers              - Assign operator to department
DELETE /departments/managers/{id}         - Remove operator assignment
GET    /departments/managers              - List all department managers
```

4. **Excel Parser:**
   - Parse columns: phone, INN, FOIV, org_name, branch_name, KPP, faculty, course, group_number, chat_name, chat_url
   - Validate required fields
   - Handle duplicates by INN+KPP
   - Create structure in transaction

### MaxBot Service

**Responsibilities:**
- Integration with MAX Messenger Bot API
- Phone number normalization and validation
- Single and batch MAX_id lookup
- User search by phone

**New Components:**

1. **Use Cases:**
   - `BatchGetUsersByPhoneUseCase`: Process up to 100 phones per request
   - `NormalizePhoneUseCase`: Convert to E.164 format

2. **gRPC Interface:**
```protobuf
service MaxBotService {
  rpc GetUserByPhone(GetUserByPhoneRequest) returns (GetUserByPhoneResponse);
  rpc BatchGetUsersByPhone(BatchGetUsersByPhoneRequest) returns (BatchGetUsersByPhoneResponse);
  rpc NormalizePhone(NormalizePhoneRequest) returns (NormalizePhoneResponse);
}

message BatchGetUsersByPhoneRequest {
  repeated string phones = 1;
}

message BatchGetUsersByPhoneResponse {
  repeated UserPhoneMapping mappings = 1;
}

message UserPhoneMapping {
  string phone = 1;
  string max_id = 2;
  bool found = 3;
}
```

### Migration Service (New)

**Responsibilities:**
- Orchestrate migration from three data sources
- Import from existing database (admin_panel source)
- Import from Google Sheets (bot_registrar source)
- Import from Excel files (academic_group source)
- Progress tracking and error reporting

**Components:**

1. **Domain Layer:**
   - `MigrationJob` entity: id, source_type, status, total, processed, failed, started_at, completed_at
   - `MigrationError` entity: job_id, record_identifier, error_message

2. **Use Cases:**
   - `MigrateFromDatabaseUseCase`: Import 6,000 chats from existing DB
   - `MigrateFromGoogleSheetsUseCase`: Import from Google Sheets API
   - `MigrateFromExcelUseCase`: Import 155,000+ chats from Excel
   - `GetMigrationStatusUseCase`: Track progress

3. **HTTP Endpoints:**
```
POST   /migration/database              - Start database migration
POST   /migration/google-sheets         - Start Google Sheets migration
POST   /migration/excel                 - Upload and start Excel migration
GET    /migration/jobs/{id}             - Get migration job status
GET    /migration/jobs                  - List all migration jobs
```

## Data Models

### Auth Service Schema

```sql
-- Roles table
CREATE TABLE roles (
  id SERIAL PRIMARY KEY,
  name TEXT NOT NULL UNIQUE, -- 'superadmin', 'curator', 'operator'
  description TEXT,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT now()
);

-- User roles with context
CREATE TABLE user_roles (
  id SERIAL PRIMARY KEY,
  user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  role_id INTEGER NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
  university_id INTEGER, -- NULL for superadmin
  branch_id INTEGER,     -- NULL for curator, set for operator
  faculty_id INTEGER,    -- NULL for curator, set for operator
  assigned_by INTEGER REFERENCES users(id),
  assigned_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
  UNIQUE(user_id, role_id, university_id, branch_id, faculty_id)
);

CREATE INDEX idx_user_roles_user_id ON user_roles(user_id);
CREATE INDEX idx_user_roles_role_id ON user_roles(role_id);
CREATE INDEX idx_user_roles_university_id ON user_roles(university_id);
```

### Employee Service Schema

```sql
-- Enhanced employees table
ALTER TABLE employees 
  ADD COLUMN role TEXT, -- 'curator', 'operator', NULL for regular employee
  ADD COLUMN user_id INTEGER, -- Reference to auth-service user
  ADD COLUMN max_id_updated_at TIMESTAMP WITH TIME ZONE;

CREATE INDEX idx_employees_role ON employees(role);
CREATE INDEX idx_employees_user_id ON employees(user_id);

-- Batch update tracking
CREATE TABLE batch_update_jobs (
  id SERIAL PRIMARY KEY,
  job_type TEXT NOT NULL, -- 'max_id_update'
  status TEXT NOT NULL, -- 'running', 'completed', 'failed'
  total INTEGER DEFAULT 0,
  processed INTEGER DEFAULT 0,
  failed INTEGER DEFAULT 0,
  started_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
  completed_at TIMESTAMP WITH TIME ZONE
);
```

### Structure Service Schema

```sql
-- Department managers (operators assigned to departments)
CREATE TABLE department_managers (
  id SERIAL PRIMARY KEY,
  employee_id INTEGER NOT NULL, -- Reference to employee-service
  branch_id INTEGER REFERENCES branches(id) ON DELETE CASCADE,
  faculty_id INTEGER REFERENCES faculties(id) ON DELETE CASCADE,
  assigned_by INTEGER, -- Reference to curator user_id
  assigned_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
  UNIQUE(employee_id, branch_id, faculty_id),
  CHECK (branch_id IS NOT NULL OR faculty_id IS NOT NULL)
);

CREATE INDEX idx_department_managers_employee_id ON department_managers(employee_id);
CREATE INDEX idx_department_managers_branch_id ON department_managers(branch_id);
CREATE INDEX idx_department_managers_faculty_id ON department_managers(faculty_id);

-- Enhanced groups table to store chat link
ALTER TABLE groups 
  ADD COLUMN chat_url TEXT,
  ADD COLUMN chat_name TEXT;
```

### Migration Service Schema

```sql
CREATE TABLE migration_jobs (
  id SERIAL PRIMARY KEY,
  source_type TEXT NOT NULL, -- 'database', 'google_sheets', 'excel'
  source_identifier TEXT, -- file path or sheet ID
  status TEXT NOT NULL, -- 'pending', 'running', 'completed', 'failed'
  total INTEGER DEFAULT 0,
  processed INTEGER DEFAULT 0,
  failed INTEGER DEFAULT 0,
  started_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
  completed_at TIMESTAMP WITH TIME ZONE
);

CREATE TABLE migration_errors (
  id SERIAL PRIMARY KEY,
  job_id INTEGER NOT NULL REFERENCES migration_jobs(id) ON DELETE CASCADE,
  record_identifier TEXT NOT NULL,
  error_message TEXT NOT NULL,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT now()
);

CREATE INDEX idx_migration_errors_job_id ON migration_errors(job_id);
```


## Correctness Properties

*A property is a characteristic or behavior that should hold true across all valid executions of a system-essentially, a formal statement about what the system should do. Properties serve as the bridge between human-readable specifications and machine-verifiable correctness guarantees.*

### Property 1: JWT tokens contain role information

*For any* authenticated user with an assigned role, the JWT token payload should include the role name, university_id, branch_id, and faculty_id (where applicable)
**Validates: Requirements 1.2**

### Property 2: Permission validation respects ABAC rules

*For any* request with role, resource, and context attributes, the permission validation should return true only when the role has access to the resource in that context
**Validates: Requirements 1.3**

### Property 3: Curator access is restricted to assigned university

*For any* curator user and resource request, the system should only grant access if the resource belongs to the curator's assigned university
**Validates: Requirements 1.4**

### Property 4: Role hierarchy grants cumulative permissions

*For any* higher role in the hierarchy (superadmin > curator > operator), the higher role should have all permissions of lower roles
**Validates: Requirements 1.5**

### Property 5: Role assignment creates auth service user

*For any* employee created with a role, a corresponding user account should be created in the auth service with the same role
**Validates: Requirements 2.3**

### Property 6: Role updates synchronize across services

*For any* employee role update, the corresponding user role in auth service should be updated to match
**Validates: Requirements 2.4**

### Property 7: Employee deletion revokes all permissions

*For any* employee deletion, all associated user roles should be removed and tokens should be invalidated
**Validates: Requirements 2.5**

### Property 8: Employee creation triggers MAX_id lookup

*For any* employee created with a phone number, a gRPC call to MaxBot Service should be made to retrieve MAX_id
**Validates: Requirements 3.1**

### Property 9: Phone numbers are normalized to E.164

*For any* phone number input, the MaxBot Service should normalize it to E.164 format (+7XXXXXXXXXX for Russian numbers)
**Validates: Requirements 3.2, 19.1**

### Property 10: MAX_id is stored when received

*For any* successful MAX_id lookup, the employee record should be updated with the received MAX_id
**Validates: Requirements 3.4**

### Property 11: Employee creation succeeds without MAX_id

*For any* employee creation where MAX_id lookup fails, the employee record should still be created without MAX_id
**Validates: Requirements 3.5**

### Property 12: Batch update processes correct employees

*For any* batch MAX_id update, only employees without MAX_id should be selected for processing
**Validates: Requirements 4.1**

### Property 13: Batch requests respect size limits

*For any* batch request to MaxBot Service, the number of phone numbers should not exceed 100
**Validates: Requirements 4.2, 4.3**

### Property 14: Batch results are accurately reported

*For any* completed batch update, the report should accurately reflect the number of successful and failed updates
**Validates: Requirements 4.5**

### Property 15: Superadmin sees all chats

*For any* superadmin user requesting chat list, all chats from all universities should be returned
**Validates: Requirements 5.1**

### Property 16: Curator sees only university chats

*For any* curator user requesting chat list, only chats from their assigned university should be returned
**Validates: Requirements 5.2**

### Property 17: Operator sees only department chats

*For any* operator user requesting chat list, only chats from their assigned branch or faculty should be returned
**Validates: Requirements 5.3**

### Property 18: Invalid tokens are rejected

*For any* request with an invalid or expired JWT token, the system should return 403 Forbidden error
**Validates: Requirements 5.4, 5.5**

### Property 19: Administrator addition requires permission

*For any* attempt to add administrator to chat, the user must have permission for that chat's university or branch
**Validates: Requirements 6.1**

### Property 20: Adding administrator triggers MAX_id lookup

*For any* administrator added by phone number, a gRPC call to MaxBot Service should be made to retrieve MAX_id
**Validates: Requirements 6.2**

### Property 21: Last administrator cannot be removed

*For any* chat with only one administrator, attempting to remove that administrator should be rejected with an error
**Validates: Requirements 6.3, 6.4**

### Property 22: Administrator changes are persisted

*For any* successful administrator addition or removal, the administrators table should be updated accordingly
**Validates: Requirements 6.5**

### Property 23: Migrated chats have correct source

*For any* chat migrated from database, Google Sheets, or Excel, the source field should be set to 'admin_panel', 'bot_registrar', or 'academic_group' respectively
**Validates: Requirements 7.3, 8.4, 9.4**

### Property 24: Universities are created or reused by INN

*For any* chat or structure import with INN, the system should create a new university if INN doesn't exist, or reuse existing university
**Validates: Requirements 7.2, 8.3**

### Property 25: Migration generates accurate reports

*For any* completed migration, the report should accurately reflect the number of imported records and errors
**Validates: Requirements 7.5, 8.5**

### Property 26: Excel import validates required columns

*For any* Excel file upload, the system should validate that all required columns are present before processing
**Validates: Requirements 9.1**

### Property 27: Structure import creates full hierarchy

*For any* Excel row with structure data, the system should create or update University, Branch, Faculty, and Group records
**Validates: Requirements 9.3**

### Property 28: Groups are linked to chats

*For any* group created with chat information, the chat_id reference should be stored in the group record
**Validates: Requirements 9.5, 10.1**

### Property 29: Structure retrieval includes chat details

*For any* university structure request, the response should include chat information for groups that have associated chats
**Validates: Requirements 10.2**

### Property 30: Chat deletion preserves groups

*For any* chat deletion, the associated group should remain with chat_id set to NULL
**Validates: Requirements 10.4**

### Property 31: Structure displays correct hierarchy

*For any* university structure, the hierarchy should be: University → Branch (optional) → Faculty → Group → Chat
**Validates: Requirements 10.5**

### Property 32: Department manager assignments are validated

*For any* operator assignment to department, the system should verify the operator employee exists in Employee Service
**Validates: Requirements 11.3**

### Property 33: Operator permissions reflect assignments

*For any* operator user, the list of assigned branches and faculties should match the department_managers records
**Validates: Requirements 11.4**

### Property 34: Excel import is transactional

*For any* Excel import, either all records should be created successfully, or none should be created if an error occurs
**Validates: Requirements 12.3**

### Property 35: Duplicate records are updated not created

*For any* import with duplicate INN+KPP, the system should update existing records instead of creating duplicates
**Validates: Requirements 12.4**

### Property 36: Import summary is accurate

*For any* completed import, the summary should accurately reflect created, updated, and failed record counts
**Validates: Requirements 12.5**

### Property 37: Structure is returned as nested JSON

*For any* university structure request, the response should be properly nested JSON with all hierarchy levels
**Validates: Requirements 13.1**

### Property 38: Structure entities are alphabetically ordered

*For any* university structure, entities at each level should be ordered alphabetically by name
**Validates: Requirements 13.5**

### Property 39: Employee search matches multiple fields

*For any* search query, the system should match employees by first_name, last_name, or university name
**Validates: Requirements 14.1**

### Property 40: Search respects role-based filtering

*For any* curator search, only employees from their university should be returned
**Validates: Requirements 14.3**

### Property 41: Search results include all required fields

*For any* search result, each employee should include full name, phone, role, and university name
**Validates: Requirements 14.4**

### Property 42: New universities are created automatically

*For any* employee creation with new INN, a university record should be created before the employee
**Validates: Requirements 15.1**

### Property 43: Existing universities are reused

*For any* employee creation with existing INN, the existing university should be reused instead of creating duplicate
**Validates: Requirements 15.2**

### Property 44: Pagination limit is capped at 100

*For any* chat list request with limit > 100, the system should cap the limit at 100
**Validates: Requirements 16.3**

### Property 45: Pagination includes total count

*For any* paginated response, the metadata should include the total count of records
**Validates: Requirements 16.4**

### Property 46: Search applies role-based filtering

*For any* chat search, role-based filtering should be applied before returning results
**Validates: Requirements 17.2**

### Property 47: Multi-word search requires all words

*For any* search query with multiple words, only chats containing all words should be returned
**Validates: Requirements 17.4**

### Property 48: gRPC calls are retried on failure

*For any* failed gRPC call, the system should retry up to 3 times with exponential backoff
**Validates: Requirements 18.4**

### Property 49: Failed retries return appropriate errors

*For any* gRPC call that fails after all retries, the system should log the error and return appropriate HTTP error
**Validates: Requirements 18.5**

### Property 50: Phone normalization handles Russian formats

*For any* phone starting with 8 or 9, the system should normalize to +7 format
**Validates: Requirements 19.2, 19.3**

### Property 51: Phone normalization removes non-digits

*For any* phone containing spaces, dashes, or parentheses, the system should remove them before validation
**Validates: Requirements 19.4**

### Property 52: Invalid phones return clear errors

*For any* invalid phone format, the system should return a validation error with a clear message
**Validates: Requirements 19.5**

## Error Handling

### Error Categories

1. **Validation Errors (400 Bad Request)**
   - Invalid phone format
   - Missing required fields
   - Invalid file format
   - Limit/offset out of range

2. **Authentication Errors (401 Unauthorized)**
   - Missing JWT token
   - Expired JWT token
   - Invalid JWT signature

3. **Authorization Errors (403 Forbidden)**
   - Insufficient permissions for resource
   - Role not allowed for action
   - Access to different university/branch

4. **Not Found Errors (404 Not Found)**
   - Employee not found
   - Chat not found
   - University not found

5. **Conflict Errors (409 Conflict)**
   - Duplicate INN+KPP
   - Cannot remove last administrator
   - Employee already has role

6. **External Service Errors (502 Bad Gateway)**
   - MAX API unavailable
   - Auth Service unavailable
   - gRPC call failed after retries

7. **Internal Errors (500 Internal Server Error)**
   - Database connection failed
   - Transaction rollback
   - Unexpected error

### Error Response Format

All errors should follow consistent JSON format:

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid phone format",
    "details": {
      "field": "phone",
      "value": "123",
      "expected": "E.164 format (+7XXXXXXXXXX)"
    }
  }
}
```

### Retry Strategy

For transient errors (network, timeout):
- Retry up to 3 times
- Exponential backoff: 1s, 2s, 4s
- Log each retry attempt
- Return error after final failure

### Circuit Breaker

For external services (MAX API, gRPC):
- Open circuit after 5 consecutive failures
- Half-open after 30 seconds
- Close circuit after 2 successful requests
- Return cached data or graceful degradation when open

## Testing Strategy

### Unit Testing

**Framework:** Go testing package with testify assertions

**Coverage targets:**
- Domain logic: 90%+
- Use cases: 85%+
- Infrastructure: 70%+

**Key areas:**
- Phone normalization logic
- Permission validation rules
- Role hierarchy checks
- Excel parsing logic
- Pagination calculations

**Example unit tests:**
- Test phone normalization with various formats
- Test ABAC permission rules with different role combinations
- Test pagination edge cases (offset > total, limit > max)
- Test Excel parser with valid and invalid data

### Property-Based Testing

**Framework:** gopter (Go property testing library)

**Configuration:**
- Minimum 100 iterations per property
- Use seed for reproducibility
- Generate realistic test data

**Generators:**
- `PhoneGenerator`: Generate valid and invalid phone formats
- `RoleGenerator`: Generate role combinations with contexts
- `EmployeeGenerator`: Generate employee records with all fields
- `ChatGenerator`: Generate chat records with administrators
- `StructureGenerator`: Generate university hierarchies

**Property tests:**

Each property-based test must be tagged with a comment referencing the correctness property:

```go
// Feature: digital-university-mvp-completion, Property 9: Phone numbers are normalized to E.164
func TestPhoneNormalization(t *testing.T) {
    properties := gopter.NewProperties(nil)
    properties.Property("all phones normalize to E.164", prop.ForAll(
        func(phone string) bool {
            normalized := NormalizePhone(phone)
            return strings.HasPrefix(normalized, "+7") && len(normalized) == 12
        },
        gen.RegexMatch(`^[89]\d{10}$`),
    ))
    properties.TestingRun(t, gopter.ConsoleReporter(false))
}
```

**Key property tests:**
- Property 9: Phone normalization (100+ random phone formats)
- Property 2: ABAC permission validation (100+ role/resource combinations)
- Property 16: Curator filtering (100+ curator/chat combinations)
- Property 21: Last admin protection (100+ admin removal scenarios)
- Property 44: Pagination limit capping (100+ limit values)

### Integration Testing

**Framework:** Go testing with Docker containers (testcontainers-go)

**Test scenarios:**
- End-to-end employee creation with MAX_id lookup
- Chat filtering with role-based access control
- Excel import with full structure creation
- Migration from all three sources
- gRPC communication between services

**Setup:**
- Spin up PostgreSQL containers for each service
- Mock MAX API responses
- Use real gRPC communication

### Migration Testing

**Test data:**
- Sample database with 100 chats (admin_panel source)
- Sample Google Sheet with 50 chats (bot_registrar source)
- Sample Excel with 200 rows (academic_group source)

**Validation:**
- Verify all chats imported with correct source
- Verify universities created/reused correctly
- Verify structure hierarchy created correctly
- Verify administrators linked to chats
- Verify error handling for invalid data

### Performance Testing

**Load testing:**
- 1000 concurrent chat list requests
- Batch MAX_id update for 10,000 employees
- Excel import with 10,000 rows
- Search with 100,000 chats in database

**Targets:**
- Chat list: < 200ms p95
- Search: < 300ms p95
- Batch update: < 5 minutes for 10,000 employees
- Excel import: < 10 minutes for 10,000 rows

## Deployment Strategy

### Database Migrations

**Tool:** golang-migrate

**Process:**
1. Run migrations in order (001, 002, 003...)
2. Each migration is transactional
3. Rollback on failure
4. Version tracking in schema_migrations table

**Migration order:**
1. Auth Service: Add roles and user_roles tables
2. Employee Service: Add role, user_id, max_id_updated_at columns
3. Structure Service: Add department_managers table
4. Migration Service: Add migration_jobs and migration_errors tables

### Service Deployment

**Strategy:** Blue-Green deployment

**Process:**
1. Deploy new version to green environment
2. Run health checks
3. Switch traffic to green
4. Keep blue for rollback
5. Decommission blue after 24 hours

**Health checks:**
- Database connectivity
- gRPC service availability
- MAX API connectivity (for MaxBot Service)

### Data Migration

**Phase 1: Database migration (6,000 chats)**
- Run during maintenance window
- Estimated time: 10 minutes
- Rollback: Delete imported chats by source='admin_panel'

**Phase 2: Google Sheets migration (TBD chats)**
- Run as background job
- Monitor progress via API
- Retry failed rows

**Phase 3: Excel migration (155,000+ chats)**
- Split into batches of 10,000 rows
- Run batches sequentially
- Estimated time: 2-3 hours
- Checkpoint after each batch

### Monitoring

**Metrics:**
- Request rate per endpoint
- Response time percentiles (p50, p95, p99)
- Error rate by error type
- gRPC call success/failure rate
- Migration progress (processed/total)
- Database connection pool usage

**Alerts:**
- Error rate > 5%
- Response time p95 > 1s
- gRPC failure rate > 10%
- Database connection pool > 80%
- Migration stalled (no progress for 5 minutes)

**Logging:**
- Structured JSON logs
- Log levels: DEBUG, INFO, WARN, ERROR
- Include request_id for tracing
- Log all gRPC calls with duration
- Log all migration operations

## Security Considerations

### Authentication

- JWT tokens with RS256 signing
- Access token expiry: 15 minutes
- Refresh token expiry: 7 days
- Token rotation on refresh

### Authorization

- ABAC with role, university, branch, faculty context
- Validate permissions on every request
- No client-side permission checks
- Audit log for permission changes

### Data Protection

- Encrypt sensitive data at rest (phone numbers, MAX_id)
- Use TLS for all gRPC communication
- Sanitize all user inputs
- Parameterized SQL queries (no string concatenation)

### Rate Limiting

- 100 requests per minute per user
- 1000 requests per minute per IP
- Separate limits for migration endpoints (10 per hour)

### Input Validation

- Validate all inputs against schema
- Sanitize file uploads (Excel, CSV)
- Limit file size to 50MB
- Validate phone format before MAX API call
- Validate INN/KPP format

## Future Enhancements

### Phase 2 (Post-MVP)

1. **Operator Role Implementation**
   - Complete operator role with branch/faculty filtering
   - Operator assignment UI
   - Operator permission management

2. **Advanced ABAC**
   - Policy-based access control
   - Dynamic permission rules
   - Permission inheritance

3. **Analytics and Reporting**
   - Chat activity metrics
   - Employee engagement reports
   - Migration audit reports

4. **Caching Layer**
   - Redis for frequently accessed data
   - Cache invalidation strategy
   - Distributed cache for multi-instance deployment

5. **Event-Driven Architecture**
   - Message queue (RabbitMQ/Kafka)
   - Async processing for migrations
   - Event sourcing for audit trail

6. **API Gateway**
   - Kong or Nginx
   - Centralized authentication
   - Rate limiting
   - Request routing

7. **Frontend WebApp**
   - React-based mini-application
   - Integration with MAX SDK
   - Real-time updates via WebSocket
