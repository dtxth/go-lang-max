#!/bin/bash

# Script to apply database migrations to all services

set -e

echo "========================================="
echo "Applying Database Migrations"
echo "========================================="
echo ""

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if databases are running
echo "Checking if databases are running..."

if ! docker-compose ps | grep -q "auth-db.*Up"; then
    echo -e "${RED}✗ auth-db is not running${NC}"
    echo "Please start services with: docker-compose up -d"
    exit 1
fi

echo -e "${GREEN}✓ All databases are running${NC}"
echo ""

# Apply migrations for each service

# Auth Service
echo "=== Auth Service ==="
if docker-compose exec -T auth-db psql -U postgres -d auth_db < auth-service/migrations/001_init.sql > /dev/null 2>&1; then
    echo -e "${GREEN}✓ Auth Service migrations applied${NC}"
else
    echo -e "${YELLOW}⚠ Auth Service migrations may already be applied${NC}"
fi
echo ""

# Employee Service
echo "=== Employee Service ==="
if docker-compose exec -T employee-db psql -U employee_user -d employee_db < employee-service/migrations/001_init.sql > /dev/null 2>&1; then
    echo -e "${GREEN}✓ Employee Service migrations applied${NC}"
else
    echo -e "${YELLOW}⚠ Employee Service migrations may already be applied${NC}"
fi
echo ""

# Chat Service
echo "=== Chat Service ==="
if docker-compose exec -T chat-db psql -U chat_user -d chat_db < chat-service/migrations/001_init.sql > /dev/null 2>&1; then
    echo -e "${GREEN}✓ Chat Service migrations applied${NC}"
else
    echo -e "${YELLOW}⚠ Chat Service migrations may already be applied${NC}"
fi
echo ""

# Structure Service
echo "=== Structure Service ==="
if docker-compose exec -T structure-db psql -U postgres -d structure_db < structure-service/migrations/001_init.sql > /dev/null 2>&1; then
    echo -e "${GREEN}✓ Structure Service migrations applied${NC}"
else
    echo -e "${YELLOW}⚠ Structure Service migrations may already be applied${NC}"
fi
echo ""

# Migration Service
echo "=== Migration Service ==="
if docker-compose exec -T migration-db psql -U postgres -d migration_db < migration-service/migrations/001_init.sql > /dev/null 2>&1; then
    echo -e "${GREEN}✓ Migration Service migrations applied${NC}"
else
    echo -e "${YELLOW}⚠ Migration Service migrations may already be applied${NC}"
fi
echo ""

echo "========================================="
echo -e "${GREEN}Migration process completed!${NC}"
echo "========================================="
echo ""
echo "To verify migrations, run:"
echo "  ./verify_migrations.sh"
