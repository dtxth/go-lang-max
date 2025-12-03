#!/bin/bash

# Migration Testing Script
# Tests all database migrations with rollback functionality

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Service configurations
declare -A SERVICES=(
    ["auth-service"]="postgresql://postgres:postgres@localhost:5432/auth_db"
    ["employee-service"]="postgresql://postgres:postgres@localhost:5433/employee_db"
    ["chat-service"]="postgresql://postgres:postgres@localhost:5434/chat_db"
    ["structure-service"]="postgresql://postgres:postgres@localhost:5435/structure_db"
    ["migration-service"]="postgresql://postgres:postgres@localhost:5436/migration_db"
)

# Function to print colored output
print_status() {
    local color=$1
    local message=$2
    echo -e "${color}${message}${NC}"
}

# Function to run SQL file
run_sql() {
    local db_url=$1
    local sql_file=$2
    
    if [ ! -f "$sql_file" ]; then
        print_status "$RED" "  ✗ File not found: $sql_file"
        return 1
    fi
    
    psql "$db_url" -f "$sql_file" > /dev/null 2>&1
    return $?
}

# Function to test migrations for a service
test_service_migrations() {
    local service=$1
    local db_url=${SERVICES[$service]}
    local migrations_dir="$service/migrations"
    
    print_status "$YELLOW" "\n=== Testing $service ==="
    
    if [ ! -d "$migrations_dir" ]; then
        print_status "$RED" "  ✗ Migrations directory not found: $migrations_dir"
        return 1
    fi
    
    # Get list of up migrations (sorted)
    local up_migrations=($(ls "$migrations_dir"/*.sql 2>/dev/null | grep -v "_down.sql" | sort))
    
    if [ ${#up_migrations[@]} -eq 0 ]; then
        print_status "$YELLOW" "  ⚠ No migrations found"
        return 0
    fi
    
    print_status "$GREEN" "  Found ${#up_migrations[@]} migration(s)"
    
    # Test each migration
    for up_migration in "${up_migrations[@]}"; do
        local migration_name=$(basename "$up_migration")
        local down_migration="${up_migration%.sql}_down.sql"
        
        print_status "$YELLOW" "\n  Testing: $migration_name"
        
        # Run up migration
        print_status "$NC" "    → Running up migration..."
        if run_sql "$db_url" "$up_migration"; then
            print_status "$GREEN" "    ✓ Up migration successful"
        else
            print_status "$RED" "    ✗ Up migration failed"
            return 1
        fi
        
        # Test down migration if it exists
        if [ -f "$down_migration" ]; then
            print_status "$NC" "    → Running down migration..."
            if run_sql "$db_url" "$down_migration"; then
                print_status "$GREEN" "    ✓ Down migration successful"
            else
                print_status "$RED" "    ✗ Down migration failed"
                return 1
            fi
            
            # Re-run up migration to restore state
            print_status "$NC" "    → Re-running up migration..."
            if run_sql "$db_url" "$up_migration"; then
                print_status "$GREEN" "    ✓ Re-applied up migration"
            else
                print_status "$RED" "    ✗ Failed to re-apply up migration"
                return 1
            fi
        else
            print_status "$YELLOW" "    ⚠ No down migration found: $(basename "$down_migration")"
        fi
    done
    
    print_status "$GREEN" "  ✓ All migrations passed for $service"
    return 0
}

# Function to check if database is accessible
check_database() {
    local service=$1
    local db_url=${SERVICES[$service]}
    
    if psql "$db_url" -c "SELECT 1" > /dev/null 2>&1; then
        return 0
    else
        return 1
    fi
}

# Main execution
main() {
    print_status "$YELLOW" "==================================="
    print_status "$YELLOW" "  Database Migration Test Suite"
    print_status "$YELLOW" "==================================="
    
    local failed_services=()
    local skipped_services=()
    
    for service in "${!SERVICES[@]}"; do
        # Check if database is accessible
        if ! check_database "$service"; then
            print_status "$YELLOW" "\n⚠ Skipping $service (database not accessible)"
            skipped_services+=("$service")
            continue
        fi
        
        # Test migrations
        if ! test_service_migrations "$service"; then
            failed_services+=("$service")
        fi
    done
    
    # Print summary
    print_status "$YELLOW" "\n==================================="
    print_status "$YELLOW" "  Test Summary"
    print_status "$YELLOW" "==================================="
    
    local total_services=${#SERVICES[@]}
    local tested_services=$((total_services - ${#skipped_services[@]}))
    local passed_services=$((tested_services - ${#failed_services[@]}))
    
    print_status "$NC" "Total services: $total_services"
    print_status "$NC" "Tested: $tested_services"
    print_status "$GREEN" "Passed: $passed_services"
    
    if [ ${#skipped_services[@]} -gt 0 ]; then
        print_status "$YELLOW" "Skipped: ${#skipped_services[@]} (${skipped_services[*]})"
    fi
    
    if [ ${#failed_services[@]} -gt 0 ]; then
        print_status "$RED" "Failed: ${#failed_services[@]} (${failed_services[*]})"
        print_status "$RED" "\n✗ Some migrations failed"
        exit 1
    else
        print_status "$GREEN" "\n✓ All migrations passed!"
        exit 0
    fi
}

# Run main function
main
