# Руководство по пересборке Docker образов

Это руководство объясняет, как правильно обновить Docker образы после изменения кода или Swagger документации.

## Проблема

После обновления Swagger аннотаций в коде, изменения не отображаются в Swagger UI, потому что Docker использует закешированный образ со старыми файлами.

## Решение

### Вариант 1: Пересборка конкретного сервиса (быстро)

```bash
# Остановите сервис
docker-compose stop auth-service

# Пересоберите образ БЕЗ кеша
docker-compose build --no-cache auth-service

# Запустите сервис
docker-compose up -d auth-service
```

### Вариант 2: Пересборка всех сервисов (полная)

```bash
# Остановите все сервисы
docker-compose down

# Пересоберите все образы БЕЗ кеша
docker-compose build --no-cache

# Запустите все сервисы
docker-compose up -d
```

### Вариант 3: Полная очистка и пересборка

```bash
# Остановите и удалите все контейнеры, сети и volumes
docker-compose down -v

# Удалите старые образы
docker-compose rm -f
docker rmi $(docker images 'go-lang-max*' -q) 2>/dev/null || true

# Пересоберите все образы
docker-compose build --no-cache

# Запустите все сервисы
docker-compose up -d
```

## Workflow обновления Swagger

### Шаг 1: Обновите Swagger аннотации в коде

```go
// @Summary      Your endpoint summary
// @Description  Detailed description
// @Tags         your-tag
// @Accept       json
// @Produce      json
// @Param        id path int true "ID parameter"
// @Success      200 {object} YourResponseType
// @Failure      400 {object} ErrorResponse
// @Router       /your/endpoint [get]
func (h *Handler) YourEndpoint(w http.ResponseWriter, r *http.Request) {
    // implementation
}
```

### Шаг 2: Регенерируйте Swagger локально (опционально)

```bash
# Обновить все сервисы
./update_swagger.sh

# Или конкретный сервис
cd auth-service
make swagger
```

### Шаг 3: Пересоберите Docker образ

```bash
# Пересоберите конкретный сервис
docker-compose build --no-cache auth-service

# Или все сервисы
docker-compose build --no-cache
```

### Шаг 4: Перезапустите сервисы

```bash
# Перезапустите конкретный сервис
docker-compose up -d auth-service

# Или все сервисы
docker-compose up -d
```

### Шаг 5: Проверьте изменения

```bash
# Откройте Swagger UI в браузере
open http://localhost:8080/swagger/index.html

# Или проверьте через curl
curl http://localhost:8080/swagger/doc.json | jq '.paths | keys'
```

## Автоматическая генерация Swagger в Docker

Все Dockerfile'ы настроены на автоматическую генерацию Swagger документации во время сборки:

```dockerfile
# Генерируем Swagger документацию
RUN go install github.com/swaggo/swag/cmd/swag@latest
RUN swag init -g cmd/auth/main.go -o internal/infrastructure/http/docs
```

Это означает, что при каждой пересборке образа Swagger будет генерироваться заново на основе текущих аннотаций в коде.

## Проверка текущей версии Swagger

### В Docker контейнере:

```bash
# Проверьте endpoints в контейнере
docker exec go-lang-max-auth-service-1 cat /root/internal/infrastructure/http/docs/swagger.json | jq '.paths | keys'
```

### Локально:

```bash
# Проверьте endpoints локально
cat auth-service/internal/infrastructure/http/docs/swagger.json | jq '.paths | keys'
```

Если версии отличаются - нужна пересборка Docker образа.

## Troubleshooting

### Проблема: Изменения не применяются после пересборки

**Решение:**
```bash
# Полная очистка
docker-compose down -v
docker system prune -a --volumes -f

# Пересборка
docker-compose build --no-cache
docker-compose up -d
```

### Проблема: Swagger UI показывает 404

**Проверьте:**
1. Сервис запущен: `docker-compose ps`
2. Логи сервиса: `docker-compose logs auth-service`
3. Swagger файлы существуют в контейнере

### Проблема: Swagger генерация падает в Docker

**Проверьте логи сборки:**
```bash
docker-compose build auth-service 2>&1 | grep -A 10 "swag init"
```

**Возможные причины:**
- Синтаксическая ошибка в Swagger аннотациях
- Отсутствует импорт пакета с типами
- Неправильный путь к main.go

### Проблема: Docker build очень медленный

**Используйте кеш для зависимостей:**

Dockerfile уже оптимизирован:
```dockerfile
# Сначала копируем только go.mod и go.sum
COPY go.mod go.sum ./
RUN go mod download

# Потом копируем код
COPY . .
```

Это позволяет кешировать слой с зависимостями.

## Best Practices

### 1. Разработка

Во время разработки используйте локальный запуск:

```bash
cd auth-service
go run cmd/auth/main.go
```

Swagger будет доступен сразу после изменений (если регенерировать через `make swagger`).

### 2. Тестирование

Перед коммитом обновите Swagger:

```bash
./update_swagger.sh
git add */internal/infrastructure/http/docs/
git commit -m "Update Swagger documentation"
```

### 3. Production

В production используйте конкретные версии образов:

```yaml
services:
  auth-service:
    image: your-registry/auth-service:v1.2.3
    build:
      context: ./auth-service
      dockerfile: Dockerfile
```

## Быстрые команды

```bash
# Пересборка одного сервиса
docker-compose build --no-cache auth-service && docker-compose up -d auth-service

# Пересборка всех сервисов
docker-compose build --no-cache && docker-compose up -d

# Полная очистка и пересборка
docker-compose down -v && docker-compose build --no-cache && docker-compose up -d

# Проверка Swagger endpoints
for port in 8080 8081 8082 8083 8084; do
  echo "Port $port:"
  curl -s http://localhost:$port/swagger/doc.json | jq -r '.paths | keys[]'
  echo ""
done
```

## Дополнительная информация

- [Docker Compose Documentation](https://docs.docker.com/compose/)
- [Swagger Documentation](./SWAGGER_ENDPOINTS.md)
- [Update Swagger Script](./update_swagger.sh)
- [README](./README.md)
