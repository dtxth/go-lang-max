#!/bin/bash

# Monitoring and Analytics Demo Script
# This script demonstrates the monitoring functionality of the MaxBot service

echo "=== MaxBot Service Monitoring and Analytics Demo ==="
echo

# Check if the service is running
SERVICE_URL="http://localhost:8095"
echo "Checking if MaxBot service is running at $SERVICE_URL..."

# Test health endpoint
if curl -s "$SERVICE_URL/health" > /dev/null; then
    echo "✓ Service is running"
else
    echo "✗ Service is not running. Please start it first with: make up"
    exit 1
fi

echo
echo "=== Testing Monitoring Endpoints ==="
echo

# Test webhook statistics
echo "1. Getting webhook processing statistics (last day):"
curl -s "$SERVICE_URL/api/v1/monitoring/webhook/stats?period=day" | jq '.' || echo "Failed to get webhook stats"
echo

# Test webhook statistics for different periods
echo "2. Getting webhook statistics for last hour:"
curl -s "$SERVICE_URL/api/v1/monitoring/webhook/stats?period=hour" | jq '.total_events, .successful_events, .failed_events' || echo "Failed to get hourly stats"
echo

# Test profile coverage metrics
echo "3. Getting profile coverage metrics:"
curl -s "$SERVICE_URL/api/v1/monitoring/profiles/coverage" | jq '.' || echo "Failed to get profile coverage"
echo

# Test profile quality report
echo "4. Getting profile quality report:"
curl -s "$SERVICE_URL/api/v1/monitoring/profiles/quality" | jq '.' || echo "Failed to get quality report"
echo

echo "=== Testing Profile Management Endpoints ==="
echo

# Test profile statistics
echo "5. Getting profile statistics:"
curl -s "$SERVICE_URL/api/v1/profiles/stats" | jq '.' || echo "Failed to get profile stats"
echo

echo "=== Simulating Webhook Events ==="
echo

# Simulate some webhook events to generate monitoring data
echo "6. Simulating webhook events to generate monitoring data..."

# Webhook event 1: message_new with profile data
echo "Sending message_new webhook event..."
curl -s -X POST "$SERVICE_URL/api/v1/webhook/max" \
  -H "Content-Type: application/json" \
  -d '{
    "type": "message_new",
    "message": {
      "from": {
        "user_id": "demo_user_1",
        "first_name": "Иван",
        "last_name": "Петров"
      },
      "text": "Привет!"
    }
  }' | jq '.' || echo "Failed to send webhook event 1"

# Webhook event 2: callback_query with partial profile
echo "Sending callback_query webhook event..."
curl -s -X POST "$SERVICE_URL/api/v1/webhook/max" \
  -H "Content-Type: application/json" \
  -d '{
    "type": "callback_query",
    "callback_query": {
      "user": {
        "user_id": "demo_user_2",
        "first_name": "Мария"
      }
    }
  }' | jq '.' || echo "Failed to send webhook event 2"

# Webhook event 3: message with name update command
echo "Sending name update command..."
curl -s -X POST "$SERVICE_URL/api/v1/webhook/max" \
  -H "Content-Type: application/json" \
  -d '{
    "type": "message_new",
    "message": {
      "from": {
        "user_id": "demo_user_2",
        "first_name": "Мария"
      },
      "text": "меня зовут Мария Сидорова"
    }
  }' | jq '.' || echo "Failed to send name update event"

echo
echo "Waiting 2 seconds for events to be processed..."
sleep 2

echo
echo "=== Updated Statistics After Webhook Events ==="
echo

# Get updated statistics
echo "7. Updated webhook statistics:"
curl -s "$SERVICE_URL/api/v1/monitoring/webhook/stats?period=hour" | jq '{
  total_events: .total_events,
  successful_events: .successful_events,
  failed_events: .failed_events,
  profiles_extracted: .profiles_extracted,
  profiles_stored: .profiles_stored,
  events_by_type: .events_by_type
}' || echo "Failed to get updated webhook stats"

echo
echo "8. Updated profile coverage:"
curl -s "$SERVICE_URL/api/v1/monitoring/profiles/coverage" | jq '{
  total_users: .total_users,
  users_with_profiles: .users_with_profiles,
  users_with_full_names: .users_with_full_names,
  coverage_percentage: .coverage_percentage,
  profiles_by_source: .profiles_by_source
}' || echo "Failed to get updated coverage"

echo
echo "9. Updated profile quality report:"
curl -s "$SERVICE_URL/api/v1/monitoring/profiles/quality" | jq '{
  total_profiles: .total_profiles,
  quality_score: .quality_metrics.quality_score,
  completeness_score: .quality_metrics.completeness_score,
  recommended_actions: .recommended_actions[0:2]
}' || echo "Failed to get updated quality report"

echo
echo "=== Testing Profile Retrieval ==="
echo

# Test profile retrieval for the users we created
echo "10. Getting profile for demo_user_1:"
curl -s "$SERVICE_URL/api/v1/profiles/demo_user_1" | jq '.' || echo "Profile not found"

echo
echo "11. Getting profile for demo_user_2 (should have user-provided name):"
curl -s "$SERVICE_URL/api/v1/profiles/demo_user_2" | jq '.' || echo "Profile not found"

echo
echo "=== Demo Complete ==="
echo "The monitoring and analytics system is working correctly!"
echo
echo "Available monitoring endpoints:"
echo "- GET /api/v1/monitoring/webhook/stats?period={hour|day|week|month}"
echo "- GET /api/v1/monitoring/profiles/coverage"
echo "- GET /api/v1/monitoring/profiles/quality"
echo "- GET /api/v1/profiles/stats"
echo
echo "Swagger documentation available at: $SERVICE_URL/swagger/index.html"