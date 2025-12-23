# Руководство по развертыванию системы "Цифровой Вуз"

Полное руководство по развертыванию микросервисной архитектуры в production окружении.

## Содержание

1. [Требования](#требования)
2. [Подготовка окружения](#подготовка-окружения)
3. [Настройка баз данных](#настройка-баз-данных)
4. [Конфигурация сервисов](#конфигурация-сервисов)
5. [Развертывание](#развертывание)
6. [Миграция данных](#миграция-данных)
7. [Мониторинг](#мониторинг)
8. [Масштабирование](#масштабирование)
9. [Backup и восстановление](#backup-и-восстановление)
10. [Безопасность](#безопасность)

## Требования

### Минимальные требования

- **CPU**: 4 cores
- **RAM**: 8GB (рекомендуется 16GB)
- **Disk**: 100GB SSD
- **OS**: Ubuntu 20.04+ / CentOS 8+ / macOS
- **Docker**: 20.10+
- **Docker Compose**: 2.0+
- **PostgreSQL**: 15+

### Рекомендуемые требования для production

- **CPU**: 8+ cores
- **RAM**: 32GB
- **Disk**: 500GB SSD (с RAID для надежности)
- **Network**: 1Gbps
- **Load Balancer**: Nginx / HAProxy / Kong
- **Monitoring**: Prometheus + Grafana
- **Logging**: ELK Stack / Loki

## Подготовка окружения

### 1. Установка Docker

```bash
# Ubuntu/Debian
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh
sudo usermod -aG docker $USER

# Установка Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose
```

### 2. Клонирование репозитория

```bash
git clone https://github.com/your-org/digital-university.git
cd digital-university
```

### 3. Создание директорий

```bash
# Создание директорий для данных
sudo mkdir -p /var/lib/digital-university/{postgres,logs,backups,uploads}
sudo chown -R $USER:$USER /var/lib/digital-university

# Создание директории для SSL сертификатов
sudo mkdir -p /etc/digital-university/certs
```

## Настройка баз данных

### 1. Установка PostgreSQL 15

```bash
# Ubuntu/Debian
sudo apt update
sudo apt install postgresql-15 postgresql-contrib-15

# Запуск PostgreSQL
sudo systemctl start postgresql
sudo systemctl enable postgresql
```

### 2. Создание баз данных

```bash
# Подключение к PostgreSQL
sudo -u postgres psql

-- Создание пользователя
CREATE USER digital_university WITH PASSWORD 'strong_password_here';

-- Создание баз данных
CREATE DATABASE auth_db OWNER digital_university;
CREATE DATABASE employee_db OWNER digital_university;
CREATE DATABASE chat_db OWNER digital_university;
CREATE DATABASE structure_db OWNER digital_university;
CREATE DATABASE migration_db OWNER digital_university;

-- Предоставление прав
GRANT ALL PRIVILEGES ON DATABASE auth_db TO digital_university;
GRANT ALL PRIVILEGES ON DATABASE employee_db TO digital_university;
GRANT ALL PRIVILEGES ON DATABASE chat_db TO digital_university;
GRANT ALL PRIVILEGES ON DATABASE structure_db TO digital_university;
GRANT ALL PRIVILEGES ON DATABASE migration_db TO digital_university;

\q
```

### 3. Настройка PostgreSQL для production

Отредактируйте `/etc/postgresql/15/main/postgresql.conf`:

```conf
# Производительность
shared_buffers = 4GB                    # 25% от RAM
effective_cache_size = 12GB             # 75% от RAM
maintenance_work_mem = 1GB
work_mem = 64MB
max_connections = 200

# Логирование
log_min_duration_statement = 1000       # Логировать запросы > 1s
log_line_prefix = '%t [%p]: [%l-1] user=%u,db=%d,app=%a,client=%h '
log_checkpoints = on
log_connections = on
log_disconnections = on
log_lock_waits = on

# Репликация (опционально)
wal_level = replica
max_wal_senders = 3
```

Перезапустите PostgreSQL:

```bash
sudo systemctl restart postgresql
```

### 4. Применение миграций

```bash
# Проверка миграций
./verify_migrations.sh

# Применение миграций для каждого сервиса
cd auth-service && make migrate-up && cd ..
cd employee-service && make migrate-up && cd ..
cd chat-service && make migrate-up && cd ..
cd structure-service && make migrate-up && cd ..
cd migration-service && make migrate-up && cd ..
```

## Конфигурация сервисов

### 1. Создание production конфигурации

Создайте `.env.production` файл в корне проекта:

```bash
# Общие настройки
ENVIRONMENT=production
LOG_LEVEL=info

# PostgreSQL
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_USER=digital_university
POSTGRES_PASSWORD=strong_password_here

# JWT секреты (сгенерируйте сильные ключи!)
JWT_ACCESS_SECRET=$(openssl rand -base64 32)
JWT_REFRESH_SECRET=$(openssl rand -base64 32)

# MAX Messenger API
MAX_BOT_TOKEN=your_production_max_bot_token
MAX_API_URL=https://api.max.ru

# Google Sheets API (для миграции)
GOOGLE_SHEETS_CREDENTIALS_PATH=/etc/digital-university/google-credentials.json

# TLS/SSL
TLS_ENABLED=true
TLS_CERT_PATH=/etc/digital-university/certs/server.crt
TLS_KEY_PATH=/etc/digital-university/certs/server.key
```

### 2. Генерация SSL сертификатов

```bash
# Самоподписанный сертификат (для тестирования)
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
  -keyout /etc/digital-university/certs/server.key \
  -out /etc/digital-university/certs/server.crt

# Для production используйте Let's Encrypt
sudo apt install certbot
sudo certbot certonly --standalone -d your-domain.com
```

### 3. Создание docker-compose.prod.yml

```yaml
version: '3.8'

services:
  auth-service:
    image: digital-university/auth-service:latest
    restart: always
    env_file: .env.production
    environment:
      - DATABASE_URL=postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${POSTGRES_HOST}:5432/auth_db?sslmode=require
      - HTTP_ADDR=:8080
      - GRPC_PORT=9090
    ports:
      - "8080:8080"
      - "9090:9090"
    volumes:
      - /var/lib/digital-university/logs:/app/logs
      - /etc/digital-university/certs:/certs:ro
    deploy:
      resources:
        limits:
          cpus: '1.0'
          memory: 1G
        reservations:
          cpus: '0.5'
          memory: 512M
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s

  employee-service:
    image: digital-university/employee-service:latest
    restart: always
    env_file: .env.production
    environment:
      - DATABASE_URL=postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${POSTGRES_HOST}:5433/employee_db?sslmode=require
      - PORT=8081
      - GRPC_PORT=9091
      - AUTH_SERVICE_GRPC=auth-service:9090
      - MAXBOT_SERVICE_GRPC=maxbot-service:9095
    ports:
      - "8081:8081"
      - "9091:9091"
    volumes:
      - /var/lib/digital-university/logs:/app/logs
    depends_on:
      - auth-service
      - maxbot-service
    deploy:
      resources:
        limits:
          cpus: '1.5'
          memory: 2G
        reservations:
          cpus: '0.75'
          memory: 1G

  chat-service:
    image: digital-university/chat-service:latest
    restart: always
    env_file: .env.production
    environment:
      - DATABASE_URL=postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${POSTGRES_HOST}:5434/chat_db?sslmode=require
      - PORT=8082
      - GRPC_PORT=9092
      - AUTH_SERVICE_GRPC=auth-service:9090
      - MAXBOT_SERVICE_GRPC=maxbot-service:9095
    ports:
      - "8082:8082"
      - "9092:9092"
    volumes:
      - /var/lib/digital-university/logs:/app/logs
    depends_on:
      - auth-service
      - maxbot-service
    deploy:
      resources:
        limits:
          cpus: '2.0'
          memory: 2G
        reservations:
          cpus: '1.0'
          memory: 1G

  structure-service:
    image: digital-university/structure-service:latest
    restart: always
    env_file: .env.production
    environment:
      - DATABASE_URL=postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${POSTGRES_HOST}:5435/structure_db?sslmode=require
      - PORT=8083
      - GRPC_PORT=9093
      - CHAT_SERVICE_GRPC=chat-service:9092
      - EMPLOYEE_SERVICE_GRPC=employee-service:9091
    ports:
      - "8083:8083"
      - "9093:9093"
    volumes:
      - /var/lib/digital-university/logs:/app/logs
      - /var/lib/digital-university/uploads:/app/uploads
    depends_on:
      - chat-service
      - employee-service
    deploy:
      resources:
        limits:
          cpus: '1.0'
          memory: 1G

  maxbot-service:
    image: digital-university/maxbot-service:latest
    restart: always
    env_file: .env.production
    environment:
      - GRPC_PORT=9095
    ports:
      - "9095:9095"
    volumes:
      - /var/lib/digital-university/logs:/app/logs
    deploy:
      resources:
        limits:
          cpus: '1.0'
          memory: 512M

  migration-service:
    image: digital-university/migration-service:latest
    restart: always
    env_file: .env.production
    environment:
      - DATABASE_URL=postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${POSTGRES_HOST}:5436/migration_db?sslmode=require
      - PORT=8084
      - CHAT_SERVICE_GRPC=chat-service:9092
      - STRUCTURE_SERVICE_GRPC=structure-service:9093
    ports:
      - "8084:8084"
    volumes:
      - /var/lib/digital-university/logs:/app/logs
      - /var/lib/digital-university/uploads:/app/uploads
    depends_on:
      - chat-service
      - structure-service
    deploy:
      resources:
        limits:
          cpus: '2.0'
          memory: 4G

networks:
  default:
    driver: bridge
```

## Развертывание

### 1. Сборка Docker образов

```bash
# Сборка всех образов
docker-compose -f docker-compose.prod.yml build

# Или сборка отдельных сервисов
docker build -t digital-university/auth-service:latest ./auth-service
docker build -t digital-university/employee-service:latest ./employee-service
docker build -t digital-university/chat-service:latest ./chat-service
docker build -t digital-university/structure-service:latest ./structure-service
docker build -t digital-university/maxbot-service:latest ./maxbot-service
docker build -t digital-university/migration-service:latest ./migration-service
```

### 2. Запуск сервисов

```bash
# Запуск всех сервисов
docker-compose -f docker-compose.prod.yml up -d

# Проверка статуса
docker-compose -f docker-compose.prod.yml ps

# Просмотр логов
docker-compose -f docker-compose.prod.yml logs -f
```

### 3. Проверка работоспособности

```bash
# Скрипт проверки всех health endpoints
#!/bin/bash
services=("8080" "8081" "8082" "8083" "8084")
for port in "${services[@]}"; do
  echo "Checking service on port $port..."
  response=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:$port/health)
  if [ "$response" == "200" ]; then
    echo "✓ Service on port $port is healthy"
  else
    echo "✗ Service on port $port is unhealthy (HTTP $response)"
    exit 1
  fi
done
echo "All services are healthy!"
```

### 4. Blue-Green Deployment

```bash
# 1. Запуск новой версии (green) параллельно со старой (blue)
docker-compose -f docker-compose.prod.yml up -d --scale auth-service=2

# 2. Проверка health checks новых инстансов
./check_health.sh

# 3. Переключение трафика (через load balancer)
# Обновите конфигурацию Nginx/HAProxy

# 4. Остановка старых инстансов
docker-compose -f docker-compose.prod.yml scale auth-service=1

# 5. Откат при необходимости
docker-compose -f docker-compose.prod.yml.backup up -d
```

## Миграция данных

### 1. Миграция из базы данных (6,000 чатов)

```bash
# Запуск миграции
curl -X POST http://localhost:8084/migration/database \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "source_db_url": "postgres://user:pass@old-host:5432/old_db"
  }'

# Мониторинг прогресса
watch -n 5 'curl -s http://localhost:8084/migration/jobs/1 | jq .'
```

### 2. Миграция из Google Sheets

```bash
# Настройка Google Sheets API credentials
# 1. Создайте service account в Google Cloud Console
# 2. Скачайте JSON credentials
# 3. Поместите в /etc/digital-university/google-credentials.json

# Запуск миграции
curl -X POST http://localhost:8084/migration/google-sheets \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "sheet_id": "1abc...xyz",
    "credentials_path": "/etc/digital-university/google-credentials.json"
  }'
```

### 3. Миграция из Excel (155,000+ чатов)

```bash
# Загрузка Excel файла
curl -X POST http://localhost:8084/migration/excel \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -F "file=@academic_groups.xlsx"

# Для больших файлов рекомендуется разбить на части
split -l 10000 academic_groups.xlsx part_

# Загрузка по частям
for file in part_*; do
  echo "Uploading $file..."
  curl -X POST http://localhost:8084/migration/excel \
    -H "Authorization: Bearer $ADMIN_TOKEN" \
    -F "file=@$file"
  sleep 60  # Пауза между загрузками
done
```

## Мониторинг

### 1. Настройка Prometheus

Создайте `prometheus.yml`:

```yaml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

scrape_configs:
  - job_name: 'auth-service'
    static_configs:
      - targets: ['localhost:8080']
    metrics_path: '/metrics'

  - job_name: 'employee-service'
    static_configs:
      - targets: ['localhost:8081']

  - job_name: 'chat-service'
    static_configs:
      - targets: ['localhost:8082']

  - job_name: 'structure-service'
    static_configs:
      - targets: ['localhost:8083']

  - job_name: 'migration-service'
    static_configs:
      - targets: ['localhost:8084']
```

### 2. Настройка Grafana

```bash
# Запуск Grafana
docker run -d -p 3000:3000 \
  -v /var/lib/digital-university/grafana:/var/lib/grafana \
  grafana/grafana

# Доступ: http://localhost:3000 (admin/admin)
```

Импортируйте дашборды:
- Go application metrics
- PostgreSQL metrics
- gRPC metrics

### 3. Настройка алертов

Создайте `alerts.yml`:

```yaml
groups:
  - name: digital_university_alerts
    interval: 30s
    rules:
      - alert: HighErrorRate
        expr: rate(http_requests_total{status=~"5.."}[5m]) > 0.05
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "High error rate detected"

      - alert: ServiceDown
        expr: up == 0
        for: 2m
        labels:
          severity: critical
        annotations:
          summary: "Service {{ $labels.job }} is down"

      - alert: HighResponseTime
        expr: histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m])) > 1
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "High response time on {{ $labels.job }}"
```

## Масштабирование

### 1. Горизонтальное масштабирование

```bash
# Увеличение количества инстансов
docker-compose -f docker-compose.prod.yml up -d --scale employee-service=3
docker-compose -f docker-compose.prod.yml up -d --scale chat-service=3
```

### 2. Настройка Load Balancer (Nginx)

Создайте `/etc/nginx/conf.d/digital-university.conf`:

```nginx
upstream auth_service {
    least_conn;
    server localhost:8080 max_fails=3 fail_timeout=30s;
    server localhost:8180 max_fails=3 fail_timeout=30s;
}

upstream employee_service {
    least_conn;
    server localhost:8081 max_fails=3 fail_timeout=30s;
    server localhost:8181 max_fails=3 fail_timeout=30s;
    server localhost:8281 max_fails=3 fail_timeout=30s;
}

upstream chat_service {
    least_conn;
    server localhost:8082 max_fails=3 fail_timeout=30s;
    server localhost:8182 max_fails=3 fail_timeout=30s;
    server localhost:8282 max_fails=3 fail_timeout=30s;
}

server {
    listen 80;
    server_name your-domain.com;

    # Redirect to HTTPS
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name your-domain.com;

    ssl_certificate /etc/letsencrypt/live/your-domain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/your-domain.com/privkey.pem;

    # Security headers
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;

    # Rate limiting
    limit_req_zone $binary_remote_addr zone=api_limit:10m rate=100r/m;
    limit_req zone=api_limit burst=20 nodelay;

    location /auth/ {
        proxy_pass http://auth_service/;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    location /employees/ {
        proxy_pass http://employee_service/;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }

    location /chats/ {
        proxy_pass http://chat_service/;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

### 3. Database Connection Pooling

Настройте connection pooling в каждом сервисе:

```go
// config/database.go
db.SetMaxOpenConns(25)           // Максимум открытых соединений
db.SetMaxIdleConns(5)            // Максимум idle соединений
db.SetConnMaxLifetime(5 * time.Minute)  // Время жизни соединения
db.SetConnMaxIdleTime(10 * time.Minute) // Время idle
```

## Backup и восстановление

### 1. Автоматический backup

Создайте `/usr/local/bin/backup-digital-university.sh`:

```bash
#!/bin/bash

BACKUP_DIR="/var/lib/digital-university/backups"
DATE=$(date +%Y%m%d_%H%M%S)
RETENTION_DAYS=30

# Создание директории для backup
mkdir -p "$BACKUP_DIR/$DATE"

# Backup всех баз данных
databases=("auth_db" "employee_db" "chat_db" "structure_db" "migration_db")
for db in "${databases[@]}"; do
    echo "Backing up $db..."
    pg_dump -U digital_university -h localhost $db | gzip > "$BACKUP_DIR/$DATE/${db}.sql.gz"
done

# Backup конфигурации
tar -czf "$BACKUP_DIR/$DATE/config.tar.gz" \
    /etc/digital-university \
    .env.production \
    docker-compose.prod.yml

# Удаление старых backup'ов
find "$BACKUP_DIR" -type d -mtime +$RETENTION_DAYS -exec rm -rf {} +

echo "Backup completed: $BACKUP_DIR/$DATE"
```

Добавьте в crontab:

```bash
# Ежедневный backup в 2:00 AM
0 2 * * * /usr/local/bin/backup-digital-university.sh >> /var/log/backup.log 2>&1
```

### 2. Восстановление из backup

```bash
#!/bin/bash

BACKUP_DATE=$1  # Например: 20240115_020000
BACKUP_DIR="/var/lib/digital-university/backups/$BACKUP_DATE"

if [ ! -d "$BACKUP_DIR" ]; then
    echo "Backup not found: $BACKUP_DIR"
    exit 1
fi

# Остановка сервисов
docker-compose -f docker-compose.prod.yml down

# Восстановление баз данных
databases=("auth_db" "employee_db" "chat_db" "structure_db" "migration_db")
for db in "${databases[@]}"; do
    echo "Restoring $db..."
    dropdb -U digital_university $db
    createdb -U digital_university $db
    gunzip < "$BACKUP_DIR/${db}.sql.gz" | psql -U digital_university $db
done

# Восстановление конфигурации
tar -xzf "$BACKUP_DIR/config.tar.gz" -C /

# Запуск сервисов
docker-compose -f docker-compose.prod.yml up -d

echo "Restore completed from: $BACKUP_DIR"
```

## Безопасность

### 1. Checklist безопасности

- [ ] Используются сильные пароли и JWT секреты (минимум 32 символа)
- [ ] Включен SSL/TLS для всех соединений
- [ ] Настроен firewall (ufw/iptables)
- [ ] Ограничен доступ к PostgreSQL (только localhost или VPN)
- [ ] Включена валидация всех входных данных
- [ ] Используются parameterized SQL queries
- [ ] Настроен rate limiting
- [ ] Включено логирование всех операций
- [ ] Настроены alerts для критических событий
- [ ] Регулярные backup баз данных
- [ ] Обновления безопасности применяются регулярно
- [ ] Используется принцип наименьших привилегий

### 2. Настройка Firewall

```bash
# UFW (Ubuntu)
sudo ufw default deny incoming
sudo ufw default allow outgoing
sudo ufw allow 22/tcp    # SSH
sudo ufw allow 80/tcp    # HTTP
sudo ufw allow 443/tcp   # HTTPS
sudo ufw enable

# Ограничение доступа к PostgreSQL
sudo ufw deny 5432/tcp
```

### 3. Настройка fail2ban

```bash
# Установка
sudo apt install fail2ban

# Создание /etc/fail2ban/jail.local
[DEFAULT]
bantime = 3600
findtime = 600
maxretry = 5

[nginx-limit-req]
enabled = true
filter = nginx-limit-req
logpath = /var/log/nginx/error.log
```

### 4. Регулярные обновления

```bash
# Создайте скрипт обновления
#!/bin/bash
docker-compose -f docker-compose.prod.yml pull
docker-compose -f docker-compose.prod.yml up -d
docker image prune -f
```

## Заключение

Следуя этому руководству, вы сможете развернуть систему "Цифровой Вуз" в production окружении с высокой доступностью, безопасностью и производительностью.

Для дополнительной информации обратитесь к:
- [README.md](./README.md) - Общая документация
- [MIGRATIONS.md](./MIGRATIONS.md) - Руководство по миграциям БД
- [MONITORING_AND_LOGGING_IMPLEMENTATION.md](./MONITORING_AND_LOGGING_IMPLEMENTATION.md) - Мониторинг и логирование
- [.kiro/specs/digital-university-mvp-completion/](./. kiro/specs/digital-university-mvp-completion/) - Спецификации системы
