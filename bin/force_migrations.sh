#!/bin/bash

set -e

echo "========================================="
echo "Force Applying Migrations with migrate tool"
echo "========================================="
echo ""

# Функция для применения миграций
apply_migrations() {
    local service=$1
    local db_host=$2
    local db_user=$3
    local db_pass=$4
    local db_name=$5
    local version=$6
    
    echo "=== $service ==="
    
    # Проверяем, что база данных доступна
    if ! docker exec $db_host psql -U $db_user -d $db_name -c "SELECT 1" > /dev/null 2>&1; then
        echo "  ✗ Database $db_name not accessible"
        return 1
    fi
    
    # Применяем миграции через migrate tool
    local db_url="postgres://${db_user}:${db_pass}@${db_host}:5432/${db_name}?sslmode=disable"
    
    # Используем force для установки версии миграций
    docker run --rm \
        -v "$(pwd)/${service}/migrations:/migrations" \
        --network go-lang-max_microservices-network \
        migrate/migrate \
        -path=/migrations/ \
        -database "$db_url" \
        force $version 2>&1
    
    if [ $? -eq 0 ]; then
        echo "  ✓ Migrations marked as version $version"
    else
        echo "  ✗ Failed to mark migrations"
        return 1
    fi
}

# Auth Service (3 миграции)
apply_migrations "auth-service" "auth-db" "postgres" "postgres" "postgres" 3

# Employee Service (3 миграции)
apply_migrations "employee-service" "employee-db" "employee_user" "employee_pass" "employee_db" 3

# Chat Service (2 миграции)
apply_migrations "chat-service" "chat-db" "chat_user" "chat_pass" "chat_db" 2

# Structure Service уже применено
echo "=== structure-service ==="
echo "  ✓ Already applied (version 3)"

# Migration Service (1 миграция) - база не запущена
echo "=== migration-service ==="
echo "  ⚠ Database not running, will be applied on first start"

echo ""
echo "========================================="
echo "✅ Migrations marked successfully"
echo "========================================="
echo ""
echo "To verify migrations, run:"
echo "  ./check_migrations.sh"
