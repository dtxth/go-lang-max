# Swagger Quick Reference

## 🚀 Быстрый доступ к Swagger UI

| Сервис | URL | Порт |
|--------|-----|------|
| Auth Service | http://localhost:8080/swagger/index.html | 8080 |
| Employee Service | http://localhost:8081/swagger/index.html | 8081 |
| Chat Service | http://localhost:8082/swagger/index.html | 8082 |
| Structure Service | http://localhost:8083/swagger/index.html | 8083 |
| Migration Service | http://localhost:8084/swagger/index.html | 8084 |

## 📝 Обновление документации

### Обновить все сервисы
```bash
./update_swagger.sh
```

### Обновить конкретный сервис
```bash
cd <service-name>
swag init -g cmd/<service>/main.go -o internal/infrastructure/http/docs
```

## 📊 Статистика типов

- **Auth Service**: 2 типа
- **Employee Service**: 8 типов
- **Chat Service**: 8 типов
- **Structure Service**: 9 типов
- **Migration Service**: 4 типа

**Всего**: 31 тип данных

## 📚 Документация

- **Полное описание типов**: [docs/SWAGGER_TYPES_SUMMARY.md](docs/SWAGGER_TYPES_SUMMARY.md)
- **История исправлений**: [docs/SWAGGER_TYPES_FIX.md](docs/SWAGGER_TYPES_FIX.md)
- **Список эндпоинтов**: [docs/SWAGGER_ENDPOINTS.md](docs/SWAGGER_ENDPOINTS.md)

## ✅ Статус

Все типы данных корректно определены в Swagger документации. Документация актуальна и соответствует коду.
