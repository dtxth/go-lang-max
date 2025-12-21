#!/bin/bash

# Тестовый скрипт для MAX API аутентификации
# Использует токен из .env файла: MAX_BOT_TOKEN=test_bot_token_12345

echo "=== Тестирование MAX API аутентификации ==="
echo

# Полный тестовый init_data с всеми полями
echo "1. Тест с полными данными пользователя:"
curl -X 'POST' \
  'http://localhost:8080/auth/max' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
    "init_data": "first_name=%D0%98%D0%B2%D0%B0%D0%BD&hash=e9ca048be27dc6671a40203fa8ab37ac8d6a764a54f70b85a708ce6f11980273&last_name=%D0%9F%D0%B5%D1%82%D1%80%D0%BE%D0%B2&max_id=123456789&username=ivan_petrov"
  }'

echo -e "\n\n2. Тест с минимальными данными:"
curl -X 'POST' \
  'http://localhost:8080/auth/max' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
    "init_data": "first_name=%D0%A2%D0%B5%D1%81%D1%82&hash=1a8a7b589bc022fca6be81661e8183f27b18c9ee1034ae095c0bf54c876b93f4&max_id=987654321"
  }'

echo -e "\n\n3. Тест с невалидным хешем (должен вернуть ошибку):"
curl -X 'POST' \
  'http://localhost:8080/auth/max' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
    "init_data": "first_name=Test&hash=invalid_hash&max_id=123"
  }'

echo -e "\n\n4. Тест с пустым init_data (должен вернуть ошибку):"
curl -X 'POST' \
  'http://localhost:8080/auth/max' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
    "init_data": ""
  }'

echo -e "\n\n=== Тестирование завершено ==="