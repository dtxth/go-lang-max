# Protocol Buffers Setup and Code Generation

This document describes the Protocol Buffers (protobuf) setup for the Digital University MVP project, including gRPC service definitions and code generation.

## Overview

The project uses Protocol Buffers for defining gRPC service interfaces and message types across all microservices. This ensures type safety, performance, and consistency in inter-service communication.

## Proto File Structure

```
├── auth-service/api/proto/
│   ├── auth.proto          # Auth service definitions
│   ├── auth.pb.go          # Generated Go structs
│   └── auth_grpc.pb.go     # Generated gRPC client/server code
├── chat-service/api/proto/
│   ├── chat.proto          # Chat service definitions
│   ├── chat.pb.go          # Generated Go structs
│   └── chat_grpc.pb.go     # Generated gRPC client/server code
├── employee-service/api/proto/
│   ├── employee.proto      # Employee service definitions
│   ├── employee.pb.go      # Generated Go structs
│   └── employee_grpc.pb.go # Generated gRPC client/server code
├── structure-service/api/proto/
│   ├── structure.proto     # Structure service definitions
│   ├── structure.pb.go     # Generated Go structs
│   └── structure_grpc.pb.go # Generated gRPC client/server code
└── gateway-service/api/proto/
    ├── gateway.proto       # Gateway service definitions
    ├── gateway.pb.go       # Generated Go structs
    └── gateway_grpc.pb.go  # Generated gRPC client/server code
```

## Service Definitions

### Auth Service (auth.proto)
- **Register**: User registration with email/phone
- **Login/LoginByPhone**: User authentication
- **Refresh/Logout**: Token management
- **AuthenticateMAX**: MAX platform integration
- **Password Management**: Reset, change password
- **Bot Operations**: GetBotMe for bot information
- **User Management**: CRUD operations for users
- **Role Management**: Assign/revoke user roles
- **Health/Metrics**: Service monitoring

### Chat Service (chat.proto)
- **Chat Operations**: GetAllChats, SearchChats, CreateChat
- **Administrator Management**: Add/remove chat administrators
- **Participants**: RefreshParticipantsCount
- **Migration Support**: AddAdministratorForMigration
- **Health**: Service monitoring

### Employee Service (employee.proto)
- **Employee Operations**: CRUD operations for employees
- **Search/Pagination**: GetAllEmployees, SearchEmployees
- **Batch Operations**: BatchUpdateMaxID with job tracking
- **University Operations**: Get university by ID/INN/KPP
- **Health**: Service monitoring

### Structure Service (structure.proto)
- **University Management**: CRUD operations for universities
- **Structure Operations**: Create/manage organizational structure
- **Excel Import**: ImportExcel for bulk data import
- **Name Updates**: Update branch/faculty/group names
- **Department Managers**: Assign/remove department operators
- **Health**: Service monitoring

### Gateway Service (gateway.proto)
- **Health**: Gateway service monitoring
- **Metrics**: Gateway performance metrics

## Code Generation

### Prerequisites

Install required tools:

```bash
# macOS
brew install protobuf

# Ubuntu
apt-get install protobuf-compiler

# Install Go plugins
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

### Generation Commands

#### Generate All Services
```bash
# From project root
make proto-gen
# or
./scripts/generate-proto.sh
```

#### Generate Individual Services
```bash
# Auth Service
cd auth-service && make proto-gen

# Chat Service  
cd chat-service && make proto-gen

# Employee Service
cd employee-service && make proto-gen

# Structure Service
cd structure-service && make proto-gen

# Gateway Service
cd gateway-service && make proto-gen
```

#### Manual Generation
```bash
# Example for auth service
cd auth-service
protoc --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
       api/proto/auth.proto
```

## Build Integration

Proto generation is integrated into the build process:

1. **Main Makefile**: `make build` automatically runs `make proto-gen`
2. **Service Makefiles**: Each service's `make build` runs proto generation
3. **Docker Builds**: Dockerfiles include proto generation steps

## Go Package Configuration

Each proto file uses consistent Go package configuration:

```protobuf
option go_package = "service-name/api/proto;proto";
```

This generates Go code in the `api/proto` package within each service.

## Message Design Patterns

### Common Patterns
1. **Request/Response Pairs**: Every RPC has dedicated request/response messages
2. **Error Handling**: All responses include an `error` field for error messages
3. **Pagination**: List operations include `page`, `limit`, `total` fields
4. **Optional Fields**: Use `optional` keyword for nullable fields
5. **Timestamps**: Use `string` type for ISO 8601 timestamps

### Example Message Structure
```protobuf
message GetAllItemsRequest {
  int32 page = 1;
  int32 limit = 2;
  string sort_by = 3;
  string sort_order = 4;
}

message GetAllItemsResponse {
  repeated Item items = 1;
  int32 total = 2;
  int32 page = 3;
  int32 limit = 4;
  string error = 5;
}
```

## Validation and Compatibility

### Backward Compatibility Rules
1. **Never remove fields**: Mark as deprecated instead
2. **Never change field numbers**: Field numbers are permanent identifiers
3. **Never change field types**: This breaks serialization
4. **Add new fields with new numbers**: Always increment field numbers

### Validation
- Proto files are validated during generation
- Build fails if proto files have syntax errors
- Generated code is checked into version control

## Integration with Gateway Service

The Gateway Service will:
1. Import all service proto packages
2. Create gRPC clients for each service
3. Translate HTTP requests to gRPC calls
4. Handle error mapping and response formatting

## Troubleshooting

### Common Issues

1. **Duplicate message definitions**: Ensure no duplicate message names in proto files
2. **Import path issues**: Verify `go_package` options are correct
3. **Missing protoc**: Install Protocol Buffers compiler
4. **Missing Go plugins**: Install protoc-gen-go and protoc-gen-go-grpc

### Regeneration
If generated files become corrupted:
```bash
# Clean generated files
find . -name "*.pb.go" -delete

# Regenerate all
make proto-gen
```

## Future Enhancements

1. **Proto validation**: Add field validation rules
2. **Documentation generation**: Generate API docs from proto files
3. **Client libraries**: Generate client libraries for other languages
4. **Schema registry**: Implement proto schema versioning
5. **Breaking change detection**: Automated compatibility checking