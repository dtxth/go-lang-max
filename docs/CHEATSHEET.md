# Шпаргалка по командам

Быстрый справочник по наиболее используемым командам.

## 🚀 Быстрый старт

```bash
# Полное развертывание (тесты + сборка + запуск)
make deploy

# Быстрое развертывание (без тестов)
make deploy-fast

# Только запустить сервисы
make up

# Остановить сервисы
make down
```

## 🧪 Тестирование

```bash
# Все тесты
make test

# Быстрая проверка
make test-quick

# С подробным выводом
make test-verbose

# С покрытием кода
make test-coverage

# Тест конкретного сервиса
make test-auth
make test-chat
make test-employee
```

## 📦 Docker

```bash
# Собрать образы
make build

# Пересобрать без кеша
make build-no-cache

# Показать образы
make images

# Удалить все
make clean
```

## 📋 Логи и мониторинг

```bash
# Все логи
make logs

# Логи конкретного сервиса
make logs-chat
make logs-employee

# Статус контейнеров
make ps

# Проверка Swagger
make swagger

# Полная проверка здоровья
make health
```

## 🔧 Разработка

```bash
# Форматировать код
make fmt

# Обновить go.mod
make mod-tidy

# Запустить сервис локально
make dev-chat
make dev-employee
```

## 📚 Swagger UI

```
http://localhost:8080/swagger/index.html  # Auth Service
http://localhost:8081/swagger/index.html  # Employee Service
http://localhost:8082/swagger/index.html  # Chat Service
http://localhost:8083/swagger/index.html  # Structure Service
http://localhost:8084/swagger/index.html  # Migration Service
```

## 🐛 Отладка

```bash
# Перезапустить сервис
docker-compose restart chat-service

# Пересобрать и перезапустить
docker-compose up -d --build chat-service

# Войти в контейнер
docker exec -it chat-service sh

# Проверить логи с ошибками
docker-compose logs chat-service | grep -i error
```

## 🔄 Типичные сценарии

### После изменения кода

```bash
make test              # Проверить тесты
make deploy-fast       # Быстро пересобрать и запустить
```

### Полная пересборка

```bash
make clean             # Очистить
make deploy-rebuild    # Пересобрать все
```

### Проверка перед коммитом

```bash
make test              # Тесты
make fmt               # Форматирование
make mod-tidy          # Обновить зависимости
```

### Отладка проблемного сервиса

```bash
make logs-chat         # Смотрим логи
docker-compose restart chat-service  # Перезапускаем
make test-chat         # Проверяем тесты
```

## 💡 Полезные алиасы

Добавьте в `~/.bashrc` или `~/.zshrc`:

```bash
# Алиасы для проекта
alias mtest='make test'
alias mdeploy='make deploy'
alias mup='make up'
alias mdown='make down'
alias mlogs='make logs'
alias mps='make ps'
```

## 📖 Документация

- [Testing and Deployment Guide](./TESTING_AND_DEPLOYMENT.md)
- [Docker Cross-Service Dependencies](./DOCKER_CROSS_SERVICE_DEPENDENCIES.md)
- [Docker Build Quick Start](./DOCKER_BUILD_QUICK_START.md)
- [README](./README.md)
