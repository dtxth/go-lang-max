#!/bin/bash

# Скрипт для переименования миграций в формат migrate tool

rename_migrations() {
    local service=$1
    local migrations_dir="$service/migrations"
    
    if [ ! -d "$migrations_dir" ]; then
        echo "⚠ Directory $migrations_dir not found"
        return
    fi
    
    echo "=== Processing $service ==="
    
    cd "$migrations_dir"
    
    # Переименовываем файлы в формат 000001_name.up.sql и 000001_name.down.sql
    for file in *.sql; do
        if [[ $file =~ ^([0-9]+)_(.+)_down\.sql$ ]]; then
            # Down migration
            num=$(printf "%06d" ${BASH_REMATCH[1]})
            name=${BASH_REMATCH[2]}
            new_name="${num}_${name}.down.sql"
            if [ "$file" != "$new_name" ]; then
                mv "$file" "$new_name"
                echo "  ✓ $file -> $new_name"
            fi
        elif [[ $file =~ ^([0-9]+)_(.+)\.sql$ ]]; then
            # Up migration
            num=$(printf "%06d" ${BASH_REMATCH[1]})
            name=${BASH_REMATCH[2]}
            new_name="${num}_${name}.up.sql"
            if [ "$file" != "$new_name" ]; then
                mv "$file" "$new_name"
                echo "  ✓ $file -> $new_name"
            fi
        fi
    done
    
    cd - > /dev/null
}

# Переименовываем миграции для всех сервисов
rename_migrations "auth-service"
rename_migrations "employee-service"
rename_migrations "chat-service"
rename_migrations "migration-service"

echo ""
echo "✅ Migration files renamed successfully"
