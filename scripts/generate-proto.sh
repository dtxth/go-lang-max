#!/bin/bash

# Proto generation script for all microservices
# This script generates Go code from proto files for all services

set -e

echo "🔧 Generating proto files for all services..."

# Check if protoc is installed
if ! command -v protoc &> /dev/null; then
    echo "❌ protoc is not installed. Please install Protocol Buffers compiler."
    echo "   macOS: brew install protobuf"
    echo "   Ubuntu: apt-get install protobuf-compiler"
    exit 1
fi

# Check if Go plugins are installed
if ! command -v protoc-gen-go &> /dev/null; then
    echo "❌ protoc-gen-go is not installed. Installing..."
    go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
fi

if ! command -v protoc-gen-go-grpc &> /dev/null; then
    echo "❌ protoc-gen-go-grpc is not installed. Installing..."
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
fi

# Function to generate proto for a service
generate_service_proto() {
    local service_name=$1
    local proto_file=$2
    
    echo "📦 Generating proto for $service_name..."
    
    cd "$service_name"
    
    # Generate Go code from proto files
    protoc --go_out=. --go_opt=paths=source_relative \
           --go-grpc_out=. --go-grpc_opt=paths=source_relative \
           "$proto_file"
    
    echo "✅ Generated proto for $service_name"
    cd ..
}

# Generate proto files for all services
generate_service_proto "auth-service" "api/proto/auth.proto"
generate_service_proto "chat-service" "api/proto/chat.proto"
generate_service_proto "employee-service" "api/proto/employee.proto"
generate_service_proto "structure-service" "api/proto/structure.proto"
generate_service_proto "gateway-service" "api/proto/gateway.proto"

echo "🎉 All proto files generated successfully!"
echo ""
echo "Generated files:"
echo "  - auth-service/api/proto/auth.pb.go"
echo "  - auth-service/api/proto/auth_grpc.pb.go"
echo "  - chat-service/api/proto/chat.pb.go"
echo "  - chat-service/api/proto/chat_grpc.pb.go"
echo "  - employee-service/api/proto/employee.pb.go"
echo "  - employee-service/api/proto/employee_grpc.pb.go"
echo "  - structure-service/api/proto/structure.pb.go"
echo "  - structure-service/api/proto/structure_grpc.pb.go"
echo "  - gateway-service/api/proto/gateway.pb.go"
echo "  - gateway-service/api/proto/gateway_grpc.pb.go"