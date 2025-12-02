package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetEmployeeByID_InvalidID(t *testing.T) {
	handler := NewHandler(nil, nil, nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/employees/invalid", nil)
	w := httptest.NewRecorder()

	handler.GetEmployeeByID(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestAddEmployee_MissingPhone(t *testing.T) {
	handler := NewHandler(nil, nil, nil, nil, nil)

	reqBody := map[string]string{
		"first_name": "Иван",
		"last_name":  "Иванов",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/employees", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.AddEmployee(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestAddEmployee_MissingName(t *testing.T) {
	handler := NewHandler(nil, nil, nil, nil, nil)

	reqBody := map[string]string{
		"phone": "+79001234567",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/employees", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.AddEmployee(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestAddEmployee_InvalidJSON(t *testing.T) {
	handler := NewHandler(nil, nil, nil, nil, nil)

	req := httptest.NewRequest(http.MethodPost, "/employees", bytes.NewReader([]byte("invalid json")))
	w := httptest.NewRecorder()

	handler.AddEmployee(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestUpdateEmployee_InvalidID(t *testing.T) {
	handler := NewHandler(nil, nil, nil, nil, nil)

	reqBody := map[string]string{
		"first_name": "Петр",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPut, "/employees/invalid", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.UpdateEmployee(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestUpdateEmployee_InvalidJSON(t *testing.T) {
	handler := NewHandler(nil, nil, nil, nil, nil)

	req := httptest.NewRequest(http.MethodPut, "/employees/1", bytes.NewReader([]byte("invalid json")))
	w := httptest.NewRecorder()

	handler.UpdateEmployee(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestDeleteEmployee_InvalidID(t *testing.T) {
	handler := NewHandler(nil, nil, nil, nil, nil)

	req := httptest.NewRequest(http.MethodDelete, "/employees/invalid", nil)
	w := httptest.NewRecorder()

	handler.DeleteEmployee(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestSearchEmployees_MissingAuth(t *testing.T) {
	handler := NewHandler(nil, nil, nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/employees?query=Иван", nil)
	w := httptest.NewRecorder()

	handler.SearchEmployees(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", w.Code)
	}
}

func TestSearchEmployees_InvalidAuthFormat(t *testing.T) {
	handler := NewHandler(nil, nil, nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/employees?query=Иван", nil)
	req.Header.Set("Authorization", "InvalidFormat")
	w := httptest.NewRecorder()

	handler.SearchEmployees(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", w.Code)
	}
}

func TestBatchUpdateMaxID_ServiceNotAvailable(t *testing.T) {
	handler := NewHandler(nil, nil, nil, nil, nil)

	req := httptest.NewRequest(http.MethodPost, "/employees/batch-update-maxid", nil)
	w := httptest.NewRecorder()

	handler.BatchUpdateMaxID(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("expected status 503, got %d", w.Code)
	}
}

func TestGetBatchStatus_ServiceNotAvailable(t *testing.T) {
	handler := NewHandler(nil, nil, nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/employees/batch-status/1", nil)
	w := httptest.NewRecorder()

	handler.GetBatchStatus(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("expected status 503, got %d", w.Code)
	}
}

func TestGetBatchStatus_InvalidID(t *testing.T) {
	handler := NewHandler(nil, nil, nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/employees/batch-status/invalid", nil)
	w := httptest.NewRecorder()

	handler.GetBatchStatus(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("expected status 503, got %d", w.Code)
	}
}

func TestGetAllBatchJobs_ServiceNotAvailable(t *testing.T) {
	handler := NewHandler(nil, nil, nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/employees/batch-status?limit=10", nil)
	w := httptest.NewRecorder()

	handler.GetAllBatchJobs(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("expected status 503, got %d", w.Code)
	}
}
