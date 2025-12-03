#!/bin/bash

# Migration Verification Script
# Verifies that all migrations have corresponding rollback scripts

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    local color=$1
    local message=$2
    echo -e "${color}${message}${NC}"
}

# Services to check
SERVICES=(
    "auth-service"
    "employee-service"
    "chat-service"
    "structure-service"
    "migration-service"
)

print_status "$YELLOW" "==================================="
print_status "$YELLOW" "  Migration Verification"
print_status "$YELLOW" "==================================="

total_migrations=0
missing_rollbacks=0
verified_migrations=0

for service in "${SERVICES[@]}"; do
    migrations_dir="$service/migrations"
    
    print_status "$YELLOW" "\n=== Checking $service ==="
    
    if [ ! -d "$migrations_dir" ]; then
        print_status "$RED" "  ✗ Migrations directory not found: $migrations_dir"
        continue
    fi
    
    # Get list of up migrations
    up_migrations=($(ls "$migrations_dir"/*.sql 2>/dev/null | grep -v "_down.sql" | sort))
    
    if [ ${#up_migrations[@]} -eq 0 ]; then
        print_status "$YELLOW" "  ⚠ No migrations found"
        continue
    fi
    
    print_status "$NC" "  Found ${#up_migrations[@]} migration(s)"
    
    # Check each migration
    for up_migration in "${up_migrations[@]}"; do
        migration_name=$(basename "$up_migration")
        down_migration="${up_migration%.sql}_down.sql"
        
        total_migrations=$((total_migrations + 1))
        
        if [ -f "$down_migration" ]; then
            print_status "$GREEN" "    ✓ $migration_name (rollback exists)"
            verified_migrations=$((verified_migrations + 1))
        else
            print_status "$RED" "    ✗ $migration_name (missing rollback)"
            missing_rollbacks=$((missing_rollbacks + 1))
        fi
    done
done

# Print summary
print_status "$YELLOW" "\n==================================="
print_status "$YELLOW" "  Verification Summary"
print_status "$YELLOW" "==================================="

print_status "$NC" "Total migrations: $total_migrations"
print_status "$GREEN" "Verified: $verified_migrations"

if [ $missing_rollbacks -gt 0 ]; then
    print_status "$RED" "Missing rollbacks: $missing_rollbacks"
    print_status "$RED" "\n✗ Some migrations are missing rollback scripts"
    exit 1
else
    print_status "$GREEN" "\n✓ All migrations have rollback scripts!"
    exit 0
fi
