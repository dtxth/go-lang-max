# Migration Service

Migration Service handles the migration of chat data from three different sources:
1. Existing database (admin_panel source) - ~6,000 chats
2. Google Sheets (bot_registrar source)
3. Excel files (academic_group source) - ~155,000 chats

## API Documentation

После запуска сервиса, **Swagger UI** доступен по адресу:

**http://localhost:8084/swagger/index.html**

### Обновление Swagger документации

```bash
# Генерация Swagger документации
make swagger

# Или напрямую
swag init -g cmd/migration/main.go -o internal/infrastructure/http/docs
```

## Features

- **Database Migration**: Import chats from existing database with INN, name, URL, and admin phone
- **Google Sheets Migration**: Import chats from Google Sheets with INN, KPP, URL, and admin phone
- **Excel Migration**: Import academic group chats with full structure hierarchy
- **Progress Tracking**: Track migration progress with job status and error reporting
- **Error Handling**: Record and report errors for failed records

## Architecture

The service follows clean architecture principles:
- **Domain Layer**: Core business entities and interfaces
- **Use Case Layer**: Business logic for each migration type
- **Infrastructure Layer**: HTTP handlers, repositories, and external service clients

## API Endpoints

### Start Database Migration
```
POST /migration/database
Content-Type: application/json

{
  "source_identifier": "legacy_db"
}
```

### Start Google Sheets Migration
```
POST /migration/google-sheets
Content-Type: application/json

{
  "spreadsheet_id": "1abc..."
}
```

### Start Excel Migration
```
POST /migration/excel
Content-Type: multipart/form-data

file: <excel_file.xlsx>
```

### Get Migration Job Status
```
GET /migration/jobs/{id}
```

Response:
```json
{
  "id": 1,
  "source_type": "excel",
  "source_identifier": "academic_groups.xlsx",
  "status": "running",
  "total": 155000,
  "processed": 50000,
  "failed": 100,
  "started_at": "2024-01-01T10:00:00Z",
  "completed_at": null
}
```

### List All Migration Jobs
```
GET /migration/jobs
```

## Configuration

Environment variables:
- `SERVER_PORT`: HTTP server port (default: 8084)
- `DB_HOST`: Database host (default: localhost)
- `DB_PORT`: Database port (default: 5436)
- `DB_USER`: Database user (default: postgres)
- `DB_PASSWORD`: Database password (default: postgres)
- `DB_NAME`: Database name (default: migration_db)
- `CHAT_SERVICE_URL`: Chat service URL (default: localhost:9092)
- `STRUCTURE_SERVICE_URL`: Structure service URL (default: localhost:9093)
- `GOOGLE_CREDENTIALS_PATH`: Path to Google service account credentials
- `GOOGLE_SPREADSHEET_ID`: Google Sheets spreadsheet ID

## Database Schema

### migration_jobs
- `id`: Primary key
- `source_type`: 'database', 'google_sheets', or 'excel'
- `source_identifier`: File path or sheet ID
- `status`: 'pending', 'running', 'completed', or 'failed'
- `total`: Total records to process
- `processed`: Successfully processed records
- `failed`: Failed records
- `started_at`: Migration start time
- `completed_at`: Migration completion time

### migration_errors
- `id`: Primary key
- `job_id`: Foreign key to migration_jobs
- `record_identifier`: Identifier for the failed record
- `error_message`: Error description
- `created_at`: Error timestamp

## Running the Service

### Local Development
```bash
# Run migrations
make migrate-up

# Start service
make run
```

### Docker
```bash
# Build image
make docker-build

# Run with docker-compose
make docker-run
```

## Migration Process

### Database Migration (admin_panel)
1. Read chat data from source database
2. Lookup university by INN
3. Create chat with source='admin_panel'
4. Add administrator if phone provided

### Google Sheets Migration (bot_registrar)
1. Authenticate with Google Sheets API
2. Read rows with INN, KPP, URL, admin phone
3. Lookup university by INN+KPP
4. Create chat with source='bot_registrar'
5. Add administrator if phone provided

### Excel Migration (academic_group)
1. Parse Excel file with structure data
2. Create structure hierarchy via Structure Service
3. Create chat with source='academic_group'
4. Link group to chat
5. Add administrator if phone provided

## Error Handling

- All errors are logged and recorded in migration_errors table
- Failed records don't stop the migration process
- Progress is updated periodically (every 50-100 records)
- Final report includes total, processed, and failed counts

## Dependencies

- PostgreSQL for migration tracking
- Chat Service for creating chats
- Structure Service for creating structure hierarchy
- Google Sheets API for Google Sheets migration
- Excelize library for Excel parsing
