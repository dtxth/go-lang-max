#!/bin/bash

# Script to run end-to-end tests

set -e

echo "========================================="
echo "Running End-to-End Tests"
echo "========================================="
echo ""

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if services are running
echo "Checking if services are running..."
if ! curl -s http://localhost:8080/health > /dev/null 2>&1; then
    echo -e "${RED}✗ Auth service is not running${NC}"
    echo "Please start services with: docker-compose up -d"
    exit 1
fi

if ! curl -s http://localhost:8081/employees/all > /dev/null 2>&1; then
    echo -e "${RED}✗ Employee service is not running${NC}"
    echo "Please start services with: docker-compose up -d"
    exit 1
fi

if ! curl -s http://localhost:8082/chats/all > /dev/null 2>&1; then
    echo -e "${RED}✗ Chat service is not running${NC}"
    echo "Please start services with: docker-compose up -d"
    exit 1
fi

if ! curl -s http://localhost:8083/universities > /dev/null 2>&1; then
    echo -e "${RED}✗ Structure service is not running${NC}"
    echo "Please start services with: docker-compose up -d"
    exit 1
fi

echo -e "${GREEN}✓ All services are running${NC}"
echo ""

# Run E2E tests
echo "Running E2E tests..."
echo ""

# Test 1: Complete User Journey
echo "=== Test 1: Complete User Journey ==="
if go test -v -run TestE2E_CompleteUserJourney -timeout 5m; then
    echo -e "${GREEN}✓ Complete User Journey test passed${NC}"
else
    echo -e "${RED}✗ Complete User Journey test failed${NC}"
    exit 1
fi
echo ""

# Test 2: Role-Based Access Control
echo "=== Test 2: Role-Based Access Control ==="
if go test -v -run TestE2E_RoleBasedAccessControl -timeout 5m; then
    echo -e "${GREEN}✓ Role-Based Access Control test passed${NC}"
else
    echo -e "${RED}✗ Role-Based Access Control test failed${NC}"
    exit 1
fi
echo ""

# Test 3: Chat Administrator Management
echo "=== Test 3: Chat Administrator Management ==="
if go test -v -run TestE2E_ChatAdministratorManagement -timeout 5m; then
    echo -e "${GREEN}✓ Chat Administrator Management test passed${NC}"
else
    echo -e "${RED}✗ Chat Administrator Management test failed${NC}"
    exit 1
fi
echo ""

# Test 4: Pagination and Search
echo "=== Test 4: Pagination and Search ==="
if go test -v -run TestE2E_PaginationAndSearch -timeout 5m; then
    echo -e "${GREEN}✓ Pagination and Search test passed${NC}"
else
    echo -e "${RED}✗ Pagination and Search test failed${NC}"
    exit 1
fi
echo ""

# Test 5: Error Handling
echo "=== Test 5: Error Handling ==="
if go test -v -run TestE2E_ErrorHandling -timeout 5m; then
    echo -e "${GREEN}✓ Error Handling test passed${NC}"
else
    echo -e "${RED}✗ Error Handling test failed${NC}"
    exit 1
fi
echo ""

# Test 6: Concurrent Operations
echo "=== Test 6: Concurrent Operations ==="
if go test -v -run TestE2E_ConcurrentOperations -timeout 5m; then
    echo -e "${GREEN}✓ Concurrent Operations test passed${NC}"
else
    echo -e "${RED}✗ Concurrent Operations test failed${NC}"
    exit 1
fi
echo ""

# Test 7: Data Consistency
echo "=== Test 7: Data Consistency ==="
if go test -v -run TestE2E_DataConsistency -timeout 5m; then
    echo -e "${GREEN}✓ Data Consistency test passed${NC}"
else
    echo -e "${RED}✗ Data Consistency test failed${NC}"
    exit 1
fi
echo ""

echo "========================================="
echo -e "${GREEN}All E2E tests passed!${NC}"
echo "========================================="
