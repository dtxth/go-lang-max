# Batch MAX_id Update Implementation

## Overview

This document describes the implementation of the batch MAX_id update feature for the Employee Service, which allows updating MAX_id for all employees who don't have one yet.

## Requirements Implemented

- **Requirement 4.1**: Query employees without MAX_id
- **Requirement 4.2**: Call MaxBot Service in batches of 100
- **Requirement 4.4**: Update employee records with received MAX_ids
- **Requirement 4.5**: Generate report with success/failure counts

## Components

### 1. Database Migration

**File**: `migrations/003_add_batch_update_jobs.sql`

Creates the `batch_update_jobs` table to track batch operations:
- `id`: Primary key
- `job_type`: Type of batch job (e.g., 'max_id_update')
- `status`: Job status ('running', 'completed', 'failed')
- `total`: Total number of records to process
- `processed`: Number of successfully processed records
- `failed`: Number of failed records
- `started_at`: Job start timestamp
- `completed_at`: Job completion timestamp

### 2. Domain Entities

**File**: `internal/domain/batch_update_job.go`

Defines:
- `BatchUpdateJob`: Represents a batch update operation
- `BatchUpdateResult`: Result of a batch update with success/failure counts
- `BatchUpdateJobRepository`: Interface for batch job operations

### 3. Repository Implementation

**File**: `internal/infrastructure/repository/batch_update_job_postgres.go`

Implements PostgreSQL repository for batch update jobs with CRUD operations.

**File**: `internal/infrastructure/repository/employee_postgres.go` (updated)

Added methods:
- `GetEmployeesWithoutMaxID(limit, offset)`: Retrieves employees without MAX_id
- `CountEmployeesWithoutMaxID()`: Counts employees without MAX_id

### 4. Use Case

**File**: `internal/usecase/batch_update_max_id.go`

Implements `BatchUpdateMaxIdUseCase` with methods:
- `StartBatchUpdate()`: Initiates batch update process
  - Counts total employees without MAX_id
  - Creates batch job record
  - Processes employees in batches of 100
  - Calls MaxBot Service for each batch
  - Updates employee records with received MAX_ids
  - Generates report with success/failure counts
- `GetBatchJobStatus(jobID)`: Retrieves status of a specific batch job
- `GetAllBatchJobs(limit, offset)`: Lists all batch jobs with pagination

### 5. MaxService Enhancement

**File**: `internal/domain/max_service.go` (updated)

Added `BatchGetMaxIDByPhone(phones []string)` method to the interface.

**File**: `internal/infrastructure/max/max_client.go` (updated)

Implemented `BatchGetMaxIDByPhone()` method:
- Limits batch size to 100 phones (Requirements 4.2)
- Currently calls `GetMaxIDByPhone()` for each phone individually
- Returns map of phone → MAX_id for successful lookups
- TODO: Will use `BatchGetUsersByPhone` gRPC method when available in MaxBot Service

### 6. HTTP Endpoints

**File**: `internal/infrastructure/http/handler.go` (updated)

Added handlers:
- `BatchUpdateMaxID()`: POST /employees/batch-update-maxid
  - Triggers batch update
  - Returns BatchUpdateResult with job_id, total, success, failed counts
- `GetBatchStatus()`: GET /employees/batch-status/{id}
  - Returns status of specific batch job
- `GetAllBatchJobs()`: GET /employees/batch-status
  - Lists all batch jobs with pagination

**File**: `internal/infrastructure/http/router.go` (updated)

Registered new routes for batch operations.

### 7. Dependency Injection

**File**: `cmd/employee/main.go` (updated)

Wired up:
- `BatchUpdateJobRepository`
- `BatchUpdateMaxIdUseCase`
- Updated handler initialization to include batch use case

## API Usage

### Trigger Batch Update

```bash
POST /employees/batch-update-maxid
```

Response:
```json
{
  "job_id": 1,
  "total": 150,
  "success": 145,
  "failed": 5,
  "errors": [
    "Failed to update employee 123: ...",
    "MaxBot service error: ..."
  ]
}
```

### Get Batch Job Status

```bash
GET /employees/batch-status/1
```

Response:
```json
{
  "id": 1,
  "job_type": "max_id_update",
  "status": "completed",
  "total": 150,
  "processed": 150,
  "failed": 5,
  "started_at": "2024-01-15T10:00:00Z",
  "completed_at": "2024-01-15T10:05:00Z"
}
```

### List All Batch Jobs

```bash
GET /employees/batch-status?limit=50&offset=0
```

Response:
```json
[
  {
    "id": 1,
    "job_type": "max_id_update",
    "status": "completed",
    "total": 150,
    "processed": 150,
    "failed": 5,
    "started_at": "2024-01-15T10:00:00Z",
    "completed_at": "2024-01-15T10:05:00Z"
  }
]
```

## Testing

**File**: `internal/usecase/batch_update_max_id_test.go`

Implemented unit tests:
- `TestBatchUpdateMaxId_EmptyDatabase`: Handles empty database
- `TestBatchUpdateMaxId_SuccessfulUpdate`: Successful batch update
- `TestBatchUpdateMaxId_PartialFailure`: Handles partial failures
- `TestBatchUpdateMaxId_BatchSizeLimit`: Processes large batches correctly

All tests pass successfully.

## Future Enhancements

1. **Async Processing**: Run batch updates in background goroutines
2. **Progress Tracking**: Real-time progress updates via WebSocket
3. **Retry Logic**: Automatic retry for failed MAX_id lookups
4. **MaxBot Batch Method**: Use `BatchGetUsersByPhone` gRPC method when implemented
5. **Scheduling**: Cron job for automatic periodic batch updates
6. **Notifications**: Email/notification when batch completes

## Notes

- Batch size is limited to 100 phones per request (Requirements 4.2, 4.3)
- Failed MAX_id lookups don't block the entire batch
- Job status is updated throughout the process
- Errors are logged and included in the final report
