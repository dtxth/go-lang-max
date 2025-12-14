#!/bin/bash

# Simple syntax validation for migration 000005
# This checks for common SQL syntax issues

set -e

echo "========================================="
echo "Validating Migration 000005 Syntax"
echo "========================================="
echo ""

GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m'

UP_FILE="migrations/000005_add_password_reset_tokens.up.sql"
DOWN_FILE="migrations/000005_add_password_reset_tokens.down.sql"

# Check files exist
if [ ! -f "$UP_FILE" ]; then
    echo -e "${RED}✗ UP migration file not found${NC}"
    exit 1
fi

if [ ! -f "$DOWN_FILE" ]; then
    echo -e "${RED}✗ DOWN migration file not found${NC}"
    exit 1
fi

echo -e "${GREEN}✓ Migration files exist${NC}"
echo ""

# Validate UP migration
echo -e "${BLUE}Validating UP migration...${NC}"

# Check for CREATE TABLE
if grep -q "CREATE TABLE.*password_reset_tokens" "$UP_FILE"; then
    echo -e "${GREEN}✓ CREATE TABLE statement found${NC}"
else
    echo -e "${RED}✗ CREATE TABLE statement missing${NC}"
    exit 1
fi

# Check for required columns
for column in "id" "user_id" "token" "expires_at" "used_at" "created_at"; do
    if grep -q "$column" "$UP_FILE"; then
        echo -e "${GREEN}✓ Column $column defined${NC}"
    else
        echo -e "${RED}✗ Column $column missing${NC}"
        exit 1
    fi
done

# Check for indexes
for index in "idx_password_reset_tokens_token" "idx_password_reset_tokens_expires_at" "idx_password_reset_tokens_user_id"; do
    if grep -q "$index" "$UP_FILE"; then
        echo -e "${GREEN}✓ Index $index defined${NC}"
    else
        echo -e "${RED}✗ Index $index missing${NC}"
        exit 1
    fi
done

# Check for UNIQUE constraint on token
if grep -q "token.*UNIQUE" "$UP_FILE"; then
    echo -e "${GREEN}✓ UNIQUE constraint on token${NC}"
else
    echo -e "${RED}✗ UNIQUE constraint on token missing${NC}"
    exit 1
fi

# Check for foreign key
if grep -q "REFERENCES users" "$UP_FILE"; then
    echo -e "${GREEN}✓ Foreign key to users table${NC}"
else
    echo -e "${RED}✗ Foreign key to users table missing${NC}"
    exit 1
fi

# Check for IF NOT EXISTS (idempotency)
if grep -q "IF NOT EXISTS" "$UP_FILE"; then
    echo -e "${GREEN}✓ Idempotent (IF NOT EXISTS)${NC}"
else
    echo -e "${RED}✗ Not idempotent (missing IF NOT EXISTS)${NC}"
    exit 1
fi

echo ""
echo -e "${BLUE}Validating DOWN migration...${NC}"

# Check for DROP TABLE
if grep -q "DROP TABLE.*password_reset_tokens" "$DOWN_FILE"; then
    echo -e "${GREEN}✓ DROP TABLE statement found${NC}"
else
    echo -e "${RED}✗ DROP TABLE statement missing${NC}"
    exit 1
fi

# Check for DROP INDEX statements
index_count=$(grep -c "DROP INDEX" "$DOWN_FILE")
if [ "$index_count" -eq 3 ]; then
    echo -e "${GREEN}✓ All 3 DROP INDEX statements found${NC}"
else
    echo -e "${RED}✗ Expected 3 DROP INDEX statements, found $index_count${NC}"
    exit 1
fi

# Check for IF EXISTS (idempotency)
if grep -q "IF EXISTS" "$DOWN_FILE"; then
    echo -e "${GREEN}✓ Idempotent (IF EXISTS)${NC}"
else
    echo -e "${RED}✗ Not idempotent (missing IF EXISTS)${NC}"
    exit 1
fi

# Check proper cleanup order (indexes before table)
table_line=$(grep -n "DROP TABLE" "$DOWN_FILE" | cut -d: -f1)
first_index_line=$(grep -n "DROP INDEX" "$DOWN_FILE" | head -1 | cut -d: -f1)

if [ "$first_index_line" -lt "$table_line" ]; then
    echo -e "${GREEN}✓ Proper cleanup order (indexes before table)${NC}"
else
    echo -e "${RED}✗ Improper cleanup order (table should be dropped after indexes)${NC}"
    exit 1
fi

echo ""
echo "========================================="
echo -e "${GREEN}✓ All syntax validations passed!${NC}"
echo "========================================="
echo ""
echo "Summary:"
echo "  ✓ Migration files exist"
echo "  ✓ All required columns present"
echo "  ✓ All required indexes present"
echo "  ✓ UNIQUE constraint on token"
echo "  ✓ Foreign key to users table"
echo "  ✓ Migrations are idempotent"
echo "  ✓ Proper cleanup order in DOWN migration"
echo ""
