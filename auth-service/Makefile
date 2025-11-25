.PHONY: build up down logs

build:
	docker build -t auth-service .

up:
	docker compose up --build

down:
	docker compose down -v

logs:
	docker compose logs -f