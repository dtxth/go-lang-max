#!/bin/bash

echo "🔍 Debugging MaxBot Service Configuration"
echo "=========================================="

echo "Environment Variables:"
echo "MAX_API_TOKEN: ${MAX_API_TOKEN:-(not set)}"
echo "MOCK_MODE: ${MOCK_MODE:-(not set)}"
echo "MAX_API_URL: ${MAX_API_URL:-(not set)}"
echo "GRPC_PORT: ${GRPC_PORT:-9095 (default)}"
echo "HTTP_PORT: ${HTTP_PORT:-8095 (default)}"

echo ""
echo "Token length: ${#MAX_API_TOKEN} characters"

if [ -z "$MAX_API_TOKEN" ]; then
    echo "❌ MAX_API_TOKEN is not set!"
    echo "   Set it with: export MAX_API_TOKEN='your-token'"
else
    echo "✅ MAX_API_TOKEN is set"
fi

if [ "$MOCK_MODE" = "true" ] || [ "$MOCK_MODE" = "1" ] || [ "$MOCK_MODE" = "yes" ]; then
    echo "⚠️  MOCK_MODE is enabled - will use mock client"
else
    echo "✅ MOCK_MODE is disabled - will use real client"
fi

echo ""
echo "Starting MaxBot service with debug logging..."
go run cmd/maxbot/main.go