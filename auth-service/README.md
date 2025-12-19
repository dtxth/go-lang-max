# Auth Service

Authentication and authorization microservice for the Digital University system.

## Features

- **User Authentication**: JWT-based authentication with access and refresh tokens
- **Role-Based Access Control**: Support for multiple roles (superadmin, curator, operator)
- **Bot Information**: `/bot/me` endpoint to retrieve MaxBot name and add bot link
- **Secure Password Management**:
  - Cryptographically secure random password generation
  - Password delivery via MAX Messenger
  - Self-service password reset with time-limited tokens
  - User-initiated password changes
  - Bcrypt password hashing
- **Audit Logging**: Comprehensive logging of all password operations
- **Monitoring**: Metrics for password operations and notification delivery
- **Health Checks**: Service and dependency health monitoring

## Quick Start

### Prerequisites

- Go 1.21+
- PostgreSQL 14+
- Docker & Docker Compose (optional)

### Local Development

1. **Clone the repository:**
   ```bash
   git clone <repository-url>
   cd auth-service
   ```

2. **Set up environment variables:**
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

3. **Start dependencies:**
   ```bash
   docker-compose up -d postgres
   ```

4. **Run migrations:**
   ```bash
   make migrate-up
   ```

5. **Run the service:**
   ```bash
   make run
   ```

6. **Run tests:**
   ```bash
   make test
   ```

### Docker Deployment

```bash
docker-compose up -d
```

## Configuration

See [Configuration Guide](./PASSWORD_MANAGEMENT_CONFIG.md) for detailed configuration options.

### Environment Variables

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `DATABASE_URL` | PostgreSQL connection string | - | Yes |
| `ACCESS_SECRET` | JWT access token secret | - | Yes |
| `REFRESH_SECRET` | JWT refresh token secret | - | Yes |
| `MAX_BOT_TOKEN` | MAX Mini App bot token for authentication | - | Yes |
| `PORT` | HTTP server port | 8080 | No |
| `GRPC_PORT` | gRPC server port | 9090 | No |
| `MIN_PASSWORD_LENGTH` | Minimum password length | 12 | No |
| `RESET_TOKEN_EXPIRATION` | Token expiration (minutes) | 15 | No |
| `TOKEN_CLEANUP_INTERVAL` | Cleanup interval (minutes) | 60 | No |
| `NOTIFICATION_SERVICE_TYPE` | Notification service (mock/max) | mock | No |
| `MAXBOT_SERVICE_ADDR` | MaxBot gRPC address | - | Conditional* |

\* Required when `NOTIFICATION_SERVICE_TYPE=max`

### Example Configuration

**Development:**
```bash
DATABASE_URL=postgres://auth:password@localhost:5432/auth_dev?sslmode=disable
ACCESS_SECRET=dev_access_secret_change_in_production
REFRESH_SECRET=dev_refresh_secret_change_in_production
MAX_BOT_TOKEN=your-max-bot-token-here
PORT=8080
GRPC_PORT=9090
NOTIFICATION_SERVICE_TYPE=mock
MIN_PASSWORD_LENGTH=8
```

**Production:**
```bash
DATABASE_URL=postgres://auth:secure_password@db.example.com:5432/auth_prod?sslmode=require
ACCESS_SECRET=<strong-random-secret>
REFRESH_SECRET=<strong-random-secret>
MAX_BOT_TOKEN=<your-production-max-bot-token>
PORT=8080
GRPC_PORT=9090
NOTIFICATION_SERVICE_TYPE=max
MAXBOT_SERVICE_ADDR=maxbot-service:9095
MIN_PASSWORD_LENGTH=12
RESET_TOKEN_EXPIRATION=15
TOKEN_CLEANUP_INTERVAL=60
```

## API Documentation

### REST API

- **Base URL**: `http://localhost:8080`
- **Swagger**: `http://localhost:8080/swagger/index.html`

#### Authentication Endpoints

- `POST /register` - Register new user
- `POST /login` - Login user
- `POST /auth/max` - MAX Mini App authentication
- `POST /refresh` - Refresh access token
- `POST /logout` - Logout user

#### Password Management Endpoints

- `POST /auth/password-reset/request` - Request password reset
- `POST /auth/password-reset/confirm` - Reset password with token
- `POST /auth/password/change` - Change password (authenticated)

#### Monitoring Endpoints

- `GET /health` - Health check
- `GET /metrics` - Service metrics

See [API Documentation](./PASSWORD_MANAGEMENT_API.md) for detailed API reference.

### gRPC API

- **Address**: `localhost:9090`
- **Proto**: `api/proto/auth.proto`

#### Available Methods

- `ValidateToken` - Validate JWT token
- `GetUser` - Get user by ID
- `GetUserPermissions` - Get user permissions
- `CreateUser` - Create new user
- `AssignRole` - Assign role to user
- `RevokeUserRoles` - Revoke user roles
- `RequestPasswordReset` - Request password reset
- `ResetPassword` - Reset password with token
- `ChangePassword` - Change user password

## Password Management

### Password Requirements

All passwords must meet the following requirements:
- **Minimum length**: 12 characters (configurable)
- **Complexity**: Must contain:
  - At least one uppercase letter (A-Z)
  - At least one lowercase letter (a-z)
  - At least one digit (0-9)
  - At least one special character (!@#$%^&*()_+-=[]{}|;:,.<>?)

### Password Reset Flow

1. User requests password reset with phone number
2. System generates time-limited reset token (15 minutes)
3. Token is sent to user's phone via MAX Messenger
4. User submits token and new password
5. System validates token and updates password
6. All refresh tokens are invalidated
7. User must log in with new password

### Password Change Flow

1. User logs in with current password
2. User requests password change
3. System verifies current password
4. System validates new password meets requirements
5. System updates password
6. All refresh tokens are invalidated
7. User must log in with new password

See [User Guide](./PASSWORD_MANAGEMENT_USER_GUIDE.md) for detailed user documentation.

## Security Features

### Password Security

- **Generation**: Cryptographically secure random passwords using `crypto/rand`
- **Storage**: Bcrypt hashing with cost factor 10
- **Validation**: Strict complexity requirements
- **Audit**: All operations logged without sensitive data

### Token Security

- **Reset Tokens**: 
  - Cryptographically random (32 bytes, hex-encoded)
  - Time-limited (15 minutes default)
  - Single-use only
  - Automatically cleaned up after expiration

### Session Security

- **JWT Tokens**: 
  - Access tokens (short-lived)
  - Refresh tokens (long-lived)
  - Automatic invalidation on password change/reset

### Audit Logging

All password operations are logged with:
- Operation type (create, reset, change)
- User ID
- Timestamp
- Success/failure status
- Sanitized phone number (last 4 digits only)

**Never logged:**
- Plaintext passwords
- Password hashes
- Reset tokens
- Full phone numbers

## Monitoring

### Metrics

Access metrics at `GET /metrics`:

```json
{
  "user_creations": 1234,
  "password_resets": 567,
  "password_changes": 890,
  "notifications_sent": 1801,
  "notifications_failed": 23,
  "tokens_generated": 567,
  "tokens_used": 543,
  "tokens_expired": 24,
  "maxbot_healthy": true,
  "notification_success_rate": 0.987
}
```

### Key Metrics to Monitor

- **notification_success_rate**: Should be > 0.95
- **notification_failure_rate**: Should be < 0.05
- **maxbot_healthy**: Should be true
- **tokens_expired**: High rate may indicate UX issues

### Health Checks

```bash
# HTTP health check
curl http://localhost:8080/health

# Check metrics
curl http://localhost:8080/metrics
```

## Database

### Migrations

Apply migrations:
```bash
make migrate-up
```

Rollback migrations:
```bash
make migrate-down
```

Create new migration:
```bash
make migrate-create NAME=migration_name
```

### Schema

Key tables:
- `users` - User accounts
- `refresh_tokens` - Active refresh tokens
- `password_reset_tokens` - Password reset tokens
- `roles` - Available roles
- `user_roles` - User role assignments

## Testing

### Run All Tests

```bash
make test
```

### Run Unit Tests

```bash
go test ./internal/...
```

### Run Property-Based Tests

```bash
go test ./test/...
```

### Run Integration Tests

```bash
cd ../integration-tests
make test-auth
```

### Test Coverage

```bash
make coverage
```

## Development

### Project Structure

```
auth-service/
├── api/proto/          # gRPC protocol definitions
├── cmd/auth/           # Application entry point
├── internal/
│   ├── app/            # Application setup
│   ├── config/         # Configuration
│   ├── domain/         # Domain models and interfaces
│   ├── infrastructure/ # Infrastructure implementations
│   │   ├── cleanup/    # Token cleanup job
│   │   ├── grpc/       # gRPC handlers
│   │   ├── http/       # HTTP handlers
│   │   ├── jwt/        # JWT token management
│   │   ├── notification/ # Notification services
│   │   ├── password/   # Password generation
│   │   └── repository/ # Database repositories
│   └── usecase/        # Business logic
├── migrations/         # Database migrations
├── test/               # Property-based tests
└── docs/               # Documentation
```

### Code Style

- Follow Go best practices
- Use `gofmt` for formatting
- Use `golint` for linting
- Write tests for all new features

### Adding New Features

1. Define domain models in `internal/domain/`
2. Implement business logic in `internal/usecase/`
3. Add infrastructure in `internal/infrastructure/`
4. Add HTTP/gRPC handlers
5. Write tests
6. Update documentation

## Troubleshooting

See [Troubleshooting Guide](./PASSWORD_MANAGEMENT_TROUBLESHOOTING.md) for common issues and solutions.

### Quick Diagnostics

```bash
# Check service health
curl http://localhost:8080/health

# Check metrics
curl http://localhost:8080/metrics

# Check logs
docker logs auth-service --tail 100

# Test database connection
psql $DATABASE_URL -c "SELECT 1"

# Check MaxBot connectivity
grpcurl -plaintext $MAXBOT_SERVICE_ADDR list
```

## Documentation

- [API Documentation](./PASSWORD_MANAGEMENT_API.md) - Complete API reference
- [User Guide](./PASSWORD_MANAGEMENT_USER_GUIDE.md) - End-user documentation
- [Configuration Guide](./PASSWORD_MANAGEMENT_CONFIG.md) - Configuration options
- [Troubleshooting Guide](./PASSWORD_MANAGEMENT_TROUBLESHOOTING.md) - Common issues

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Write tests
5. Update documentation
6. Submit a pull request

## License

[License information]

## Support

For issues and questions:
- Check the [Troubleshooting Guide](./PASSWORD_MANAGEMENT_TROUBLESHOOTING.md)
- Review the [API Documentation](./PASSWORD_MANAGEMENT_API.md)
- Contact the development team
