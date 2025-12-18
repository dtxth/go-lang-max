#!/bin/bash

echo "🧪 Testing .env file loading"
echo "============================"

# Check if .env exists
if [ -f ".env" ]; then
    echo "✅ .env file exists"
    
    # Show MAX_API_TOKEN from file (first 10 chars)
    token_from_file=$(grep "^MAX_API_TOKEN=" .env | cut -d'=' -f2)
    echo "📄 Token in .env: ${token_from_file:0:10}..."
    
    # Check current environment
    echo "🌍 Current environment MAX_API_TOKEN: ${MAX_API_TOKEN:0:10:-not set}..."
    
    # Test maxbot-service loading
    echo ""
    echo "🤖 Testing MaxBot service .env loading..."
    cd maxbot-service
    
    # Run with debug output
    timeout 10s go run cmd/maxbot/main.go 2>&1 | head -20
    
    echo ""
    echo "✅ Test completed!"
else
    echo "❌ .env file not found"
fi