#!/bin/bash

echo "Тестирование загрузки Excel файла в migration-service..."

# Перезапускаем migration-service для свежего соединения с БД
echo "Перезапуск migration-service..."
docker-compose restart migration-service

# Ждем запуска сервиса
echo "Ожидание запуска сервиса..."
sleep 15

# Проверяем health
echo "Проверка health endpoint..."
curl -s http://localhost:8084/health

echo -e "\n"

# Загружаем Excel файл
echo "Загрузка Excel файла..."
response=$(curl -X POST -F "file=@data/Минобр 2025-12-14.xlsx" http://localhost:8084/migration/excel 2>/dev/null)
echo "Ответ сервера: $response"

# Ждем немного для обработки
sleep 5

# Проверяем статус миграции
echo -e "\nПроверка статуса миграций..."
curl -s http://localhost:8084/migration/jobs | jq . 2>/dev/null || curl -s http://localhost:8084/migration/jobs

echo -e "\n"

# Проверяем логи migration-service
echo "Последние логи migration-service:"
docker-compose logs migration-service | tail -10

echo -e "\nТест завершен."