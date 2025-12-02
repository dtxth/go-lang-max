#!/bin/bash

# Скрипт для запуска всех тестов перед запуском сервисов
# Использование: ./run_tests.sh [--verbose] [--coverage]

set -e  # Остановить выполнение при первой ошибке

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Флаги
VERBOSE=false
COVERAGE=false
FAILED_TESTS=()
TOTAL_TESTS=0
PASSED_TESTS=0

# Парсинг аргументов
for arg in "$@"; do
    case $arg in
        --verbose|-v)
            VERBOSE=true
            shift
            ;;
        --coverage|-c)
            COVERAGE=true
            shift
            ;;
        --help|-h)
            echo "Использование: ./run_tests.sh [опции]"
            echo ""
            echo "Опции:"
            echo "  --verbose, -v    Подробный вывод тестов"
            echo "  --coverage, -c   Генерация отчета о покрытии кода"
            echo "  --help, -h       Показать эту справку"
            exit 0
            ;;
    esac
done

# Функция для вывода заголовка
print_header() {
    echo -e "\n${BLUE}========================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}========================================${NC}\n"
}

# Функция для вывода результата
print_result() {
    if [ $1 -eq 0 ]; then
        echo -e "${GREEN}✓ $2${NC}"
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        echo -e "${RED}✗ $2${NC}"
        FAILED_TESTS+=("$2")
    fi
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
}

# Функция для запуска тестов в сервисе
run_service_tests() {
    local service_name=$1
    local service_path=$2
    
    echo -e "${YELLOW}Тестирование: $service_name${NC}"
    
    if [ ! -d "$service_path" ]; then
        echo -e "${RED}Директория $service_path не найдена${NC}"
        return 1
    fi
    
    cd "$service_path"
    
    # Проверяем наличие тестов
    if ! find . -name "*_test.go" -type f | grep -q .; then
        echo -e "${YELLOW}Тесты не найдены в $service_name${NC}"
        cd - > /dev/null
        return 0
    fi
    
    # Запускаем тесты
    if [ "$COVERAGE" = true ]; then
        if [ "$VERBOSE" = true ]; then
            go test -v -race -coverprofile=coverage.out ./... 2>&1
        else
            go test -race -coverprofile=coverage.out ./... 2>&1
        fi
        test_result=$?
        
        if [ $test_result -eq 0 ]; then
            coverage=$(go tool cover -func=coverage.out | grep total | awk '{print $3}')
            echo -e "${GREEN}Покрытие кода: $coverage${NC}"
        fi
    else
        if [ "$VERBOSE" = true ]; then
            go test -v -race ./... 2>&1
        else
            go test -race ./... 2>&1
        fi
        test_result=$?
    fi
    
    cd - > /dev/null
    return $test_result
}

# Начало выполнения
print_header "🧪 ЗАПУСК ВСЕХ ТЕСТОВ"

echo -e "${BLUE}Режим:${NC}"
echo -e "  Подробный вывод: $VERBOSE"
echo -e "  Покрытие кода: $COVERAGE"
echo ""

# Сохраняем текущую директорию
ORIGINAL_DIR=$(pwd)

# 1. Auth Service
print_header "1/6 Auth Service"
run_service_tests "Auth Service" "auth-service"
print_result $? "Auth Service"

# 2. Chat Service
print_header "2/6 Chat Service"
run_service_tests "Chat Service" "chat-service"
print_result $? "Chat Service"

# 3. Employee Service
print_header "3/6 Employee Service"
run_service_tests "Employee Service" "employee-service"
print_result $? "Employee Service"

# 4. Structure Service
print_header "4/6 Structure Service"
run_service_tests "Structure Service" "structure-service"
print_result $? "Structure Service"

# 5. MaxBot Service
print_header "5/6 MaxBot Service"
run_service_tests "MaxBot Service" "maxbot-service"
print_result $? "MaxBot Service"

# 6. Migration Service
print_header "6/6 Migration Service"
run_service_tests "Migration Service" "migration-service"
print_result $? "Migration Service"

# Возвращаемся в исходную директорию
cd "$ORIGINAL_DIR"

# Итоговый отчет
print_header "📊 ИТОГОВЫЙ ОТЧЕТ"

echo -e "${BLUE}Всего сервисов протестировано:${NC} $TOTAL_TESTS"
echo -e "${GREEN}Успешно:${NC} $PASSED_TESTS"
echo -e "${RED}Провалено:${NC} ${#FAILED_TESTS[@]}"

if [ ${#FAILED_TESTS[@]} -gt 0 ]; then
    echo -e "\n${RED}Провалившиеся тесты:${NC}"
    for test in "${FAILED_TESTS[@]}"; do
        echo -e "  ${RED}✗${NC} $test"
    done
    echo ""
    echo -e "${RED}❌ ТЕСТЫ НЕ ПРОШЛИ! Запуск сервисов отменен.${NC}"
    exit 1
else
    echo ""
    echo -e "${GREEN}✅ ВСЕ ТЕСТЫ ПРОШЛИ УСПЕШНО!${NC}"
    echo -e "${GREEN}Можно запускать сервисы: docker-compose up -d${NC}"
    exit 0
fi
