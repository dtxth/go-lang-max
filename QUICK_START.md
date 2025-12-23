# Быстрый старт - Цифровой Вуз

## Предварительные требования

- Docker и Docker Compose
- Make (для удобных команд)
- Go 1.24+ (для локальной разработки)

## Запуск за 3 шага

### 1. Настройка окружения

```bash
# Клонируйте репозиторий
git clone <repository-url>
cd digital-university-mvp

# Скопируйте и настройте переменные окружения
cp .env.example .env

# Отредактируйте .env файл:
# - Измените пароли баз данных
# - Установите MAX_BOT_TOKEN (если есть)
# - Для production: измените все SECRET значения
```

### 2. Запуск системы

```bash
# Полный запуск с тестами (рекомендуется)
make deploy

# Или быстрый запуск без тестов
make deploy-fast
```

### 3. Проверка работы

```bash
# Проверка статуса сервисов
make health

# Просмотр логов
make logs

# Остановка системы
make down
```

## Доступные сервисы

После запуска будут доступны:

- **Auth Service**: http://localhost:8080
- **Employee Service**: http://localhost:8081  
- **Chat Service**: http://localhost:8082
- **Structure Service**: http://localhost:8083
- **Migration Service**: http://localhost:8084

## Swagger документация

- Auth API: http://localhost:8080/swagger/index.html
- Employee API: http://localhost:8081/swagger/index.html
- Chat API: http://localhost:8082/swagger/index.html
- Structure API: http://localhost:8083/swagger/index.html
- Migration API: http://localhost:8084/swagger/index.html

## Базы данных

Доступны на портах:
- auth-db: localhost:5432
- employee-db: localhost:5433
- chat-db: localhost:5434
- structure-db: localhost:5435
- migration-db: localhost:5436

## Полезные команды

```bash
# Пересборка всех сервисов
make deploy-rebuild

# Запуск только тестов
make test

# Форматирование кода
make fmt

# Обновление зависимостей
make mod-tidy

# Генерация Swagger документации
make swagger
```

## Troubleshooting

### Проблемы с портами
```bash
# Проверьте занятые порты
lsof -i :8080-8084
lsof -i :5432-5436

# Остановите конфликтующие сервисы
make down
```

### Проблемы с базами данных
```bash
# Очистка volumes
docker-compose down -v
make deploy-rebuild
```

### Проблемы с сетью
```bash
# Пересоздание Docker сети
docker network prune
make deploy-rebuild
```

## Дополнительная документация

- [Полное руководство по развертыванию](./docs/DEPLOYMENT_GUIDE.md)
- [API Reference](./docs/API_REFERENCE.md)
- [Архитектура системы](./docs/SETUP_INSTRUCTIONS.md)
- [Миграция данных](./docs/MIGRATIONS.md)