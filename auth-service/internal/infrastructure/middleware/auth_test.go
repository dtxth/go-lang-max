package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAuthMiddleware_MissingAuthorizationHeader(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	w := httptest.NewRecorder()

	// Test that missing header returns 401
	middleware := AuthMiddleware(nil)
	middleware(handler).ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", w.Code)
	}
}

func TestAuthMiddleware_InvalidAuthorizationFormat(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	req.Header.Set("Authorization", "InvalidFormat")
	w := httptest.NewRecorder()

	middleware := AuthMiddleware(nil)
	middleware(handler).ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", w.Code)
	}
}

func TestAuthMiddleware_MissingBearerToken(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	req.Header.Set("Authorization", "Bearer")
	w := httptest.NewRecorder()

	middleware := AuthMiddleware(nil)
	middleware(handler).ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", w.Code)
	}
}

func TestGetRequestID_WithRequestID(t *testing.T) {
	ctx := context.WithValue(context.Background(), RequestIDKey, "test-request-id")
	
	requestID := GetRequestID(ctx)
	
	if requestID != "test-request-id" {
		t.Errorf("expected request ID 'test-request-id', got '%s'", requestID)
	}
}

func TestGetRequestID_WithoutRequestID(t *testing.T) {
	ctx := context.Background()
	
	requestID := GetRequestID(ctx)
	
	if requestID != "unknown" {
		t.Errorf("expected request ID 'unknown', got '%s'", requestID)
	}
}
