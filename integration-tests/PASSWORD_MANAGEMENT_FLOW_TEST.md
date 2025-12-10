# Password Management Integration Test

## Overview

This document describes the integration test for the secure password management flow, which validates the complete Employee Service → Auth Service → MaxBot Service integration.

## Test: TestEmployeeServiceToAuthServiceToMaxBotFlow

### Purpose

Validates the complete user creation flow with secure password generation, hashing, and notification delivery.

### Requirements Validated

- **Requirement 1.1**: Password is at least 12 characters (validated by password generator)
- **Requirement 1.2**: Password contains uppercase, lowercase, digit, and special character (validated by password generator)
- **Requirement 1.3**: Password is hashed with bcrypt before storage
- **Requirement 2.1**: Password notification is sent via MaxBot Service

### Test Flow

```
1. Employee Service receives employee creation request with role
   ↓
2. Employee Service generates cryptographically secure random password
   ↓
3. Employee Service calls Auth Service to create user
   ↓
4. Auth Service hashes password with bcrypt
   ↓
5. Auth Service stores user with hashed password
   ↓
6. Auth Service assigns role to user
   ↓
7. Employee Service attempts to send password notification via MaxBot Service
   ↓
8. Employee creation succeeds (even if notification fails)
```

### What the Test Verifies

#### Step 1: Employee Service Integration
- Employee creation request returns 201 Created
- Response includes employee ID, phone, and role
- Response includes user_id (indicating Auth Service integration succeeded)

#### Step 2: Auth Service User Creation
- User record exists in Auth Service database
- User ID matches the one returned by Employee Service
- Phone number is stored correctly

#### Step 3: Password Hashing (Requirement 1.3)
- Password hash is not empty
- Password hash starts with `$2a` or `$2b` (bcrypt format)
- Password hash is exactly 60 characters (bcrypt standard)
- Password is NOT stored in plaintext

#### Step 4: Role Assignment
- User has exactly one role assigned
- Assigned role matches the requested role (curator/operator)

#### Step 5: Graceful Error Handling
- User creation succeeds even if notification delivery fails
- Employee can be retrieved after creation
- All data persists correctly

### What the Test Does NOT Verify

The following are verified by other tests:

1. **Password Length (Requirement 1.1)**: Verified by property-based tests in `auth-service/test/password_reset_properties_test.go`

2. **Password Complexity (Requirement 1.2)**: Verified by property-based tests in `auth-service/test/password_reset_properties_test.go`

3. **Actual Notification Delivery (Requirement 2.1)**: Cannot be verified in integration tests without a real MaxBot Service instance. Verified by:
   - Unit tests in `auth-service/internal/infrastructure/notification/max_service_test.go`
   - Property tests in `auth-service/test/password_reset_properties_test.go`

### Running the Test

```bash
# Start all services
docker-compose up -d

# Wait for services to be ready
sleep 30

# Run the specific test
cd integration-tests
go test -v -run TestEmployeeServiceToAuthServiceToMaxBotFlow

# Or run all password management tests
go test -v -run TestPassword
```

### Test Data Cleanup

The test automatically cleans up:
- Employee records from Employee Service database
- University records from Employee Service database
- User records from Auth Service database
- User role assignments from Auth Service database

### Expected Output

```
=== RUN   TestEmployeeServiceToAuthServiceToMaxBotFlow
    password_management_integration_test.go:XXX: Step 1: Creating employee via Employee Service
    password_management_integration_test.go:XXX: Step 2: Verifying Auth Service created user
    password_management_integration_test.go:XXX: Step 3: Verifying password is hashed with bcrypt
    password_management_integration_test.go:XXX: Step 4: Verifying role assignment in Auth Service
    password_management_integration_test.go:XXX: Step 5: Verifying graceful handling of notification flow
    password_management_integration_test.go:XXX: Cleaning up test data
    password_management_integration_test.go:XXX: Integration test completed successfully
--- PASS: TestEmployeeServiceToAuthServiceToMaxBotFlow (X.XXs)
PASS
```

### Troubleshooting

#### Test fails with "Service not ready"
- Ensure all services are running: `docker-compose ps`
- Check service health: `curl http://localhost:8080/health`
- Wait longer for services to start (increase sleep time)

#### Test fails with "404 page not found"
- Verify Employee Service is running on port 8081
- Check docker-compose.yml for correct port mappings
- Verify routes are registered in Employee Service

#### Test fails with "user_id should be set"
- Check Auth Service logs for errors
- Verify Auth Service database is accessible
- Check gRPC connection between Employee and Auth services

#### Test fails with "Password should be hashed with bcrypt"
- Verify bcrypt is being used in Auth Service
- Check Auth Service password hashing implementation
- Review Auth Service logs for password storage errors

## Related Tests

- `TestUserCreationWithPasswordGeneration` - Basic password generation test
- `TestPasswordGenerationUniqueness` - Verifies passwords are unique
- `TestPasswordNotificationFlow` - Verifies notification flow doesn't block creation
- `TestEmployeeCreationWithoutRole` - Verifies no password for employees without roles

## Related Files

- `employee-service/internal/usecase/create_employee_with_role.go` - Employee creation logic
- `auth-service/internal/usecase/auth_service.go` - User creation and password hashing
- `auth-service/internal/infrastructure/password/generator.go` - Password generation
- `auth-service/internal/infrastructure/notification/max_service.go` - Notification delivery
