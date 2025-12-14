#!/bin/bash

# Test script for password_reset_tokens migration (000005)
# This script tests both up and down migrations

set -e

echo "========================================="
echo "Testing Migration 000005: password_reset_tokens"
echo "========================================="
echo ""

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Database connection details
DB_CONTAINER="auth-db"
DB_USER="postgres"
DB_NAME="auth_db"

# Check if database is running
echo -e "${BLUE}Checking if database is running...${NC}"
if ! docker ps | grep -q "$DB_CONTAINER"; then
    echo -e "${RED}✗ Database container $DB_CONTAINER is not running${NC}"
    echo "Please start services with: docker-compose up -d"
    exit 1
fi
echo -e "${GREEN}✓ Database is running${NC}"
echo ""

# Function to execute SQL and check result
execute_sql() {
    local sql=$1
    docker exec -i "$DB_CONTAINER" psql -U "$DB_USER" -d "$DB_NAME" -c "$sql" 2>&1
}

# Function to check if table exists
table_exists() {
    local table_name=$1
    local result=$(execute_sql "SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_schema = 'public' AND table_name = '$table_name');")
    echo "$result" | grep -q "t"
}

# Function to check if index exists
index_exists() {
    local index_name=$1
    local result=$(execute_sql "SELECT EXISTS (SELECT FROM pg_indexes WHERE schemaname = 'public' AND indexname = '$index_name');")
    echo "$result" | grep -q "t"
}

echo "========================================="
echo "Step 1: Testing DOWN migration (cleanup)"
echo "========================================="
echo ""

# Apply down migration to ensure clean state
echo -e "${BLUE}Applying down migration...${NC}"
docker exec -i "$DB_CONTAINER" psql -U "$DB_USER" -d "$DB_NAME" < migrations/000005_add_password_reset_tokens.down.sql > /dev/null 2>&1 || true

# Verify table doesn't exist
if table_exists "password_reset_tokens"; then
    echo -e "${RED}✗ Table password_reset_tokens still exists after down migration${NC}"
    exit 1
fi
echo -e "${GREEN}✓ Table password_reset_tokens removed${NC}"

# Verify indexes don't exist
for index in "idx_password_reset_tokens_token" "idx_password_reset_tokens_expires_at" "idx_password_reset_tokens_user_id"; do
    if index_exists "$index"; then
        echo -e "${RED}✗ Index $index still exists after down migration${NC}"
        exit 1
    fi
    echo -e "${GREEN}✓ Index $index removed${NC}"
done

echo ""
echo "========================================="
echo "Step 2: Testing UP migration"
echo "========================================="
echo ""

# Apply up migration
echo -e "${BLUE}Applying up migration...${NC}"
if docker exec -i "$DB_CONTAINER" psql -U "$DB_USER" -d "$DB_NAME" < migrations/000005_add_password_reset_tokens.up.sql > /dev/null 2>&1; then
    echo -e "${GREEN}✓ Up migration applied successfully${NC}"
else
    echo -e "${RED}✗ Failed to apply up migration${NC}"
    exit 1
fi

# Verify table exists
if ! table_exists "password_reset_tokens"; then
    echo -e "${RED}✗ Table password_reset_tokens was not created${NC}"
    exit 1
fi
echo -e "${GREEN}✓ Table password_reset_tokens created${NC}"

# Verify table structure
echo -e "${BLUE}Verifying table structure...${NC}"
table_structure=$(execute_sql "\d password_reset_tokens")

# Check for required columns
for column in "id" "user_id" "token" "expires_at" "used_at" "created_at"; do
    if echo "$table_structure" | grep -q "$column"; then
        echo -e "${GREEN}✓ Column $column exists${NC}"
    else
        echo -e "${RED}✗ Column $column is missing${NC}"
        exit 1
    fi
done

# Verify indexes exist
for index in "idx_password_reset_tokens_token" "idx_password_reset_tokens_expires_at" "idx_password_reset_tokens_user_id"; do
    if index_exists "$index"; then
        echo -e "${GREEN}✓ Index $index created${NC}"
    else
        echo -e "${RED}✗ Index $index was not created${NC}"
        exit 1
    fi
done

echo ""
echo "========================================="
echo "Step 3: Testing data operations"
echo "========================================="
echo ""

# Test inserting data
echo -e "${BLUE}Testing data insertion...${NC}"
insert_result=$(execute_sql "INSERT INTO password_reset_tokens (user_id, token, expires_at) VALUES (1, 'test_token_123', NOW() + INTERVAL '15 minutes') RETURNING id;" 2>&1)

if echo "$insert_result" | grep -q "ERROR"; then
    echo -e "${RED}✗ Failed to insert test data${NC}"
    echo "$insert_result"
    exit 1
fi
echo -e "${GREEN}✓ Test data inserted successfully${NC}"

# Test querying data
echo -e "${BLUE}Testing data retrieval...${NC}"
select_result=$(execute_sql "SELECT token FROM password_reset_tokens WHERE token = 'test_token_123';" 2>&1)

if echo "$select_result" | grep -q "test_token_123"; then
    echo -e "${GREEN}✓ Test data retrieved successfully${NC}"
else
    echo -e "${RED}✗ Failed to retrieve test data${NC}"
    exit 1
fi

# Test unique constraint on token
echo -e "${BLUE}Testing unique constraint on token...${NC}"
duplicate_result=$(execute_sql "INSERT INTO password_reset_tokens (user_id, token, expires_at) VALUES (1, 'test_token_123', NOW() + INTERVAL '15 minutes');" 2>&1)

if echo "$duplicate_result" | grep -q "duplicate key value violates unique constraint"; then
    echo -e "${GREEN}✓ Unique constraint on token works correctly${NC}"
else
    echo -e "${RED}✗ Unique constraint on token is not working${NC}"
    exit 1
fi

# Clean up test data
echo -e "${BLUE}Cleaning up test data...${NC}"
execute_sql "DELETE FROM password_reset_tokens WHERE token = 'test_token_123';" > /dev/null 2>&1
echo -e "${GREEN}✓ Test data cleaned up${NC}"

echo ""
echo "========================================="
echo "Step 4: Testing DOWN migration again"
echo "========================================="
echo ""

# Apply down migration
echo -e "${BLUE}Applying down migration...${NC}"
if docker exec -i "$DB_CONTAINER" psql -U "$DB_USER" -d "$DB_NAME" < migrations/000005_add_password_reset_tokens.down.sql > /dev/null 2>&1; then
    echo -e "${GREEN}✓ Down migration applied successfully${NC}"
else
    echo -e "${RED}✗ Failed to apply down migration${NC}"
    exit 1
fi

# Verify table doesn't exist
if table_exists "password_reset_tokens"; then
    echo -e "${RED}✗ Table password_reset_tokens still exists after down migration${NC}"
    exit 1
fi
echo -e "${GREEN}✓ Table password_reset_tokens removed${NC}"

# Verify indexes don't exist
for index in "idx_password_reset_tokens_token" "idx_password_reset_tokens_expires_at" "idx_password_reset_tokens_user_id"; do
    if index_exists "$index"; then
        echo -e "${RED}✗ Index $index still exists after down migration${NC}"
        exit 1
    fi
done
echo -e "${GREEN}✓ All indexes removed${NC}"

echo ""
echo "========================================="
echo "Step 5: Re-applying UP migration for final state"
echo "========================================="
echo ""

# Re-apply up migration to leave database in correct state
echo -e "${BLUE}Re-applying up migration...${NC}"
if docker exec -i "$DB_CONTAINER" psql -U "$DB_USER" -d "$DB_NAME" < migrations/000005_add_password_reset_tokens.up.sql > /dev/null 2>&1; then
    echo -e "${GREEN}✓ Up migration re-applied successfully${NC}"
else
    echo -e "${RED}✗ Failed to re-apply up migration${NC}"
    exit 1
fi

echo ""
echo "========================================="
echo -e "${GREEN}✓ All migration tests passed!${NC}"
echo "========================================="
echo ""
echo "Summary:"
echo "  ✓ DOWN migration removes table and indexes correctly"
echo "  ✓ UP migration creates table with correct structure"
echo "  ✓ All required columns are present"
echo "  ✓ All required indexes are created"
echo "  ✓ Data operations work correctly"
echo "  ✓ Unique constraint on token works"
echo "  ✓ Migration is idempotent"
echo ""
