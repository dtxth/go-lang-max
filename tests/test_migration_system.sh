#!/bin/bash

# Тест новой системы миграций
# Проверяет, что все сервисы корректно запускают миграции

set -e

echo "🔄 Тестирование новой системы миграций..."

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Функция для проверки логов миграций
check_migration_logs() {
    local service=$1
    echo -e "${BLUE}Проверка миграций для $service...${NC}"
    
    # Ждем запуска сервиса
    sleep 5
    
    # Проверяем логи на наличие сообщений о миграциях
    if docker-compose logs $service 2>/dev/null | grep -q "Starting database migrations"; then
        echo -e "${GREEN}✓ $service: Миграции запущены${NC}"
    else
        echo -e "${RED}✗ $service: Миграции не найдены в логах${NC}"
        return 1
    fi
    
    if docker-compose logs $service 2>/dev/null | grep -q "Database connection established"; then
        echo -e "${GREEN}✓ $service: Подключение к БД установлено${NC}"
    else
        echo -e "${RED}✗ $service: Нет подключения к БД${NC}"
        return 1
    fi
    
    if docker-compose logs $service 2>/dev/null | grep -q -E "(Successfully migrated|Database is up to date)"; then
        echo -e "${GREEN}✓ $service: Миграции выполнены успешно${NC}"
    else
        echo -e "${RED}✗ $service: Миграции не выполнены${NC}"
        return 1
    fi
}

# Функция для проверки готовности сервиса
check_service_health() {
    local service=$1
    local port=$2
    echo -e "${BLUE}Проверка готовности $service на порту $port...${NC}"
    
    for i in {1..30}; do
        if curl -s -f "http://localhost:$port/health" >/dev/null 2>&1 || \
           curl -s -f "http://localhost:$port/" >/dev/null 2>&1; then
            echo -e "${GREEN}✓ $service готов${NC}"
            return 0
        fi
        sleep 2
    done
    
    echo -e "${YELLOW}⚠ $service не отвечает на порту $port (возможно, нет health endpoint)${NC}"
    return 0  # Не критично для теста миграций
}

echo -e "${BLUE}1. Остановка существующих контейнеров...${NC}"
docker-compose down >/dev/null 2>&1 || true

echo -e "${BLUE}2. Запуск только баз данных...${NC}"
docker-compose up -d auth-db employee-db chat-db structure-db migration-db redis

echo -e "${BLUE}3. Ожидание готовности баз данных...${NC}"
sleep 10

echo -e "${BLUE}4. Запуск сервисов с новой системой миграций...${NC}"
docker-compose up -d auth-service employee-service chat-service structure-service migration-service maxbot-service

echo -e "${BLUE}5. Ожидание запуска сервисов...${NC}"
sleep 15

echo -e "${BLUE}6. Проверка миграций для каждого сервиса...${NC}"
echo ""

# Проверяем миграции для каждого сервиса
services=("auth-service" "employee-service" "chat-service" "structure-service" "migration-service")
failed_services=()

for service in "${services[@]}"; do
    if ! check_migration_logs "$service"; then
        failed_services+=("$service")
    fi
    echo ""
done

echo -e "${BLUE}7. Проверка готовности HTTP endpoints...${NC}"
echo ""

# Проверяем HTTP endpoints
check_service_health "auth-service" "8080"
check_service_health "employee-service" "8081" 
check_service_health "chat-service" "8082"
check_service_health "structure-service" "8083"
check_service_health "migration-service" "8084"
check_service_health "maxbot-service" "8095"

echo ""
echo -e "${BLUE}8. Проверка состояния контейнеров...${NC}"
docker-compose ps

echo ""
echo -e "${BLUE}9. Результаты тестирования:${NC}"

if [ ${#failed_services[@]} -eq 0 ]; then
    echo -e "${GREEN}🎉 Все сервисы успешно запустили миграции!${NC}"
    echo -e "${GREEN}✓ Новая система миграций работает корректно${NC}"
    exit 0
else
    echo -e "${RED}❌ Проблемы с миграциями в сервисах: ${failed_services[*]}${NC}"
    echo -e "${YELLOW}Проверьте логи командой: make logs${NC}"
    exit 1
fi