# Test Coverage Diagram

## Архитектура тестирования

```
┌─────────────────────────────────────────────────────────────────┐
│                    Цифровой Вуз - Test Coverage                 │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│                      API Handler Tests (49)                      │
│                    ✅ Unit Tests for HTTP Layer                  │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐          │
│  │ Auth Service │  │   Employee   │  │ Chat Service │          │
│  │  11 tests    │  │   Service    │  │   8 tests    │          │
│  │              │  │  13 tests    │  │              │          │
│  └──────────────┘  └──────────────┘  └──────────────┘          │
│                                                                  │
│  ┌──────────────┐  ┌──────────────┐                            │
│  │  Structure   │  │  Migration   │                            │
│  │   Service    │  │   Service    │                            │
│  │   7 tests    │  │  10 tests    │                            │
│  └──────────────┘  └──────────────┘                            │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│                      Use Case Tests (20+)                        │
│                  ✅ Business Logic Unit Tests                    │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  • validate_permission_test.go                                  │
│  • create_employee_with_role_test.go                            │
│  • batch_update_max_id_test.go             
┌─────────────────────────────────────?ts_test.go                                          │
│  • add_administrator_wi│                 test.go              │
│  • assign_operator_to_department_test.go                        │
│  • normalize_phone_test.go                                       │
│  • batch_get_users_by_phone_test.go                             │
│                                                                  │
└─?───────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│                    Integration Tests (15+)                       │
│                  ✅ End-to-End Service Tests                     │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  • chat_integration_test.go                                      │
│    - Role-based filtering (Superadmin, Curator, Operator)       │
│    - Pagination and search                                       │
│    - Administrator management                                    │
│                                                                  │
│  • employee_integration_test.go                                  │
│    - Employee creation with roles                                │
│    - MAX_id integration                                          │
│                                                                  │
│  • structure_integration_test.go                                 │
│    - Excel import with full hierarchy                            │
│    - University structure retrieval                              │
│                                                                  │
│  • migration_integration_test.go                                 │
│    - Database migration (6,000 chats)                            │
│    - Google Sheets migration                                     │
│    - Excel migration (155,000+ chats)                            │
│                                                                  │
│  • grpc_integration_test.go                                      │
│    - Cross-service gRPC communication                            │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

## Test Coverage by Layer

```
┌─────────────────────────────────────────────────────────────────┐
│                         Test Pyramid                             │
└─────────────────────────────────────────────────────────────────┘

                          ▲
                         ╱ ╲
                        ╱   ╲
                       ╱  E2E ╲
                      ╱  Tests ╲
                     ╱───────────╲
                    ╱ Integration ╲
                   ╱     Tests     ╲
                  ╱─────────────────╲
                 ╱   API Handler     ╲
                ╱       Tests         ╲
               ╱─────────────────────────╲
              ╱      Use Case Tests       ╲
             ╱         (Unit Tests)        ╲
            ╱───────────────────────────────╲
           ╱                                 ╲
          ╱                                   ╲
         ╱─────────────────────────────────────╲

         Fast ←──────────────────────────→ Slow
         Many ←──────────────────────────→ Few
```

## Coverage Statistics

```
┌─────────────────────────────────────────────────────────────────┐
│                      Coverage by Service                         │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  Auth Service         ████████████████████ 11 tests             │
│  Employee Service     ██████████████████████ 13 tests           │
│  Chat Service         ████████████████ 8 tests                  │
│  Structure Service    ██████████████ 7 tests                    │
│  Migration Service    ████████████████████ 10 tests             │
│                                                                  │
│  Total: 49 API Handler Tests                                    │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│                      Test Types Distribution                     │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  Validation Tests     ████████████████████████ 45%              │
│  Error Handling       ████████████████████ 35%                  │
│  Authorization        ████████████ 15%                          │
│  HTTP Methods         ██████ 5%                                 │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

## Test Execution Flow

```
┌─────────────────────────────────────────────────────────────────┐
│                    ./test_api_handlers.sh                        │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
              ┌───────────────────────────────┐
              │  Run tests for each service   │
              └───────────────────────────────┘
                              │
        ┌─────────────────────┼─────────────────────┐
        │                     │                     │
        ▼                     ▼                     ▼
┌──────────────┐      ┌──────────────┐     ┌──────────────┐
│ auth-service │      │   employee   │     │ chat-service │
│              │      │   -service   │     │              │
│  ✅ PASS     │      │  ✅ PASS     │     │  ✅ PASS     │
└──────────────┘      └──────────────┘     └──────────────┘
        │                     │                     │
        └─────────────────────┼─────────────────────┘
                              │
        ┌─────────────────────┼─────────────────────┐
        │                     │                     │
        ▼                     ▼                     ▼
┌──────────────┐      ┌──────────────┐     ┌──────────────┐
│  structure   │      │  migration   │     │    Report    │
│  -service    │      │  -service    │     │   Results    │
│  ✅ PASS     │      │  ✅ PASS     │     │              │
└──────────────┘      └──────────────┘     └──────────────┘
```

## Test Categories

```
┌─────────────────────────────────────────────────────────────────┐
│                        Test Categories                           │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  1. Input Validation                                             │
│     ├─ Required fields (email, password, phone, etc.)           │
│     ├─ Data format (JSON, multipart/form-data)                  │
│     ├─ Data types (ID, strings, numbers)                        │
│     └─ Field constraints (length, pattern)                      │
│                                                                  │
│  2. Error Handling                                               │
│     ├─ Invalid JSON                                              │
│     ├─ Invalid IDs                                               │
│     ├─ Missing parameters                                        │
│     └─ Wrong HTTP methods                                        │
│                                                                  │
│  3. Authorization & Authentication                               │
│     ├─ Missing Authorization header                              │
│     ├─ Invalid token format                                      │
│     ├─ Expired tokens                                            │
│     └─ Insufficient permissions                                  │
│                                                                  │
│  4. File Upload                                                  │
│     ├─ Missing file                                              │
│     ├─ Invalid file format                                       │
│     └─ File size limits                                          │
│                                                                  │
│  5. Service Availability                                         │
│     ├─ Service not initialized                                   │
│     ├─ Dependency unavailable                                    │
│     └─ Graceful degradation                                      │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

## Quick Commands

```bash
# Run all API handler tests
./test_api_handlers.sh

# Run specific service tests
cd auth-service && go test -v ./internal/infrastructure/http/
cd employee-service && go test -v ./internal/infrastructure/http/
cd chat-service && go test -v ./internal/infrastructure/http/
cd structure-service && go test -v ./internal/infrastructure/http/
cd migration-service && go test -v ./internal/infrastructure/http/

# Run all tests (including integration)
./run_tests.sh

# Run integration tests only
cd integration-tests && go test -v
```

## Documentation

- [API_TESTS_COVERAGE.md](./API_TESTS_COVERAGE.md) - Detailed coverage
- [API_TESTING_SUMMARY.md](./API_TESTING_SUMMARY.md) - Work summary
- [API_TESTS_QUICK_REFERENCE.md](./API_TESTS_QUICK_REFERENCE.md) - Quick reference
- [README.md](./README.md) - Main documentation
