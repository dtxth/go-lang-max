package http

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetUniversity_InvalidID(t *testing.T) {
	handler := NewHandler(nil, nil, nil, nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/universities/invalid", nil)
	w := httptest.NewRecorder()

	handler.GetUniversity(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestCreateUniversity_InvalidJSON(t *testing.T) {
	handler := NewHandler(nil, nil, nil, nil, nil, nil)

	req := httptest.NewRequest(http.MethodPost, "/universities", bytes.NewReader([]byte("invalid json")))
	w := httptest.NewRecorder()

	handler.CreateUniversity(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestGetStructure_InvalidID(t *testing.T) {
	handler := NewHandler(nil, nil, nil, nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/universities/invalid/structure", nil)
	w := httptest.NewRecorder()

	handler.GetStructure(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestAssignOperator_InvalidJSON(t *testing.T) {
	handler := NewHandler(nil, nil, nil, nil, nil, nil)

	req := httptest.NewRequest(http.MethodPost, "/departments/managers", bytes.NewReader([]byte("invalid json")))
	w := httptest.NewRecorder()

	handler.AssignOperator(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestRemoveOperator_InvalidID(t *testing.T) {
	handler := NewHandler(nil, nil, nil, nil, nil, nil)

	req := httptest.NewRequest(http.MethodDelete, "/departments/managers/invalid", nil)
	w := httptest.NewRecorder()

	handler.RemoveOperator(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestImportExcel_InvalidMethod(t *testing.T) {
	handler := NewHandler(nil, nil, nil, nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/import/excel", nil)
	w := httptest.NewRecorder()

	handler.ImportExcel(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

func TestImportExcel_MissingFile(t *testing.T) {
	handler := NewHandler(nil, nil, nil, nil, nil, nil)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/import/excel", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	handler.ImportExcel(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}


