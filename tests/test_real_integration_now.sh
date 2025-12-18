#!/bin/bash

echo "🚀 Testing Real Integration with .env loading"
echo "============================================="

# Start MaxBot service (it will load .env automatically)
echo "1. Starting MaxBot service..."
cd maxbot-service
go run cmd/maxbot/main.go > /tmp/maxbot-real.log 2>&1 &
MAXBOT_PID=$!
echo "   MaxBot PID: $MAXBOT_PID"

# Wait for startup
sleep 5

# Check if MaxBot started successfully
if kill -0 $MAXBOT_PID 2>/dev/null; then
    echo "   ✅ MaxBot service started"
    
    # Show relevant logs
    echo "   📋 MaxBot logs:"
    grep -E "(Loading environment|MAX_API_TOKEN|MOCK_MODE|Successfully retrieved bot info)" /tmp/maxbot-real.log | tail -5
else
    echo "   ❌ MaxBot service failed to start"
    echo "   📋 Error logs:"
    cat /tmp/maxbot-real.log
    exit 1
fi

# Start auth-service
echo ""
echo "2. Starting auth-service..."
cd ../auth-service

# Set MaxBot address
export MAXBOT_SERVICE_ADDR="localhost:9095"

go run cmd/auth/main.go > /tmp/auth-real.log 2>&1 &
AUTH_PID=$!
echo "   Auth PID: $AUTH_PID"

# Wait for startup
sleep 3

# Check if auth-service started
if kill -0 $AUTH_PID 2>/dev/null; then
    echo "   ✅ Auth service started"
    
    # Show relevant logs
    echo "   📋 Auth logs:"
    grep -E "(MaxBot client|Initialized)" /tmp/auth-real.log | tail -3
else
    echo "   ❌ Auth service failed to start"
    echo "   📋 Error logs:"
    cat /tmp/auth-real.log
    kill $MAXBOT_PID 2>/dev/null
    exit 1
fi

# Test the endpoint
echo ""
echo "3. Testing /bot/me endpoint..."
sleep 2

response=$(curl -s -w "\n%{http_code}" http://localhost:8080/bot/me)
http_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | head -n -1)

echo "   HTTP Status: $http_code"
echo "   Response: $body"

if [ "$http_code" = "200" ]; then
    # Check if it's real data
    if echo "$body" | grep -q "Digital University Support Bot"; then
        echo "   🎉 SUCCESS! Receiving data from REAL MaxBot client!"
        echo "   ✅ .env file loaded correctly"
        echo "   ✅ Real MAX API token being used"
    elif echo "$body" | grep -q "Digital University Bot"; then
        echo "   ⚠️  Still using mock client - check configuration"
    else
        echo "   ❓ Unknown response format"
    fi
else
    echo "   ❌ HTTP request failed"
fi

# Cleanup
echo ""
echo "4. Cleaning up..."
kill $MAXBOT_PID $AUTH_PID 2>/dev/null
wait $MAXBOT_PID $AUTH_PID 2>/dev/null

echo ""
echo "📋 Log files:"
echo "   MaxBot: /tmp/maxbot-real.log"
echo "   Auth: /tmp/auth-real.log"

echo ""
echo "✅ Test completed!"