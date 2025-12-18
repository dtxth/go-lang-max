#!/bin/bash

# Integration test for real MaxBot service connection

echo "🤖 Testing real MaxBot integration..."

# Check if MaxBot service is running
echo "1. Checking if MaxBot service is available..."
if curl -s --connect-timeout 3 http://localhost:9095 > /dev/null 2>&1; then
    echo "✅ MaxBot service is running on port 9095"
else
    echo "❌ MaxBot service is not running on port 9095"
    echo "   Start it with: cd maxbot-service && go run cmd/maxbot/main.go"
    exit 1
fi

# Test gRPC connection
echo "2. Testing gRPC connection to MaxBot service..."
if command -v grpcurl > /dev/null 2>&1; then
    grpc_response=$(grpcurl -plaintext localhost:9095 maxbot.MaxBotService/GetMe 2>/dev/null)
    if [ $? -eq 0 ]; then
        echo "✅ gRPC connection successful"
        echo "   Response: $grpc_response"
    else
        echo "❌ gRPC connection failed"
    fi
else
    echo "⚠️  grpcurl not installed, skipping gRPC test"
fi

# Test auth-service with MaxBot integration
echo "3. Testing auth-service with MaxBot integration..."
export MAXBOT_SERVICE_ADDR="localhost:9095"

# Start auth-service in background
echo "   Starting auth-service with MaxBot integration..."
cd auth-service
go run cmd/auth/main.go > /tmp/auth-service.log 2>&1 &
AUTH_PID=$!

# Wait for auth-service to start
sleep 3

# Test the endpoint
echo "   Testing /bot/me endpoint..."
response=$(curl -s -w "\n%{http_code}" http://localhost:8080/bot/me)
http_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | head -n -1)

if [ "$http_code" = "200" ]; then
    echo "✅ Success! HTTP $http_code"
    echo "   Response: $body"
    
    # Check if it's real data (not mock)
    if echo "$body" | grep -q "Digital University Bot"; then
        echo "⚠️  Still using mock data (MaxBot service might not be properly connected)"
    else
        echo "✅ Receiving real data from MaxBot service!"
    fi
else
    echo "❌ Failed! HTTP $http_code"
    echo "   Response: $body"
    echo "   Check logs: tail /tmp/auth-service.log"
fi

# Cleanup
kill $AUTH_PID 2>/dev/null
wait $AUTH_PID 2>/dev/null

echo ""
echo "🏁 Integration test completed!"
echo ""
echo "📋 To run with real MaxBot service:"
echo "   1. Start MaxBot: cd maxbot-service && MAX_API_TOKEN=your-token go run cmd/maxbot/main.go"
echo "   2. Start auth-service: cd auth-service && MAXBOT_SERVICE_ADDR=localhost:9095 go run cmd/auth/main.go"
echo "   3. Test: curl http://localhost:8080/bot/me"