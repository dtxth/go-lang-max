# Implementation Plan

- [x] 1. Configuration setup (COMPLETED)
  - Configuration already properly loads MAX_BOT_TOKEN, MAX_API_URL, and MAX_API_TIMEOUT
  - Domain interface MaxAPIClient exists with correct methods
  - Usecase layer properly structured
  - gRPC handler with error mapping complete
  - _Requirements: 6.1, 6.2, 6.3, 6.4, 6.5_

- [x] 2. Add max-bot-api-client-go dependency to project
  - Update go.mod to include github.com/max-messenger/max-bot-api-client-go
  - Run go mod tidy to download the library and update go.sum
  - Verify the project builds successfully with the new dependency
  - _Requirements: 7.1, 7.2, 7.3, 7.4_

- [x] 3. Add gopter dependency for property-based testing
  - Add github.com/leanovate/gopter to go.mod
  - Run go mod tidy
  - Verify gopter is available for test files
  - _Requirements: Testing Strategy_

- [x] 4. Replace stub implementation with real Max API client
  - Update client.go to use max-bot-api-client-go library
  - Initialize official Max API client in NewClient function
  - Keep existing phone normalization helper functions (normalizePhone, etc.)
  - Implement error mapping from Max API errors to domain errors
  - Add logging for API calls and errors (with privacy considerations)
  - _Requirements: 1.1, 1.2, 1.3, 1.4, 4.1, 4.2, 4.3, 5.1, 5.2, 5.3, 5.4, 5.5_

- [x] 4.1 Update GetMaxIDByPhone to call real Max API
  - Replace TODO stub with actual Max API call to get user by phone
  - Extract Max ID from API response
  - Map API errors to domain errors (not found, auth errors, timeouts, etc.)
  - Return Max ID or appropriate error
  - _Requirements: 2.1, 2.2, 2.3, 2.4, 2.5_

- [ ]* 4.2 Write unit tests for phone normalization
  - Test "8" to "7" conversion for 11-digit numbers
  - Test "7" prepending for 10-digit numbers
  - Test non-digit character removal
  - Test invalid length rejection
  - Test edge cases (empty string, very long numbers, special characters)
  - _Requirements: 3.1, 3.2, 3.3, 3.4, 3.5_

- [ ]* 4.3 Write property test for phone normalization consistency
  - **Property 5: Phone normalization consistency**
  - **Validates: Requirements 3.1**

- [ ]* 4.4 Write property test for eight-to-seven conversion
  - **Property 6: Eight-to-seven conversion**
  - **Validates: Requirements 3.2**

- [ ]* 4.5 Write property test for ten-digit prepending
  - **Property 7: Ten-digit prepending**
  - **Validates: Requirements 3.3**

- [ ]* 4.6 Write property test for length validation
  - **Property 8: Length validation**
  - **Validates: Requirements 3.4**

- [ ]* 4.7 Write property test for non-digit removal
  - **Property 9: Non-digit removal**
  - **Validates: Requirements 3.5**

- [ ]* 4.8 Write unit tests for error mapping
  - Test Max API not found error maps to domain.ErrMaxIDNotFound
  - Test Max API auth error maps to internal error
  - Test Max API timeout maps to timeout error
  - Test unexpected errors are wrapped properly
  - _Requirements: 5.1, 5.2, 5.3, 5.4_

- [ ]* 4.9 Write property test for error messages presence
  - **Property 10: Error messages presence**
  - **Validates: Requirements 5.5**

- [x] 5. Update main.go with validation and error handling
  - Add validation for MAX_BOT_TOKEN (fail fast if missing)
  - Handle client initialization errors gracefully
  - Add appropriate logging for startup
  - _Requirements: 1.1, 1.2, 1.3_

- [ ]* 5.1 Write property test for client initialization
  - **Property 1: Client initialization with valid tokens**
  - **Validates: Requirements 1.1**

- [ ]* 5.2 Write property test for invalid phone rejection
  - **Property 4: Invalid phone rejection**
  - **Validates: Requirements 2.4**

- [ ]* 5.3 Write integration tests for complete flow
  - Test GetMaxIDByPhone with valid phone (requires test Max API or mock)
  - Test GetMaxIDByPhone with invalid phone
  - Test GetMaxIDByPhone with non-existent user
  - Test ValidatePhone with various phone formats
  - Verify usecase and gRPC handler work correctly with new client
  - _Requirements: 2.1, 2.2, 2.3, 2.4, 2.5, 4.4_

- [x] 6. Checkpoint - Ensure all tests pass
  - Ensure all tests pass, ask the user if questions arise.

- [x] 7. Update documentation
  - Update maxbot-service README.md with new Max API integration details
  - Document required environment variables (MAX_BOT_TOKEN, MAX_API_URL, MAX_API_TIMEOUT)
  - Add examples of how to configure the service
  - Document available Max API features and future extension points
  - _Requirements: 8.4_

- [x] 8. Clean up stub implementation comments
  - Remove TODO comments and stub-specific notes
  - Ensure code is production-ready
  - _Requirements: 1.4_
