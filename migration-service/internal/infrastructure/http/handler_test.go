package http

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestStartDatabaseMigration_InvalidMethod(t *testing.T) {
	handler := NewHandler(nil, nil, nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/migration/database", nil)
	w := httptest.NewRecorder()

	handler.StartDatabaseMigration(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

func TestStartDatabaseMigration_InvalidJSON(t *testing.T) {
	handler := NewHandler(nil, nil, nil, nil, nil)

	req := httptest.NewRequest(http.MethodPost, "/migration/database", bytes.NewReader([]byte("invalid json")))
	w := httptest.NewRecorder()

	handler.StartDatabaseMigration(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestStartGoogleSheetsMigration_InvalidMethod(t *testing.T) {
	handler := NewHandler(nil, nil, nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/migration/google-sheets", nil)
	w := httptest.NewRecorder()

	handler.StartGoogleSheetsMigration(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

func TestStartGoogleSheetsMigration_InvalidJSON(t *testing.T) {
	handler := NewHandler(nil, nil, nil, nil, nil)

	req := httptest.NewRequest(http.MethodPost, "/migration/google-sheets", bytes.NewReader([]byte("invalid json")))
	w := httptest.NewRecorder()

	handler.StartGoogleSheetsMigration(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestStartGoogleSheetsMigration_MissingSpreadsheetID(t *testing.T) {
	handler := NewHandler(nil, nil, nil, nil, nil)

	reqBody := StartGoogleSheetsMigrationRequest{}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/migration/google-sheets", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.StartGoogleSheetsMigration(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestStartExcelMigration_InvalidMethod(t *testing.T) {
	handler := NewHandler(nil, nil, nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/migration/excel", nil)
	w := httptest.NewRecorder()

	handler.StartExcelMigration(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

func TestStartExcelMigration_MissingFile(t *testing.T) {
	handler := NewHandler(nil, nil, nil, nil, nil)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/migration/excel", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	handler.StartExcelMigration(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestGetMigrationJob_InvalidMethod(t *testing.T) {
	handler := NewHandler(nil, nil, nil, nil, nil)

	req := httptest.NewRequest(http.MethodPost, "/migration/jobs/1", nil)
	w := httptest.NewRecorder()

	handler.GetMigrationJob(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

func TestGetMigrationJob_InvalidID(t *testing.T) {
	handler := NewHandler(nil, nil, nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/migration/jobs/invalid", nil)
	w := httptest.NewRecorder()

	handler.GetMigrationJob(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestListMigrationJobs_InvalidMethod(t *testing.T) {
	handler := NewHandler(nil, nil, nil, nil, nil)

	req := httptest.NewRequest(http.MethodPost, "/migration/jobs", nil)
	w := httptest.NewRecorder()

	handler.ListMigrationJobs(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}
