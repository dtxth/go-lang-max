#!/bin/bash

# Integration Test Runner Script
# This script starts all services and runs integration tests

set -e

echo "=== Digital University Integration Tests ==="
echo ""

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo "Error: Docker is not running. Please start Docker and try again."
    exit 1
fi

# Navigate to project root
cd "$(dirname "$0")/.."

echo "Step 1: Starting all services with Docker Compose..."
docker-compose up -d

echo ""
echo "Step 2: Waiting for services to be healthy..."
sleep 30

# Check service health
echo "Checking Auth Service..."
curl -f http://localhost:8080/health || echo "Auth Service not ready"

echo "Checking Employee Service..."
curl -f http://localhost:8081/health || echo "Employee Service not ready"

echo "Checking Chat Service..."
curl -f http://localhost:8082/health || echo "Chat Service not ready"

echo "Checking Structure Service..."
curl -f http://localhost:8083/health || echo "Structure Service not ready"

echo "Checking Migration Service..."
curl -f http://localhost:8084/health || echo "Migration Service not ready"

echo ""
echo "Step 3: Running integration tests..."
cd integration-tests

# Run tests with verbose output
go test -v -timeout 10m ./... 2>&1 | tee test_results.log

TEST_EXIT_CODE=${PIPESTATUS[0]}

echo ""
echo "Step 4: Test Results"
if [ $TEST_EXIT_CODE -eq 0 ]; then
    echo "✅ All integration tests passed!"
else
    echo "❌ Some integration tests failed. Check test_results.log for details."
fi

echo ""
echo "Step 5: Cleanup (optional)"
read -p "Do you want to stop all services? (y/n) " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    cd ..
    docker-compose down
    echo "Services stopped."
fi

exit $TEST_EXIT_CODE
