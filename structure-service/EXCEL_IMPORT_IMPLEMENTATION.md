# Excel Import Implementation

## Overview

This document describes the implementation of Excel import functionality for the Structure Service, which allows importing university structure hierarchies from Excel files.

## Changes Made

### 1. Database Migration

**File:** `migrations/003_add_chat_info_to_groups.sql`

Added two new columns to the `groups` table:
- `chat_url TEXT` - URL of the chat associated with the group
- `chat_name TEXT` - Name of the chat associated with the group
- Added index on `chat_url` for faster lookups

### 2. Domain Model Updates

**File:** `internal/domain/structure.go`

- Updated `Group` entity to include `ChatURL` and `ChatName` fields
- Added `ImportResult` type to represent import operation results:
  - `Created` - count of created records
  - `Updated` - count of updated records
  - `Failed` - count of failed records
  - `Errors` - list of error messages

### 3. Repository Interface Extensions

**File:** `internal/domain/structure_repository.go`

Added new methods for finding existing entities:
- `GetUniversityByINNAndKPP(inn, kpp string)` - Find university by INN and KPP
- `GetBranchByUniversityAndName(universityID int64, name string)` - Find branch by university and name
- `GetFacultyByBranchAndName(branchID *int64, name string)` - Find faculty by branch and name
- `GetGroupByFacultyAndNumber(facultyID int64, course int, number string)` - Find group by faculty, course, and number

### 4. Repository Implementation

**File:** `internal/infrastructure/repository/structure_postgres.go`

- Implemented all new repository methods
- Updated CRUD operations for groups to handle `chat_url` and `chat_name` columns
- All methods properly handle NULL values for optional fields

### 5. Import Use Case

**File:** `internal/usecase/import_structure_from_excel.go`

Created `ImportStructureFromExcelUseCase` with the following features:

**Transaction Management:**
- All import operations are wrapped in a database transaction
- Rollback on any error to maintain data consistency
- Commit only when all rows are processed

**Duplicate Handling:**
- Universities: Found by INN+KPP or INN alone
- Branches: Found by university_id + name
- Faculties: Found by branch_id + name (or NULL branch_id)
- Groups: Found by faculty_id + course + number
- Existing records are updated if data has changed
- New records are created if not found

**Caching:**
- In-memory cache for universities, branches, and faculties within a single import
- Reduces database queries for repeated entities
- Cache is per-import operation

**Error Handling:**
- Validates required fields (INN, Organization, Faculty, GroupNumber)
- Collects all errors with row numbers
- Continues processing after individual row failures
- Returns detailed error report

**Result Tracking:**
- Counts created, updated, and failed records
- Provides detailed error messages for debugging

### 6. HTTP Handler Updates

**File:** `internal/infrastructure/http/handler.go`

Updated `ImportExcel` endpoint:
- Increased file size limit to 50MB (from 10MB)
- Added validation for empty Excel files
- Returns `ImportResult` with detailed statistics
- Proper error messages for all failure cases

### 7. Dependency Injection

**File:** `cmd/structure/main.go`

- Wired up `ImportStructureFromExcelUseCase` with repository and database connection
- Passed use case to HTTP handler

## API Endpoint

### POST /import/excel

Imports university structure from an Excel file.

**Request:**
- Content-Type: `multipart/form-data`
- Field: `file` (Excel file, .xlsx or .xls)
- Max file size: 50MB

**Response:**
```json
{
  "created": 150,
  "updated": 25,
  "failed": 5,
  "errors": [
    "row 10: missing required fields (INN, Organization, Faculty, GroupNumber)",
    "row 25: failed to create group: duplicate key"
  ]
}
```

**Status Codes:**
- 200 OK - Import completed (check result for details)
- 400 Bad Request - Invalid file format or validation error
- 500 Internal Server Error - Database or transaction error

## Excel File Format

The Excel file should contain the following columns (case-insensitive):

**Required:**
- INN (ИНН) - University tax ID
- Organization (Наименование организации) - University name
- Faculty (Факультет) - Faculty/institute name
- Group Number (Номер группы) - Group identifier

**Optional:**
- KPP (КПП) - Tax registration reason code
- FOIV (ФОИВ) - Federal executive body
- Branch (Филиал) - Branch/subdivision name
- Course (Курс обучения) - Course year
- Chat Name (Название чата) - Chat name
- Chat URL (Ссылка на чат) - Chat URL

## Features

### Transactional Import
All operations within a single import are wrapped in a transaction. If any critical error occurs, all changes are rolled back.

### Duplicate Detection
The system intelligently detects duplicates:
- Universities by INN+KPP combination
- Branches by university and name
- Faculties by branch (or NULL) and name
- Groups by faculty, course, and number

### Update vs Create
- If an entity exists, it's updated with new data
- If an entity doesn't exist, it's created
- Updates only occur if data has actually changed

### Error Resilience
- Individual row failures don't stop the entire import
- All errors are collected and reported
- Row numbers are included for easy debugging

### Performance
- In-memory caching reduces database queries
- Batch processing of related entities
- Efficient duplicate detection

## Requirements Satisfied

This implementation satisfies the following requirements from the specification:

- **12.1** - Validate file format and size limit (50MB)
- **12.2** - Parse Excel rows into structure entities
- **12.3** - Create University, Branch, Faculty, Group in transaction
- **12.4** - Handle duplicates by updating existing records
- **12.5** - Return summary with created/updated/failed counts

## Testing

To test the implementation:

1. Start the structure-service
2. Prepare an Excel file with the required columns
3. Send POST request to `/import/excel` with the file
4. Check the response for import statistics
5. Verify data in the database

Example using curl:
```bash
curl -X POST http://localhost:8083/import/excel \
  -F "file=@structure.xlsx" \
  -H "Content-Type: multipart/form-data"
```

## Future Enhancements

Potential improvements for future iterations:

1. **Async Processing** - Process large files in background
2. **Progress Tracking** - Real-time progress updates via WebSocket
3. **Validation Preview** - Validate file before import
4. **Rollback Support** - Manual rollback of completed imports
5. **Import History** - Track all import operations
6. **Partial Imports** - Resume failed imports from last successful row
