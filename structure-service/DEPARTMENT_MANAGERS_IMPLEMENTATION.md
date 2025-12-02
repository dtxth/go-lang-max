# Department Managers Implementation

## Overview

This document describes the implementation of department managers functionality in the Structure Service, which allows curators to assign operators to specific branches or faculties.

## Implementation Details

### 1. Database Migration

**File:** `migrations/002_add_department_managers.sql`

Created the `department_managers` table with the following structure:
- `id`: Primary key
- `employee_id`: Reference to employee in Employee Service
- `branch_id`: Optional reference to branch
- `faculty_id`: Optional reference to faculty
- `assigned_by`: User ID of curator who made the assignment
- `assigned_at`: Timestamp of assignment

**Constraints:**
- At least one of `branch_id` or `faculty_id` must be specified
- Unique constraint on (employee_id, branch_id, faculty_id)
- Cascade delete when branch or faculty is deleted

### 2. Domain Layer

**Files:**
- `internal/domain/department_manager.go`: Entity definition
- `internal/domain/department_manager_repository.go`: Repository interface
- `internal/domain/employee_service.go`: Employee Service interface for gRPC integration
- `internal/domain/errors.go`: Added new error types

**New Entities:**
- `DepartmentManager`: Represents operator assignment to department
- `Employee`: Represents employee from Employee Service

**New Errors:**
- `ErrDepartmentManagerNotFound`
- `ErrEmployeeNotFound`
- `ErrInvalidDepartment`

### 3. Repository Layer

**File:** `internal/infrastructure/repository/department_manager_postgres.go`

Implemented PostgreSQL repository with methods:
- `CreateDepartmentManager`: Create new assignment
- `GetDepartmentManagerByID`: Get assignment by ID
- `GetDepartmentManagersByEmployeeID`: Get all assignments for employee
- `GetDepartmentManagersByBranchID`: Get all operators for branch
- `GetDepartmentManagersByFacultyID`: Get all operators for faculty
- `GetAllDepartmentManagers`: Get all assignments
- `DeleteDepartmentManager`: Remove assignment

### 4. gRPC Integration

**Files:**
- `internal/infrastructure/employee/employee_client.go`: gRPC client for Employee Service
- `api/proto/employee/employee.proto`: Proto definitions (copied from employee-service)

**Added to Employee Service:**
- `GetEmployeeByID` gRPC method in proto definition
- Implementation in `employee-service/internal/infrastructure/grpc/employee_handler.go`

### 5. Use Case Layer

**File:** `internal/usecase/assign_operator_to_department.go`

Implemented `AssignOperatorToDepartmentUseCase` with the following logic:
1. Validate that at least one department (branch or faculty) is specified
2. Verify employee exists via Employee Service gRPC call
3. Verify employee has "operator" role
4. Create department_managers record

### 6. HTTP API

**Files:**
- `internal/infrastructure/http/handler.go`: Added handler methods
- `internal/infrastructure/http/router.go`: Added routes

**New Endpoints:**
- `POST /departments/managers`: Assign operator to department
- `DELETE /departments/managers/{id}`: Remove operator assignment
- `GET /departments/managers`: List all department managers

**Request Format (POST):**
```json
{
  "employee_id": 1,
  "branch_id": 1,      // optional
  "faculty_id": null,  // optional
  "assigned_by": 2     // optional
}
```

### 7. Configuration

**File:** `internal/config/config.go`

Added `EmployeeService` configuration field for gRPC address.

**Environment Variable:**
- `EMPLOYEE_SERVICE_GRPC`: Address of Employee Service gRPC (default: localhost:9091)

### 8. Main Application

**File:** `cmd/structure/main.go`

Updated initialization to:
- Create department manager repository
- Initialize Employee Service gRPC client
- Create assign operator use case
- Wire up dependencies to HTTP handler

## Testing

**File:** `internal/usecase/assign_operator_to_department_test.go`

Created unit tests covering:
- Successful assignment to branch
- Successful assignment to faculty
- Error when no department specified
- Error when employee not found
- Error when employee is not operator

All tests pass successfully.

## Requirements Validation

This implementation satisfies the following requirements:

**Requirement 11.1:** Curator can assign Operator to Branch
- ✅ Implemented via POST /departments/managers with branch_id

**Requirement 11.2:** Curator can assign Operator to Faculty
- ✅ Implemented via POST /departments/managers with faculty_id

**Requirement 11.3:** System verifies Operator employee exists
- ✅ Implemented via gRPC call to Employee Service in use case

**Requirement 11.4:** System returns list of assigned Branches and Faculties
- ✅ Implemented via GET /departments/managers endpoint

**Requirement 11.5:** Operator can be removed from department
- ✅ Implemented via DELETE /departments/managers/{id}

## Usage Example

### Assign Operator to Branch

```bash
curl -X POST http://localhost:8083/departments/managers \
  -H "Content-Type: application/json" \
  -d '{
    "employee_id": 1,
    "branch_id": 1,
    "assigned_by": 2
  }'
```

### Assign Operator to Faculty

```bash
curl -X POST http://localhost:8083/departments/managers \
  -H "Content-Type: application/json" \
  -d '{
    "employee_id": 1,
    "faculty_id": 1,
    "assigned_by": 2
  }'
```

### List All Department Managers

```bash
curl http://localhost:8083/departments/managers
```

### Remove Operator Assignment

```bash
curl -X DELETE http://localhost:8083/departments/managers/1
```

## Next Steps

The following optional property-based tests are defined but not implemented:
- 11.2: Property test for operator validation
- 11.3: Property test for operator permissions

These can be implemented later if comprehensive testing is required.
