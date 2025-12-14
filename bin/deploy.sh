#!/bin/bash

# Полный скрипт развертывания: тесты → сборка → запуск
# Использование: ./bin/deploy.sh [--skip-tests] [--no-cache] [--verbose]

set -e  # Остановить выполнение при первой ошибке

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Флаги
SKIP_TESTS=false
NO_CACHE=false
VERBOSE=false
COVERAGE=false

# Парсинг аргументов
for arg in "$@"; do
    case $arg in
        --skip-tests)
            SKIP_TESTS=true
            shift
            ;;
        --no-cache)
            NO_CACHE=true
            shift
            ;;
        --verbose|-v)
            VERBOSE=true
            shift
            ;;
        --coverage|-c)
            COVERAGE=true
            shift
            ;;
        --help|-h)
            echo "Использование: ./bin/deploy.sh [опции]"
            echo ""
            echo "Опции:"
            echo "  --skip-tests     Пропустить запуск тестов"
            echo "  --no-cache       Пересобрать Docker образы без кеша"
            echo "  --verbose, -v    Подробный вывод"
            echo "  --coverage, -c   Генерация отчета о покрытии кода"
            echo "  --help, -h       Показать эту справку"
            echo ""
            echo "Примеры:"
            echo "  ./bin/deploy.sh                    # Полное развертывание с тестами"
            echo "  ./bin/deploy.sh --skip-tests       # Быстрое развертывание без тестов"
            echo "  ./bin/deploy.sh --no-cache         # Полная пересборка"
            echo "  ./bin/deploy.sh --verbose --coverage  # С подробным выводом и покрытием"
            exit 0
            ;;
    esac
done

# Функция для вывода заголовка
print_header() {
    echo -e "\n${CYAN}╔════════════════════════════════════════╗${NC}"
    echo -e "${CYAN}║${NC} $1"
    echo -e "${CYAN}╚════════════════════════════════════════╝${NC}\n"
}

# Функция для вывода шага
print_step() {
    echo -e "${BLUE}▶${NC} $1"
}

# Функция для вывода успеха
print_success() {
    echo -e "${GREEN}✓${NC} $1"
}

# Функция для вывода ошибки
print_error() {
    echo -e "${RED}✗${NC} $1"
}

# Функция для вывода предупреждения
print_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

# Начало развертывания
clear
print_header "🚀 РАЗВЕРТЫВАНИЕ МИКРОСЕРВИСОВ"

echo -e "${BLUE}Конфигурация:${NC}"
echo -e "  Запуск тестов: $([ "$SKIP_TESTS" = true ] && echo "${YELLOW}НЕТ${NC}" || echo "${GREEN}ДА${NC}")"
echo -e "  Пересборка без кеша: $([ "$NO_CACHE" = true ] && echo "${GREEN}ДА${NC}" || echo "${YELLOW}НЕТ${NC}")"
echo -e "  Подробный вывод: $([ "$VERBOSE" = true ] && echo "${GREEN}ДА${NC}" || echo "${YELLOW}НЕТ${NC}")"
echo -e "  Покрытие кода: $([ "$COVERAGE" = true ] && echo "${GREEN}ДА${NC}" || echo "${YELLOW}НЕТ${NC}")"
echo ""

# Шаг 1: Остановка существующих контейнеров
print_header "📦 ШАГ 1/5: ОСТАНОВКА КОНТЕЙНЕРОВ"
print_step "Остановка существующих контейнеров..."

if docker-compose ps -q | grep -q .; then
    docker-compose down
    print_success "Контейнеры остановлены"
else
    print_warning "Контейнеры не запущены"
fi

# Шаг 2: Запуск тестов
if [ "$SKIP_TESTS" = false ]; then
    print_header "🧪 ШАГ 2/5: ЗАПУСК ТЕСТОВ"
    
    TEST_ARGS=""
    [ "$VERBOSE" = true ] && TEST_ARGS="$TEST_ARGS --verbose"
    [ "$COVERAGE" = true ] && TEST_ARGS="$TEST_ARGS --coverage"
    
    if ./tests/run_tests.sh $TEST_ARGS; then
        print_success "Все тесты прошли успешно"
    else
        print_error "Тесты провалились!"
        echo ""
        echo -e "${RED}Развертывание отменено.${NC}"
        echo -e "${YELLOW}Исправьте ошибки в тестах или используйте --skip-tests для пропуска.${NC}"
        exit 1
    fi
else
    print_header "🧪 ШАГ 2/5: ЗАПУСК ТЕСТОВ"
    print_warning "Тесты пропущены (--skip-tests)"
fi

# Шаг 3: Сборка Docker образов
print_header "🔨 ШАГ 3/5: СБОРКА DOCKER ОБРАЗОВ"

BUILD_ARGS=""
if [ "$NO_CACHE" = true ]; then
    BUILD_ARGS="--no-cache --progress=plain"
    print_step "Пересборка всех образов без кеша..."
    print_warning "Это может занять 5-10 минут. Вывод сборки показан ниже..."
    echo ""
    # При --no-cache всегда показываем вывод, так как процесс долгий
    docker-compose build $BUILD_ARGS
elif [ "$VERBOSE" = true ]; then
    BUILD_ARGS="--progress=plain"
    print_step "Сборка образов..."
    docker-compose build $BUILD_ARGS
else
    print_step "Сборка образов (это может занять время)..."
    docker-compose build $BUILD_ARGS > /dev/null 2>&1
fi

if [ $? -eq 0 ]; then
    echo ""
    print_success "Все образы собраны успешно"
    
    # Показываем размеры образов
    echo ""
    echo -e "${BLUE}Размеры образов:${NC}"
    docker images | grep "go-lang-max" | awk '{printf "  %-35s %s\n", $1, $7$8}'
else
    print_error "Ошибка при сборке образов"
    exit 1
fi

# Шаг 4: Запуск сервисов
print_header "🚀 ШАГ 4/5: ЗАПУСК СЕРВИСОВ"
print_step "Запуск всех сервисов..."

docker-compose up -d

if [ $? -eq 0 ]; then
    print_success "Все сервисы запущены"
else
    print_error "Ошибка при запуске сервисов"
    exit 1
fi

# Шаг 5: Проверка здоровья сервисов
print_header "🏥 ШАГ 5/5: ПРОВЕРКА ЗДОРОВЬЯ"
print_step "Ожидание запуска сервисов (15 секунд)..."

sleep 15

echo ""
echo -e "${BLUE}Статус контейнеров:${NC}"
docker-compose ps

echo ""
echo -e "${BLUE}Проверка Swagger endpoints:${NC}"

check_endpoint() {
    local port=$1
    local service=$2
    
    if curl -s -f "http://localhost:$port/swagger/doc.json" > /dev/null 2>&1; then
        echo -e "  ${GREEN}✓${NC} $service (http://localhost:$port/swagger/index.html)"
        return 0
    else
        echo -e "  ${RED}✗${NC} $service (http://localhost:$port/swagger/index.html)"
        return 1
    fi
}

HEALTHY_COUNT=0
TOTAL_SERVICES=5

check_endpoint 8080 "Auth Service" && HEALTHY_COUNT=$((HEALTHY_COUNT + 1))
check_endpoint 8081 "Employee Service" && HEALTHY_COUNT=$((HEALTHY_COUNT + 1))
check_endpoint 8082 "Chat Service" && HEALTHY_COUNT=$((HEALTHY_COUNT + 1))
check_endpoint 8083 "Structure Service" && HEALTHY_COUNT=$((HEALTHY_COUNT + 1))
check_endpoint 8084 "Migration Service" && HEALTHY_COUNT=$((HEALTHY_COUNT + 1))

# Итоговый отчет
print_header "📊 ИТОГОВЫЙ ОТЧЕТ"

echo -e "${BLUE}Сервисы:${NC}"
echo -e "  Всего: $TOTAL_SERVICES"
echo -e "  Работают: ${GREEN}$HEALTHY_COUNT${NC}"
echo -e "  Не отвечают: ${RED}$((TOTAL_SERVICES - HEALTHY_COUNT))${NC}"

if [ $HEALTHY_COUNT -eq $TOTAL_SERVICES ]; then
    echo ""
    echo -e "${GREEN}╔════════════════════════════════════════╗${NC}"
    echo -e "${GREEN}║  ✅ РАЗВЕРТЫВАНИЕ ЗАВЕРШЕНО УСПЕШНО!  ║${NC}"
    echo -e "${GREEN}╚════════════════════════════════════════╝${NC}"
    echo ""
    echo -e "${CYAN}Полезные команды:${NC}"
    echo -e "  ${YELLOW}docker-compose logs -f [service]${NC}  - Просмотр логов"
    echo -e "  ${YELLOW}docker-compose ps${NC}                 - Статус контейнеров"
    echo -e "  ${YELLOW}docker-compose down${NC}               - Остановка всех сервисов"
    echo -e "  ${YELLOW}./tests/run_tests.sh${NC}              - Запуск тестов"
    echo ""
    exit 0
else
    echo ""
    echo -e "${YELLOW}╔════════════════════════════════════════╗${NC}"
    echo -e "${YELLOW}║  ⚠ РАЗВЕРТЫВАНИЕ ЗАВЕРШЕНО С ОШИБКАМИ ║${NC}"
    echo -e "${YELLOW}╚════════════════════════════════════════╝${NC}"
    echo ""
    echo -e "${CYAN}Проверьте логи проблемных сервисов:${NC}"
    echo -e "  ${YELLOW}docker-compose logs [service-name]${NC}"
    echo ""
    exit 1
fi
