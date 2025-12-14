# E2E Tests Architecture Diagram

## Test Flow Visualization

```
┌─────────────────────────────────────────────────────────────────┐
│                    E2E Test Architecture                         │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│                      Test Execution Flow                         │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  1. Start Services                                               │
│     ├─ docker-compose up -d                                      │
│     └─ Wait for health checks                                    │
│                                                                  │
│  2. Run E2E Tests                                                │
│     ├─ Complete User Journey                                     │
│     ├─ Role-Based Access Control                                 │
│     ├─ Chat Administrator Management                             │
│     ├─ Pagination and Search                                     │
│     ├─ Error Handling                                            │
│     ├─ Concurrent Operations                                     │
│     └─ Data Consistency                                          │
│                                                                  │
│  3. Verify Results                                               │
│     ├─ Check HTTP status codes                                   │
│     ├─ Validate response data                                    │
│     └─ Verify data consistency                                   │
│                                                                  │
│  4. Cleanup (optional)                                           │
│     └─ docker-compose down -v                                    │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

## Complete User Journey Flow

```
┌─────────────────────────────────────────────────────────────────┐
│              TestE2E_CompleteUserJourney                         │
└─────────────────────────────────────────────────────────────────┘

Step 1: Register Superadmin
    │
    ▼
┌──────────────┐
│ Auth Service │ POST /register
│   :8080      │ ─────────────► User created
└──────────────┘                Token returned
    │
    ▼
Step 2: Create University
    │
    ▼
┌──────────────┐
│  Structure   │ POST /universities
│   Service    │ ─────────────► University created
│   :8083      │                ID: 1
└──────────────┘
    │
    ▼
Step 3: Register Curator
    │
    ▼
┌──────────────┐
│ Auth Service │ POST /register
│   :8080      │ ─────────────► Curator created
└──────────────┘                Token returned
    │
    ▼
Step 4: Create Employee
    │
    ▼
┌──────────────┐
│  Employee    │ POST /employees
│   Service    │ ─────────────► Employee created
│   :8081      │                ID: 1, Phone: +79001234567
└──────────────┘
    │
    ▼
Step 5: Create Chat
    │
    ▼
┌──────────────┐
│ Chat Service │ POST /chats
│   :8082      │ ─────────────► Chat created
└──────────────┘                ID: 1, University: 1
    │
    ▼
Step 6: Add Administrator
    │
    ▼
┌──────────────┐
│ Chat Service │ POST /chats/1/administrators
│   :8082      │ ─────────────► Admin added
└──────────────┘                Phone: +79001234567
    │
    ▼
Step 7-8: Search Chats
    │
    ├─► Superadmin: GET /chats?query=Математика
    │   └─► Sees all chats (1 chat)
    │
    └─► Curator: GET /chats?query=Математика
        └─► Sees only university chats (0-1 chats)
    │
    ▼
Step 9: Get Employee
    │
    ▼
┌──────────────┐
│  Employee    │ GET /employees/1
│   Service    │ ─────────────► Employee details
│   :8081      │                Name, Phone, etc.
└──────────────┘
    │
    ▼
Step 10: Get Structure
    │
    ▼
┌──────────────┐
│  Structure   │ GET /universities/1/structure
│   Service    │ ─────────────► Full hierarchy
│   :8083      │                University → Branches → ...
└──────────────┘
    │
    ▼
✅ Test PASSED
```

## Role-Based Access Control Flow

```
┌─────────────────────────────────────────────────────────────────┐
│           TestE2E_RoleBasedAccessControl                         │
└─────────────────────────────────────────────────────────────────┘

Create Users:
┌──────────────┐  ┌──────────────┐  ┌──────────────┐
│  Superadmin  │  │   Curator    │  │   Operator   │
│    Token     │  │    Token     │  │    Token     │
└──────────────┘  └──────────────┘  └──────────────┘
       │                 │                 │
       ▼                 ▼                 ▼
Create Test Data:
┌──────────────────────────────────────────────────┐
│  University (ID: 1)                              │
│    └─ Chat (ID: 1)                               │
└──────────────────────────────────────────────────┘
       │
       ▼
Test Access:
       │
       ├─► Superadmin: GET /chats
       │   └─► ✅ Sees ALL chats
       │
       ├─► Curator: GET /chats
       │   └─► ✅ Sees only University 1 chats
       │
       ├─► Operator: GET /chats
       │   └─► ✅ Sees only Department chats
       │
       └─► No Auth: GET /chats
           └─► ❌ 401 Unauthorized
```

## Error Handling Test Matrix

```
┌─────────────────────────────────────────────────────────────────┐
│                TestE2E_ErrorHandling                             │
└─────────────────────────────────────────────────────────────────┘

┌──────────────────────┬──────────────┬─────────────────────────┐
│ Test Case            │ Expected     │ Endpoint                │
├──────────────────────┼──────────────┼─────────────────────────┤
│ Invalid JSON         │ 400          │ POST /employees         │
│ Missing phone        │ 400          │ POST /employees         │
│ Missing name         │ 400          │ POST /employees         │
│ Invalid ID           │ 400          │ GET /employees/invalid  │
│ Not found            │ 404          │ GET /employees/999999   │
│ Duplicate admin      │ 409          │ POST /chats/1/admins    │
│ Missing auth         │ 401          │ GET /employees          │
│ Invalid token        │ 401          │ GET /employees          │
└──────────────────────┴──────────────┴─────────────────────────┘

Test Flow:
    │
    ├─► Test 1: Invalid JSON
    │   └─► POST with "invalid json" → 400 ✅
    │
    ├─► Test 2: Missing Fields
    │   ├─► POST without phone → 400 ✅
    │   └─► POST without name → 400 ✅
    │
    ├─► Test 3: Invalid IDs
    │   ├─► GET /employees/invalid → 400 ✅
    │   ├─► GET /chats/invalid → 400 ✅
    │   └─► GET /universities/invalid → 400 ✅
    │
    ├─► Test 4: Not Found
    │   ├─► GET /employees/999999 → 404 ✅
    │   ├─► GET /chats/999999 → 404 ✅
    │   └─► GET /universities/999999 → 404 ✅
    │
    ├─► Test 5: Duplicates
    │   └─► Add same admin twice → 409 ✅
    │
    └─► Test 6: Authorization
        ├─► No header → 401 ✅
        ├─► Invalid format → 401 ✅
        └─► Invalid token → 401 ✅
```

## Concurrent Operations Flow

```
┌─────────────────────────────────────────────────────────────────┐
│            TestE2E_ConcurrentOperations                          │
└─────────────────────────────────────────────────────────────────┘

Setup:
┌──────────────────────────────────────────────────┐
│  Token: superadmin                               │
│  University ID: 1                                │
└──────────────────────────────────────────────────┘
       │
       ▼
Concurrent Chat Creation:
       │
       ├─► Goroutine 1: Create "Chat 0"
       ├─► Goroutine 2: Create "Chat 1"
       ├─► Goroutine 3: Create "Chat 2"
       ├─► Goroutine 4: Create "Chat 3"
       └─► Goroutine 5: Create "Chat 4"
       │
       ▼
Wait for completion:
       │
       ├─► Chat 0 created ✅
       ├─► Chat 1 created ✅
       ├─► Chat 2 created ✅
       ├─► Chat 3 created ✅
       └─► Chat 4 created ✅
       │
       ▼
Verify:
       │
       └─► No race conditions ✅
           └─► All chats created successfully ✅
```

## Service Interaction Map

```
┌─────────────────────────────────────────────────────────────────┐
│              E2E Tests Service Interactions                      │
└─────────────────────────────────────────────────────────────────┘

┌──────────────┐
│   E2E Test   │
│   Runner     │
└──────┬───────┘
       │
       ├──────────────────────────────────────────────┐
       │                                              │
       ▼                                              ▼
┌──────────────┐                              ┌──────────────┐
│ Auth Service │◄─────────────────────────────┤  Employee    │
│   :8080      │  Validate Token              │   Service    │
└──────┬───────┘                              │   :8081      │
       │                                      └──────┬───────┘
       │                                             │
       │                                             │
       ▼                                             ▼
┌──────────────┐                              ┌──────────────┐
│ Chat Service │◄─────────────────────────────┤  Structure   │
│   :8082      │  Get University Info         │   Service    │
└──────────────┘                              │   :8083      │
                                              └──────────────┘

Data Flow:
1. Test → Auth: Register/Login
2. Test → Structure: Create University
3. Test → Employee: Create Employee
4. Test → Chat: Create Chat
5. Chat → Auth: Validate Token
6. Chat → Structure: Get University
7. Test → All: Verify Data
```

## Test Coverage Matrix

```
┌─────────────────────────────────────────────────────────────────┐
│                    E2E Test Coverage                             │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  Service          │ Endpoints Tested │ Scenarios               │
│  ────────────────┼──────────────────┼─────────────────────────│
│  Auth Service    │ 3/5              │ Register, Login, Token  │
│  Employee Service│ 5/8              │ CRUD, Search, Batch     │
│  Chat Service    │ 6/6              │ CRUD, Search, Admins    │
│  Structure Service│ 4/6             │ CRUD, Hierarchy         │
│                                                                  │
│  Total Coverage: 18/25 endpoints (72%)                          │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘

Feature Coverage:
┌────────────────────────────────────────┐
│ ✅ User Registration & Authentication  │
│ ✅ Role-Based Access Control (RBAC)    │
│ ✅ CRUD Operations                     │
│ ✅ Search & Pagination                 │
│ ✅ Error Handling                      │
│ ✅ Concurrent Operations               │
│ ✅ Data Consistency                    │
│ ✅ Administrator Management            │
│ ⚠️  Batch Operations (partial)         │
│ ⚠️  Migration Service (not covered)    │
└────────────────────────────────────────┘
```

## Test Execution Timeline

```
Time: 0s ──────────────────────────────────────────────► 300s

├─ 0s:   Start services check
├─ 5s:   TestE2E_CompleteUserJourney starts
├─ 45s:  TestE2E_CompleteUserJourney completes ✅
├─ 46s:  TestE2E_RoleBasedAccessControl starts
├─ 76s:  TestE2E_RoleBasedAccessControl completes ✅
├─ 77s:  TestE2E_ChatAdministratorManagement starts
├─ 117s: TestE2E_ChatAdministratorManagement completes ✅
├─ 118s: TestE2E_PaginationAndSearch starts
├─ 148s: TestE2E_PaginationAndSearch completes ✅
├─ 149s: TestE2E_ErrorHandling starts
├─ 179s: TestE2E_ErrorHandling completes ✅
├─ 180s: TestE2E_ConcurrentOperations starts
├─ 210s: TestE2E_ConcurrentOperations completes ✅
├─ 211s: TestE2E_DataConsistency starts
└─ 241s: TestE2E_DataConsistency completes ✅

Total Time: ~4 minutes
```

## Quick Reference

### Run All E2E Tests
```bash
cd integration-tests && ./run_e2e_tests.sh
```

### Run Specific Test
```bash
go test -v -run TestE2E_CompleteUserJourney -timeout 5m
```

### Check Services
```bash
curl http://localhost:8080/health  # Auth
curl http://localhost:8081/employees/all  # Employee
curl http://localhost:8082/chats/all  # Chat
curl http://localhost:8083/universities  # Structure
```

### View Logs
```bash
docker-compose logs -f auth-service
docker-compose logs -f employee-service
docker-compose logs -f chat-service
docker-compose logs -f structure-service
```
