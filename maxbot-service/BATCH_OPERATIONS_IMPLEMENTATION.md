# MaxBot Service - Batch Operations Implementation

## Overview

This document describes the implementation of phone normalization and batch operations in the MaxBot Service, as specified in tasks 5 and 5.1 of the digital-university-mvp-completion spec.

## Implemented Features

### 1. Phone Normalization (Task 5.1)

**File:** `internal/usecase/normalize_phone.go`

The `NormalizePhoneUseCase` handles phone number normalization to E.164 format for Russian phone numbers.

**Features:**
- Handles Russian phone formats starting with 8, 9, +7, or 7
- Removes all non-digit characters (spaces, dashes, parentheses)
- Validates E.164 format (+7XXXXXXXXXX - 12 characters total)
- Returns `domain.ErrInvalidPhone` for invalid inputs

**Examples:**
- `89991234567` → `+79991234567`
- `9991234567` → `+79991234567`
- `+7 (999) 123-45-67` → `+79991234567`
- `8 (999) 123-45-67` → `+79991234567`

### 2. Batch MAX_id Lookup (Task 5)

**File:** `internal/usecase/batch_get_users_by_phone.go`

The `BatchGetUsersByPhoneUseCase` processes up to 100 phone numbers in a single batch request.

**Features:**
- Validates batch size (maximum 100 phones)
- Normalizes all phone numbers before lookup
- Preserves original phone format in response
- Returns array of `UserPhoneMapping` with phone, MAX_id, and found status
- Handles invalid phones gracefully (skips them)

**Response Structure:**
```go
type UserPhoneMapping struct {
    Phone string  // Original phone format
    MaxID string  // MAX_id if found, empty otherwise
    Found bool    // true if user exists in MAX Messenger
}
```

### 3. gRPC API Updates

**File:** `api/proto/maxbot.proto`

Added two new gRPC methods:

```protobuf
// NormalizePhone нормализует номер телефона в формат E.164
rpc NormalizePhone(NormalizePhoneRequest) returns (NormalizePhoneResponse);

// BatchGetUsersByPhone получает MAX ID для списка номеров телефонов (до 100)
rpc BatchGetUsersByPhone(BatchGetUsersByPhoneRequest) returns (BatchGetUsersByPhoneResponse);
```

### 4. Infrastructure Implementation

**File:** `internal/infrastructure/maxapi/client.go`

Added `BatchGetUsersByPhone` method to the MAX API client that:
- Validates batch size limit (100 phones)
- Normalizes all phone numbers
- Uses MAX API's `ListExist` method to check phone existence
- Maps results back to original phone formats
- Returns comprehensive mappings with found/not-found status

## Testing

### Unit Tests

**File:** `internal/usecase/normalize_phone_test.go`
- Tests all Russian phone format variations
- Tests invalid phone handling
- Tests E.164 format validation
- 10 test cases covering edge cases

**File:** `internal/usecase/batch_get_users_by_phone_test.go`
- Tests empty batch handling
- Tests batch size limit enforcement
- Tests mixed found/not-found scenarios
- Tests phone normalization in batch context
- Tests original phone preservation
- 6 test cases with comprehensive coverage

All tests pass successfully.

## Integration with Employee Service

The batch operations are designed to be used by the Employee Service for:
1. **Single employee creation:** Use `GetMaxIDByPhone` with normalized phone
2. **Batch MAX_id updates:** Use `BatchGetUsersByPhone` to process up to 100 employees at once

Example usage from Employee Service:
```go
// Batch update MAX_ids for employees without them
phones := []string{"89991234567", "89991234568", "89991234569"}
mappings, err := maxbotClient.BatchGetUsersByPhone(ctx, phones)

for _, mapping := range mappings {
    if mapping.Found {
        // Update employee record with mapping.MaxID
    } else {
        // Log that MAX_id was not found for mapping.Phone
    }
}
```

## Requirements Validation

✅ **Requirement 3.2:** Phone numbers are normalized to E.164 format
✅ **Requirement 4.3:** Batch requests respect 100 phone limit
✅ **Requirement 19.1:** Phone normalization to E.164
✅ **Requirement 19.2:** Handle phones starting with 8 (replace with +7)
✅ **Requirement 19.3:** Handle phones starting with 9 (prepend +7)
✅ **Requirement 19.4:** Remove non-digit characters before validation

## API Documentation

### NormalizePhone

**Request:**
```json
{
  "phone": "89991234567"
}
```

**Response:**
```json
{
  "normalized_phone": "+79991234567",
  "error_code": 0,
  "error": ""
}
```

### BatchGetUsersByPhone

**Request:**
```json
{
  "phones": [
    "89991234567",
    "+7 (999) 123-45-68",
    "9991234569"
  ]
}
```

**Response:**
```json
{
  "mappings": [
    {
      "phone": "89991234567",
      "max_id": "+79991234567",
      "found": true
    },
    {
      "phone": "+7 (999) 123-45-68",
      "max_id": "+79991234568",
      "found": true
    },
    {
      "phone": "9991234569",
      "max_id": "",
      "found": false
    }
  ],
  "error_code": 0,
  "error": ""
}
```

## Performance Considerations

- Batch processing reduces API calls by up to 100x compared to individual lookups
- Phone normalization is performed in-memory with no external calls
- Invalid phones are filtered out early to avoid unnecessary API calls
- Original phone formats are preserved for client convenience

## Error Handling

- Invalid phone formats return `ERROR_CODE_INVALID_PHONE`
- Batch size exceeding 100 returns error immediately
- Individual phone lookup failures don't fail the entire batch
- MAX API errors are properly mapped to domain errors
