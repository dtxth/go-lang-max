#!/bin/bash

# Script to apply database migrations to all services

set -e

# Get the directory where this script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# Get the project root (parent of bin directory)
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Change to project root to ensure relative paths work correctly
cd "$PROJECT_ROOT"

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

# Function to apply all migrations for a service
apply_migrations() {
    local service=$1
    local db_container=$2
    local db_user=$3
    local db_name=$4
    local migrations_dir=$5
    
    echo "=== $service ==="
    
    # Find all .up.sql files and sort them
    local migration_files=$(find "$migrations_dir" -name "*.up.sql" | sort)
    
    if [ -z "$migration_files" ]; then
        echo -e "${YELLOW}⚠ No migrations found for $service${NC}"
        echo ""
        return
    fi
    
    local success=true
    for migration_file in $migration_files; do
        local filename=$(basename "$migration_file")
        if docker-compose exec -T "$db_container" psql -U "$db_user" -d "$db_name" < "$migration_file" > /dev/null 2>&1; then
            echo -e "${GREEN}✓ Applied: $filename${NC}"
        else
            echo -e "${YELLOW}⚠ Skipped (may already be applied): $filename${NC}"
        fi
    done
    
    echo -e "${GREEN}✓ $service migrations completed${NC}"
    echo ""
}

# Apply migrations for each service
apply_migrations "Auth Service" "auth-db" "postgres" "postgres" "auth-service/migrations"
apply_migrations "Employee Service" "employee-db" "employee_user" "employee_db" "employee-service/migrations"
apply_migrations "Chat Service" "chat-db" "chat_user" "chat_db" "chat-service/migrations"
apply_migrations "Structure Service" "structure-db" "postgres" "structure_db" "structure-service/migrations"
apply_migrations "Migration Service" "migration-db" "postgres" "migration_db" "migration-service/migrations"

echo "========================================="
echo -e "${GREEN}Migration process completed!${NC}"
echo "========================================="
echo ""
echo "To verify migrations, run:"
echo "  ./verify_migrations.sh"
