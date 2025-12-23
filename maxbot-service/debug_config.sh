#!/bin/bash

echo "🔍 Debugging MaxBot Service Configuration"
echo "=========================================="

echo "Environment Variables:"
echo "MAX_BOT_TOKEN: ${MAX_BOT_TOKEN:-(not set)}"
echo "MOCK_MODE: ${MOCK_MODE:-(not set)}"
echo "MAX_API_URL: ${MAX_API_URL:-(not set)}"
echo "GRPC_PORT: ${GRPC_PORT:-9095 (default)}"
echo "HTTP_PORT: ${HTTP_PORT:-8095 (default)}"

echo ""
echo "Token length: ${#MAX_BOT_TOKEN} characters"

if [ -z "$MAX_BOT_TOKEN" ]; then
    echo "❌ MAX_BOT_TOKEN is not set!"
    echo "   Set it with: export MAX_BOT_TOKEN='your-token'"
else
    echo "✅ MAX_BOT_TOKEN is set"
fi

if [ "$MOCK_MODE" = "true" ] || [ "$MOCK_MODE" = "1" ] || [ "$MOCK_MODE" = "yes" ]; then
    echo "⚠️  MOCK_MODE is enabled - will use mock client"
else
    echo "✅ MOCK_MODE is disabled - will use real client"
fi

echo ""
echo "Starting MaxBot service with debug logging..."
go run cmd/maxbot/main.go