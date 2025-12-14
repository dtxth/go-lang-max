# Implementation Plan

- [x] 1. Setup Auth Service role infrastructure
  - Create database migration for roles and user_roles tables
  - Implement Role and UserRole domain entities
  - Create repository interfaces and PostgreSQL implementations
  - _Requirements: 1.1_

- [ ]* 1.1 Write property test for role hierarchy
  - **Property 4: Role hierarchy grants cumulative permissions**
  - **Validates: Requirements 1.5**

- [x] 1.2 Implement JWT token enhancement with role information
  - Modify JWT token generation to include role, university_id, branch_id, faculty_id
  - Update token validation to extract role information
  - _Requirements: 1.2_

- [ ]* 1.3 Write property test for JWT role inclusion
  - **Property 1: JWT tokens contain role information**
  - **Validates: Requirements 1.2**

- [x] 1.4 Implement ABAC permission validation use case
  - Create Permission value object
  - Implement ValidatePermissionUseCase with role and context checking
  - Add permission validation logic for resource access
  - _Requirements: 1.3, 1.4_

- [ ]* 1.5 Write property test for ABAC validation
  - **Property 2: Permission validation respects ABAC rules**
  - **Validates: Requirements 1.3**

- [ ]* 1.6 Write property test for curator restrictions
  - **Property 3: Curator access is restricted to assigned university**
  - **Validates: Requirements 1.4**

- [x] 1.7 Implement gRPC service for token validation
  - Define ValidateToken proto method
  - Implement gRPC handler to return user_id, role, and context
  - Add GetUserPermissions gRPC method
  - _Requirements: 1.3, 18.1_

- [ ]* 1.8 Write property test for invalid token rejection
  - **Property 18: Invalid tokens are rejected**
  - **Validates: Requirements 5.4, 5.5**

- [x] 2. Enhance Employee Service with role management
  - Add role, user_id, max_id_updated_at columns to employees table via migration
  - Update Employee domain entity with new fields
  - _Requirements: 2.1, 2.2_

- [x] 2.1 Implement role assignment integration with Auth Service
  - Add gRPC client for Auth Service
  - Implement CreateEmployeeWithRoleUseCase that calls Auth Service to assign role
  - Handle role validation (Curator can only assign Operator)
  - _Requirements: 2.1, 2.2, 2.3_

- [ ]* 2.2 Write property test for role assignment
  - **Property 5: Role assignment creates auth service user**
  - **Validates: Requirements 2.3**

- [x] 2.3 Implement role update synchronization
  - Update employee use case to sync role changes with Auth Service
  - Handle role change validation based on requester's role
  - _Requirements: 2.4_

- [ ]* 2.4 Write property test for role synchronization
  - **Property 6: Role updates synchronize across services**
  - **Validates: Requirements 2.4**

- [x] 2.5 Implement employee deletion with permission revocation
  - Update delete use case to call Auth Service to revoke roles
  - Invalidate user tokens via Auth Service
  - _Requirements: 2.5_

- [ ]* 2.6 Write property test for deletion cleanup
  - **Property 7: Employee deletion revokes all permissions**
  - **Validates: Requirements 2.5**

- [x] 3. Integrate MAX_id lookup in Employee Service
  - Add gRPC client for MaxBot Service
  - Update CreateEmployeeUseCase to call MaxBot Service for MAX_id
  - Store MAX_id in employee record when received
  - Handle MAX_id lookup failures gracefully
  - _Requirements: 3.1, 3.4, 3.5_

- [ ]* 3.1 Write property test for MAX_id lookup trigger
  - **Property 8: Employee creation triggers MAX_id lookup**
  - **Validates: Requirements 3.1**

- [ ]* 3.2 Write property test for MAX_id storage
  - **Property 10: MAX_id is stored when received**
  - **Validates: Requirements 3.4**

- [ ]* 3.3 Write property test for graceful failure
  - **Property 11: Employee creation succeeds without MAX_id**
  - **Validates: Requirements 3.5**

- [x] 4. Implement batch MAX_id update in Employee Service
  - Create batch_update_jobs table via migration
  - Implement BatchUpdateMaxIdUseCase
  - Query employees without MAX_id
  - Call MaxBot Service in batches of 100
  - Update employee records with received MAX_ids
  - Generate report with success/failure counts
  - _Requirements: 4.1, 4.2, 4.4, 4.5_

- [ ]* 4.1 Write property test for batch employee selection
  - **Property 12: Batch update processes correct employees**
  - **Validates: Requirements 4.1**

- [ ]* 4.2 Write property test for batch size limits
  - **Property 13: Batch requests respect size limits**
  - **Validates: Requirements 4.2, 4.3**

- [ ]* 4.3 Write property test for batch reporting
  - **Property 14: Batch results are accurately reported**
  - **Validates: Requirements 4.5**

- [x] 4.4 Add HTTP endpoints for batch operations
  - POST /employees/batch-update-maxid - Trigger batch update
  - GET /employees/batch-status - Get batch update status
  - _Requirements: 4.5_

- [x] 5. Enhance MaxBot Service with batch operations
  - Define BatchGetUsersByPhone proto method
  - Implement batch processing with 100 phone limit
  - Return array of phone-to-MAX_id mappings
  - _Requirements: 4.3_

- [x] 5.1 Implement phone normalization in MaxBot Service
  - Create NormalizePhoneUseCase
  - Handle Russian phone formats (8, 9, +7)
  - Remove non-digit characters
  - Validate E.164 format
  - _Requirements: 3.2, 19.1, 19.2, 19.3, 19.4_

- [ ]* 5.2 Write property test for phone normalization
  - **Property 9: Phone numbers are normalized to E.164**
  - **Validates: Requirements 3.2, 19.1**

- [ ]* 5.3 Write property test for Russian phone formats
  - **Property 50: Phone normalization handles Russian formats**
  - **Validates: Requirements 19.2, 19.3**

- [ ]* 5.4 Write property test for non-digit removal
  - **Property 51: Phone normalization removes non-digits**
  - **Validates: Requirements 19.4**

- [ ]* 5.5 Write property test for invalid phone errors
  - **Property 52: Invalid phones return clear errors**
  - **Validates: Requirements 19.5**

- [x] 6. Checkpoint - Ensure all tests pass
  - Ensure all tests pass, ask the user if questions arise.

- [x] 7. Implement role-based filtering in Chat Service
  - Add AuthMiddleware to validate JWT and extract role information
  - Call Auth Service gRPC to validate token
  - Implement ChatFilter value object with role and context
  - _Requirements: 5.4_

- [x] 7.1 Implement ListChatsWithRoleFilterUseCase
  - Filter chats for Superadmin (all chats)
  - Filter chats for Curator (only their university)
  - Filter chats for Operator (only their branch/faculty)
  - _Requirements: 5.1, 5.2, 5.3_

- [ ]* 7.2 Write property test for superadmin access
  - **Property 15: Superadmin sees all chats**
  - **Validates: Requirements 5.1**

- [ ]* 7.3 Write property test for curator filtering
  - **Property 16: Curator sees only university chats**
  - **Validates: Requirements 5.2**

- [ ]* 7.4 Write property test for operator filtering
  - **Property 17: Operator sees only department chats**
  - **Validates: Requirements 5.3**

- [x] 7.5 Update chat list endpoints with role filtering
  - Modify GET /chats to apply role-based filtering
  - Modify GET /chats/all to apply role-based filtering
  - Return 403 for invalid roles
  - _Requirements: 5.5_

- [x] 8. Implement administrator management with permission checks
  - Implement AddAdministratorWithPermissionCheckUseCase
  - Verify user has permission for chat's university/branch
  - Call MaxBot Service to get MAX_id for phone
  - Create administrator record
  - _Requirements: 6.1, 6.2_

- [ ]* 8.1 Write property test for administrator permission check
  - **Property 19: Administrator addition requires permission**
  - **Validates: Requirements 6.1**

- [ ]* 8.2 Write property test for administrator MAX_id lookup
  - **Property 20: Adding administrator triggers MAX_id lookup**
  - **Validates: Requirements 6.2**

- [x] 8.2 Implement RemoveAdministratorWithValidationUseCase
  - Check that at least one other administrator exists
  - Reject removal if last administrator
  - Delete administrator record
  - _Requirements: 6.3, 6.4_

- [ ]* 8.3 Write property test for last admin protection
  - **Property 21: Last administrator cannot be removed**
  - **Validates: Requirements 6.3, 6.4**

- [ ]* 8.4 Write property test for administrator persistence
  - **Property 22: Administrator changes are persisted**
  - **Validates: Requirements 6.5**

- [x] 9. Create Migration Service infrastructure
  - Initialize new migration-service directory with Go module
  - Create migration_jobs and migration_errors tables
  - Implement MigrationJob and MigrationError domain entities
  - Create repository interfaces and PostgreSQL implementations
  - _Requirements: 7.5, 8.5, 20.1, 20.4_

- [x] 9.1 Implement database migration use case
  - Create MigrateFromDatabaseUseCase
  - Read chat data from existing database (INN, name, URL, admin phone)
  - Lookup or create University by INN
  - Create chat records with source='admin_panel'
  - Create administrator records
  - Generate migration report
  - _Requirements: 7.1, 7.2, 7.3, 7.4, 7.5_

- [ ]* 9.2 Write property test for university creation/reuse
  - **Property 24: Universities are created or reused by INN**
  - **Validates: Requirements 7.2, 8.3**

- [ ]* 9.3 Write property test for migration source
  - **Property 23: Migrated chats have correct source**
  - **Validates: Requirements 7.3, 8.4, 9.4**

- [ ]* 9.4 Write property test for migration reporting
  - **Property 25: Migration generates accurate reports**
  - **Validates: Requirements 7.5, 8.5**

- [x] 9.5 Implement Google Sheets migration use case
  - Create MigrateFromGoogleSheetsUseCase
  - Authenticate with Google Sheets API using service account
  - Parse columns: INN, KPP, URL, admin phone
  - Lookup or create University by INN+KPP
  - Create chat records with source='bot_registrar'
  - Log processed rows and errors
  - _Requirements: 8.1, 8.2, 8.3, 8.4, 8.5_

- [x] 9.6 Implement Excel migration use case
  - Create MigrateFromExcelUseCase
  - Validate Excel file format and required columns
  - Parse all columns: phone, INN, FOIV, org name, branch, KPP, faculty, course, group, chat name, URL
  - Call Structure Service to create hierarchy
  - Call Chat Service to create chats with source='academic_group'
  - Link groups to chats
  - _Requirements: 9.1, 9.2, 9.3, 9.4, 9.5_

- [ ]* 9.7 Write property test for Excel validation
  - **Property 26: Excel import validates required columns**
  - **Validates: Requirements 9.1**

- [x] 9.8 Add HTTP endpoints for migration
  - POST /migration/database - Start database migration
  - POST /migration/google-sheets - Start Google Sheets migration
  - POST /migration/excel - Upload and start Excel migration
  - GET /migration/jobs/{id} - Get migration job status
  - GET /migration/jobs - List all migration jobs
  - _Requirements: 7.1, 8.1, 9.1, 20.5_

- [x] 10. Checkpoint - Ensure all tests pass
  - Ensure all tests pass, ask the user if questions arise.

- [x] 11. Enhance Structure Service with department managers
  - Create department_managers table via migration
  - Implement DepartmentManager domain entity
  - Create repository interface and PostgreSQL implementation
  - _Requirements: 11.1, 11.2_

- [x] 11.1 Implement operator assignment use case
  - Create AssignOperatorToDepartmentUseCase
  - Verify operator employee exists via Employee Service gRPC
  - Create department_managers record for branch or faculty
  - _Requirements: 11.1, 11.2, 11.3_

- [ ]* 11.2 Write property test for operator validation
  - **Property 32: Department manager assignments are validated**
  - **Validates: Requirements 11.3**

- [ ]* 11.3 Write property test for operator permissions
  - **Property 33: Operator permissions reflect assignments**
  - **Validates: Requirements 11.4**

- [x] 11.4 Add HTTP endpoints for department managers
  - POST /departments/managers - Assign operator to department
  - DELETE /departments/managers/{id} - Remove operator assignment
  - GET /departments/managers - List all department managers
  - _Requirements: 11.1, 11.5_

- [x] 12. Implement Excel import for structure
  - Enhance groups table with chat_url and chat_name columns
  - Create ImportStructureFromExcelUseCase
  - Validate file format and size limit
  - Parse Excel rows into structure entities
  - Create University, Branch, Faculty, Group in transaction
  - Handle duplicates by updating existing records
  - Return summary with created/updated/failed counts
  - _Requirements: 12.1, 12.2, 12.3, 12.4, 12.5_

- [ ]* 12.1 Write property test for structure hierarchy creation
  - **Property 27: Structure import creates full hierarchy**
  - **Validates: Requirements 9.3**

- [ ]* 12.2 Write property test for transactional import
  - **Property 34: Excel import is transactional**
  - **Validates: Requirements 12.3**

- [ ]* 12.3 Write property test for duplicate handling
  - **Property 35: Duplicate records are updated not created**
  - **Validates: Requirements 12.4**

- [ ]* 12.4 Write property test for import summary
  - **Property 36: Import summary is accurate**
  - **Validates: Requirements 12.5**

- [x] 12.5 Add Excel import endpoint
  - POST /import/excel - Import structure from Excel
  - Validate file upload
  - Process in background
  - Return import job ID
  - _Requirements: 12.1_

- [x] 13. Implement structure hierarchy retrieval
  - Create GetUniversityStructureUseCase
  - Build nested hierarchy: University → Branch → Faculty → Group → Chat
  - Call Chat Service gRPC for chat details when chat_id present
  - Order entities alphabetically at each level
  - Handle optional branches
  - _Requirements: 10.2, 10.3, 10.5, 13.1, 13.2, 13.3, 13.5_

- [ ]* 13.1 Write property test for group-chat linking
  - **Property 28: Groups are linked to chats**
  - **Validates: Requirements 9.5, 10.1**

- [ ]* 13.2 Write property test for structure with chat details
  - **Property 29: Structure retrieval includes chat details**
  - **Validates: Requirements 10.2**

- [ ]* 13.3 Write property test for chat deletion handling
  - **Property 30: Chat deletion preserves groups**
  - **Validates: Requirements 10.4**

- [ ]* 13.4 Write property test for hierarchy structure
  - **Property 31: Structure displays correct hierarchy**
  - **Validates: Requirements 10.5**

- [ ]* 13.5 Write property test for nested JSON format
  - **Property 37: Structure is returned as nested JSON**
  - **Validates: Requirements 13.1**

- [ ]* 13.6 Write property test for alphabetical ordering
  - **Property 38: Structure entities are alphabetically ordered**
  - **Validates: Requirements 13.5**

- [x] 13.7 Update structure endpoints
  - GET /universities/{id}/structure - Get full hierarchy
  - Include chat details in response
  - _Requirements: 13.4_

- [x] 14. Implement employee search with role filtering
  - Create SearchEmployeesWithRoleFilterUseCase
  - Search by first_name, last_name, university name
  - Apply role-based filtering (Superadmin: all, Curator: their university)
  - Include full name, phone, role, university name in results
  - _Requirements: 14.1, 14.2, 14.3, 14.4_

- [ ]* 14.1 Write property test for multi-field search
  - **Property 39: Employee search matches multiple fields**
  - **Validates: Requirements 14.1**

- [ ]* 14.2 Write property test for search filtering
  - **Property 40: Search respects role-based filtering**
  - **Validates: Requirements 14.3**

- [ ]* 14.3 Write property test for search result format
  - **Property 41: Search results include all required fields**
  - **Validates: Requirements 14.4**

- [x] 14.4 Update employee search endpoint
  - GET /employees - Apply role-based filtering
  - Return empty array for no matches
  - _Requirements: 14.5_

- [x] 15. Implement automatic university creation in Employee Service
  - Update CreateEmployeeUseCase to check if University exists by INN
  - Create University if not exists
  - Reuse existing University if exists
  - Store name, INN, KPP
  - _Requirements: 15.1, 15.2, 15.3, 15.5_

- [ ]* 15.1 Write property test for university auto-creation
  - **Property 42: New universities are created automatically**
  - **Validates: Requirements 15.1**

- [ ]* 15.2 Write property test for university reuse
  - **Property 43: Existing universities are reused**
  - **Validates: Requirements 15.2**

- [x] 16. Implement pagination for chat lists
  - Update chat list endpoints to accept limit and offset
  - Use default limit of 50
  - Cap limit at 100
  - Include total count in response metadata
  - Return empty array for offset > total
  - _Requirements: 16.1, 16.2, 16.3, 16.4, 16.5_

- [ ]* 16.1 Write property test for limit capping
  - **Property 44: Pagination limit is capped at 100**
  - **Validates: Requirements 16.3**

- [ ]* 16.2 Write property test for total count
  - **Property 45: Pagination includes total count**
  - **Validates: Requirements 16.4**

- [x] 17. Implement chat search functionality
  - Update search endpoint to use full-text search on chat name
  - Configure Russian language text search
  - Apply role-based filtering before returning results
  - Support multi-word search (all words must match)
  - Return empty array for no matches
  - _Requirements: 17.1, 17.2, 17.3, 17.4, 17.5_

- [ ]* 17.1 Write property test for search with filtering
  - **Property 46: Search applies role-based filtering**
  - **Validates: Requirements 17.2**

- [ ]* 17.2 Write property test for multi-word search
  - **Property 47: Multi-word search requires all words**
  - **Validates: Requirements 17.4**

- [x] 18. Implement gRPC retry logic across all services
  - Create gRPC client wrapper with retry logic
  - Retry up to 3 times with exponential backoff (1s, 2s, 4s)
  - Log each retry attempt
  - Return error after final failure with appropriate HTTP status
  - _Requirements: 18.4, 18.5_

- [ ]* 18.1 Write property test for gRPC retries
  - **Property 48: gRPC calls are retried on failure**
  - **Validates: Requirements 18.4**

- [ ]* 18.2 Write property test for retry failure handling
  - **Property 49: Failed retries return appropriate errors**
  - **Validates: Requirements 18.5**

- [x] 19. Add comprehensive error handling
  - Implement consistent error response format across all services
  - Add error codes for each error category
  - Include detailed error messages and context
  - Log all errors with request context
  - _Requirements: 19.5_

- [x] 20. Checkpoint - Ensure all tests pass
  - Ensure all tests pass, ask the user if questions arise.

- [x] 21. Setup monitoring and logging
  - Add structured JSON logging to all services
  - Include request_id for tracing
  - Log all gRPC calls with duration
  - Log all migration operations with progress
  - Expose health check endpoints
  - _Requirements: 20.1, 20.2, 20.3, 20.4, 20.5_

- [x] 22. Create database migration scripts
  - Write migration for Auth Service (roles, user_roles)
  - Write migration for Employee Service (role, user_id, max_id_updated_at, batch_update_jobs)
  - Write migration for Structure Service (department_managers, groups enhancements)
  - Write migration for Migration Service (migration_jobs, migration_errors)
  - Test migrations with rollback
  - _Requirements: 1.1_

- [x] 23. Update Docker Compose configuration
  - Add migration-service to docker-compose.yml
  - Add migration-db database
  - Configure service dependencies
  - Add environment variables for all services
  - _Requirements: All_

- [x] 24. Write integration tests
  - Test end-to-end employee creation with role and MAX_id
  - Test chat filtering with different roles
  - Test Excel import with full structure creation
  - Test migration from all three sources
  - Test gRPC communication between services
  - _Requirements: All_

- [x] 25. Final Checkpoint - Ensure all tests pass
  - Ensure all tests pass, ask the user if questions arise.

- [x] 26. Update documentation
  - Update README.md with new services and features
  - Document new API endpoints
  - Document migration process
  - Document role-based access control
  - Add deployment guide
  - _Requirements: All_
