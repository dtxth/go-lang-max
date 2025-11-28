package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGenerateRequestID(t *testing.T) {
	id1 := GenerateRequestID()
	id2 := GenerateRequestID()
	
	if id1 == "" {
		t.Error("Expected non-empty request ID")
	}
	
	if id1 == id2 {
		t.Error("Expected unique request IDs")
	}
	
	if len(id1) != 32 {
		t.Errorf("Expected request ID length 32, got %d", len(id1))
	}
}

func TestGetRequestID(t *testing.T) {
	requestID := "test-request-id"
	ctx := context.WithValue(context.Background(), RequestIDKey, requestID)
	
	retrieved := GetRequestID(ctx)
	if retrieved != requestID {
		t.Errorf("Expected request ID '%s', got '%s'", requestID, retrieved)
	}
}

func TestGetRequestIDMissing(t *testing.T) {
	ctx := context.Background()
	
	retrieved := GetRequestID(ctx)
	if retrieved != "unknown" {
		t.Errorf("Expected 'unknown' for missing request ID, got '%s'", retrieved)
	}
}

func TestRequestIDMiddleware(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := GetRequestID(r.Context())
		if requestID == "" {
			t.Error("Expected request ID in context")
		}
		w.WriteHeader(http.StatusOK)
	})
	
	middleware := RequestIDMiddleware(handler)
	
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	
	middleware.ServeHTTP(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	requestID := w.Header().Get("X-Request-ID")
	if requestID == "" {
		t.Error("Expected X-Request-ID header in response")
	}
}

func TestRequestIDMiddlewareWithExistingID(t *testing.T) {
	existingID := "existing-request-id"
	
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := GetRequestID(r.Context())
		if requestID != existingID {
			t.Errorf("Expected request ID '%s', got '%s'", existingID, requestID)
		}
		w.WriteHeader(http.StatusOK)
	})
	
	middleware := RequestIDMiddleware(handler)
	
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Request-ID", existingID)
	w := httptest.NewRecorder()
	
	middleware.ServeHTTP(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	requestID := w.Header().Get("X-Request-ID")
	if requestID != existingID {
		t.Errorf("Expected X-Request-ID '%s', got '%s'", existingID, requestID)
	}
}
