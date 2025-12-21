# Отчет об исправлении endpoint /login-phone

## Проблема
Endpoint `/login-phone` возвращал ошибку 401 "invalid email or password" при попытке входа с корректными данными.

## Причина
Endpoint `/register` создавал пользователей только с email, но не с телефоном. При попытке входа через `/login-phone` система не могла найти пользователя по номеру телефона.

## Решение

### 1. Обновлен endpoint `/register`
- ✅ Добавлена поддержка регистрации по телефону
- ✅ Теперь принимает либо `email`, либо `phone` (один из них обязателен)
- ✅ Добавлена нормализация номера телефона

### 2. Добавлен новый метод в AuthService
- ✅ `RegisterByPhone(phone, password, role)` - регистрация пользователя по телефону

### 3. Обновлена Swagger документация
- ✅ Endpoint `/register` теперь документирует поддержку `phone` поля
- ✅ Обновлено описание: "Provide either email or phone"

## Тестирование

### Регистрация с телефоном:
```bash
curl -X POST 'http://localhost:8080/register' \
  -H 'Content-Type: application/json' \
  -d '{
    "phone": "+79999999999",
    "password": "tEst123!",
    "role": "operator"
  }'
```

**Результат**: ✅ Успешно создан пользователь с ID 24

### Вход по телефону:
```bash
curl -X POST 'http://localhost:8080/login-phone' \
  -H 'Content-Type: application/json' \
  -d '{
    "password": "tEst123!",
    "phone": "+79999999999"
  }'
```

**Результат**: ✅ Успешно получены access_token и refresh_token

## Изменения в коде

### Файлы изменены:
1. `auth-service/internal/infrastructure/http/handler.go`
   - Обновлен метод `Register()` для поддержки телефона
   - Добавлен импорт `domain` пакета

2. `auth-service/internal/usecase/auth_service.go`
   - Добавлен метод `RegisterByPhone()`

3. `auth-service/internal/infrastructure/http/docs/*`
   - Обновлена Swagger документация

### Конфигурация:
- Исправлена конфигурация базы данных в `.env`
- Сервис запущен через docker-compose для корректной работы с базой данных

## Статус: ✅ ИСПРАВЛЕНО

Endpoint `/login-phone` теперь работает корректно. Пользователи могут регистрироваться с телефоном и входить в систему используя номер телефона и пароль.