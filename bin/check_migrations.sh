#!/bin/bash

echo "========================================="
echo "Checking Database Migrations"
echo "========================================="
echo ""

# Функция для проверки миграций
check_migrations() {
    local service=$1
    local db_host=$2
    local db_user=$3
    local db_name=$4
    local expected_version=$5
    
    echo "=== $service ==="
    
    # Проверяем доступность базы
    if ! docker exec $db_host psql -U $db_user -d $db_name -c "SELECT 1" > /dev/null 2>&1; then
        echo "  ✗ Database not accessible"
        return 1
    fi
    
    # Проверяем наличие таблицы schema_migrations
    local has_table=$(docker exec $db_host psql -U $db_user -d $db_name -t -c "SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'schema_migrations');" 2>/dev/null | tr -d ' ')
    
    if [ "$has_table" != "t" ]; then
        echo "  ✗ schema_migrations table not found"
        return 1
    fi
    
    # Получаем версию миграций
    local version=$(docker exec $db_host psql -U $db_user -d $db_name -t -c "SELECT version FROM schema_migrations;" 2>/dev/null | tr -d ' ')
    local dirty=$(docker exec $db_host psql -U $db_user -d $db_name -t -c "SELECT dirty FROM schema_migrations;" 2>/dev/null | tr -d ' ')
    
    if [ -z "$version" ]; then
        echo "  ✗ No migration version found"
        return 1
    fi
    
    if [ "$dirty" = "t" ]; then
        echo "  ✗ Migrations are in dirty state (version: $version)"
        return 1
    fi
    
    if [ "$version" = "$expected_version" ]; then
        echo "  ✓ Version $version (expected: $expected_version)"
        
        # Показываем список таблиц
        echo "  Tables:"
        docker exec $db_host psql -U $db_user -d $db_name -t -c "\dt" 2>/dev/null | grep "public" | awk '{print "    - " $3}'
    else
        echo "  ⚠ Version $version (expected: $expected_version)"
    fi
}

# Проверяем все сервисы
check_migrations "auth-service" "auth-db" "postgres" "postgres" 3
echo ""
check_migrations "employee-service" "employee-db" "employee_user" "employee_db" 3
echo ""
check_migrations "chat-service" "chat-db" "chat_user" "chat_db" 2
echo ""
check_migrations "structure-service" "structure-db" "postgres" "structure_db" 3
echo ""

# Migration service
echo "=== migration-service ==="
if docker ps | grep -q migration-db; then
    check_migrations "migration-service" "migration-db" "postgres" "migration_db" 1
else
    echo "  ⚠ Database not running"
fi

echo ""
echo "========================================="
echo "✅ Migration check completed"
echo "========================================="
