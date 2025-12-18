#!/bin/bash

# Скрипт для загрузки Excel файла в migration-service

if [ $# -eq 0 ]; then
    echo "Использование: $0 <путь_к_excel_файлу>"
    echo "Пример: $0 data/Минобр\ 2025-12-14.xlsx"
    exit 1
fi

EXCEL_FILE="$1"

if [ ! -f "$EXCEL_FILE" ]; then
    echo "Ошибка: Файл '$EXCEL_FILE' не найден"
    exit 1
fi

echo "Загрузка Excel файла: $EXCEL_FILE"
echo "=================================="

# Проверяем, что migration-service работает
echo "Проверка статуса migration-service..."
if ! curl -s http://localhost:8084/health > /dev/null; then
    echo "Ошибка: migration-service недоступен на localhost:8084"
    exit 1
fi

echo "✓ Migration-service работает"

# Загружаем файл
echo "Загрузка файла..."
response=$(curl -X POST -F "file=@$EXCEL_FILE" http://localhost:8084/migration/excel 2>/dev/null)
echo "Ответ сервера: $response"

# Ждем немного для обработки
echo "Ожидание обработки..."
sleep 5

# Проверяем статус миграций
echo "Статус миграций:"
echo "================"
curl -s http://localhost:8084/migration/jobs | jq '.[0]' 2>/dev/null || curl -s http://localhost:8084/migration/jobs

echo -e "\n"
echo "Готово! Проверьте статус миграции выше."
echo "Для просмотра всех миграций: curl -s http://localhost:8084/migration/jobs | jq ."