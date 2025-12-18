#!/bin/bash

echo "Тестирование structure-service..."

# Тестируем создание университета
echo "Создание университета..."
curl -X POST http://localhost:8083/universities \
  -H "Content-Type: application/json" \
  -d '{
    "name": "НГТУ НЭТИ",
    "inn": "5404105174",
    "kpp": "540401001",
    "foiv": "Новосибирская область"
  }'

echo -e "\n"

# Проверяем, что университет создался
echo "Проверка университетов..."
curl -s http://localhost:8083/universities | jq .

echo -e "\n"

# Проверяем в БД
echo "Проверка в БД:"
docker-compose exec structure-db psql -U postgres -d structure_db -c "SELECT * FROM universities;"