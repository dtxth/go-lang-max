# Requirements Document

## Introduction

This document specifies the requirements for integrating the official max-bot-api-client-go library (https://github.com/max-messenger/max-bot-api-client-go) into the maxbot-service. Currently, the service uses a stub implementation with only phone normalization logic. The integration will replace the stub with the official client library to enable real interactions with the Max Messenger API, including user lookup, message sending, and other bot operations.

## Glossary

- **MaxBot Service**: The gRPC microservice that provides an interface for interacting with Max Messenger API
- **Max API Client**: The official Go client library for Max Messenger Bot API (max-bot-api-client-go)
- **Max ID**: A unique identifier for a user in the Max Messenger system
- **Phone Number**: A user's phone number used to identify them in Max Messenger
- **Bot Token**: An authentication token required to interact with Max Messenger API
- **Stub Implementation**: The current placeholder implementation that only normalizes phone numbers without making real API calls
- **Domain Interface**: The MaxAPIClient interface that defines the contract for Max API interactions
- **Infrastructure Layer**: The layer containing concrete implementations of domain interfaces

## Requirements

### Requirement 1

**User Story:** As a developer, I want to replace the stub Max API client with the official max-bot-api-client-go library, so that the service can make real API calls to Max Messenger.

#### Acceptance Criteria

1. WHEN the maxbot-service starts THEN the system SHALL initialize the Max API client using the official max-bot-api-client-go library with the configured bot token
2. WHEN the Max API client is created THEN the system SHALL use the bot token from the MAX_API_TOKEN environment variable
3. WHEN the Max API client initialization fails THEN the system SHALL return a descriptive error and prevent service startup
4. WHEN the service makes API calls THEN the system SHALL use the official client library instead of the stub implementation

### Requirement 2

**User Story:** As a service consumer, I want to retrieve a user's Max ID by their phone number, so that I can identify users in the Max Messenger system.

#### Acceptance Criteria

1. WHEN a valid phone number is provided THEN the MaxBot Service SHALL call the Max API to retrieve the corresponding Max ID
2. WHEN the Max API returns a user THEN the MaxBot Service SHALL return the user's Max ID to the caller
3. WHEN the Max API does not find a user for the phone number THEN the MaxBot Service SHALL return an ERROR_CODE_MAX_ID_NOT_FOUND error
4. WHEN the phone number format is invalid THEN the MaxBot Service SHALL return an ERROR_CODE_INVALID_PHONE error before making an API call
5. WHEN the Max API call fails due to network or server errors THEN the MaxBot Service SHALL return an ERROR_CODE_INTERNAL error with details

### Requirement 3

**User Story:** As a developer, I want the phone validation logic to remain functional, so that invalid phone numbers are rejected before making API calls.

#### Acceptance Criteria

1. WHEN a phone number is provided for validation THEN the system SHALL normalize it according to Russian phone number format rules
2. WHEN a phone number starts with "8" and has 11 digits THEN the system SHALL convert it to start with "7"
3. WHEN a phone number has 10 digits THEN the system SHALL prepend "7" to create an 11-digit number
4. WHEN a phone number has fewer than 10 or more than 15 digits THEN the system SHALL reject it as invalid
5. WHEN a phone number contains non-digit characters THEN the system SHALL remove them before validation

### Requirement 4

**User Story:** As a developer, I want to maintain the existing domain interface, so that the integration does not break existing service contracts.

#### Acceptance Criteria

1. WHEN the new Max API client is implemented THEN the system SHALL maintain compatibility with the existing MaxAPIClient domain interface
2. WHEN the GetMaxIDByPhone method is called THEN the system SHALL return the same signature as the current interface
3. WHEN the ValidatePhone method is called THEN the system SHALL return the same signature as the current interface
4. WHEN the service layer calls the client THEN the system SHALL work without modifications to the usecase layer

### Requirement 5

**User Story:** As a developer, I want proper error handling for Max API interactions, so that failures are communicated clearly to service consumers.

#### Acceptance Criteria

1. WHEN the Max API returns an authentication error THEN the system SHALL log the error and return ERROR_CODE_INTERNAL
2. WHEN the Max API returns a rate limit error THEN the system SHALL log the error and return ERROR_CODE_INTERNAL
3. WHEN the Max API request times out THEN the system SHALL return ERROR_CODE_INTERNAL with a timeout message
4. WHEN the Max API returns an unexpected error THEN the system SHALL log the full error details and return ERROR_CODE_INTERNAL
5. WHEN an error occurs THEN the system SHALL include a descriptive error message in the response

### Requirement 6

**User Story:** As a system administrator, I want to configure the Max API client through environment variables, so that I can deploy the service in different environments without code changes.

#### Acceptance Criteria

1. WHEN the service starts THEN the system SHALL read the MAX_API_TOKEN environment variable for authentication
2. WHEN the service starts THEN the system SHALL read the MAX_API_URL environment variable for the API base URL
3. WHEN the service starts THEN the system SHALL read the MAX_API_TIMEOUT environment variable for request timeout configuration
4. WHEN an environment variable is missing THEN the system SHALL use a sensible default value or fail with a clear error message
5. WHEN the configuration is loaded THEN the system SHALL validate that required values are present

### Requirement 7

**User Story:** As a developer, I want to add the max-bot-api-client-go dependency to the project, so that the library is available for use.

#### Acceptance Criteria

1. WHEN the go.mod file is updated THEN the system SHALL include the max-bot-api-client-go library as a dependency
2. WHEN go mod tidy is executed THEN the system SHALL download the library and update go.sum
3. WHEN the project is built THEN the system SHALL successfully compile with the new dependency
4. WHEN the dependency is added THEN the system SHALL use a stable version of the library

### Requirement 8

**User Story:** As a developer, I want to explore additional Max API capabilities provided by the client library, so that I can extend the service functionality in the future.

#### Acceptance Criteria

1. WHEN reviewing the max-bot-api-client-go library THEN the system documentation SHALL identify available API methods beyond user lookup
2. WHEN the client is initialized THEN the system SHALL be structured to easily add new API method wrappers
3. WHEN new methods are needed THEN the system SHALL allow extending the MaxAPIClient interface without breaking existing functionality
4. WHEN the integration is complete THEN the system SHALL document which Max API features are currently implemented and which are available for future use
