#!/bin/bash

# Диагностика Excel файла

echo "🔍 Диагностика Excel файла в migration-service..."
echo ""

# Проверяем файлы
echo "📁 Файлы в /tmp/migration-uploads/:"
docker-compose exec migration-service ls -lh /tmp/migration-uploads/
echo ""

# Проверяем последний job
echo "📊 Последний completed job:"
curl -s http://localhost:8084/migration/jobs | jq '.[] | select(.status == "completed") | {id, source_identifier, total, processed, failed, started_at}' | head -20
echo ""

# Проверяем логи
echo "📝 Логи migration-service (последние 50 строк с фильтром):"
docker-compose logs migration-service --tail=50 | grep -E "Excel|rows|total|sheet|Skipping|completed"
echo ""

echo "✅ Диагностика завершена"
