#!/bin/bash

# SQL Syntax Validation Script
# Validates SQL syntax without executing migrations

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

# Function to validate SQL syntax
validate_sql() {
    local sql_file=$1
    
    # Check if file exists
    if [ ! -f "$sql_file" ]; then
        return 1
    fi
    
    # Basic syntax checks
    local errors=0
    
    # Check for common SQL syntax issues
    if grep -q "CREATE TABLE.*(" "$sql_file"; then
        # Check for missing semicolons at end of statements
        if ! grep -q ";" "$sql_file"; then
            print_status "$RED" "      ⚠ Warning: No semicolons found"
            errors=$((errors + 1))
        fi
    fi
    
    # Check for balanced parentheses
    local open_parens=$(grep -o "(" "$sql_file" | wc -l)
    local close_parens=$(grep -o ")" "$sql_file" | wc -l)
    
    if [ "$open_parens" -ne "$close_parens" ]; then
        print_status "$RED" "      ✗ Unbalanced parentheses (open: $open_parens, close: $close_parens)"
        errors=$((errors + 1))
    fi
    
    return $errors
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
print_status "$YELLOW" "  SQL Syntax Validation"
print_status "$YELLOW" "==================================="

total_files=0
valid_files=0
invalid_files=0

for service in "${SERVICES[@]}"; do
    migrations_dir="$service/migrations"
    
    print_status "$YELLOW" "\n=== Validating $service ==="
    
    if [ ! -d "$migrations_dir" ]; then
        print_status "$RED" "  ✗ Migrations directory not found: $migrations_dir"
        continue
    fi
    
    # Get all SQL files
    sql_files=($(ls "$migrations_dir"/*.sql 2>/dev/null | sort))
    
    if [ ${#sql_files[@]} -eq 0 ]; then
        print_status "$YELLOW" "  ⚠ No SQL files found"
        continue
    fi
    
    print_status "$NC" "  Found ${#sql_files[@]} SQL file(s)"
    
    # Validate each file
    for sql_file in "${sql_files[@]}"; do
        file_name=$(basename "$sql_file")
        total_files=$((total_files + 1))
        
        if validate_sql "$sql_file"; then
            print_status "$GREEN" "    ✓ $file_name"
            valid_files=$((valid_files + 1))
        else
            print_status "$RED" "    ✗ $file_name (syntax issues detected)"
            invalid_files=$((invalid_files + 1))
        fi
    done
done

# Print summary
print_status "$YELLOW" "\n==================================="
print_status "$YELLOW" "  Validation Summary"
print_status "$YELLOW" "==================================="

print_status "$NC" "Total SQL files: $total_files"
print_status "$GREEN" "Valid: $valid_files"

if [ $invalid_files -gt 0 ]; then
    print_status "$RED" "Invalid: $invalid_files"
    print_status "$RED" "\n✗ Some SQL files have syntax issues"
    exit 1
else
    print_status "$GREEN" "\n✓ All SQL files passed basic validation!"
    print_status "$YELLOW" "\nNote: This is a basic syntax check. Run actual migrations to verify full correctness."
    exit 0
fi
