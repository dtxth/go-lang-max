# Quick Migration Guide

Fast reference for database migrations in the Digital University MVP project.

## Quick Commands

### Verify All Migrations
```bash
./verify_migrations.sh
```
✅ Checks that all migrations have rollback scripts

### Validate SQL Syntax
```bash
./validate_sql_syntax.sh
```
✅ Basic syntax validation for all SQL files

### Test Migrations (requires running databases)
```bash
./test_migrations.sh
```
✅ Tests up and down migrations with actual databases

## Migration Files Location

```
auth-service/migrations/
├── 001_init.sql
├── 001_init_down.sql
├── 002_add_role_to_users.sql
├── 002_add_role_to_users_down.sql
├── 003_add_roles_and_user_roles.sql
└── 003_add_roles_and_user_roles_down.sql

employee-service/migrations/
├── 001_init.sql
├── 001_init_down.sql
├── 002_add_role_to_employees.sql
├── 002_add_role_to_employees_down.sql
├── 003_add_batch_update_jobs.sql
└── 003_add_batch_update_jobs_down.sql

chat-service/migrations/
├── 001_init.sql
└── 001_init_down.sql

structure-service/migrations/
├── 001_init.sql
├── 001_init_down.sql
├── 002_add_department_managers.sql
├── 002_add_department_managers_down.sql
├── 003_add_chat_info_to_groups.sql
└── 003_add_chat_info_to_groups_down.sql

migration-service/migrations/
├── 001_init.sql
└── 001_init_down.sql
```

## Manual Migration Commands

### Auth Service
```bash
# Up
psql postgresql://postgres:postgres@localhost:5432/auth_db -f auth-service/migrations/001_init.sql
psql postgresql://postgres:postgres@localhost:5432/auth_db -f auth-service/migrations/002_add_role_to_users.sql
psql postgresql://postgres:postgres@localhost:5432/auth_db -f auth-service/migrations/003_add_roles_and_user_roles.sql

# Down (reverse order)
psql postgresql://postgres:postgres@localhost:5432/auth_db -f auth-service/migrations/003_add_roles_and_user_roles_down.sql
psql postgresql://postgres:postgres@localhost:5432/auth_db -f auth-service/migrations/002_add_role_to_users_down.sql
psql postgresql://postgres:postgres@localhost:5432/auth_db -f auth-service/migrations/001_init_down.sql
```

### Employee Service
```bash
# Up
psql postgresql://postgres:postgres@localhost:5433/employee_db -f employee-service/migrations/001_init.sql
psql postgresql://postgres:postgres@localhost:5433/employee_db -f employee-service/migrations/002_add_role_to_employees.sql
psql postgresql://postgres:postgres@localhost:5433/employee_db -f employee-service/migrations/003_add_batch_update_jobs.sql

# Down (reverse order)
psql postgresql://postgres:postgres@localhost:5433/employee_db -f employee-service/migrations/003_add_batch_update_jobs_down.sql
psql postgresql://postgres:postgres@localhost:5433/employee_db -f employee-service/migrations/002_add_role_to_employees_down.sql
psql postgresql://postgres:postgres@localhost:5433/employee_db -f employee-service/migrations/001_init_down.sql
```

### Chat Service
```bash
# Up
psql postgresql://postgres:postgres@localhost:5434/chat_db -f chat-service/migrations/001_init.sql

# Down
psql postgresql://postgres:postgres@localhost:5434/chat_db -f chat-service/migrations/001_init_down.sql
```

### Structure Service
```bash
# Up
psql postgresql://postgres:postgres@localhost:5435/structure_db -f structure-service/migrations/001_init.sql
psql postgresql://postgres:postgres@localhost:5435/structure_db -f structure-service/migrations/002_add_department_managers.sql
psql postgresql://postgres:postgres@localhost:5435/structure_db -f structure-service/migrations/003_add_chat_info_to_groups.sql

# Down (reverse order)
psql postgresql://postgres:postgres@localhost:5435/structure_db -f structure-service/migrations/003_add_chat_info_to_groups_down.sql
psql postgresql://postgres:postgres@localhost:5435/structure_db -f structure-service/migrations/002_add_department_managers_down.sql
psql postgresql://postgres:postgres@localhost:5435/structure_db -f structure-service/migrations/001_init_down.sql
```

### Migration Service
```bash
# Up
psql postgresql://postgres:postgres@localhost:5436/migration_db -f migration-service/migrations/001_init.sql

# Down
psql postgresql://postgres:postgres@localhost:5436/migration_db -f migration-service/migrations/001_init_down.sql
```

## Database Ports

| Service | Database | Port |
|---------|----------|------|
| auth-service | auth_db | 5432 |
| employee-service | employee_db | 5433 |
| chat-service | chat_db | 5434 |
| structure-service | structure_db | 5435 |
| migration-service | migration_db | 5436 |

## Key Tables Created

### Auth Service
- `users` - User accounts
- `refresh_tokens` - JWT refresh tokens
- `roles` - System roles (superadmin, curator, operator)
- `user_roles` - User role assignments with ABAC context

### Employee Service
- `universities` - University records
- `employees` - Employee records with role and MAX_id
- `batch_update_jobs` - Batch operation tracking

### Chat Service
- `universities` - University records
- `chats` - Chat records with source tracking
- `administrators` - Chat administrators

### Structure Service
- `universities` - University records
- `branches` - Branch/campus records
- `faculties` - Faculty/institute records
- `groups` - Academic group records
- `department_managers` - Operator assignments

### Migration Service
- `migration_jobs` - Migration job tracking
- `migration_errors` - Migration error logging

## Common Issues

### "relation already exists"
Migration was partially applied. Run down migration first.

### "column does not exist"
Previous migration didn't complete. Check migration order.

### "foreign key constraint violation"
Data exists that depends on the schema. Clean up data first.

## Documentation

- **MIGRATIONS.md** - Comprehensive migration guide
- **MIGRATION_SUMMARY.md** - Detailed migration breakdown
- **QUICK_MIGRATION_GUIDE.md** - This file

## Status

✅ 11 up migrations created  
✅ 11 down migrations created  
✅ All migrations verified  
✅ All SQL syntax validated  
✅ Ready for testing with live databases
