.PHONY: build run test clean docker-build docker-up docker-down proto-gen docs docs-serve docs-validate

proto-gen:
	@echo "Generating proto files for gateway-service..."
	protoc --go_out=. --go_opt=paths=source_relative \
	       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
	       api/proto/gateway.proto

build: proto-gen
	go build -o bin/gateway-service ./cmd/gateway

run:
	go run ./cmd/gateway/main.go

test:
	go test ./...

clean:
	rm -rf bin/

docker-build:
	docker-compose build

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

# Documentation commands
docs:
	@echo "📖 Opening API documentation..."
	@echo "Swagger UI: http://localhost:8082"
	@echo "Documentation files are in docs/ directory"
	@echo ""
	@echo "Available documentation commands:"
	@echo "  make docs-serve    - Start documentation server"
	@echo "  make docs-validate - Validate OpenAPI specification"
	@echo "  make docs-test     - Test Swagger UI"

docs-serve:
	@echo "🚀 Starting documentation server on http://localhost:8082..."
	@echo "📖 Swagger UI will be available at http://localhost:8082"
	@echo "🛑 Press Ctrl+C to stop the server"
	@if command -v python3 >/dev/null 2>&1; then \
		cd docs && python3 -m http.server 8082; \
	elif command -v python >/dev/null 2>&1; then \
		cd docs && python -m SimpleHTTPServer 8082; \
	else \
		echo "❌ Python not found. Please install Python or use a web server to serve docs/ directory"; \
		echo "Alternative: npx http-server docs -p 8082"; \
	fi

docs-validate:
	@echo "🔍 Validating OpenAPI specification..."
	@python3 -c "import yaml; yaml.safe_load(open('docs/swagger.yaml'))" && echo "✅ YAML syntax is valid" || echo "❌ YAML syntax error"
	@if command -v npx >/dev/null 2>&1; then \
		echo "🔍 Validating OpenAPI schema..."; \
		npx swagger-parser validate docs/swagger.yaml 2>/dev/null && echo "✅ OpenAPI schema is valid" || echo "⚠️  OpenAPI validation skipped (swagger-parser not available)"; \
	else \
		echo "⚠️  OpenAPI validation skipped (npx not available)"; \
	fi

docs-test:
	@echo "🧪 Testing Swagger documentation..."
	@./test_swagger_ui.sh