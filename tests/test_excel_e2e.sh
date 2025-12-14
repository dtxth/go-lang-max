#!/bin/bash

# E2E тест для Excel импорта

set -e

echo "🧪 E2E тест Excel импорта..."
echo ""

# Цвета
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Проверка что сервисы запущены
echo "1️⃣ Проверка сервисов..."
if ! docker-compose ps | grep -q "migration-service.*Up"; then
    echo -e "${RED}❌ migration-service не запущен${NC}"
    exit 1
fi
echo -e "${GREEN}✅ Сервисы запущены${NC}"
echo ""

# Очистка БД перед тестом
echo "2️⃣ Очистка БД..."
docker-compose exec -T chat-db psql -U chat_user -d chat_db -c "TRUNCATE chats, administrators CASCADE;" > /dev/null 2>&1 || true
docker-compose exec -T structure-db psql -U postgres -d postgres -c "TRUNCATE universities, branches, faculties, groups CASCADE;" > /dev/null 2>&1 || true
echo -e "${GREEN}✅ БД очищены${NC}"
echo ""

# Создание тестового Excel файла
echo "3️⃣ Создание тестового Excel файла..."
cat > /tmp/create_test_excel.py << 'PYTHON_EOF'
#!/usr/bin/env python3
from openpyxl import Workbook

wb = Workbook()
ws = wb.active

# Header
headers = [
    "Phone1", "MaxID", "INN_Ref", "FOIV", "OrgName", "Branch",
    "INN", "KPP", "Faculty", "Course", "Group", "ChatName",
    "Phone2", "FileName", "ChatID", "Link", "AddUser", "AddAdmin",
]
ws.append(headers)

# Data rows
data_rows = [
    [
        "79884753064", "496728250", "105014177", "Минобрнауки России",
        "МГТУ Тест E2E", "Головной филиал", "105014177", "10501001",
        "Политехнический колледж МГТУ", "2", "Колледж ИП-22",
        "Колледж ИП-22 (2024 ОФО МГТУ", "79884753064", "file.xlsx",
        "-69257108032233", "https://max.ru/join/test1", "ИСТИНА", "ИСТИНА",
    ],
    [
        "79001234567", "123456789", "105014177", "Минобрнауки России",
        "МГТУ Тест E2E", "Головной филиал", "105014177", "10501001",
        "Факультет информатики", "3", "ИВТ-31",
        "Группа ИВТ-31", "79001234567", "file.xlsx",
        "-12345678", "https://max.ru/join/test2", "TRUE", "TRUE",
    ],
]

for row in data_rows:
    ws.append(row)

wb.save("/tmp/test_e2e.xlsx")
print("✅ Test Excel file created: /tmp/test_e2e.xlsx")
PYTHON_EOF

python3 /tmp/create_test_excel.py 2>/dev/null || {
    echo -e "${YELLOW}⚠️  Python/openpyxl не установлен, используем существующий файл${NC}"
    if [ ! -f "/tmp/test_real.xlsx" ]; then
        echo -e "${RED}❌ Нет тестового файла${NC}"
        exit 1
    fi
    cp /tmp/test_real.xlsx /tmp/test_e2e.xlsx
}
echo -e "${GREEN}✅ Тестовый файл создан${NC}"
echo ""

# Загрузка файла
echo "4️⃣ Загрузка Excel файла..."
RESPONSE=$(curl -s -X POST http://localhost:8084/migration/excel \
  -F "file=@/tmp/test_e2e.xlsx")

echo "Ответ: $RESPONSE"

# Извлечение job_id (если есть в ответе)
JOB_ID=$(curl -s http://localhost:8084/migration/jobs | jq -r '.[0].id' 2>/dev/null || echo "")

if [ -z "$JOB_ID" ] || [ "$JOB_ID" = "null" ]; then
    echo -e "${RED}❌ Не удалось получить job_id${NC}"
    echo "Проверьте логи: docker-compose logs migration-service --tail=50"
    exit 1
fi

echo -e "${GREEN}✅ Файл загружен, Job ID: $JOB_ID${NC}"
echo ""

# Ожидание завершения
echo "5️⃣ Ожидание завершения обработки..."
MAX_WAIT=60
WAITED=0

while [ $WAITED -lt $MAX_WAIT ]; do
    STATUS=$(curl -s http://localhost:8084/migration/jobs/$JOB_ID | jq -r '.status' 2>/dev/null || echo "unknown")
    TOTAL=$(curl -s http://localhost:8084/migration/jobs/$JOB_ID | jq -r '.total' 2>/dev/null || echo "0")
    PROCESSED=$(curl -s http://localhost:8084/migration/jobs/$JOB_ID | jq -r '.processed' 2>/dev/null || echo "0")
    FAILED=$(curl -s http://localhost:8084/migration/jobs/$JOB_ID | jq -r '.failed' 2>/dev/null || echo "0")
    
    echo "  Статус: $STATUS, Total: $TOTAL, Processed: $PROCESSED, Failed: $FAILED"
    
    if [ "$STATUS" = "completed" ]; then
        echo -e "${GREEN}✅ Обработка завершена${NC}"
        break
    fi
    
    if [ "$STATUS" = "failed" ]; then
        echo -e "${RED}❌ Обработка провалилась${NC}"
        exit 1
    fi
    
    sleep 2
    WAITED=$((WAITED + 2))
done

if [ $WAITED -ge $MAX_WAIT ]; then
    echo -e "${YELLOW}⚠️  Превышено время ожидания${NC}"
fi
echo ""

# Проверка логов
echo "6️⃣ Проверка логов migration-service..."
docker-compose logs migration-service --tail=30 | grep -E "Excel|Streaming|progress|completed" || echo "Нет релевантных логов"
echo ""

# Проверка данных в БД
echo "7️⃣ Проверка данных в БД..."
echo ""

echo "📊 Structure DB:"
UNIV_COUNT=$(docker-compose exec -T structure-db psql -U postgres -d postgres -t -c "SELECT COUNT(*) FROM universities;" 2>/dev/null | tr -d ' ')
BRANCH_COUNT=$(docker-compose exec -T structure-db psql -U postgres -d postgres -t -c "SELECT COUNT(*) FROM branches;" 2>/dev/null | tr -d ' ')
FACULTY_COUNT=$(docker-compose exec -T structure-db psql -U postgres -d postgres -t -c "SELECT COUNT(*) FROM faculties;" 2>/dev/null | tr -d ' ')
GROUP_COUNT=$(docker-compose exec -T structure-db psql -U postgres -d postgres -t -c "SELECT COUNT(*) FROM groups;" 2>/dev/null | tr -d ' ')

echo "  Universities: $UNIV_COUNT"
echo "  Branches: $BRANCH_COUNT"
echo "  Faculties: $FACULTY_COUNT"
echo "  Groups: $GROUP_COUNT"
echo ""

echo "📊 Chat DB:"
CHAT_COUNT=$(docker-compose exec -T chat-db psql -U chat_user -d chat_db -t -c "SELECT COUNT(*) FROM chats;" 2>/dev/null | tr -d ' ')
ADMIN_COUNT=$(docker-compose exec -T chat-db psql -U chat_user -d chat_db -t -c "SELECT COUNT(*) FROM administrators;" 2>/dev/null | tr -d ' ')
EXTERNAL_ID_COUNT=$(docker-compose exec -T chat-db psql -U chat_user -d chat_db -t -c "SELECT COUNT(*) FROM chats WHERE external_chat_id IS NOT NULL;" 2>/dev/null | tr -d ' ')

echo "  Chats: $CHAT_COUNT"
echo "  Administrators: $ADMIN_COUNT"
echo "  Chats with external_chat_id: $EXTERNAL_ID_COUNT"
echo ""

# Проверка результатов
echo "8️⃣ Проверка результатов..."
ERRORS=0

if [ "$TOTAL" -eq 0 ]; then
    echo -e "${RED}❌ Total = 0 (файл не был обработан)${NC}"
    ERRORS=$((ERRORS + 1))
else
    echo -e "${GREEN}✅ Total > 0 ($TOTAL строк)${NC}"
fi

if [ "$CHAT_COUNT" -eq 0 ]; then
    echo -e "${RED}❌ Нет чатов в БД${NC}"
    ERRORS=$((ERRORS + 1))
else
    echo -e "${GREEN}✅ Чаты созданы ($CHAT_COUNT)${NC}"
fi

if [ "$UNIV_COUNT" -eq 0 ]; then
    echo -e "${RED}❌ Нет университетов в БД${NC}"
    ERRORS=$((ERRORS + 1))
else
    echo -e "${GREEN}✅ Университеты созданы ($UNIV_COUNT)${NC}"
fi

echo ""

if [ $ERRORS -eq 0 ]; then
    echo -e "${GREEN}✅ E2E тест пройден успешно!${NC}"
    exit 0
else
    echo -e "${RED}❌ E2E тест провален ($ERRORS ошибок)${NC}"
    echo ""
    echo "Для диагностики:"
    echo "  docker-compose logs migration-service --tail=100"
    echo "  docker-compose logs chat-service --tail=50"
    echo "  docker-compose logs structure-service --tail=50"
    exit 1
fi
