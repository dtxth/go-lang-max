#!/bin/bash

# Тест gRPC метода AddAdministratorForMigration

echo "=== Testing gRPC AddAdministratorForMigration ==="

# 1. Создать тестовый чат через HTTP
echo "1. Creating test chat..."
CHAT_RESPONSE=$(curl -s -X POST http://localhost:8082/chats \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test Chat for gRPC",
    "url": "https://max.ru/test",
    "source": "admin_panel",
    "participants_count": 0
  }')

echo "Chat response: $CHAT_RESPONSE"
CHAT_ID=$(echo $CHAT_RESPONSE | jq -r '.id')
echo "Created chat ID: $CHAT_ID"

# 2. Проверить, что чат создан
if [ "$CHAT_ID" == "null" ] || [ -z "$CHAT_ID" ]; then
    echo "ERROR: Failed to create chat"
    exit 1
fi

# 3. Подождать немного
sleep 1

# 4. Проверить администраторов до добавления
echo ""
echo "2. Checking administrators before adding..."
docker-compose exec -T chat-db psql -U chat_user -d chat_db -c \
  "SELECT COUNT(*) as count FROM administrators WHERE chat_id = $CHAT_ID;"

# 5. Добавить администратора через HTTP (с skip_phone_validation)
echo ""
echo "3. Adding administrator via HTTP with skip_phone_validation..."
ADMIN_RESPONSE=$(curl -s -X POST http://localhost:8082/chats/$CHAT_ID/administrators \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "+79001234567",
    "max_id": "123456",
    "add_user": true,
    "add_admin": true,
    "skip_phone_validation": true
  }')

echo "Administrator response: $ADMIN_RESPONSE"

# 6. Подождать немного
sleep 1

# 7. Проверить администраторов после добавления
echo ""
echo "4. Checking administrators after adding..."
docker-compose exec -T chat-db psql -U chat_user -d chat_db -c \
  "SELECT id, chat_id, phone, max_id, add_user, add_admin FROM administrators WHERE chat_id = $CHAT_ID;"

# 8. Проверить общее количество
echo ""
echo "5. Total administrators count:"
docker-compose exec -T chat-db psql -U chat_user -d chat_db -c \
  "SELECT COUNT(*) as total_count FROM administrators;"

echo ""
echo "=== Test completed ==="
