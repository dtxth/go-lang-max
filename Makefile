.PHONY: help test test-quick test-verbose test-coverage deploy deploy-fast deploy-rebuild up down logs ps clean build

# Цвета для вывода
BLUE := \033[0;34m
GREEN := \033[0;32m
YELLOW := \033[1;33m
NC := \033[0m

help: ## Показать эту справку
	@echo "$(BLUE)Доступные команды:$(NC)"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(GREEN)%-20s$(NC) %s\n", $$1, $$2}'
	@echo ""
	@echo "$(YELLOW)Примеры:$(NC)"
	@echo "  make test           # Запустить все тесты"
	@echo "  make deploy         # Полное развертывание (тесты + сборка + запуск)"
	@echo "  make deploy-fast    # Быстрое развертывание (без тестов)"
	@echo "  make logs           # Просмотр логов всех сервисов"

test: ## Запустить все тесты с race detector
	@echo "$(BLUE)Запуск всех тестов...$(NC)"
	@./tests/run_tests.sh

test-quick: ## Быстрая проверка тестов (без race detector)
	@echo "$(BLUE)Быстрая проверка тестов...$(NC)"
	@./tests/test_quick.sh

test-verbose: ## Запустить тесты с подробным выводом
	@echo "$(BLUE)Запуск тестов (подробный режим)...$(NC)"
	@./tests/run_tests.sh --verbose

test-coverage: ## Запустить тесты с отчетом о покрытии кода
	@echo "$(BLUE)Запуск тестов с покрытием кода...$(NC)"
	@./tests/run_tests.sh --coverage

deploy: ## Полное развертывание: тесты → сборка → запуск
	@echo "$(BLUE)Полное развертывание...$(NC)"
	@./bin/deploy.sh

deploy-fast: ## Быстрое развертывание без тестов
	@echo "$(YELLOW)Быстрое развертывание (без тестов)...$(NC)"
	@./bin/deploy.sh --skip-tests

deploy-rebuild: ## Полная пересборка с тестами (медленно, 5-10 минут)
	@echo "$(BLUE)Полная пересборка...$(NC)"
	@./bin/deploy.sh --no-cache

deploy-rebuild-fast: ## Полная пересборка без тестов
	@echo "$(YELLOW)Полная пересборка без тестов...$(NC)"
	@./bin/deploy.sh --no-cache --skip-tests

deploy-verbose: ## Развертывание с подробным выводом
	@echo "$(BLUE)Развертывание (подробный режим)...$(NC)"
	@./bin/deploy.sh --verbose

build: ## Собрать Docker образы
	@echo "$(BLUE)Сборка Docker образов...$(NC)"
	@docker-compose build

build-no-cache: ## Собрать Docker образы без кеша
	@echo "$(BLUE)Пересборка Docker образов без кеша...$(NC)"
	@docker-compose build --no-cache

up: ## Запустить все сервисы
	@echo "$(BLUE)Запуск сервисов...$(NC)"
	@docker-compose up -d
	@echo "$(GREEN)Сервисы запущены!$(NC)"
	@make ps

down: ## Остановить все сервисы
	@echo "$(YELLOW)Остановка сервисов...$(NC)"
	@docker-compose down
	@echo "$(GREEN)Сервисы остановлены!$(NC)"

restart: ## Перезапустить все сервисы
	@echo "$(YELLOW)Перезапуск сервисов...$(NC)"
	@docker-compose restart
	@echo "$(GREEN)Сервисы перезапущены!$(NC)"

logs: ## Просмотр логов всех сервисов
	@docker-compose logs -f

logs-auth: ## Просмотр логов auth-service
	@docker-compose logs -f auth-service

logs-chat: ## Просмотр логов chat-service
	@docker-compose logs -f chat-service

logs-employee: ## Просмотр логов employee-service
	@docker-compose logs -f employee-service

logs-structure: ## Просмотр логов structure-service
	@docker-compose logs -f structure-service

logs-maxbot: ## Просмотр логов maxbot-service
	@docker-compose logs -f maxbot-service

logs-migration: ## Просмотр логов migration-service
	@docker-compose logs -f migration-service

ps: ## Показать статус контейнеров
	@echo "$(BLUE)Статус контейнеров:$(NC)"
	@docker-compose ps

images: ## Показать размеры Docker образов
	@echo "$(BLUE)Docker образы:$(NC)"
	@docker images | grep "go-lang-max" || echo "$(YELLOW)Образы не найдены$(NC)"

clean: ## Удалить все контейнеры и образы
	@echo "$(YELLOW)Удаление контейнеров и образов...$(NC)"
	@docker-compose down -v
	@docker images | grep "go-lang-max" | awk '{print $$3}' | xargs -r docker rmi -f
	@echo "$(GREEN)Очистка завершена!$(NC)"

clean-volumes: ## Удалить все контейнеры, образы и volumes
	@echo "$(RED)Удаление контейнеров, образов и volumes...$(NC)"
	@docker-compose down -v
	@docker images | grep "go-lang-max" | awk '{print $$3}' | xargs -r docker rmi -f
	@docker volume prune -f
	@echo "$(GREEN)Полная очистка завершена!$(NC)"

swagger: ## Проверить Swagger endpoints
	@echo "$(BLUE)Проверка Swagger endpoints:$(NC)"
	@for port in 8080 8081 8082 8083 8084; do \
		echo -n "  Port $$port: "; \
		curl -s -f "http://localhost:$$port/swagger/doc.json" > /dev/null 2>&1 && \
			echo "$(GREEN)✓ http://localhost:$$port/swagger/index.html$(NC)" || \
			echo "$(RED)✗ Не доступен$(NC)"; \
	done

health: ## Проверить здоровье всех сервисов
	@echo "$(BLUE)Проверка здоровья сервисов:$(NC)"
	@make ps
	@echo ""
	@make swagger

# =============================================================================
# Profile Integration Monitoring
# =============================================================================

profile-health: ## Проверить состояние профильной интеграции
	@echo "$(BLUE)Проверка профильной интеграции:$(NC)"
	@echo -n "  MaxBot Service: "
	@curl -s -f "http://localhost:8095/health" > /dev/null 2>&1 && \
		echo "$(GREEN)✓ Работает$(NC)" || \
		echo "$(RED)✗ Недоступен$(NC)"
	@echo -n "  Redis Cache: "
	@docker exec redis redis-cli ping > /dev/null 2>&1 && \
		echo "$(GREEN)✓ Подключен$(NC)" || \
		echo "$(RED)✗ Недоступен$(NC)"
	@echo -n "  Webhook Endpoint: "
	@curl -s -f "http://localhost:8095/webhook/max" -X POST -H "Content-Type: application/json" -d '{}' > /dev/null 2>&1 && \
		echo "$(GREEN)✓ Доступен$(NC)" || \
		echo "$(RED)✗ Недоступен$(NC)"

profile-stats: ## Показать статистику профилей
	@echo "$(BLUE)Статистика профилей:$(NC)"
	@curl -s "http://localhost:8095/monitoring/profile-stats" 2>/dev/null | jq '.' || \
		echo "$(RED)Не удается получить статистику профилей$(NC)"

webhook-stats: ## Показать статистику webhook событий
	@echo "$(BLUE)Статистика webhook событий:$(NC)"
	@curl -s "http://localhost:8095/monitoring/webhook-stats" 2>/dev/null | jq '.' || \
		echo "$(RED)Не удается получить статистику webhook$(NC)"

cache-health: ## Проверить состояние кэша профилей
	@echo "$(BLUE)Состояние кэша профилей:$(NC)"
	@curl -s "http://localhost:8095/monitoring/cache-health" 2>/dev/null | jq '.' || \
		echo "$(RED)Не удается получить состояние кэша$(NC)"

test-webhook: ## Тестировать webhook endpoint
	@echo "$(BLUE)Тестирование webhook endpoint...$(NC)"
	@curl -X POST "http://localhost:8095/webhook/max" \
		-H "Content-Type: application/json" \
		-d '{"type":"message_new","message":{"from":{"user_id":"test123","first_name":"Тест","last_name":"Пользователь"},"text":"Hello"}}' \
		-w "\nHTTP Status: %{http_code}\n" || \
		echo "$(RED)Ошибка при тестировании webhook$(NC)"

profile-monitor: ## Мониторинг профильной интеграции в реальном времени
	@echo "$(BLUE)Мониторинг профильной интеграции (Ctrl+C для выхода):$(NC)"
	@while true; do \
		clear; \
		echo "$(BLUE)=== Profile Integration Monitor ===$(NC)"; \
		echo "$(YELLOW)Время: $$(date)$(NC)"; \
		echo ""; \
		make profile-health; \
		echo ""; \
		make profile-stats; \
		echo ""; \
		echo "$(BLUE)Обновление через 10 секунд...$(NC)"; \
		sleep 10; \
	done

deploy-profile: ## Развернуть только компоненты профильной интеграции
	@echo "$(BLUE)Развертывание профильной интеграции...$(NC)"
	@docker-compose up -d redis maxbot-service employee-service
	@echo "$(GREEN)Профильная интеграция развернута!$(NC)"
	@make profile-health

validate-profile-config: ## Проверить конфигурацию профильной интеграции
	@echo "$(BLUE)Проверка конфигурации профильной интеграции...$(NC)"
	@./bin/validate_profile_config.sh

# Тесты для отдельных сервисов
test-auth: ## Тесты auth-service
	@echo "$(BLUE)Тестирование auth-service...$(NC)"
	@cd auth-service && go test -v -race ./...

test-chat: ## Тесты chat-service
	@echo "$(BLUE)Тестирование chat-service...$(NC)"
	@cd chat-service && go test -v -race ./...

test-employee: ## Тесты employee-service
	@echo "$(BLUE)Тестирование employee-service...$(NC)"
	@cd employee-service && go test -v -race ./...

test-structure: ## Тесты structure-service
	@echo "$(BLUE)Тестирование structure-service...$(NC)"
	@cd structure-service && go test -v -race ./...

test-maxbot: ## Тесты maxbot-service
	@echo "$(BLUE)Тестирование maxbot-service...$(NC)"
	@cd maxbot-service && go test -v -race ./...

test-migration: ## Тесты migration-service
	@echo "$(BLUE)Тестирование migration-service...$(NC)"
	@cd migration-service && go test -v -race ./...

# Локальная разработка
dev-auth: ## Запустить auth-service локально
	@cd auth-service && go run cmd/auth/main.go

dev-chat: ## Запустить chat-service локально
	@cd chat-service && go run cmd/chat/main.go

dev-employee: ## Запустить employee-service локально
	@cd employee-service && go run cmd/employee/main.go

dev-structure: ## Запустить structure-service локально
	@cd structure-service && go run cmd/structure/main.go

# Утилиты
fmt: ## Форматировать код
	@echo "$(BLUE)Форматирование кода...$(NC)"
	@find . -name "*.go" -not -path "./vendor/*" -exec gofmt -w {} \;
	@echo "$(GREEN)Форматирование завершено!$(NC)"

lint: ## Запустить линтер (требует golangci-lint)
	@echo "$(BLUE)Запуск линтера...$(NC)"
	@for dir in auth-service chat-service employee-service structure-service maxbot-service migration-service; do \
		echo "Linting $$dir..."; \
		cd $$dir && golangci-lint run ./... && cd ..; \
	done

mod-tidy: ## Обновить go.mod для всех сервисов
	@echo "$(BLUE)Обновление go.mod...$(NC)"
	@for dir in auth-service chat-service employee-service structure-service maxbot-service migration-service; do \
		echo "  $$dir"; \
		cd $$dir && go mod tidy && cd ..; \
	done
	@echo "$(GREEN)go.mod обновлены!$(NC)"

# =============================================================================
# Безопасные операции с данными
# =============================================================================

backup-auth: ## Создать резервную копию базы auth-service
	@echo "$(BLUE)Создание резервной копии auth-db...$(NC)"
	@mkdir -p backups
	@docker exec auth-db pg_dump -U postgres postgres > backups/auth_backup_$$(date +%Y%m%d_%H%M%S).sql
	@echo "$(GREEN)Резервная копия создана в backups/$(NC)"

restore-auth: ## Восстановить базу auth-service из резервной копии
	@echo "$(BLUE)Доступные резервные копии:$(NC)"
	@ls -la backups/auth_backup_*.sql 2>/dev/null || echo "$(YELLOW)Резервные копии не найдены$(NC)"
	@echo -n "$(YELLOW)Введите имя файла для восстановления: $(NC)" && read backup_file && \
	docker exec -i auth-db psql -U postgres postgres < "$$backup_file"

check-volumes: ## Проверить состояние volumes и данных
	@echo "$(BLUE)Docker volumes:$(NC)"
	@docker volume ls | grep -E "(auth|chat|employee|structure|migration)" || echo "$(YELLOW)Volumes не найдены$(NC)"
	@echo ""
	@echo "$(BLUE)Таблицы в auth-db:$(NC)"
	@docker exec auth-db psql -U postgres -d postgres -c "\dt" 2>/dev/null || echo "$(RED)Не удается подключиться к auth-db$(NC)"
	@echo ""
	@echo "$(BLUE)Количество пользователей:$(NC)"
	@docker exec auth-db psql -U postgres -d postgres -c "SELECT COUNT(*) as users_count FROM users;" 2>/dev/null || echo "$(RED)Таблица users не найдена$(NC)"

safe-restart: ## Безопасный перезапуск сервисов (без удаления данных)
	@echo "$(BLUE)Безопасный перезапуск сервисов...$(NC)"
	@docker-compose restart
	@echo "$(GREEN)Сервисы перезапущены без потери данных!$(NC)"

# Переопределяем опасную команду с предупреждением
clean-volumes-dangerous: ## ⚠️  ОПАСНО: Удалить все volumes (ЗАТИРАЕТ ВСЕ ДАННЫЕ!)
	@echo "$(RED)⚠️  ВНИМАНИЕ: Эта команда УДАЛИТ ВСЕ ДАННЫЕ из баз данных!$(NC)"
	@echo -n "$(YELLOW)Введите 'DELETE_ALL_DATA' для подтверждения: $(NC)" && read confirmation && [ "$$confirmation" = "DELETE_ALL_DATA" ]
	@echo "$(RED)Удаление всех volumes...$(NC)"
	@docker-compose down -v
	@docker volume prune -f
	@echo "$(GREEN)Все данные удалены!$(NC)"