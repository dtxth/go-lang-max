# Руководство администратора: Получение паролей из логов

## Обзор

При создании новых сотрудников система автоматически генерирует безопасные пароли и выводит их в логи для администраторов. Это руководство объясняет, как получить эти пароли.

## Когда пароли логируются

### 1. Создание нового сотрудника с ролью
```bash
# Лог в employee-service
Generated password for new employee with phone ending in ****4567: Abc123!@#Def
```

### 2. Создание нового пользователя
```bash
# Лог в auth-service  
Generated password for new user with phone ending in ****4567: Xyz789$%^Ghi
```

### 3. Запрос сброса пароля
```bash
# Лог в auth-service
Generated password reset token for user with phone ending in ****4567: a1b2c3d4e5f6...
```

## Как получить пароли

### Docker Compose (локальная разработка)

```bash
# Просмотр логов employee-service
docker logs employee-service | grep "Generated password"

# Просмотр логов auth-service
docker logs auth-service | grep "Generated password"

# Просмотр токенов сброса
docker logs auth-service | grep "Generated password reset token"

# Поиск по номеру телефона (последние 4 цифры)
docker logs employee-service | grep "****4567"
```

### Kubernetes

```bash
# Получить список подов
kubectl get pods | grep -E "(employee|auth)-service"

# Просмотр логов
kubectl logs deployment/employee-service | grep "Generated password"
kubectl logs deployment/auth-service | grep "Generated password"

# Поиск по номеру телефона
kubectl logs deployment/employee-service | grep "****4567"
```

### Файловые логи (production)

```bash
# Поиск паролей в логах
grep "Generated password" /var/log/employee-service.log
grep "Generated password" /var/log/auth-service.log

# Поиск по номеру телефона
grep "****4567" /var/log/*.log

# Поиск за определенную дату
grep "Generated password" /var/log/*.log | grep "2024-01-15"
```

## Примеры использования

### Сценарий 1: Новый сотрудник не получил пароль

1. **Найдите пароль в логах:**
   ```bash
   docker logs employee-service | grep "****$(echo '+79161234567' | tail -c 5)"
   ```

2. **Результат:**
   ```
   Generated password for new employee with phone ending in ****4567: Abc123!@#Def
   ```

3. **Передайте пароль сотруднику безопасным способом**

### Сценарий 2: Пользователь не получил токен сброса

1. **Найдите токен в логах:**
   ```bash
   docker logs auth-service | grep "Generated password reset token" | grep "****4567"
   ```

2. **Результат:**
   ```
   Generated password reset token for user with phone ending in ****4567: a1b2c3d4e5f6g7h8...
   ```

3. **Используйте токен для сброса пароля через API**

### Сценарий 3: Массовый поиск паролей

```bash
# Все пароли за последний час
docker logs employee-service --since 1h | grep "Generated password"

# Экспорт в файл для анализа
docker logs employee-service | grep "Generated password" > passwords_$(date +%Y%m%d).log
```

## Безопасность

### ⚠️ Важные предупреждения

1. **Доступ к логам** - Ограничьте доступ к логам только администраторам
2. **Ротация логов** - Настройте автоматическое удаление старых логов
3. **Передача паролей** - Используйте безопасные каналы для передачи паролей
4. **Аудит** - Ведите учет кому и когда передавались пароли

### Рекомендуемые настройки безопасности

```bash
# Ограничение доступа к логам
chmod 640 /var/log/employee-service.log
chmod 640 /var/log/auth-service.log
chown root:admin /var/log/*.log

# Настройка ротации логов
cat > /etc/logrotate.d/digital-university << EOF
/var/log/employee-service.log /var/log/auth-service.log {
    daily
    rotate 7
    compress
    delaycompress
    missingok
    notifempty
    create 0640 root admin
}
EOF
```

## Автоматизация

### Скрипт для поиска паролей

```bash
#!/bin/bash
# get_password.sh

PHONE_SUFFIX="$1"

if [ -z "$PHONE_SUFFIX" ]; then
    echo "Usage: $0 <last_4_digits>"
    echo "Example: $0 4567"
    exit 1
fi

echo "Searching for passwords for phone ending in ****$PHONE_SUFFIX"
echo

echo "=== Employee Service ==="
docker logs employee-service 2>/dev/null | grep "Generated password" | grep "****$PHONE_SUFFIX" | tail -5

echo
echo "=== Auth Service ==="
docker logs auth-service 2>/dev/null | grep "Generated password" | grep "****$PHONE_SUFFIX" | tail -5

echo
echo "=== Reset Tokens ==="
docker logs auth-service 2>/dev/null | grep "Generated password reset token" | grep "****$PHONE_SUFFIX" | tail -5
```

### Использование скрипта

```bash
chmod +x get_password.sh
./get_password.sh 4567
```

## Мониторинг

### Настройка алертов

```bash
# Алерт при частом создании пользователей (возможная атака)
if [ $(docker logs employee-service --since 1h | grep "Generated password" | wc -l) -gt 100 ]; then
    echo "ALERT: Too many password generations in last hour" | mail -s "Security Alert" admin@company.com
fi
```

### Метрики для мониторинга

- Количество созданных паролей в час
- Количество запросов сброса пароля
- Соотношение успешных/неуспешных уведомлений

## Troubleshooting

### Проблема: Пароль не найден в логах

**Возможные причины:**
1. Логи были ротированы
2. Сервис был перезапущен
3. Ошибка при создании пользователя

**Решение:**
1. Проверьте архивные логи
2. Создайте пользователя заново
3. Используйте функцию сброса пароля

### Проблема: Слишком много логов

**Решение:**
```bash
# Фильтрация по времени
docker logs employee-service --since 2024-01-15T10:00:00 --until 2024-01-15T11:00:00 | grep "Generated password"

# Ограничение количества строк
docker logs employee-service --tail 1000 | grep "Generated password"
```

## Связанная документация

- [Password Logging Implementation](./PASSWORD_LOGGING_IMPLEMENTATION.md) - Техническая документация
- [Password Management API](../auth-service/PASSWORD_MANAGEMENT_API.md) - API для работы с паролями
- [Security Best Practices](./SECURITY_BEST_PRACTICES.md) - Рекомендации по безопасности