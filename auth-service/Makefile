.PHONY: build up down logs proto-gen

proto-gen:
	@echo "Generating proto files for auth-service..."
	protoc --go_out=. --go_opt=paths=source_relative \
	       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
	       api/proto/auth.proto

build: proto-gen
	docker build -t auth-service .

up:
	docker compose up --build

down:
	docker compose down -v

logs:
	docker compose logs -f