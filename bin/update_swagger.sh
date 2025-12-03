#!/bin/bash

# Script to update Swagger documentation for all services

echo "🔄 Updating Swagger documentation for all services..."
echo ""

# Check if swag is installed
if ! command -v swag &> /dev/null && ! command -v ~/go/bin/swag &> /dev/null; then
    echo "❌ swag is not installed. Installing..."
    go install github.com/swaggo/swag/cmd/swag@latest
    echo "✅ swag installed successfully"
    echo ""
fi

# Use swag from go/bin if not in PATH
SWAG_CMD="swag"
if ! command -v swag &> /dev/null; then
    SWAG_CMD="$HOME/go/bin/swag"
fi

success_count=0
fail_count=0

# List of services
services="auth-service employee-service chat-service structure-service migration-service"

# Update Swagger for each service
for service in $services; do
    echo "📝 Updating $service..."
    
    if [ ! -d "$service" ]; then
        echo "⚠️  Directory $service not found, skipping..."
        fail_count=$((fail_count + 1))
        continue
    fi
    
    cd "$service" || continue
    
    # Determine main.go path
    service_name=$(echo "$service" | sed 's/-service//')
    main_path="cmd/$service_name/main.go"
    
    if [ ! -f "$main_path" ]; then
        echo "⚠️  Main file $main_path not found in $service, skipping..."
        cd ..
        fail_count=$((fail_count + 1))
        continue
    fi
    
    # Generate Swagger documentation
    if $SWAG_CMD init -g "$main_path" -o internal/infrastructure/http/docs 2>&1 | grep -q "create docs.go"; then
        echo "✅ $service swagger updated successfully"
        success_count=$((success_count + 1))
    else
        echo "❌ Failed to update $service swagger"
        fail_count=$((fail_count + 1))
    fi
    
    cd ..
    echo ""
done

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "📊 Summary:"
echo "   ✅ Success: $success_count"
echo "   ❌ Failed: $fail_count"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

if [ $fail_count -eq 0 ]; then
    echo "🎉 All Swagger documentation updated successfully!"
    exit 0
else
    echo "⚠️  Some services failed to update. Please check the logs above."
    exit 1
fi
