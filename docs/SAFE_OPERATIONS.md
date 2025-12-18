# Безопасные операции с базой данных

## ✅ БЕЗОПАСНЫЕ команды (НЕ затирают данные):

```bash
# Перезапуск сервисов без удаления данных
make restart
docker-compose restart

# Остановка без удаления volumes
make down
docker-compose down

# Просмотр логов
make logs
make logs-auth

# Проверка статуса
make ps
docker-compose ps
```

## ⚠️ ОСТОРОЖНО - могут затереть данные:

```bash
# Удаляет контейнеры, но НЕ volumes
docker-compose down

# Пересборка образов (безопасно для данных)
make deploy-rebuild
docker-compose build --no-cache
```

## 🔴 ОПАСНЫЕ команды (ЗАТИРАЮТ данные):

```bash
# УДАЛЯЕТ ВСЕ VOLUMES!
make clean-volumes
docker-compose down -v

# Удаляет конкретный volume
docker volume rm go-lang-max_auth_pgdata

# Полная очистка
make clean
```

## 🛡️ Резервное копирование базы:

```bash
# Создать бэкап
docker exec auth-db pg_dump -U postgres postgres > auth_backup_$(date +%Y%m%d_%H%M%S).sql

# Восстановить из бэкапа
docker exec -i auth-db psql -U postgres postgres < auth_backup_20231217_120000.sql
```

## 🔍 Диагностика проблем:

```bash
# Проверить volumes
docker volume ls | grep auth

# Проверить таблицы в базе
docker exec auth-db psql -U postgres -d postgres -c "\dt"

# Проверить пользователей
docker exec auth-db psql -U postgres -d postgres -c "SELECT id, phone, role FROM users;"

# Проверить размер базы
docker exec auth-db psql -U postgres -d postgres -c "SELECT pg_size_pretty(pg_database_size('postgres'));"
```