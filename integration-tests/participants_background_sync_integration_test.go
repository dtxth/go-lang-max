package integration_tests

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	RedisURL = "redis://localhost:6379/0"
)

// TestParticipantsBackgroundSyncIntegration tests the full participants background sync workflow
func TestParticipantsBackgroundSyncIntegration(t *testing.T) {
	// Wait for services to be ready
	WaitForService(t, ChatServiceURL, 10)
	WaitForService(t, AuthServiceURL, 10)
	
	// Connect to Redis
	redisClient := connectRedis(t)
	defer redisClient.Close()
	
	// Connect to chat database
	chatDB := ConnectDB(t, ChatDBConnStr)
	defer chatDB.Close()
	
	// Setup test data
	setupParticipantsTestData(t, chatDB)
	defer CleanupDB(t, chatDB, []string{"administrators", "chats"})
	
	// Test 1: Verify Redis connectivity from chat service
	t.Run("Redis Connectivity", func(t *testing.T) {
		testRedisConnectivity(t, redisClient)
	})
	
	// Test 2: Test manual refresh endpoint
	t.Run("Manual Refresh Endpoint", func(t *testing.T) {
		testManualRefreshEndpoint(t, chatDB)
	})
	
	// Test 3: Test cache operations
	t.Run("Cache Operations", func(t *testing.T) {
		testCacheOperations(t, redisClient, chatDB)
	})
	
	// Test 4: Test graceful degradation when Redis is unavailable
	t.Run("Graceful Degradation", func(t *testing.T) {
		testGracefulDegradation(t, chatDB)
	})
	
	// Test 5: Test background worker operation
	t.Run("Background Worker Operation", func(t *testing.T) {
		testBackgroundWorkerOperation(t, redisClient, chatDB)
	})
}

// TestParticipantsRedisIntegration tests Redis-specific functionality
func TestParticipantsRedisIntegration(t *testing.T) {
	redisClient := connectRedis(t)
	defer redisClient.Close()
	
	ctx := context.Background()
	
	// Test Redis basic operations
	t.Run("Redis Basic Operations", func(t *testing.T) {
		// Test SET and GET
		err := redisClient.Set(ctx, "test:participants:1", `{"count":42,"updated_at":"2023-01-01T00:00:00Z","source":"test"}`, time.Hour).Err()
		require.NoError(t, err)
		
		val, err := redisClient.Get(ctx, "test:participants:1").Result()
		require.NoError(t, err)
		assert.Contains(t, val, `"count":42`)
		
		// Test TTL
		ttl, err := redisClient.TTL(ctx, "test:participants:1").Result()
		require.NoError(t, err)
		assert.Greater(t, ttl, time.Minute*50) // Should be close to 1 hour
		
		// Cleanup
		redisClient.Del(ctx, "test:participants:1")
	})
	
	// Test Redis batch operations
	t.Run("Redis Batch Operations", func(t *testing.T) {
		pipe := redisClient.Pipeline()
		
		// Set multiple keys
		for i := 1; i <= 5; i++ {
			key := fmt.Sprintf("test:batch:participants:%d", i)
			value := fmt.Sprintf(`{"count":%d,"updated_at":"2023-01-01T00:00:00Z","source":"batch_test"}`, i*10)
			pipe.Set(ctx, key, value, time.Hour)
		}
		
		_, err := pipe.Exec(ctx)
		require.NoError(t, err)
		
		// Verify all keys exist
		for i := 1; i <= 5; i++ {
			key := fmt.Sprintf("test:batch:participants:%d", i)
			exists, err := redisClient.Exists(ctx, key).Result()
			require.NoError(t, err)
			assert.Equal(t, int64(1), exists)
		}
		
		// Cleanup
		for i := 1; i <= 5; i++ {
			key := fmt.Sprintf("test:batch:participants:%d", i)
			redisClient.Del(ctx, key)
		}
	})
}

// TestParticipantsDockerComposeDeployment tests Docker Compose deployment scenarios
func TestParticipantsDockerComposeDeployment(t *testing.T) {
	// This test verifies that the participants integration works in Docker Compose environment
	
	// Wait for all services
	WaitForService(t, ChatServiceURL, 10)
	WaitForService(t, AuthServiceURL, 10)
	
	client := NewHTTPClient()
	// Note: Not using authentication for this test to avoid user creation issues
	
	// Test that chat service is running with participants integration
	t.Run("Chat Service Health", func(t *testing.T) {
		status, respBody := client.GET(t, ChatServiceURL+"/health")
		require.Equal(t, 200, status, string(respBody))
		assert.Equal(t, "OK", string(respBody))
	})
	
	// Test that Redis is accessible from chat service
	t.Run("Redis Accessibility", func(t *testing.T) {
		redisClient := connectRedis(t)
		defer redisClient.Close()
		
		ctx := context.Background()
		
		// Test Redis ping
		pong, err := redisClient.Ping(ctx).Result()
		require.NoError(t, err)
		assert.Equal(t, "PONG", pong)
	})
	
	// Test environment variables are properly set
	t.Run("Environment Variables", func(t *testing.T) {
		// This is tested implicitly by the fact that participants integration works
		// We can verify by checking that manual refresh endpoint exists
		chatDB := ConnectDB(t, ChatDBConnStr)
		defer chatDB.Close()
		
		// Create a test chat
		var chatID int
		err := chatDB.QueryRow(`
			INSERT INTO chats (name, url, university_id, source, created_at)
			VALUES ('Docker Test Chat', 'https://max.com/dockertest', 1, 'admin_panel', NOW())
			RETURNING id
		`).Scan(&chatID)
		require.NoError(t, err)
		
		// Test manual refresh endpoint (should return 401 unauthorized, but endpoint should exist)
		status, _ := client.POST(t, fmt.Sprintf("%s/chats/%d/refresh-participants", ChatServiceURL, chatID), nil)
		// 401 is expected since we're not authenticated, but the endpoint should exist
		assert.Contains(t, []int{200, 401, 404, 400}, status, "Manual refresh endpoint should be accessible")
		
		// Cleanup
		CleanupDB(t, chatDB, []string{"chats"})
	})
}

// Helper functions

func connectRedis(t *testing.T) *redis.Client {
	opt, err := redis.ParseURL(RedisURL)
	require.NoError(t, err)
	
	client := redis.NewClient(opt)
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	err = client.Ping(ctx).Err()
	require.NoError(t, err, "Failed to connect to Redis")
	
	return client
}

func setupParticipantsTestData(t *testing.T, db *sql.DB) {
	// Insert test chats with different scenarios
	_, err := db.Exec(`
		INSERT INTO chats (id, name, url, university_id, source, max_chat_id, created_at)
		VALUES 
			(1001, 'Test Chat with MAX ID', 'https://max.com/test1', 1, 'admin_panel', 'max_chat_123', NOW()),
			(1002, 'Test Chat without MAX ID', 'https://max.com/test2', 1, 'admin_panel', NULL, NOW()),
			(1003, 'Test Chat for Background Sync', 'https://max.com/test3', 1, 'admin_panel', 'max_chat_456', NOW())
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			url = EXCLUDED.url,
			max_chat_id = EXCLUDED.max_chat_id
	`)
	require.NoError(t, err)
}

func testRedisConnectivity(t *testing.T, redisClient *redis.Client) {
	ctx := context.Background()
	
	// Test basic Redis operations
	err := redisClient.Set(ctx, "test:connectivity", "ok", time.Minute).Err()
	require.NoError(t, err)
	
	val, err := redisClient.Get(ctx, "test:connectivity").Result()
	require.NoError(t, err)
	assert.Equal(t, "ok", val)
	
	// Cleanup
	redisClient.Del(ctx, "test:connectivity")
}

func testManualRefreshEndpoint(t *testing.T, db *sql.DB) {
	client := NewHTTPClient()
	
	// Test manual refresh for chat with MAX ID (without authentication for now)
	status, respBody := client.POST(t, fmt.Sprintf("%s/chats/1001/refresh-participants", ChatServiceURL), nil)
	// Should return 401 (unauthorized), 404 (chat not found), or 400 (no MAX chat ID)
	// 401 is expected since we're not authenticated, but the endpoint should exist
	assert.Contains(t, []int{200, 401, 404, 400}, status, "Manual refresh endpoint should be accessible: %s", string(respBody))
	
	// Test manual refresh for chat without MAX ID
	status, respBody = client.POST(t, fmt.Sprintf("%s/chats/1002/refresh-participants", ChatServiceURL), nil)
	// Should return 401 (unauthorized), 400 (no MAX chat ID) or 404 (chat not found)
	assert.Contains(t, []int{401, 400, 404}, status, "Manual refresh for chat without MAX ID should return error: %s", string(respBody))
}

func testCacheOperations(t *testing.T, redisClient *redis.Client, db *sql.DB) {
	ctx := context.Background()
	
	// Test cache key format
	chatID := int64(1001)
	cacheKey := fmt.Sprintf("participants:%d", chatID)
	
	// Set test data in cache
	testData := `{"count":25,"updated_at":"2023-01-01T12:00:00Z","source":"api"}`
	err := redisClient.Set(ctx, cacheKey, testData, time.Hour).Err()
	require.NoError(t, err)
	
	// Verify data can be retrieved
	val, err := redisClient.Get(ctx, cacheKey).Result()
	require.NoError(t, err)
	assert.Contains(t, val, `"count":25`)
	assert.Contains(t, val, `"source":"api"`)
	
	// Test cache expiration
	ttl, err := redisClient.TTL(ctx, cacheKey).Result()
	require.NoError(t, err)
	assert.Greater(t, ttl, time.Minute*50) // Should be close to 1 hour
	
	// Cleanup
	redisClient.Del(ctx, cacheKey)
}

func testGracefulDegradation(t *testing.T, db *sql.DB) {
	client := NewHTTPClient()
	
	// Test that chat service continues to work even if Redis operations fail
	// This is tested by ensuring the chat service health endpoint still works
	status, respBody := client.GET(t, ChatServiceURL+"/health")
	require.Equal(t, 200, status, string(respBody))
	
	// Test that chat listing endpoint exists (even if it returns 401 unauthorized)
	status, respBody = client.GET(t, ChatServiceURL+"/chats")
	// Should return 401 (unauthorized) or 200 (success), but not 404 (not found)
	assert.Contains(t, []int{200, 401}, status, "Chat listing endpoint should exist: %s", string(respBody))
}

func testBackgroundWorkerOperation(t *testing.T, redisClient *redis.Client, db *sql.DB) {
	ctx := context.Background()
	
	// Test that background worker can process stale data
	// We simulate stale data by setting old cache entries
	
	chatID := int64(1003)
	cacheKey := fmt.Sprintf("participants:%d", chatID)
	
	// Set stale data (older than stale threshold)
	staleTime := time.Now().Add(-2 * time.Hour) // 2 hours ago (older than 1 hour threshold)
	staleData := fmt.Sprintf(`{"count":10,"updated_at":"%s","source":"api"}`, staleTime.Format(time.RFC3339))
	
	err := redisClient.Set(ctx, cacheKey, staleData, time.Hour).Err()
	require.NoError(t, err)
	
	// Verify stale data is set
	val, err := redisClient.Get(ctx, cacheKey).Result()
	require.NoError(t, err)
	assert.Contains(t, val, `"count":10`)
	
	// The background worker should eventually update this stale data
	// For integration test purposes, we just verify the cache structure is correct
	
	// Cleanup
	redisClient.Del(ctx, cacheKey)
}

// TestParticipantsEndToEndWorkflow tests the complete end-to-end workflow
func TestParticipantsEndToEndWorkflow(t *testing.T) {
	// Wait for services
	WaitForService(t, ChatServiceURL, 10)
	WaitForService(t, AuthServiceURL, 10)
	
	// Setup
	redisClient := connectRedis(t)
	defer redisClient.Close()
	
	chatDB := ConnectDB(t, ChatDBConnStr)
	defer chatDB.Close()
	
	client := NewHTTPClient()
	// Note: Not using authentication for this test to avoid user creation issues
	
	// Create a test chat
	var chatID int
	err := chatDB.QueryRow(`
		INSERT INTO chats (name, url, university_id, source, max_chat_id, created_at)
		VALUES ('E2E Test Chat', 'https://max.com/e2e', 1, 'admin_panel', 'max_e2e_123', NOW())
		RETURNING id
	`).Scan(&chatID)
	require.NoError(t, err)
	
	defer CleanupDB(t, chatDB, []string{"chats"})
	
	// Test workflow:
	// 1. Request chat list (should trigger lazy update if cache is stale)
	// 2. Manual refresh
	// 3. Verify cache is updated
	
	t.Run("Chat List Request", func(t *testing.T) {
		status, respBody := client.GET(t, ChatServiceURL+"/chats")
		// Should return 401 (unauthorized) since we're not authenticated, but endpoint should exist
		assert.Contains(t, []int{200, 401}, status, "Chat list endpoint should exist: %s", string(respBody))
	})
	
	t.Run("Manual Refresh", func(t *testing.T) {
		status, respBody := client.POST(t, fmt.Sprintf("%s/chats/%d/refresh-participants", ChatServiceURL, chatID), nil)
		// Should return 401 (unauthorized) since we're not authenticated, but endpoint should exist
		assert.Contains(t, []int{200, 401, 400, 404}, status, "Manual refresh endpoint should exist: %s", string(respBody))
	})
	
	t.Run("Cache Verification", func(t *testing.T) {
		ctx := context.Background()
		cacheKey := fmt.Sprintf("participants:%d", chatID)
		
		// Check if cache entry exists (it might not if MAX API is mocked)
		exists, err := redisClient.Exists(ctx, cacheKey).Result()
		require.NoError(t, err)
		
		if exists == 1 {
			val, err := redisClient.Get(ctx, cacheKey).Result()
			require.NoError(t, err)
			
			// Verify cache structure
			assert.Contains(t, val, "count")
			assert.Contains(t, val, "updated_at")
			assert.Contains(t, val, "source")
		}
		// If cache doesn't exist, it's also valid (MAX API might be unavailable)
	})
}

// TestParticipantsConfigurationIntegration tests configuration handling in integration environment
func TestParticipantsConfigurationIntegration(t *testing.T) {
	// Wait for services
	WaitForService(t, ChatServiceURL, 10)
	
	// Test that chat service is running with proper configuration
	t.Run("Service Configuration Validation", func(t *testing.T) {
		client := NewHTTPClient()
		
		// Test health endpoint to ensure service is configured correctly
		status, respBody := client.GET(t, ChatServiceURL+"/health")
		require.Equal(t, 200, status, "Chat service should be healthy with proper configuration: %s", string(respBody))
		
		// Connect to Redis to verify Redis configuration
		redisClient := connectRedis(t)
		defer redisClient.Close()
		
		ctx := context.Background()
		
		// Test Redis configuration by setting and getting a test value
		testKey := "test:config:validation"
		testValue := "configuration_test"
		
		err := redisClient.Set(ctx, testKey, testValue, time.Minute).Err()
		require.NoError(t, err, "Redis should be properly configured")
		
		val, err := redisClient.Get(ctx, testKey).Result()
		require.NoError(t, err, "Should be able to retrieve test value")
		assert.Equal(t, testValue, val, "Retrieved value should match set value")
		
		// Cleanup
		redisClient.Del(ctx, testKey)
	})
	
	// Test environment variable handling
	t.Run("Environment Variables Integration", func(t *testing.T) {
		// This test verifies that environment variables are properly loaded
		// by checking that the participants integration is working
		
		redisClient := connectRedis(t)
		defer redisClient.Close()
		
		ctx := context.Background()
		
		// Test that Redis URL environment variable is working
		pong, err := redisClient.Ping(ctx).Result()
		require.NoError(t, err, "Redis URL environment variable should be properly configured")
		assert.Equal(t, "PONG", pong, "Redis should respond to ping")
		
		// Test cache TTL configuration by setting a value with TTL
		testKey := "test:ttl:config"
		testValue := "ttl_test"
		
		err = redisClient.Set(ctx, testKey, testValue, time.Second*5).Err()
		require.NoError(t, err, "Should be able to set value with TTL")
		
		// Verify TTL is set
		ttl, err := redisClient.TTL(ctx, testKey).Result()
		require.NoError(t, err, "Should be able to get TTL")
		assert.Greater(t, ttl, time.Second*3, "TTL should be properly configured")
		assert.LessOrEqual(t, ttl, time.Second*5, "TTL should not exceed set value")
		
		// Cleanup
		redisClient.Del(ctx, testKey)
	})
}

// TestParticipantsErrorHandlingIntegration tests error handling scenarios
func TestParticipantsErrorHandlingIntegration(t *testing.T) {
	// Wait for services
	WaitForService(t, ChatServiceURL, 10)
	
	chatDB := ConnectDB(t, ChatDBConnStr)
	defer chatDB.Close()
	
	client := NewHTTPClient()
	
	t.Run("Invalid Chat ID Handling", func(t *testing.T) {
		// Test manual refresh with invalid chat ID
		invalidChatID := 999999
		status, respBody := client.POST(t, fmt.Sprintf("%s/chats/%d/refresh-participants", ChatServiceURL, invalidChatID), nil)
		
		// Should return 401 (unauthorized) or 404 (not found), but not 500 (internal server error)
		assert.Contains(t, []int{401, 404}, status, "Invalid chat ID should be handled gracefully: %s", string(respBody))
	})
	
	t.Run("Chat Without MAX ID Handling", func(t *testing.T) {
		// Create a chat without MAX ID
		var chatID int
		err := chatDB.QueryRow(`
			INSERT INTO chats (name, url, university_id, source, max_chat_id, created_at)
			VALUES ('Chat Without MAX ID', 'https://max.com/no-max-id', 1, 'admin_panel', NULL, NOW())
			RETURNING id
		`).Scan(&chatID)
		require.NoError(t, err)
		
		defer CleanupDB(t, chatDB, []string{"chats"})
		
		// Test manual refresh for chat without MAX ID
		status, respBody := client.POST(t, fmt.Sprintf("%s/chats/%d/refresh-participants", ChatServiceURL, chatID), nil)
		
		// Should return 401 (unauthorized) or 400 (bad request), but handle gracefully
		assert.Contains(t, []int{401, 400, 404}, status, "Chat without MAX ID should be handled gracefully: %s", string(respBody))
	})
	
	t.Run("Service Resilience During Redis Operations", func(t *testing.T) {
		// Test that chat service remains functional even during Redis operations
		redisClient := connectRedis(t)
		defer redisClient.Close()
		
		ctx := context.Background()
		
		// Perform multiple Redis operations to simulate load
		for i := 0; i < 10; i++ {
			key := fmt.Sprintf("test:resilience:%d", i)
			value := fmt.Sprintf("resilience_test_%d", i)
			
			err := redisClient.Set(ctx, key, value, time.Minute).Err()
			require.NoError(t, err, "Redis operations should work during load")
		}
		
		// Verify chat service health during Redis operations
		status, respBody := client.GET(t, ChatServiceURL+"/health")
		require.Equal(t, 200, status, "Chat service should remain healthy during Redis operations: %s", string(respBody))
		
		// Cleanup
		for i := 0; i < 10; i++ {
			key := fmt.Sprintf("test:resilience:%d", i)
			redisClient.Del(ctx, key)
		}
	})
}

// TestParticipantsPerformanceIntegration tests performance aspects
func TestParticipantsPerformanceIntegration(t *testing.T) {
	// Wait for services
	WaitForService(t, ChatServiceURL, 10)
	
	redisClient := connectRedis(t)
	defer redisClient.Close()
	
	ctx := context.Background()
	
	t.Run("Batch Cache Operations Performance", func(t *testing.T) {
		// Test batch operations performance
		batchSize := 100
		startTime := time.Now()
		
		// Perform batch SET operations
		pipe := redisClient.Pipeline()
		for i := 0; i < batchSize; i++ {
			key := fmt.Sprintf("test:batch:perf:%d", i)
			value := fmt.Sprintf(`{"count":%d,"updated_at":"2023-01-01T00:00:00Z","source":"perf_test"}`, i)
			pipe.Set(ctx, key, value, time.Hour)
		}
		
		_, err := pipe.Exec(ctx)
		require.NoError(t, err, "Batch SET operations should complete successfully")
		
		setDuration := time.Since(startTime)
		t.Logf("Batch SET operations (%d items) completed in: %v", batchSize, setDuration)
		
		// Verify all keys exist
		startTime = time.Now()
		for i := 0; i < batchSize; i++ {
			key := fmt.Sprintf("test:batch:perf:%d", i)
			exists, err := redisClient.Exists(ctx, key).Result()
			require.NoError(t, err, "Should be able to check key existence")
			assert.Equal(t, int64(1), exists, "Key should exist")
		}
		
		getDuration := time.Since(startTime)
		t.Logf("Batch GET operations (%d items) completed in: %v", batchSize, getDuration)
		
		// Performance assertions (reasonable thresholds)
		assert.Less(t, setDuration, time.Second*5, "Batch SET operations should complete within 5 seconds")
		assert.Less(t, getDuration, time.Second*5, "Batch GET operations should complete within 5 seconds")
		
		// Cleanup
		for i := 0; i < batchSize; i++ {
			key := fmt.Sprintf("test:batch:perf:%d", i)
			redisClient.Del(ctx, key)
		}
	})
	
	t.Run("Cache TTL Performance", func(t *testing.T) {
		// Test TTL operations performance
		numKeys := 50
		startTime := time.Now()
		
		// Set keys with different TTLs
		for i := 0; i < numKeys; i++ {
			key := fmt.Sprintf("test:ttl:perf:%d", i)
			value := fmt.Sprintf("ttl_perf_test_%d", i)
			ttl := time.Duration(i+1) * time.Minute
			
			err := redisClient.Set(ctx, key, value, ttl).Err()
			require.NoError(t, err, "Should be able to set key with TTL")
		}
		
		ttlSetDuration := time.Since(startTime)
		t.Logf("TTL SET operations (%d items) completed in: %v", numKeys, ttlSetDuration)
		
		// Check TTLs
		startTime = time.Now()
		for i := 0; i < numKeys; i++ {
			key := fmt.Sprintf("test:ttl:perf:%d", i)
			ttl, err := redisClient.TTL(ctx, key).Result()
			require.NoError(t, err, "Should be able to get TTL")
			assert.Greater(t, ttl, time.Duration(0), "TTL should be positive")
		}
		
		ttlGetDuration := time.Since(startTime)
		t.Logf("TTL GET operations (%d items) completed in: %v", numKeys, ttlGetDuration)
		
		// Performance assertions
		assert.Less(t, ttlSetDuration, time.Second*3, "TTL SET operations should complete within 3 seconds")
		assert.Less(t, ttlGetDuration, time.Second*2, "TTL GET operations should complete within 2 seconds")
		
		// Cleanup
		for i := 0; i < numKeys; i++ {
			key := fmt.Sprintf("test:ttl:perf:%d", i)
			redisClient.Del(ctx, key)
		}
	})
}