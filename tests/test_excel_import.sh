#!/bin/bash

# Скрипт для тестирования Excel импорта

set -e

echo "🧪 Тестирование Excel импорта..."
echo ""

# Цвета для вывода
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Функция для вывода результата
print_result() {
    if [ $1 -eq 0 ]; then
        echo -e "${GREEN}✅ $2${NC}"
    else
        echo -e "${RED}❌ $2${NC}"
        exit 1
    fi
}

# 1. Unit-тесты для Excel обработки
echo "📋 Запуск unit-тестов для Excel обработки..."
cd migration-service
go test -v ./internal/usecase/migrate_from_excel_test.go ./internal/usecase/migrate_from_excel.go \
    -run "TestReadFromExcel|TestNormalizePhone|TestProcessRow" 2>&1 | tee /tmp/excel_unit_tests.log
UNIT_RESULT=$?
cd ..
print_result $UNIT_RESULT "Unit-тесты Excel обработки"
echo ""

# 2. Тесты загрузки файлов
echo "📤 Запуск тестов загрузки файлов..."
cd migration-service
go test -v ./internal/infrastructure/http/upload_test.go \
    -run "TestUpload|TestMultipart|TestExcelFile" 2>&1 | tee /tmp/excel_upload_tests.log
UPLOAD_RESULT=$?
cd ..
print_result $UPLOAD_RESULT "Тесты загрузки файлов"
echo ""

# 3. Проверка зависимостей
echo "📦 Проверка зависимостей..."
cd migration-service
go mod tidy
go mod verify
DEP_RESULT=$?
cd ..
print_result $DEP_RESULT "Зависимости"
echo ""

# 4. Проверка компиляции
echo "🔨 Проверка компиляции migration-service..."
cd migration-service
go build -o /tmp/migration-service ./cmd/main.go
BUILD_RESULT=$?
cd ..
print_result $BUILD_RESULT "Компиляция migration-service"
echo ""

# 5. Статистика тестов
echo "📊 Статистика тестов:"
echo ""

if [ -f /tmp/excel_unit_tests.log ]; then
    UNIT_PASSED=$(grep -c "PASS:" /tmp/excel_unit_tests.log || echo "0")
    UNIT_FAILED=$(grep -c "FAIL:" /tmp/excel_unit_tests.log || echo "0")
    echo "  Unit-тесты:"
    echo "    ✅ Пройдено: $UNIT_PASSED"
    echo "    ❌ Провалено: $UNIT_FAILED"
fi

if [ -f /tmp/excel_upload_tests.log ]; then
    UPLOAD_PASSED=$(grep -c "PASS:" /tmp/excel_upload_tests.log || echo "0")
    UPLOAD_FAILED=$(grep -c "FAIL:" /tmp/excel_upload_tests.log || echo "0")
    echo "  Тесты загрузки:"
    echo "    ✅ Пройдено: $UPLOAD_PASSED"
    echo "    ❌ Провалено: $UPLOAD_FAILED"
fi

echo ""
echo -e "${GREEN}✅ Все тесты пройдены успешно!${NC}"
echo ""
echo "📝 Логи сохранены в:"
echo "  - /tmp/excel_unit_tests.log"
echo "  - /tmp/excel_upload_tests.log"
echo ""
echo "🚀 Готово к импорту Excel файлов!"
