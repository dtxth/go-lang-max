#!/bin/bash

# Test script for /bot/me endpoint

echo "Testing /bot/me endpoint..."

# Test with mock client (no MaxBot service required)
echo "1. Testing with mock client..."
response=$(curl -s -w "\n%{http_code}" http://localhost:8080/bot/me)
http_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | head -n -1)

if [ "$http_code" = "200" ]; then
    echo "✅ Success! HTTP $http_code"
    echo "Response: $body"
    
    # Parse JSON to check fields
    name=$(echo "$body" | grep -o '"name":"[^"]*"' | cut -d'"' -f4)
    add_link=$(echo "$body" | grep -o '"add_link":"[^"]*"' | cut -d'"' -f4)
    
    if [ -n "$name" ] && [ -n "$add_link" ]; then
        echo "✅ Bot name: $name"
        echo "✅ Add link: $add_link"
        
        # Check if using mock or real data
        if [ "$name" = "Digital University Bot" ]; then
            echo "ℹ️  Using mock client (MaxBot service not connected)"
        else
            echo "🎉 Using real MaxBot service data!"
        fi
    else
        echo "❌ Missing required fields in response"
    fi
else
    echo "❌ Failed! HTTP $http_code"
    echo "Response: $body"
fi

echo ""
echo "2. Testing Swagger documentation..."
swagger_response=$(curl -s -w "\n%{http_code}" http://localhost:8080/swagger/index.html)
swagger_code=$(echo "$swagger_response" | tail -n1)

if [ "$swagger_code" = "200" ]; then
    echo "✅ Swagger documentation available at http://localhost:8080/swagger/index.html"
else
    echo "❌ Swagger documentation not available (HTTP $swagger_code)"
fi

echo ""
echo "Test completed!"