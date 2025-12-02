# API Tests Quick Reference

## Быстрый запуск

```bash
# Все тесты API хендлеров
./test_api_handlers.sh

# Конкретный сервис
cd <service-name> && go test -v ./internal/infrastructure/http/
```

## Покрытие тестами

### Auth Service (11 тестов)
| Endpoint | Метод | Тесты |
|----------|-------|-------|
| `/register` | POST | Invalid JSON, Missing email, Missing password |
| `/login` | POST | Invalid JSON, Missing email, Missing password |
| `/refresh` | POST | Invalid JSON, Missing token |
| `/logout` | POST | Invalid JSON, Missing token |
| `/health` | GET | Success |

### Employee Service (13 тестов)
| Endpoint | Метод | Тесты |
|----------|-------|-------|
| `/employees` | GET | Missing auth, Invalid auth format |
| `/employees` | POST | Missing phone, Missing name, Invalid JSON |
| `/employees/{id}` | GET | Invalid ID |
| `/employees/{id}` | PUT | Invalid ID, Invalid JSON |
| `/employees/{id}` | DELETE | Invalid ID |
| `/employees/batch-update-maxid` | POST | Service not available |
| `/employees/batch-status` | GET | Service not available |
| `/employees/batch-status/{id}` | GET | Service not available, Invalid ID |

### Chat Service (8 тестов)
| Endpoint | Метод | Тесты |
|----------|-------|-------|
| `/chats` | GET | Unauthorized |
| `/chats/all` | GET | Unauthorized |
| `/chats/{id}` | GET | Invalid ID |
| `/chats/{chat_id}/administrators` | POST | Missing phone, Invalid chat ID, Invalid JSON, Invalid path |
| `/administrators/{admin_id}` | DELETE | Invalid ID |

### Structure Service (7 тестов)
| Endpoint | Метод | Тесты |
|----------|-------|-------|
| `/universities` | POST | Invalid JSON |
| `/universities/{id}` | GET | Invalid ID |
| `/universities/{university_id}/structure` | GET | Invalid ID |
| `/departments/managers` | POST | Invalid JSON |
| `/departments/managers/{id}` | DELETE | Invalid ID |
| `/import/excel` | POST | Invalid method, Missing file |

### Migration Service (10 тестов)
| Endpoint | Метод | Тесты |
|----------|-------|-------|
| `/migration/database` | POST | Invalid method, Invalid JSON |
| `/migration/google-sheets` | POST | Invalid method, Invalid JSON, Missing spreadsheet_id |
| `/migration/excel` | POST | Invalid method, Missing file |
| `/migration/jobs/{id}` | GET | Invalid method, Invalid ID |
| `/migration/jobs` | GET | Invalid method |

## Итого

- **Сервисов:** 5
- **Endpoints:** 29
- **Тестов:** 49
- **Статус:** ✅ All tests passing

## Документация

- [API_TESTS_COVERAGE.md](./API_TESTS_COVERAGE.md) - Полная документация
- [API_TESTING_SUMMARY.md](./API_TESTING_SUMMARY.md) - Сводка по работе
- [README.md](./README.md) - Основная документация проекта
