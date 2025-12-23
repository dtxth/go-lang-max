package maxapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_GetInternalUsers_RealAPI(t *testing.T) {
	// Create a test server that simulates MAX API /internal/users
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/internal/users", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "test-token", r.Header.Get("Authorization"))

		// Parse request body
		var req InternalUsersRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		// Create mock response
		response := InternalUsersResponse{
			Users: []InternalUserAPIResponse{
				{
					UserID:        123456789,
					FirstName:     "Иван",
					LastName:      "Иванов",
					IsBot:         false,
					Username:      "ivan_ivanov",
					AvatarURL:     "https://max.ru/avatars/123_small.jpg",
					FullAvatarURL: "https://max.ru/avatars/123_full.jpg",
					Link:          "max.ru/ivan_ivanov",
					PhoneNumber:   "+79991234567",
				},
				{
					UserID:        987654321,
					FirstName:     "Мария",
					LastName:      "Петрова",
					IsBot:         false,
					Username:      "",
					AvatarURL:     "https://max.ru/avatars/987_small.jpg",
					FullAvatarURL: "https://max.ru/avatars/987_full.jpg",
					Link:          "max.ru/u/abc123hash",
					PhoneNumber:   "+79995678901",
				},
			},
			FailedPhoneNumbers: []string{"invalid_phone"},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create client with test server URL
	client := &Client{
		baseURL: server.URL,
		token:   "test-token",
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}

	ctx := context.Background()
	phones := []string{"+79991234567", "+79995678901", "invalid_phone"}

	// Test the method
	users, failedPhones, err := client.callInternalUsersAPI(ctx, phones)

	// Verify results
	require.NoError(t, err)
	assert.Len(t, users, 2)
	assert.Len(t, failedPhones, 1)
	assert.Contains(t, failedPhones, "invalid_phone")

	// Verify first user
	user1 := users[0]
	assert.Equal(t, int64(123456789), user1.UserID)
	assert.Equal(t, "Иван", user1.FirstName)
	assert.Equal(t, "Иванов", user1.LastName)
	assert.Equal(t, "+79991234567", user1.PhoneNumber)
	assert.Equal(t, "ivan_ivanov", user1.Username)
	assert.Equal(t, "max.ru/ivan_ivanov", user1.Link)
	assert.False(t, user1.IsBot)

	// Verify second user (no username)
	user2 := users[1]
	assert.Equal(t, int64(987654321), user2.UserID)
	assert.Equal(t, "Мария", user2.FirstName)
	assert.Equal(t, "Петрова", user2.LastName)
	assert.Equal(t, "+79995678901", user2.PhoneNumber)
	assert.Empty(t, user2.Username)
	assert.Equal(t, "max.ru/u/abc123hash", user2.Link)
}

func TestClient_GetInternalUsers_APIError(t *testing.T) {
	// Create a test server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer server.Close()

	// Create client with test server URL
	client := &Client{
		baseURL: server.URL,
		token:   "test-token",
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}

	ctx := context.Background()
	phones := []string{"+79991234567"}

	// Test the method
	users, failedPhones, err := client.callInternalUsersAPI(ctx, phones)

	// Verify error handling
	require.Error(t, err)
	assert.Nil(t, users)
	assert.Nil(t, failedPhones)
	assert.Contains(t, err.Error(), "MAX API returned status 500")
}

func TestClient_GetInternalUsers_Fallback(t *testing.T) {
	// Create a mock client that will use fallback
	mockClient := NewMockClient()
	
	// Override the CheckPhoneNumbers method to simulate existing phones
	ctx := context.Background()
	phones := []string{"+79991234567", "+79995678901"}

	// Test fallback method directly
	users, failedPhones, err := mockClient.GetInternalUsers(ctx, phones)

	// Verify fallback works
	require.NoError(t, err)
	assert.Len(t, users, 2)
	assert.Empty(t, failedPhones)

	// Verify fallback data structure
	for _, user := range users {
		assert.Greater(t, user.UserID, int64(0))
		assert.NotEmpty(t, user.PhoneNumber)
		assert.NotEmpty(t, user.Link)
		// In mock, names are populated
		assert.NotEmpty(t, user.FirstName)
		assert.NotEmpty(t, user.LastName)
	}
}

func TestClient_GetInternalUsers_Integration(t *testing.T) {
	// This test demonstrates the full flow with fallback
	
	// Create a server that simulates MAX API being unavailable
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate network timeout or server error
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	// Create real client that will fallback to mock behavior
	client := &Client{
		baseURL: server.URL,
		token:   "test-token",
		client: &http.Client{
			Timeout: 50 * time.Millisecond, // Short timeout to trigger fallback
		},
	}

	ctx := context.Background()
	phones := []string{"+79991234567"}

	// Test the method - should fallback gracefully
	users, failedPhones, err := client.GetInternalUsers(ctx, phones)

	// Should succeed with fallback (all phones marked as failed)
	require.NoError(t, err)
	assert.Empty(t, users)
	assert.Len(t, failedPhones, 1)
	assert.Contains(t, failedPhones, "+79991234567")
}

func TestInternalUsersRequest_JSON(t *testing.T) {
	// Test JSON marshaling/unmarshaling
	req := InternalUsersRequest{
		PhoneNumbers: []string{"+79991234567", "+79995678901"},
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(req)
	require.NoError(t, err)

	// Unmarshal back
	var unmarshaled InternalUsersRequest
	err = json.Unmarshal(jsonData, &unmarshaled)
	require.NoError(t, err)

	// Verify
	assert.Equal(t, req.PhoneNumbers, unmarshaled.PhoneNumbers)
}

func TestInternalUsersResponse_JSON(t *testing.T) {
	// Test JSON marshaling/unmarshaling
	resp := InternalUsersResponse{
		Users: []InternalUserAPIResponse{
			{
				UserID:        123456789,
				FirstName:     "Иван",
				LastName:      "Иванов",
				IsBot:         false,
				Username:      "ivan_ivanov",
				AvatarURL:     "https://max.ru/avatars/123_small.jpg",
				FullAvatarURL: "https://max.ru/avatars/123_full.jpg",
				Link:          "max.ru/ivan_ivanov",
				PhoneNumber:   "+79991234567",
			},
		},
		FailedPhoneNumbers: []string{"invalid_phone"},
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(resp)
	require.NoError(t, err)

	// Unmarshal back
	var unmarshaled InternalUsersResponse
	err = json.Unmarshal(jsonData, &unmarshaled)
	require.NoError(t, err)

	// Verify
	assert.Len(t, unmarshaled.Users, 1)
	assert.Equal(t, resp.Users[0].UserID, unmarshaled.Users[0].UserID)
	assert.Equal(t, resp.Users[0].FirstName, unmarshaled.Users[0].FirstName)
	assert.Equal(t, resp.FailedPhoneNumbers, unmarshaled.FailedPhoneNumbers)
}