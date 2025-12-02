#!/bin/bash

# Script to check migration status for all databases

echo "========================================="
echo "Checking Database Migration Status"
echo "========================================="
echo ""

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Auth Service
echo "=== Auth Service Database ==="
echo "Tables:"
docker-compose exec -T auth-db psql -U postgres -d auth_db -c "\dt" 2>/dev/null | grep -E "users|refresh_tokens|user_roles" || echo "No tables found"
echo ""

# Employee Service
echo "=== Employee Service Database ==="
echo "Tables:"
docker-compose exec -T employee-db psql -U employee_user -d employee_db -c "\dt" 2>/dev/null | grep -E "employees|universities|batch_update_jobs" || echo "No tables found"
echo ""

# Chat Service
echo "=== Chat Service Database ==="
echo "Tables:"
docker-compose exec -T chat-db psql -U chat_user -d chat_db -c "\dt" 2>/dev/null | grep -E "chats|administrators" || echo "No tables found"
echo ""

# Structure Service
echo "=== Structure Service Database ==="
echo "Tables:"
docker-compose exec -T structure-db psql -U postgres -d structure_db -c "\dt" 2>/dev/null | grep -E "universities|branches|faculties|groups|department_managers" || echo "No tables found"
echo ""

# Migration Service
echo "=== Migration Service Database ==="
echo "Tables:"
docker-compose exec -T migration-db psql -U postgres -d migration_db -c "\dt" 2>/dev/null | grep -E "migration_jobs|migration_errors" || echo "No tables found"
echo ""

echo "========================================="
echo "Check completed!"
echo "========================================="
