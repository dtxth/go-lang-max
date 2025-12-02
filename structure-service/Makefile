.PHONY: build up down logs copy-proto

# Копируем proto файл из chat-service перед сборкой и модифицируем go_package
copy-proto:
	@mkdir -p api/proto/chat
	@cp ../chat-service/api/proto/chat.proto api/proto/chat/chat.proto
	@sed -i '' 's|option go_package = "chat-service/api/proto;proto";|option go_package = "structure-service/api/proto/chat;chatproto";|' api/proto/chat/chat.proto || \
	 sed -i 's|option go_package = "chat-service/api/proto;proto";|option go_package = "structure-service/api/proto/chat;chatproto";|' api/proto/chat/chat.proto
	@echo "Proto file copied and modified successfully"

build: copy-proto
	docker build -t structure-service .

up:
	docker compose up --build

down:
	docker compose down -v

logs:
	docker compose logs -f

