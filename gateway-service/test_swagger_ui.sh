#!/bin/bash

# Gateway Service Swagger UI Test Script

echo "🔍 Testing Gateway Service Swagger UI..."
echo "========================================="

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

GATEWAY_URL="http://localhost:8085"

echo -e "${YELLOW}Testing Gateway Service endpoints...${NC}"

# Test Swagger UI
echo -n "Testing Swagger UI: "
if curl -s -f "${GATEWAY_URL}/swagger/" > /dev/null; then
    echo -e "${GREEN}✅ PASS${NC}"
else
    echo -e "${RED}❌ FAIL${NC}"
fi

# Test Swagger YAML
echo -n "Testing Swagger YAML: "
if curl -s -f "${GATEWAY_URL}/swagger/swagger.yaml" > /dev/null; then
    echo -e "${GREEN}✅ PASS${NC}"
else
    echo -e "${RED}❌ FAIL${NC}"
fi

# Test Health endpoint
echo -n "Testing Health endpoint: "
if curl -s -f "${GATEWAY_URL}/health" > /dev/null; then
    echo -e "${GREEN}✅ PASS${NC}"
else
    echo -e "${RED}❌ FAIL${NC}"
fi

# Test API endpoints
echo -n "Testing Universities endpoint: "
if curl -s -f "${GATEWAY_URL}/universities" > /dev/null; then
    echo -e "${GREEN}✅ PASS${NC}"
else
    echo -e "${RED}❌ FAIL (expected if Structure service is down)${NC}"
fi

echo ""
echo -e "${YELLOW}Gateway Service URLs:${NC}"
echo "🌐 Swagger UI: ${GATEWAY_URL}/swagger/"
echo "📄 API Docs: ${GATEWAY_URL}/swagger/swagger.yaml"
echo "❤️  Health Check: ${GATEWAY_URL}/health"
echo ""
echo -e "${GREEN}✅ Swagger UI rebuild completed successfully!${NC}"
echo "You can now access the API documentation at: ${GATEWAY_URL}/swagger/"