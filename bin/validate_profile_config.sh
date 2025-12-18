#!/bin/bash

# Profile Integration Configuration Validation Script
# This script validates that all required configuration is in place

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}=== Profile Integration Configuration Validation ===${NC}"
echo ""

# Check if .env file exists
if [ ! -f ".env" ]; then
    echo -e "${RED}❌ .env file not found${NC}"
    echo -e "${YELLOW}   Run: cp .env.example .env${NC}"
    exit 1
else
    echo -e "${GREEN}✓ .env file exists${NC}"
fi

# Source environment variables
source .env

# Function to check environment variable
check_env_var() {
    local var_name=$1
    local var_value=${!var_name}
    local required=$2
    local description=$3
    
    if [ -z "$var_value" ]; then
        if [ "$required" = "true" ]; then
            echo -e "${RED}❌ $var_name is not set${NC}"
            echo -e "${YELLOW}   Description: $description${NC}"
            return 1
        else
            echo -e "${YELLOW}⚠️  $var_name is not set (optional)${NC}"
            return 0
        fi
    else
        echo -e "${GREEN}✓ $var_name = $var_value${NC}"
        return 0
    fi
}

echo -e "${BLUE}Checking MaxBot Service Configuration:${NC}"
check_env_var "MAXBOT_HTTP_PORT" "true" "HTTP port for MaxBot service webhook endpoint"
check_env_var "MAX_API_TOKEN" "false" "MAX API token (can be empty for mock mode)"
check_env_var "MOCK_MODE" "true" "Enable/disable mock mode for development"

echo ""
echo -e "${BLUE}Checking Redis Configuration:${NC}"
check_env_var "REDIS_ADDR" "true" "Redis server address for profile cache"
check_env_var "REDIS_PASSWORD" "false" "Redis password (optional)"
check_env_var "REDIS_DB" "true" "Redis database number for profile cache"
check_env_var "PROFILE_TTL" "true" "Profile cache TTL (time to live)"

echo ""
echo -e "${BLUE}Checking Webhook Configuration:${NC}"
check_env_var "WEBHOOK_SECRET" "false" "Webhook secret for security (optional but recommended)"

echo ""
echo -e "${BLUE}Checking Monitoring Configuration:${NC}"
check_env_var "MONITORING_ENABLED" "true" "Enable monitoring and alerts"
check_env_var "PROFILE_QUALITY_ALERT_THRESHOLD" "true" "Alert threshold for profile quality"
check_env_var "WEBHOOK_ERROR_ALERT_THRESHOLD" "true" "Alert threshold for webhook errors"

echo ""
echo -e "${BLUE}Checking Employee Service Configuration:${NC}"
check_env_var "PROFILE_CACHE_ENABLED" "true" "Enable profile cache integration in employee service"
check_env_var "PROFILE_CACHE_TIMEOUT" "true" "Timeout for profile cache requests"

echo ""
echo -e "${BLUE}Checking Service Communication:${NC}"
check_env_var "MAXBOT_GRPC_ADDR" "true" "MaxBot service gRPC address"
check_env_var "MAXBOT_TIMEOUT" "true" "MaxBot service timeout"

# Validate configuration values
echo ""
echo -e "${BLUE}Validating Configuration Values:${NC}"

# Check if ports are numeric
if [[ ! "$MAXBOT_HTTP_PORT" =~ ^[0-9]+$ ]]; then
    echo -e "${RED}❌ MAXBOT_HTTP_PORT must be numeric${NC}"
    exit 1
else
    echo -e "${GREEN}✓ MAXBOT_HTTP_PORT is valid${NC}"
fi

# Check if Redis DB is numeric
if [[ ! "$REDIS_DB" =~ ^[0-9]+$ ]]; then
    echo -e "${RED}❌ REDIS_DB must be numeric${NC}"
    exit 1
else
    echo -e "${GREEN}✓ REDIS_DB is valid${NC}"
fi

# Check if thresholds are valid floats between 0 and 1
if ! echo "$PROFILE_QUALITY_ALERT_THRESHOLD" | grep -qE '^0\.[0-9]+$|^1\.0$|^1$|^0$'; then
    echo -e "${RED}❌ PROFILE_QUALITY_ALERT_THRESHOLD must be between 0.0 and 1.0${NC}"
    exit 1
else
    echo -e "${GREEN}✓ PROFILE_QUALITY_ALERT_THRESHOLD is valid${NC}"
fi

if ! echo "$WEBHOOK_ERROR_ALERT_THRESHOLD" | grep -qE '^0\.[0-9]+$|^1\.0$|^1$|^0$'; then
    echo -e "${RED}❌ WEBHOOK_ERROR_ALERT_THRESHOLD must be between 0.0 and 1.0${NC}"
    exit 1
else
    echo -e "${GREEN}✓ WEBHOOK_ERROR_ALERT_THRESHOLD is valid${NC}"
fi

# Check Docker Compose configuration
echo ""
echo -e "${BLUE}Checking Docker Compose Configuration:${NC}"

if [ ! -f "docker-compose.yml" ]; then
    echo -e "${RED}❌ docker-compose.yml not found${NC}"
    exit 1
else
    echo -e "${GREEN}✓ docker-compose.yml exists${NC}"
fi

# Validate Docker Compose syntax
if docker-compose config > /dev/null 2>&1; then
    echo -e "${GREEN}✓ docker-compose.yml syntax is valid${NC}"
else
    echo -e "${RED}❌ docker-compose.yml has syntax errors${NC}"
    exit 1
fi

# Check if required services are defined
required_services=("redis" "maxbot-service" "employee-service")
for service in "${required_services[@]}"; do
    if docker-compose config --services | grep -q "^$service$"; then
        echo -e "${GREEN}✓ Service '$service' is defined${NC}"
    else
        echo -e "${RED}❌ Service '$service' is not defined in docker-compose.yml${NC}"
        exit 1
    fi
done

# Check if MaxBot service has Redis dependency
if docker-compose config | grep -A 10 "maxbot-service:" | grep -q "redis:"; then
    echo -e "${GREEN}✓ MaxBot service depends on Redis${NC}"
else
    echo -e "${YELLOW}⚠️  MaxBot service should depend on Redis for profile caching${NC}"
fi

# Check network configuration
echo ""
echo -e "${BLUE}Checking Network Configuration:${NC}"

if docker-compose config | grep -q "microservices-network"; then
    echo -e "${GREEN}✓ Microservices network is configured${NC}"
else
    echo -e "${RED}❌ Microservices network is not configured${NC}"
    exit 1
fi

# Generate webhook URL examples
echo ""
echo -e "${BLUE}Webhook URL Configuration:${NC}"
echo -e "${GREEN}Local development URL:${NC}"
echo "  http://localhost:${MAXBOT_HTTP_PORT}/webhook/max"
echo ""
echo -e "${GREEN}Production URL (replace with your domain):${NC}"
echo "  https://your-domain.com/webhook/max"
echo ""
echo -e "${YELLOW}For local development with ngrok:${NC}"
echo "  1. Install ngrok: npm install -g ngrok"
echo "  2. Run: ngrok http ${MAXBOT_HTTP_PORT}"
echo "  3. Use the generated URL: https://abc123.ngrok.io/webhook/max"

# Check if documentation files exist
echo ""
echo -e "${BLUE}Checking Documentation:${NC}"

docs=("PROFILE_INTEGRATION_DEPLOYMENT.md" "maxbot-service/WEBHOOK_CONFIGURATION.md" "maxbot-service/MONITORING_ALERTS.md")
for doc in "${docs[@]}"; do
    if [ -f "$doc" ]; then
        echo -e "${GREEN}✓ $doc exists${NC}"
    else
        echo -e "${YELLOW}⚠️  $doc not found${NC}"
    fi
done

echo ""
echo -e "${GREEN}=== Configuration Validation Complete ===${NC}"
echo ""
echo -e "${BLUE}Next Steps:${NC}"
echo "1. Review and update .env file with your specific values"
echo "2. Configure MAX bot webhook URL in bot settings"
echo "3. Deploy services: make deploy-profile"
echo "4. Test webhook integration: make test-webhook"
echo "5. Monitor profile integration: make profile-monitor"
echo ""
echo -e "${YELLOW}For detailed setup instructions, see:${NC}"
echo "  - PROFILE_INTEGRATION_DEPLOYMENT.md"
echo "  - maxbot-service/WEBHOOK_CONFIGURATION.md"
echo "  - maxbot-service/MONITORING_ALERTS.md"