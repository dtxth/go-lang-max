# Migration Summary

This document provides a comprehensive overview of all database migrations for the Digital University MVP project.

## Overview

- **Total Services**: 5
- **Total Migrations**: 11 (up migrations)
- **Total Rollback Scripts**: 11 (down migrations)
- **Total SQL Files**: 22

## Migration Status

✅ All migrations have corresponding rollback scripts  
✅ All SQL files pass basic syntax validation  
✅ All migrations are idempotent (use IF EXISTS/IF NOT EXISTS)

## Service-by-Service Breakdown

### 1. Auth Service (3 migrations)

#### 001_init.sql
**Purpose**: Initialize authentication database  
**Creates**:
- `users` table (id, email, password_hash, created_at)
- `refresh_tokens` table (id, jti, user_id, expires_at, revoked, created_at)
- Indexes for efficient querying

**Rollback**: `001_init_down.sql` - Drops all tables and indexes

---

#### 002_add_role_to_users.sql
**Purpose**: Add basic role support to users  
**Modifies**:
- Adds `role` column to users table (default: 'operator')
- Creates index on role column

**Rollback**: `002_add_role_to_users_down.sql` - Removes role column and index

**Requirements**: 1.1

---

#### 003_add_roles_and_user_roles.sql
**Purpose**: Implement ABAC role model  
**Creates**:
- `roles` table (id, name, description, created_at)
- `user_roles` table with ABAC context (user_id, role_id, university_id, branch_id, faculty_id)
- Inserts default roles: superadmin, curator, operator
- Indexes for efficient permission checking

**Rollback**: `003_add_roles_and_user_roles_down.sql` - Drops both tables and indexes

**Requirements**: 1.1, 1.2, 1.3, 1.4, 1.5

---

### 2. Employee Service (3 migrations)

#### 001_init.sql
**Purpose**: Initialize employee management database  
**Creates**:
- `universities` table (id, name, inn, kpp, created_at, updated_at)
- `employees` table (id, first_name, last_name, middle_name, phone, max_id, inn, kpp, university_id)
- Indexes for searching and filtering
- Auto-update triggers for updated_at columns

**Rollback**: `001_init_down.sql` - Drops all tables, indexes, triggers, and functions

---

#### 002_add_role_to_employees.sql
**Purpose**: Add role management to employees  
**Modifies**:
- Adds `role` column (curator, operator, or NULL)
- Adds `user_id` column (reference to auth-service)
- Adds `max_id_updated_at` column (timestamp of last MAX_id update)
- Creates indexes for new columns

**Rollback**: `002_add_role_to_employees_down.sql` - Removes columns and indexes

**Requirements**: 2.1, 2.2

---

#### 003_add_batch_update_jobs.sql
**Purpose**: Track batch MAX_id update operations  
**Creates**:
- `batch_update_jobs` table (id, job_type, status, total, processed, failed, started_at, completed_at)
- Indexes for querying job status and history

**Rollback**: `003_add_batch_update_jobs_down.sql` - Drops table and indexes

**Requirements**: 4.1, 4.2, 4.4, 4.5

---

### 3. Chat Service (1 migration)

#### 001_init.sql
**Purpose**: Initialize chat management database  
**Creates**:
- `universities` table (id, name, inn, kpp, created_at, updated_at)
- `chats` table (id, name, url, max_chat_id, participants_count, university_id, department, source)
- `administrators` table (id, chat_id, phone, max_id, created_at, updated_at)
- Full-text search index on chat names (Russian language)
- Auto-update triggers

**Rollback**: `001_init_down.sql` - Drops all tables, indexes, triggers, and functions

**Requirements**: 5.1, 5.2, 5.3, 6.1, 6.2, 7.3, 8.4, 9.4

---

### 4. Structure Service (3 migrations)

#### 001_init.sql
**Purpose**: Initialize university structure database  
**Creates**:
- `universities` table (id, name, inn, kpp, foiv, created_at, updated_at)
- `branches` table (id, university_id, name, created_at, updated_at)
- `faculties` table (id, branch_id, name, created_at, updated_at)
- `groups` table (id, faculty_id, course, number, chat_id, created_at, updated_at)
- Hierarchical relationships with foreign keys
- Auto-update triggers

**Rollback**: `001_init_down.sql` - Drops all tables, indexes, triggers, and functions

---

#### 002_add_department_managers.sql
**Purpose**: Link operators to departments  
**Creates**:
- `department_managers` table (id, employee_id, branch_id, faculty_id, assigned_by, assigned_at)
- Constraint: at least one of branch_id or faculty_id must be set
- Unique constraint on (employee_id, branch_id, faculty_id)
- Indexes for efficient querying

**Rollback**: `002_add_department_managers_down.sql` - Drops table and indexes

**Requirements**: 11.1, 11.2, 11.3

---

#### 003_add_chat_info_to_groups.sql
**Purpose**: Store chat information in groups  
**Modifies**:
- Adds `chat_url` column to groups table
- Adds `chat_name` column to groups table
- Creates index on chat_url

**Rollback**: `003_add_chat_info_to_groups_down.sql` - Removes columns and index

**Requirements**: 9.5, 10.1

---

### 5. Migration Service (1 migration)

#### 001_init.sql
**Purpose**: Initialize migration tracking database  
**Creates**:
- `migration_jobs` table (id, source_type, source_identifier, status, total, processed, failed, started_at, completed_at)
- `migration_errors` table (id, job_id, record_identifier, error_message, created_at)
- Indexes for querying jobs and errors

**Rollback**: `001_init_down.sql` - Drops both tables and indexes

**Requirements**: 7.5, 8.5, 20.1, 20.4

---

## Migration Execution Order

### Forward Migrations (Up)

Execute in this order:

1. **Auth Service**
   - 001_init.sql
   - 002_add_role_to_users.sql
   - 003_add_roles_and_user_roles.sql

2. **Employee Service**
   - 001_init.sql
   - 002_add_role_to_employees.sql
   - 003_add_batch_update_jobs.sql

3. **Chat Service**
   - 001_init.sql

4. **Structure Service**
   - 001_init.sql
   - 002_add_department_managers.sql
   - 003_add_chat_info_to_groups.sql

5. **Migration Service**
   - 001_init.sql

### Rollback Migrations (Down)

Execute in reverse order:

1. **Migration Service**
   - 001_init_down.sql

2. **Structure Service**
   - 003_add_chat_info_to_groups_down.sql
   - 002_add_department_managers_down.sql
   - 001_init_down.sql

3. **Chat Service**
   - 001_init_down.sql

4. **Employee Service**
   - 003_add_batch_update_jobs_down.sql
   - 002_add_role_to_employees_down.sql
   - 001_init_down.sql

5. **Auth Service**
   - 003_add_roles_and_user_roles_down.sql
   - 002_add_role_to_users_down.sql
   - 001_init_down.sql

## Key Features

### Idempotency
All migrations use `IF EXISTS` or `IF NOT EXISTS` clauses to ensure they can be run multiple times safely.

### Foreign Key Constraints
Proper foreign key relationships are established with appropriate `ON DELETE` actions:
- `CASCADE`: Delete dependent records
- `SET NULL`: Set reference to NULL

### Indexes
Strategic indexes are created for:
- Foreign keys (for join performance)
- Search fields (name, phone, INN)
- Filter fields (role, status, source)
- Full-text search (chat names in Russian)

### Triggers
Auto-update triggers maintain `updated_at` timestamps automatically.

### Constraints
- `UNIQUE` constraints prevent duplicates
- `CHECK` constraints enforce business rules
- `NOT NULL` constraints ensure data integrity

## Testing

### Verification Script
```bash
./verify_migrations.sh
```
Checks that all up migrations have corresponding down migrations.

### Syntax Validation
```bash
./validate_sql_syntax.sh
```
Performs basic SQL syntax validation on all migration files.

### Full Migration Test
```bash
./test_migrations.sh
```
Tests all migrations with rollback (requires running databases).

## Requirements Coverage

This migration set satisfies the following requirements:

- **1.1**: ABAC role model infrastructure
- **1.2-1.5**: Role-based access control
- **2.1-2.2**: Employee role management
- **4.1-4.5**: Batch MAX_id updates
- **5.1-5.3**: Role-based chat filtering
- **6.1-6.2**: Administrator management
- **7.3, 8.4, 9.4**: Migration source tracking
- **9.5, 10.1**: Group-chat linking
- **11.1-11.3**: Department manager assignments
- **20.1, 20.4**: Migration job tracking

## Next Steps

1. ✅ All migrations created
2. ✅ All rollback scripts created
3. ✅ Verification scripts created
4. ⏳ Test migrations with running databases
5. ⏳ Deploy to staging environment
6. ⏳ Deploy to production

## Notes

- All migrations are PostgreSQL-specific
- Russian language support is configured for full-text search
- Timestamps use `TIMESTAMP WITH TIME ZONE` for proper timezone handling
- All tables include `created_at` and most include `updated_at` columns
- Indexes are created with `IF NOT EXISTS` to prevent errors on re-runs
