#!/bin/bash

# Скрипт для применения миграций для поддержки Excel импорта

set -e

echo "🔧 Применение миграций для поддержки Excel импорта..."

# Применяем миграцию для chat-service
echo "📊 Применение миграции для chat-service..."
docker-compose exec -T chat-db psql -U chat_user -d chat_db < chat-service/migrations/002_add_excel_fields.sql

echo "✅ Миграции успешно применены!"
echo ""
echo "📋 Добавленные поля:"
echo "  - chats.external_chat_id (TEXT) - ID чата из Excel (колонка 14)"
echo "  - administrators.add_user (BOOLEAN) - Флаг добавления пользователей (колонка 16)"
echo "  - administrators.add_admin (BOOLEAN) - Флаг добавления администраторов (колонка 17)"
echo ""
echo "🚀 Теперь можно пересобрать сервисы:"
echo "   docker-compose up -d --build chat-service migration-service"
