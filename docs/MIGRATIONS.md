# Database Migrations Guide

This document describes the database migration strategy for the Digital University MVP project.

## Overview

Each microservice has its own database and migration scripts located in the `migrations/` directory within each service folder.

## Migration Structure

### Services and Databases

| Service | Database | Port | Migrations Directory |
|---------|----------|------|---------------------|
| auth-service | auth_db | 5432 | auth-service/migrations/ |
| employee-service | employee_db | 5433 | employee-service/migrations/ |
| chat-service | chat_db | 5434 | chat-service/migrations/ |
| structure-service | structure_db | 5435 | structure-service/migrations/ |
| migration-service | migration_db | 5436 | migration-service/migrations/ |

### Migration Files

Each migration consists of two files:
- `XXX_migration_name.sql` - Forward migration (up)
- `XXX_migration_name_down.sql` - Rollback migration (down)

Migrations are numbered sequentially (001, 002, 003, etc.) and executed in order.

## Auth Service Migrations

### 001_init.sql
- Creates `users` table with email, password_hash
- Creates `refresh_tokens` table for JWT refresh tokens
- **Rollback**: Drops both tables and indexes

### 002_add_role_to_users.sql
- Adds `role` column to users table
- Creates index on role column
- **Rollback**: Removes role column and index

### 003_add_roles_and_user_roles.sql
- Creates `roles` table with predefined roles (superadmin, curator, operator)
- Creates `user_roles` table with ABAC context (university_id, branch_id, faculty_id)
- Inserts default roles
- **Rollback**: Drops both tables and all indexes
- **Validates**: Requirements 1.1, 1.2, 1.3, 1.4, 1.5

## Employee Service Migrations

### 001_init.sql
- Creates `universities` table with INN, KPP
- Creates `employees` table with personal info and phone
- Creates indexes for efficient querying
- Sets up auto-update triggers for updated_at columns
- **Rollback**: Drops all tables, indexes, triggers, and functions

### 002_add_role_to_employees.sql
- Adds `role`, `user_id`, `max_id_updated_at` columns to employees
- Creates indexes for new columns
- **Rollback**: Removes columns and indexes
- **Validates**: Requirements 2.1, 2.2

### 003_add_batch_update_jobs.sql
- Creates `batch_update_jobs` table for tracking batch MAX_id updates
- Tracks job status, progress, and errors
- **Rollback**: Drops table and indexes
- **Validates**: Requirements 4.1, 4.2, 4.4, 4.5

## Structure Service Migrations

### 001_init.sql
- Creates `universities` table
- Creates `branches` table (филиалы)
- Creates `faculties` table (факультеты)
- Creates `groups` table (академические группы)
- Sets up hierarchical relationships with foreign keys
- Creates auto-update triggers
- **Rollback**: Drops all tables, indexes, triggers, and functions

### 002_add_department_managers.sql
- Creates `department_managers` table
- Links operators (employees) to branches or faculties
- Enforces constraint that at least one of branch_id or faculty_id must be set
- **Rollback**: Drops table and indexes
- **Validates**: Requirements 11.1, 11.2, 11.3

### 003_add_chat_info_to_groups.sql
- Adds `chat_url` and `chat_name` columns to groups table
- Creates index on chat_url
- **Rollback**: Removes columns and index
- **Validates**: Requirements 9.5, 10.1

## Migration Service Migrations

### 001_init.sql
- Creates `migration_jobs` table for tracking migration operations
- Creates `migration_errors` table for detailed error logging
- Supports three source types: database, google_sheets, excel
- **Rollback**: Drops both tables and indexes
- **Validates**: Requirements 7.5, 8.5, 20.1, 20.4

## Running Migrations

### Manual Execution

To run migrations manually for a specific service:

```bash
# Forward migration
psql postgresql://postgres:postgres@localhost:5432/auth_db -f auth-service/migrations/001_init.sql

# Rollback migration
psql postgresql://postgres:postgres@localhost:5432/auth_db -f auth-service/migrations/001_init_down.sql
```

### Automated Testing

Use the provided test script to test all migrations with rollback:

```bash
./test_migrations.sh
```

This script will:
1. Check database connectivity for each service
2. Run each up migration
3. Run the corresponding down migration (rollback)
4. Re-run the up migration to restore state
5. Report success/failure for each service

### Using Docker Compose

Start all databases:

```bash
docker-compose up -d
```

Run migrations for all services:

```bash
for service in auth-service employee-service chat-service structure-service migration-service; do
    echo "Running migrations for $service..."
    # Add your migration tool command here
done
```

## Migration Best Practices

### Writing Migrations

1. **Always use IF EXISTS/IF NOT EXISTS** to make migrations idempotent
2. **Create down migrations** for every up migration
3. **Test rollback** before deploying to production
4. **Use transactions** where possible (PostgreSQL default)
5. **Add comments** to explain complex migrations
6. **Version control** all migration files

### Migration Order

Migrations must be run in numerical order:
1. 001_init.sql
2. 002_*.sql
3. 003_*.sql
...

### Rollback Order

Rollbacks must be run in reverse order:
1. 003_*_down.sql
2. 002_*_down.sql
3. 001_*_down.sql

### Data Safety

- **Backup database** before running migrations in production
- **Test migrations** in staging environment first
- **Use transactions** to ensure atomicity
- **Monitor migration progress** for large datasets
- **Have rollback plan** ready

## Troubleshooting

### Migration Fails

1. Check database connectivity
2. Verify migration file syntax
3. Check for conflicting schema changes
4. Review database logs
5. Run rollback if needed

### Rollback Fails

1. Check for data dependencies
2. Verify foreign key constraints
3. Manual cleanup may be required
4. Restore from backup if necessary

### Common Issues

**Issue**: "relation already exists"
- **Solution**: Migration was partially applied. Check schema and run down migration.

**Issue**: "column does not exist"
- **Solution**: Previous migration didn't complete. Check migration order.

**Issue**: "foreign key constraint violation"
- **Solution**: Data exists that depends on the schema. Clean up data first.

## Production Deployment

### Pre-deployment Checklist

- [ ] All migrations tested in staging
- [ ] Database backup created
- [ ] Rollback plan documented
- [ ] Maintenance window scheduled
- [ ] Team notified

### Deployment Process

1. **Backup**: Create full database backup
2. **Maintenance**: Put service in maintenance mode
3. **Migrate**: Run migrations in order
4. **Verify**: Check schema and data integrity
5. **Deploy**: Deploy new service version
6. **Test**: Run smoke tests
7. **Monitor**: Watch logs and metrics
8. **Restore**: If issues, rollback migrations and restore service

### Post-deployment

- Verify all services are healthy
- Check migration logs
- Monitor error rates
- Document any issues
- Update runbook if needed

## Migration Tools

### Recommended Tools

- **golang-migrate**: CLI tool for Go projects
- **Flyway**: Java-based migration tool
- **Liquibase**: Database-independent migrations
- **Custom scripts**: Shell scripts for simple cases

### Using golang-migrate

```bash
# Install
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Run migrations
migrate -path ./auth-service/migrations -database "postgresql://postgres:postgres@localhost:5432/auth_db?sslmode=disable" up

# Rollback
migrate -path ./auth-service/migrations -database "postgresql://postgres:postgres@localhost:5432/auth_db?sslmode=disable" down 1
```

## References

- [PostgreSQL Documentation](https://www.postgresql.org/docs/)
- [golang-migrate](https://github.com/golang-migrate/migrate)
- [Database Migration Best Practices](https://www.liquibase.org/get-started/best-practices)
