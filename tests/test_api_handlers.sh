#!/bin/bash

# Script to run all API handler tests

set -e

echo "========================================="
echo "Running API Handler Tests"
echo "========================================="
echo ""

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Track overall status
FAILED=0

# Test auth-service
echo "=== Testing auth-service ==="
cd auth-service
if go test -v ./internal/infrastructure/http/ > /tmp/auth-test.log 2>&1; then
    echo -e "${GREEN}✓ auth-service tests passed${NC}"
    grep "PASS" /tmp/auth-test.log | tail -1
else
    echo -e "${RED}✗ auth-service tests failed${NC}"
    cat /tmp/auth-test.log
    FAILED=1
fi
cd ..
echo ""

# Test employee-service
echo "=== Testing employee-service ==="
cd employee-service
if go test -v ./internal/infrastructure/http/ > /tmp/employee-test.log 2>&1; then
    echo -e "${GREEN}✓ employee-service tests passed${NC}"
    grep "PASS" /tmp/employee-test.log | tail -1
else
    echo -e "${RED}✗ employee-service tests failed${NC}"
    cat /tmp/employee-test.log
    FAILED=1
fi
cd ..
echo ""

# Test chat-service
echo "=== Testing chat-service ==="
cd chat-service
if go test -v ./internal/infrastructure/http/ > /tmp/chat-test.log 2>&1; then
    echo -e "${GREEN}✓ chat-service tests passed${NC}"
    grep "PASS" /tmp/chat-test.log | tail -1
else
    echo -e "${RED}✗ chat-service tests failed${NC}"
    cat /tmp/chat-test.log
    FAILED=1
fi
cd ..
echo ""

# Test structure-service
echo "=== Testing structure-service ==="
cd structure-service
if go test -v ./internal/infrastructure/http/ > /tmp/structure-test.log 2>&1; then
    echo -e "${GREEN}✓ structure-service tests passed${NC}"
    grep "PASS" /tmp/structure-test.log | tail -1
else
    echo -e "${RED}✗ structure-service tests failed${NC}"
    cat /tmp/structure-test.log
    FAILED=1
fi
cd ..
echo ""

# Test migration-service
echo "=== Testing migration-service ==="
cd migration-service
if go test -v ./internal/infrastructure/http/ > /tmp/migration-test.log 2>&1; then
    echo -e "${GREEN}✓ migration-service tests passed${NC}"
    grep "PASS" /tmp/migration-test.log | tail -1
else
    echo -e "${RED}✗ migration-service tests failed${NC}"
    cat /tmp/migration-test.log
    FAILED=1
fi
cd ..
echo ""

echo "========================================="
if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}All API handler tests passed!${NC}"
else
    echo -e "${RED}Some tests failed. Check logs above.${NC}"
    exit 1
fi
echo "========================================="
