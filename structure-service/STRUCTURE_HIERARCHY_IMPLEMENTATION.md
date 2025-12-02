# Structure Hierarchy Retrieval Implementation

## Overview

This document describes the implementation of the structure hierarchy retrieval feature for the Structure Service, which provides a nested view of university organizational structure with integrated chat details.

## Implementation Details

### 1. GetUniversityStructureUseCase

**File:** `internal/usecase/get_university_structure.go`

**Purpose:** Retrieves the full university structure hierarchy with nested entities and chat details.

**Key Features:**
- Builds nested hierarchy: University → Branch → Faculty → Group → Chat
- Handles optional branches (direct university-to-faculty structure)
- Fetches chat details from Chat Service via gRPC when chat_id is present
- Orders all entities alphabetically at each level
- Gracefully handles deleted chats (Requirement 10.4)

**Requirements Addressed:**
- 10.2: Structure retrieval includes chat details
- 10.3: Calls Chat Service gRPC for chat information
- 10.5: Displays correct hierarchy with optional branches
- 13.1: Returns nested JSON structure
- 13.2: Includes chat details in response
- 13.3: Handles chat deletion gracefully
- 13.5: Alphabetically orders entities

### 2. ChatServiceAdapter

**File:** `internal/infrastructure/grpc/chat_service_adapter.go`

**Purpose:** Adapts the gRPC chat client to the domain ChatService interface.

**Key Features:**
- Implements the ChatService interface defined in the use case
- Converts protobuf Chat messages to domain Chat entities
- Provides clean separation between infrastructure and domain layers

### 3. Updated HTTP Handler

**File:** `internal/infrastructure/http/handler.go`

**Changes:**
- Added `getUniversityStructureUseCase` field to Handler struct
- Updated `GetStructure` endpoint to use the new use case
- Passes request context to enable gRPC calls

### 4. Updated Main Application

**File:** `cmd/structure/main.go`

**Changes:**
- Created ChatServiceAdapter instance
- Initialized GetUniversityStructureUseCase with repository and chat service
- Wired up the new use case to the HTTP handler

## API Endpoint

### GET /universities/{id}/structure

**Description:** Retrieves the full hierarchical structure of a university with nested entities and chat details.

**Path Parameters:**
- `id` (int64): University ID

**Response Format:**
```json
{
  "type": "university",
  "id": 1,
  "name": "Moscow State University",
  "children": [
    {
      "type": "branch",
      "id": 1,
      "name": "Main Campus",
      "children": [
        {
          "type": "faculty",
          "id": 1,
          "name": "Computer Science",
          "children": [
            {
              "type": "group",
              "id": 1,
              "name": "CS-101",
              "course": 1,
              "group_num": "CS-101",
              "chat": {
                "id": 1,
                "name": "CS-101 Chat",
                "url": "https://max.chat/...",
                "max_id": "chat123"
              }
            }
          ]
        }
      ]
    }
  ]
}
```

**Response Codes:**
- 200: Success
- 400: Invalid university ID
- 404: University not found
- 500: Internal server error

## Structure Hierarchy Rules

### With Branches
```
University
  └── Branch (alphabetically ordered)
      └── Faculty (alphabetically ordered)
          └── Group (ordered by course, then alphabetically)
              └── Chat (optional)
```

### Without Branches
```
University
  └── Faculty (alphabetically ordered)
      └── Group (ordered by course, then alphabetically)
          └── Chat (optional)
```

## Chat Integration

### Chat Details Retrieval
- When a group has a `chat_id`, the use case calls Chat Service via gRPC
- Chat details include: id, name, url, max_id
- If chat is not found or deleted, the group is still returned without chat details
- Errors are logged but don't fail the entire request

### Graceful Degradation
- If Chat Service is unavailable, groups are returned without chat details
- If a specific chat is deleted, the group's chat_id remains but no chat object is included
- This preserves the group structure even when chats are removed (Requirement 10.4)

## Alphabetical Ordering

All entities are ordered alphabetically at each level:
- Branches: by name
- Faculties: by name
- Groups: by course (ascending), then by number (alphabetically)

This ensures consistent and predictable structure presentation (Requirement 13.5).

## Testing

The implementation can be tested by:
1. Creating a university with branches, faculties, and groups
2. Linking some groups to chats
3. Calling GET /universities/{id}/structure
4. Verifying the nested structure and chat details
5. Deleting a chat and verifying the group still appears without chat details

## Dependencies

- Chat Service gRPC client for fetching chat details
- Structure Repository for database access
- Context for request lifecycle management

## Error Handling

- University not found: Returns 404
- Invalid university ID: Returns 400
- Database errors: Returns 500
- Chat Service errors: Logged, but request continues without chat details
- Missing chat: Group returned without chat object

## Future Enhancements

- Caching of structure hierarchy for frequently accessed universities
- Pagination for large structures
- Filtering by branch or faculty
- Include department manager information in the structure
