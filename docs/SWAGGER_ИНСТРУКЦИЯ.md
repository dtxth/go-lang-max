# 📚 Swagger - Краткая Инструкция

## 🚀 Быстрый старт

### 1. Запустите сервисы
```bash
docker-compose up -d
```

### 2. Откройте Swagger UI

| Сервис | Адрес |
|--------|-------|
| Auth | http://localhost:8080/swagger/index.html |
| Employee | http://localhost:8081/swagger/index.html |
| Chat | http://localhost:8082/swagger/index.html |
| Structure | http://localhost:8083/swagger/index.html |
| Migration | http://localhost:8084/swagger/index.html |

## 📝 Типы данных

Всего определено **31 тип** в 5 сервисах:

- Auth Service: 2 типа
- Employee Service: 8 типов
- Chat Service: 8 типов
- Structure Service: 9 типов
- Migration Service: 4 типа

**Полное описание:** [docs/SWAGGER_TYPES_SUMMARY.md](docs/SWAGGER_TYPES_SUMMARY.md)

## 🔄 Обновление документации

После изменения кода:

```bash
./update_swagger.sh
```

Или для конкретного сервиса:

```bash
cd employee-service
swag init -g cmd/employee/main.go -o internal/infrastructure/http/docs
```

## 📖 Документация

- [Swagger Types Summary](docs/SWAGGER_TYPES_SUMMARY.md) - Все типы
- [Swagger Endpoints](docs/SWAGGER_ENDPOINTS.md) - Все эндпоинты
- [Swagger Quick Reference](SWAGGER_QUICK_REFERENCE.md) - Быстрая справка
- [Swagger Validation Report](docs/SWAGGER_VALIDATION_REPORT.md) - Отчет

## ✅ Статус

Все типы данных корректно определены и актуальны.
