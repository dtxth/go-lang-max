# Docker Build Quick Start

## Быстрая пересборка всех сервисов

```bash
# 1. Пересобрать все сервисы без кеша
docker-compose build --no-cache

# 2. Запустить все сервисы
docker-compose up -d

# 3. Проверить статус
docker ps

# 4. Проверить логи конкретного сервиса
docker logs chat-service
```

## Пересборка отдельного сервиса

```bash
# Пересобрать только chat-service
docker-compose build --no-cache chat-service

# Перезапустить сервис
docker-compose up -d chat-service
```

## Проверка Swagger endpoints

```bash
# Проверить все Swagger endpoints
for port in 8080 8081 8082 8083 8084; do
  echo "Port $port:"
  curl -s http://localhost:$port/swagger/doc.json | jq -r '.info.title' 2>/dev/null || echo "Not ready"
  echo ""
done
```

## Полная очистка и пересборка

```bash
# 1. Остановить и удалить все контейнеры
docker-compose down

# 2. Удалить все образы
docker images | grep "go-lang-max" | awk '{print $3}' | xargs docker rmi -f

# 3. Пересобрать все с нуля
docker-compose build --no-cache

# 4. Запустить
docker-compose up -d
```

## Проверка образов

```bash
# Посмотреть все собранные образы
docker images | grep "go-lang-max"

# Ожидаемый результат:
# go-lang-max-migration-service   latest   44.5MB
# go-lang-max-structure-service   latest   40.8MB
# go-lang-max-chat-service        latest   37.9MB
# go-lang-max-employee-service    latest   37.9MB
# go-lang-max-auth-service        latest   36.9MB
# go-lang-max-maxbot-service      latest   25.8MB
```

## Отладка проблем сборки

```bash
# Посмотреть логи сборки конкретного сервиса
docker-compose build chat-service 2>&1 | tee build.log

# Проверить последние 50 строк логов
docker-compose build chat-service 2>&1 | tail -50

# Проверить ошибки в логах
docker-compose build chat-service 2>&1 | grep -i error
```

## Локальная разработка

```bash
# После изменения зависимостей в go.mod
cd chat-service
go mod tidy

# Проверить, что код компилируется локально
go build ./cmd/chat

# Запустить тесты
go test ./...
```

## Структура зависимостей

```
chat-service
├── зависит от: auth-service, maxbot-service
└── использует: auth.proto, maxbot.proto

employee-service
├── зависит от: auth-service, maxbot-service
└── использует: auth.proto, maxbot.proto

structure-service
├── зависит от: chat-service, employee-service
└── использует: chat.proto, employee.proto

migration-service
└── не имеет gRPC зависимостей
```

## Troubleshooting

### Ошибка: "package X is not in std"

**Причина:** Go не может найти локальные пакеты из других сервисов.

**Решение:**
1. Проверьте, что в go.mod есть replace директивы
2. Запустите `go mod tidy`
3. Пересоберите с `--no-cache`

### Ошибка: "replacement directory does not exist"

**Причина:** Replace директивы указывают на неправильные пути в Docker.

**Решение:**
1. Проверьте, что контекст сборки в docker-compose.yml - корневая директория (`.`)
2. Проверьте, что в Dockerfile есть sed команды для замены путей
3. Пересоберите образ

### Ошибка: "proto file not found"

**Причина:** Proto файлы не скопированы в Docker образ.

**Решение:**
1. Проверьте, что в Dockerfile есть COPY команды для proto файлов
2. Проверьте, что proto файлы генерируются с правильными путями
3. Пересоберите образ

## Полезные команды

```bash
# Посмотреть размеры образов
docker images | grep "go-lang-max" | awk '{print $1, $7$8}'

# Удалить неиспользуемые образы
docker image prune -a

# Посмотреть использование диска Docker
docker system df

# Очистить все (осторожно!)
docker system prune -a --volumes
```
