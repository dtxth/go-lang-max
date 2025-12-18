#!/bin/bash

echo "🚀 Starting services with real MAX API integration"
echo "=================================================="

# Check if token is provided
if [ -z "$1" ]; then
    echo "❌ Usage: $0 <MAX_API_TOKEN>"
    echo "   Example: $0 your-real-max-api-token"
    exit 1
fi

MAX_API_TOKEN="$1"
echo "✅ Using provided MAX API token (${#MAX_API_TOKEN} characters)"

# Set environment variables
export MAX_API_TOKEN="$MAX_API_TOKEN"
export MAXBOT_SERVICE_ADDR="localhost:9095"
unset MOCK_MODE

echo ""
echo "📋 Configuration:"
echo "   MAX_API_TOKEN: ${MAX_API_TOKEN:0:10}..."
echo "   MAXBOT_SERVICE_ADDR: $MAXBOT_SERVICE_ADDR"
echo "   MOCK_MODE: disabled"

# Start MaxBot service
echo ""
echo "🤖 Starting MaxBot service..."
cd maxbot-service
go run cmd/maxbot/main.go > /tmp/maxbot.log 2>&1 &
MAXBOT_PID=$!
echo "   MaxBot PID: $MAXBOT_PID"

# Wait for MaxBot to start
echo "   Waiting for MaxBot service to start..."
sleep 5

# Check if MaxBot is running
if kill -0 $MAXBOT_PID 2>/dev/null; then
    echo "   ✅ MaxBot service started successfully"
else
    echo "   ❌ MaxBot service failed to start"
    echo "   Check logs: tail /tmp/maxbot.log"
    exit 1
fi

# Start auth-service
echo ""
echo "🔐 Starting auth-service..."
cd ../auth-service
go run cmd/auth/main.go > /tmp/auth.log 2>&1 &
AUTH_PID=$!
echo "   Auth service PID: $AUTH_PID"

# Wait for auth-service to start
echo "   Waiting for auth-service to start..."
sleep 3

# Check if auth-service is running
if kill -0 $AUTH_PID 2>/dev/null; then
    echo "   ✅ Auth service started successfully"
else
    echo "   ❌ Auth service failed to start"
    echo "   Check logs: tail /tmp/auth.log"
    kill $MAXBOT_PID 2>/dev/null
    exit 1
fi

# Test the endpoint
echo ""
echo "🧪 Testing /bot/me endpoint..."
sleep 2

response=$(curl -s -w "\n%{http_code}" http://localhost:8080/bot/me)
http_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | head -n -1)

if [ "$http_code" = "200" ]; then
    echo "   ✅ Success! HTTP $http_code"
    echo "   Response: $body"
    
    # Check if it's real data
    if echo "$body" | grep -q "Digital University Bot"; then
        echo "   ⚠️  Still receiving mock data - check MaxBot integration"
        echo "   Check MaxBot logs: tail /tmp/maxbot.log"
        echo "   Check Auth logs: tail /tmp/auth.log"
    else
        echo "   🎉 SUCCESS! Receiving real data from MAX API!"
    fi
else
    echo "   ❌ Failed! HTTP $http_code"
    echo "   Response: $body"
fi

echo ""
echo "📊 Services status:"
echo "   MaxBot service: http://localhost:8095 (PID: $MAXBOT_PID)"
echo "   Auth service: http://localhost:8080 (PID: $AUTH_PID)"
echo "   Swagger: http://localhost:8080/swagger/index.html"

echo ""
echo "🛑 To stop services:"
echo "   kill $MAXBOT_PID $AUTH_PID"

echo ""
echo "📋 Useful commands:"
echo "   Test endpoint: curl http://localhost:8080/bot/me"
echo "   MaxBot logs: tail -f /tmp/maxbot.log"
echo "   Auth logs: tail -f /tmp/auth.log"
echo "   Full debug: ./debug_full_chain.sh"

# Keep script running
echo ""
echo "✨ Services are running! Press Ctrl+C to stop..."
trap "echo 'Stopping services...'; kill $MAXBOT_PID $AUTH_PID 2>/dev/null; exit 0" INT
wait