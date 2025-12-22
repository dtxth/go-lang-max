#!/bin/bash

# Proto validation script
# This script validates that all proto files are syntactically correct

set -e

echo "🔍 Validating proto files..."

# Function to validate a proto file
validate_proto() {
    local service_name=$1
    local proto_file=$2
    
    echo "📋 Validating $service_name proto file..."
    
    cd "$service_name"
    
    # Validate proto syntax
    protoc --proto_path=. --descriptor_set_out=/dev/null "$proto_file"
    
    echo "✅ $service_name proto file is valid"
    cd ..
}

# Validate all proto files
validate_proto "auth-service" "api/proto/auth.proto"
validate_proto "chat-service" "api/proto/chat.proto"
validate_proto "employee-service" "api/proto/employee.proto"
validate_proto "structure-service" "api/proto/structure.proto"
validate_proto "gateway-service" "api/proto/gateway.proto"

echo "🎉 All proto files are valid!"

# Check if generated files exist
echo ""
echo "🔍 Checking generated files..."

check_generated_files() {
    local service_name=$1
    local proto_name=$2
    
    if [[ -f "$service_name/api/proto/${proto_name}.pb.go" && -f "$service_name/api/proto/${proto_name}_grpc.pb.go" ]]; then
        echo "✅ $service_name generated files exist"
    else
        echo "❌ $service_name generated files missing"
        return 1
    fi
}

check_generated_files "auth-service" "auth"
check_generated_files "chat-service" "chat"
check_generated_files "employee-service" "employee"
check_generated_files "structure-service" "structure"
check_generated_files "gateway-service" "gateway"

echo "🎉 All generated files are present!"