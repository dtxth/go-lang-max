#!/bin/bash

# Скрипт для генерации gRPC кода из proto файлов
# Требуется установленный protoc и protoc-gen-go, protoc-gen-go-grpc

echo "Generating proto files..."

# Auth Service
echo "Generating auth-service proto..."
cd auth-service
protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    api/proto/auth.proto
cd ..

# Chat Service
echo "Generating chat-service proto..."
cd chat-service
protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    api/proto/chat.proto
cd ..

# Employee Service
echo "Generating employee-service proto..."
cd employee-service
protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    api/proto/employee.proto
cd ..

# Structure Service
echo "Generating structure-service proto..."
cd structure-service
protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    api/proto/structure.proto
cd ..

# MaxBot Service
echo "Generating maxbot-service proto..."
cd maxbot-service
protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    api/proto/maxbot.proto
cd ..

echo "Proto generation complete!"

