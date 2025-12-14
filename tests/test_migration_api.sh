#!/bin/bash

# Простой тест API migration-service

echo "🧪 Тестирование Migration Service API..."
echo ""

# 1. Проверка здоровья
echo "1️⃣ Проверка здоровья сервиса..."
curl -s http://localhost:8084/health || echo "❌ Сервис недоступен"
echo ""
echo ""

# 2. Проверка списка jobs
echo "2️⃣ Получение списка migration jobs..."
curl -s http://localhost:8084/migration/jobs | jq '.' || echo "[]"
echo ""
echo ""

# 3. Проверка что chat-service доступен
echo "3️⃣ Проверка chat-service..."
curl -s http://localhost:8082/chats | jq '.total' || echo "❌ Chat service недоступен"
echo ""
echo ""

# 4. Проверка что structure-service доступен
echo "4️⃣ Проверка structure-service..."
curl -s http://localhost:8083/universities | jq 'length' || echo "❌ Structure service недоступен"
echo ""
echo ""

# 5. Проверка данных в БД
echo "5️⃣ Проверка данных в БД..."
echo ""

echo "📊 Chat DB:"
echo "  Chats:"
docker-compose exec -T chat-db psql -U chat_user -d chat_db -c "SELECT id, name, external_chat_id FROM chats LIMIT 5;" 2>/dev/null || echo "    Нет данных"
echo ""
echo "  Administrators:"
docker-compose exec -T chat-db psql -U chat_user -d chat_db -c "SELECT id, chat_id, phone, max_id, add_user, add_admin FROM administrators LIMIT 5;" 2>/dev/null || echo "    Нет данных"
echo ""

echo "📊 Structure DB:"
echo "  Universities:"
docker-compose exec -T structure-db psql -U postgres -d postgres -c "SELECT id, name, inn FROM universities LIMIT 5;" 2>/dev/null || echo "    Нет данных"
echo ""

echo "✅ Проверка завершена"
