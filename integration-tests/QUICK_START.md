# Integration Tests - Quick Start

## 🚀 Quick Start (5 minutes)

### 1. Start Services
```bash
# From project root
docker-compose up -d
```

### 2. Wait for Services
```bash
# Wait 30 seconds for all services to be ready
sleep 30
```

### 3. Run Tests
```bash
# From project root
cd integration-tests
go test -v ./...
```

## 📊 Test Results

Tests validate:
- ✅ Employee creation with role and MAX_id
- ✅ Chat filtering by role (Superadmin, Curator, Operator)
- ✅ Excel import with full structure hierarchy
- ✅ Migration from database, Google Sheets, and Excel
- ✅ gRPC communication between all services

## 🔧 Common Commands

```bash
# Run specific test suite
go test -v -run TestEmployee
go test -v -run TestChat
go test -v -run TestStructure
go test -v -run TestMigration
go test -v -run TestGRPC

# Run with timeout
go test -v -timeout 10m ./...

# Run with coverage
go test -v -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Using Make
make test              # Run all tests
make test-employee     # Employee tests only
make test-chat         # Chat tests only
make coverage          # Generate coverage report
```

## 🐛 Troubleshooting

### Services not ready?
```bash
# Check service health
curl http://localhost:8080/health  # Auth
curl http://localhost:8081/health  # Employee
curl http://localhost:8082/health  # Chat
curl http://localhost:8083/health  # Structure
curl http://localhost:8084/health  # Migration

# View logs
docker-compose logs -f auth-service
docker-compose logs -f employee-service
```

### Database issues?
```bash
# Check databases
docker ps | grep db

# Test connectivity
docker exec -it auth-db psql -U postgres -c "SELECT 1"
docker exec -it employee-db psql -U employee_user -d employee_db -c "SELECT 1"
```

### Tests failing?
```bash
# Restart services
docker-compose down
docker-compose up -d
sleep 30

# Clean databases
docker-compose down -v  # Remove volumes
docker-compose up -d
```

## 📝 Test Coverage Summary

| Test Suite | Tests | Coverage |
|------------|-------|----------|
| Employee   | 4     | Employee creation, role sync, deletion, batch update |
| Chat       | 6     | Filtering, administrators, pagination, search |
| Structure  | 4     | Excel import, hierarchy, managers, ordering |
| Migration  | 5     | Database, Google Sheets, Excel, tracking |
| gRPC       | 8     | All service connections, retries, pooling |

**Total: 27 integration tests**

## 🎯 What's Tested

### End-to-End Flows
1. **Employee → Auth → MaxBot**: Create employee with role and MAX_id lookup
2. **Chat → Auth**: Role-based filtering and permission checks
3. **Structure → Chat**: Excel import with chat linking
4. **Migration → Chat + Structure**: Multi-source data migration

### Service Integration
- HTTP REST APIs
- gRPC inter-service communication
- Database persistence
- Error handling and retries
- Pagination and search

### Business Logic
- ABAC role-based access control
- Administrator management
- Batch operations
- Data migration
- Structure hierarchy

## 📚 More Information

- Full guide: `INTEGRATION_TEST_GUIDE.md`
- Test helpers: `helpers.go`
- Requirements: `../.kiro/specs/digital-university-mvp-completion/requirements.md`
- Design: `../.kiro/specs/digital-university-mvp-completion/design.md`

## ✨ Success Criteria

All tests should pass with:
- ✅ No connection errors
- ✅ No database errors
- ✅ No timeout errors
- ✅ All assertions passing
- ✅ Clean test data cleanup

Expected output:
```
=== RUN   TestEmployeeCreationWithRoleAndMaxID
--- PASS: TestEmployeeCreationWithRoleAndMaxID (2.34s)
=== RUN   TestChatFilteringBySuperadmin
--- PASS: TestChatFilteringBySuperadmin (1.23s)
...
PASS
ok      github.com/digital-university/integration-tests    45.678s
```
