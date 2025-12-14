#!/bin/bash

# Быстрая проверка тестов (без race detector для скорости)
# Использование: ./tests/test_quick.sh [service-name]

set -e

# Цвета
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

SERVICE=$1

if [ -z "$SERVICE" ]; then
    echo -e "${BLUE}Быстрая проверка всех тестов...${NC}\n"
    
    SERVICES=("auth-service" "chat-service" "employee-service" "structure-service" "maxbot-service" "migration-service")
    FAILED=0
    
    for service in "${SERVICES[@]}"; do
        if [ -d "$service" ]; then
            echo -e "${YELLOW}Testing $service...${NC}"
            cd "$service"
            if go test ./... > /dev/null 2>&1; then
                echo -e "${GREEN}✓ $service${NC}"
            else
                echo -e "${RED}✗ $service${NC}"
                FAILED=$((FAILED + 1))
            fi
            cd ..
        fi
    done
    
    echo ""
    if [ $FAILED -eq 0 ]; then
        echo -e "${GREEN}✅ Все тесты прошли!${NC}"
        exit 0
    else
        echo -e "${RED}❌ Провалено сервисов: $FAILED${NC}"
        exit 1
    fi
else
    if [ ! -d "$SERVICE" ]; then
        echo -e "${RED}Сервис $SERVICE не найден${NC}"
        exit 1
    fi
    
    echo -e "${BLUE}Тестирование $SERVICE...${NC}\n"
    cd "$SERVICE"
    go test -v ./...
fi
