#!/bin/bash

# Скрипт для тестирования загрузки Excel файла

set -e

echo "🧪 Тестирование загрузки Excel файла..."
echo ""

# Создаем тестовый Excel файл с помощью Go
cat > /tmp/create_test_excel.go << 'EOF'
package main

import (
	"fmt"
	"github.com/xuri/excelize/v2"
)

func main() {
	f := excelize.NewFile()
	defer f.Close()

	sheetName := "Sheet1"

	// Header row
	headers := []string{
		"Phone1", "MaxID", "INN_Ref", "FOIV", "OrgName", "Branch",
		"INN", "KPP", "Faculty", "Course", "Group", "ChatName",
		"Phone2", "FileName", "ChatID", "Link", "AddUser", "AddAdmin",
	}

	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheetName, cell, header)
	}

	// Data row
	dataRow := []interface{}{
		"79884753064", "496728250", "105014177", "Минобрнауки России",
		"МГТУ", "Головной филиал", "105014177", "10501001",
		"Политехнический колледж МГТУ", "2", "Колледж ИП-22",
		"Колледж ИП-22 (2024 ОФО МГТУ", "79884753064", "file.xlsx",
		"-69257108032233", "https://max.ru/join/test", "ИСТИНА", "ИСТИНА",
	}

	for colIdx, value := range dataRow {
		cell, _ := excelize.CoordinatesToCellName(colIdx+1, 2)
		f.SetCellValue(sheetName, cell, value)
	}

	if err := f.SaveAs("/tmp/test_import.xlsx"); err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println("✅ Test Excel file created: /tmp/test_import.xlsx")
}
EOF

# Создаем файл
echo "📝 Создание тестового Excel файла..."
cd /tmp && go run create_test_excel.go

if [ ! -f /tmp/test_import.xlsx ]; then
    echo "❌ Не удалось создать тестовый файл"
    exit 1
fi

echo "✅ Тестовый файл создан"
echo ""

# Загружаем файл
echo "📤 Загрузка файла в migration-service..."
RESPONSE=$(curl -s -X POST http://localhost:8084/migration/excel \
  -F "file=@/tmp/test_import.xlsx")

echo "Ответ сервера:"
echo "$RESPONSE" | jq '.' 2>/dev/null || echo "$RESPONSE"
echo ""

# Извлекаем job_id
JOB_ID=$(echo "$RESPONSE" | jq -r '.job_id' 2>/dev/null || echo "")

if [ -z "$JOB_ID" ] || [ "$JOB_ID" = "null" ]; then
    echo "❌ Не удалось получить job_id"
    echo "Проверьте логи: docker-compose logs migration-service"
    exit 1
fi

echo "✅ Job ID: $JOB_ID"
echo ""

# Ждем завершения
echo "⏳ Ожидание завершения импорта..."
for i in {1..30}; do
    sleep 2
    STATUS=$(curl -s http://localhost:8084/migration/jobs/$JOB_ID | jq -r '.status' 2>/dev/null || echo "")
    PROCESSED=$(curl -s http://localhost:8084/migration/jobs/$JOB_ID | jq -r '.processed' 2>/dev/null || echo "0")
    FAILED=$(curl -s http://localhost:8084/migration/jobs/$JOB_ID | jq -r '.failed' 2>/dev/null || echo "0")
    
    echo "  Статус: $STATUS, Обработано: $PROCESSED, Ошибок: $FAILED"
    
    if [ "$STATUS" = "completed" ] || [ "$STATUS" = "failed" ]; then
        break
    fi
done

echo ""
echo "📊 Финальный статус:"
curl -s http://localhost:8084/migration/jobs/$JOB_ID | jq '.'
echo ""

# Проверяем что попало в БД
echo "🔍 Проверка данных в БД..."
echo ""

echo "📊 Structure DB:"
docker-compose exec -T structure-db psql -U postgres -d postgres -c "SELECT COUNT(*) as universities FROM universities;"
docker-compose exec -T structure-db psql -U postgres -d postgres -c "SELECT COUNT(*) as branches FROM branches;"
docker-compose exec -T structure-db psql -U postgres -d postgres -c "SELECT COUNT(*) as faculties FROM faculties;"
docker-compose exec -T structure-db psql -U postgres -d postgres -c "SELECT COUNT(*) as groups FROM groups;"
echo ""

echo "📊 Chat DB:"
docker-compose exec -T chat-db psql -U chat_user -d chat_db -c "SELECT COUNT(*) as chats FROM chats;"
docker-compose exec -T chat-db psql -U chat_user -d chat_db -c "SELECT COUNT(*) as administrators FROM administrators;"
echo ""

echo "✅ Тест завершен!"
