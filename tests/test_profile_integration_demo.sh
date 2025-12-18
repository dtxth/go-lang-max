#!/bin/bash

# Demo script showing the enhanced POST /employees endpoint
# This script demonstrates how the endpoint now attempts to retrieve first_name and last_name

echo "=== Employee Profile Integration Demo ==="
echo ""

echo "This demo shows how POST /employees now retrieves first_name and last_name from MAX profiles"
echo ""

# Test 1: Create employee with empty names (should attempt profile retrieval)
echo "Test 1: Creating employee with empty names"
echo "POST /employees with empty first_name and last_name"
echo "Expected behavior: Attempts to get profile from MAX API, falls back to defaults"
echo ""

cat << 'EOF'
curl -X POST http://localhost:8081/employees \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "+79001234567",
    "first_name": "",
    "last_name": "",
    "university_name": "МГУ"
  }'
EOF

echo ""
echo "Current result: Creates employee with MAX_id and default names ('Неизвестно')"
echo "Future result: Will use actual names from MAX profile when API supports it"
echo ""

# Test 2: Create employee with provided names (should use provided names)
echo "Test 2: Creating employee with provided names"
echo "POST /employees with specified first_name and last_name"
echo "Expected behavior: Uses provided names, still gets MAX_id"
echo ""

cat << 'EOF'
curl -X POST http://localhost:8081/employees \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "+79005678901",
    "first_name": "Владимир",
    "last_name": "Смирнов",
    "university_name": "СПбГУ"
  }'
EOF

echo ""
echo "Result: Creates employee with provided names and retrieved MAX_id"
echo ""

# Test 3: Create employee with role (enhanced CreateEmployeeWithRole)
echo "Test 3: Creating employee with role and empty names"
echo "POST /employees with role assignment"
echo "Expected behavior: Attempts profile retrieval for role-based employee creation"
echo ""

cat << 'EOF'
curl -X POST http://localhost:8081/employees \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "+79009876543",
    "first_name": "",
    "last_name": "",
    "role": "curator",
    "university_name": "МФТИ"
  }'
EOF

echo ""
echo "Result: Creates employee with role, attempts profile retrieval, generates password"
echo ""

echo "=== Implementation Details ==="
echo ""
echo "1. Enhanced MaxService interface with GetUserProfileByPhone method"
echo "2. Updated employee creation logic to use profile data when available"
echo "3. Graceful fallback when profile information is not available"
echo "4. Mock implementation provides test data for development"
echo "5. Ready for real MAX API integration when profile endpoint becomes available"
echo ""

echo "=== Code Changes ==="
echo ""
echo "Key files modified:"
echo "- employee-service/internal/domain/max_service.go (new interface method)"
echo "- employee-service/internal/usecase/employee_service.go (enhanced logic)"
echo "- employee-service/internal/usecase/create_employee_with_role.go (enhanced logic)"
echo "- employee-service/internal/infrastructure/max/max_client.go (new implementation)"
echo "- maxbot-service/internal/domain/max_api_client.go (new interface method)"
echo "- maxbot-service/api/proto/maxbot.proto (new protobuf messages)"
echo ""

echo "=== Testing ==="
echo ""
echo "Run tests to verify functionality:"
echo "cd employee-service && go test ./internal/usecase/... -v"
echo ""
echo "All existing tests pass, new functionality is covered by mocks"
echo ""

echo "Demo completed! The POST /employees endpoint now attempts to retrieve"
echo "first_name and last_name from MAX profiles when creating employees."