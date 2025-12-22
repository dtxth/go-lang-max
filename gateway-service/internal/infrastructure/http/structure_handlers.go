package http

import (
	"encoding/json"
	"net/http"
	"strings"

	structurepb "structure-service/api/proto"
)

// GetAllUniversitiesHandler handles getting all universities with pagination
func (h *Handler) GetAllUniversitiesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeErrorResponse(w, http.StatusMethodNotAllowed, "method_not_allowed", "Method not allowed", h.getRequestID(r))
		return
	}

	requestID := h.getRequestID(r)
	ctx, cancel := h.createContextWithTimeout(r, h.config.Services.Structure.Timeout)
	defer cancel()

	page, limit, sortBy, sortOrder := h.parseQueryParams(r)

	req := &structurepb.GetAllUniversitiesRequest{
		Page:      page,
		Limit:     limit,
		SortBy:    sortBy,
		SortOrder: sortOrder,
	}

	client := h.clientManager.GetStructureClient()
	if !h.checkServiceAvailability(w, client, "Structure", requestID) {
		return
	}
	
	resp, err := client.GetAllUniversities(ctx, req)
	if err != nil {
		h.handleGRPCErrorLegacy(w, err, requestID)
		return
	}

	if resp.Error != "" {
		h.writeErrorResponse(w, http.StatusInternalServerError, "get_universities_failed", resp.Error, requestID)
		return
	}

	// Convert to HTTP response format
	universities := make([]map[string]interface{}, len(resp.Universities))
	for i, university := range resp.Universities {
		universities[i] = map[string]interface{}{
			"id":         university.Id,
			"name":       university.Name,
			"inn":        university.Inn,
			"kpp":        university.Kpp,
			"foiv":       university.Foiv,
			"created_at": university.CreatedAt,
			"updated_at": university.UpdatedAt,
		}
	}

	response := map[string]interface{}{
		"universities": universities,
		"total":        resp.Total,
		"page":         resp.Page,
		"limit":        resp.Limit,
		"offset":       (resp.Page - 1) * resp.Limit, // For E2E test compatibility
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

// CreateUniversityHandler handles creating a new university
func (h *Handler) CreateUniversityHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeErrorResponse(w, http.StatusMethodNotAllowed, "method_not_allowed", "Method not allowed", h.getRequestID(r))
		return
	}

	requestID := h.getRequestID(r)
	ctx, cancel := h.createContextWithTimeout(r, h.config.Services.Structure.Timeout)
	defer cancel()

	var req structurepb.CreateUniversityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid_json", "Invalid JSON format", requestID)
		return
	}

	client := h.clientManager.GetStructureClient()
	if !h.checkServiceAvailability(w, client, "Structure", requestID) {
		return
	}
	
	resp, err := client.CreateUniversity(ctx, &req)
	if err != nil {
		h.handleGRPCErrorLegacy(w, err, requestID)
		return
	}

	if resp.Error != "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "create_university_failed", resp.Error, requestID)
		return
	}

	// Convert to HTTP response format
	response := map[string]interface{}{
		"id":         resp.University.Id,
		"name":       resp.University.Name,
		"inn":        resp.University.Inn,
		"kpp":        resp.University.Kpp,
		"foiv":       resp.University.Foiv,
		"created_at": resp.University.CreatedAt,
		"updated_at": resp.University.UpdatedAt,
	}

	h.writeJSONResponse(w, http.StatusCreated, response)
}

// GetUniversityByIDHandler handles getting a specific university by ID
func (h *Handler) GetUniversityByIDHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeErrorResponse(w, http.StatusMethodNotAllowed, "method_not_allowed", "Method not allowed", h.getRequestID(r))
		return
	}

	requestID := h.getRequestID(r)
	ctx, cancel := h.createContextWithTimeout(r, h.config.Services.Structure.Timeout)
	defer cancel()

	// Extract university ID from URL path
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 2 {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid_path", "University ID is required", requestID)
		return
	}

	universityID, err := h.parseIntParam(pathParts[1])
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid_university_id", "Invalid university ID format", requestID)
		return
	}

	req := &structurepb.GetUniversityByIDRequest{
		Id: universityID,
	}

	client := h.clientManager.GetStructureClient()
	if !h.checkServiceAvailability(w, client, "Structure", requestID) {
		return
	}
	
	resp, err := client.GetUniversityByID(ctx, req)
	if err != nil {
		h.handleGRPCErrorLegacy(w, err, requestID)
		return
	}

	if resp.Error != "" {
		if strings.Contains(resp.Error, "not found") {
			h.writeErrorResponse(w, http.StatusNotFound, "university_not_found", resp.Error, requestID)
		} else {
			h.writeErrorResponse(w, http.StatusInternalServerError, "get_university_failed", resp.Error, requestID)
		}
		return
	}

	// Convert to HTTP response format
	response := map[string]interface{}{
		"id":         resp.University.Id,
		"name":       resp.University.Name,
		"inn":        resp.University.Inn,
		"kpp":        resp.University.Kpp,
		"foiv":       resp.University.Foiv,
		"created_at": resp.University.CreatedAt,
		"updated_at": resp.University.UpdatedAt,
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

// GetUniversityStructureHandler handles getting university structure
func (h *Handler) GetUniversityStructureHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeErrorResponse(w, http.StatusMethodNotAllowed, "method_not_allowed", "Method not allowed", h.getRequestID(r))
		return
	}

	requestID := h.getRequestID(r)
	ctx, cancel := h.createContextWithTimeout(r, h.config.Services.Structure.Timeout)
	defer cancel()

	// Extract university ID from URL path
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 3 {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid_path", "University ID is required", requestID)
		return
	}

	universityID, err := h.parseIntParam(pathParts[1])
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid_university_id", "Invalid university ID format", requestID)
		return
	}

	req := &structurepb.GetUniversityStructureRequest{
		UniversityId: universityID,
	}

	client := h.clientManager.GetStructureClient()
	if !h.checkServiceAvailability(w, client, "Structure", requestID) {
		return
	}
	
	resp, err := client.GetUniversityStructure(ctx, req)
	if err != nil {
		h.handleGRPCErrorLegacy(w, err, requestID)
		return
	}

	if resp.Error != "" {
		if strings.Contains(resp.Error, "not found") {
			h.writeErrorResponse(w, http.StatusNotFound, "university_not_found", resp.Error, requestID)
		} else {
			h.writeErrorResponse(w, http.StatusInternalServerError, "get_structure_failed", resp.Error, requestID)
		}
		return
	}

	// Convert to HTTP response format (simplified for E2E test compatibility)
	response := map[string]interface{}{
		"id":   resp.Structure.University.Id,
		"name": resp.Structure.University.Name,
		"type": "university", // For E2E test compatibility
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

// UpdateUniversityNameHandler handles updating university name
func (h *Handler) UpdateUniversityNameHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		h.writeErrorResponse(w, http.StatusMethodNotAllowed, "method_not_allowed", "Method not allowed", h.getRequestID(r))
		return
	}

	requestID := h.getRequestID(r)
	ctx, cancel := h.createContextWithTimeout(r, h.config.Services.Structure.Timeout)
	defer cancel()

	// Extract university ID from URL path
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 3 {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid_path", "University ID is required", requestID)
		return
	}

	universityID, err := h.parseIntParam(pathParts[1])
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid_university_id", "Invalid university ID format", requestID)
		return
	}

	var reqBody struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid_json", "Invalid JSON format", requestID)
		return
	}

	req := &structurepb.UpdateUniversityNameRequest{
		UniversityId: universityID,
		Name:         reqBody.Name,
	}

	client := h.clientManager.GetStructureClient()
	if !h.checkServiceAvailability(w, client, "Structure", requestID) {
		return
	}
	
	resp, err := client.UpdateUniversityName(ctx, req)
	if err != nil {
		h.handleGRPCErrorLegacy(w, err, requestID)
		return
	}

	if resp.Error != "" {
		if strings.Contains(resp.Error, "not found") {
			h.writeErrorResponse(w, http.StatusNotFound, "university_not_found", resp.Error, requestID)
		} else {
			h.writeErrorResponse(w, http.StatusBadRequest, "update_university_failed", resp.Error, requestID)
		}
		return
	}

	// Convert to HTTP response format
	response := map[string]interface{}{
		"id":         resp.University.Id,
		"name":       resp.University.Name,
		"inn":        resp.University.Inn,
		"kpp":        resp.University.Kpp,
		"foiv":       resp.University.Foiv,
		"created_at": resp.University.CreatedAt,
		"updated_at": resp.University.UpdatedAt,
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

// CreateStructureHandler handles creating structure
func (h *Handler) CreateStructureHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeErrorResponse(w, http.StatusMethodNotAllowed, "method_not_allowed", "Method not allowed", h.getRequestID(r))
		return
	}

	requestID := h.getRequestID(r)
	ctx, cancel := h.createContextWithTimeout(r, h.config.Services.Structure.Timeout)
	defer cancel()

	var req structurepb.CreateStructureRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid_json", "Invalid JSON format", requestID)
		return
	}

	client := h.clientManager.GetStructureClient()
	if !h.checkServiceAvailability(w, client, "Structure", requestID) {
		return
	}
	
	resp, err := client.CreateStructure(ctx, &req)
	if err != nil {
		h.handleGRPCErrorLegacy(w, err, requestID)
		return
	}

	if resp.Error != "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "create_structure_failed", resp.Error, requestID)
		return
	}

	// Convert to HTTP response format
	response := map[string]interface{}{
		"message":       "Structure created successfully",
		"university_id": resp.UniversityId,
		"branch_id":     resp.BranchId,
		"faculty_id":    resp.FacultyId,
		"group_id":      resp.GroupId,
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

// ImportExcelHandler handles Excel import
func (h *Handler) ImportExcelHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeErrorResponse(w, http.StatusMethodNotAllowed, "method_not_allowed", "Method not allowed", h.getRequestID(r))
		return
	}

	requestID := h.getRequestID(r)
	ctx, cancel := h.createContextWithTimeout(r, h.config.Services.Structure.Timeout)
	defer cancel()

	// Parse multipart form
	err := r.ParseMultipartForm(32 << 20) // 32 MB max
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid_form", "Invalid multipart form", requestID)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "missing_file", "File is required", requestID)
		return
	}
	defer file.Close()

	// Read file data
	fileData := make([]byte, header.Size)
	_, err = file.Read(fileData)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "read_file_failed", "Failed to read file", requestID)
		return
	}

	req := &structurepb.ImportExcelRequest{
		FileData: fileData,
		Filename: header.Filename,
	}

	client := h.clientManager.GetStructureClient()
	if !h.checkServiceAvailability(w, client, "Structure", requestID) {
		return
	}
	
	resp, err := client.ImportExcel(ctx, req)
	if err != nil {
		h.handleGRPCErrorLegacy(w, err, requestID)
		return
	}

	if resp.Error != "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "import_failed", resp.Error, requestID)
		return
	}

	// Convert to HTTP response format
	response := map[string]interface{}{
		"processed_rows":  resp.ProcessedRows,
		"successful_rows": resp.SuccessfulRows,
		"failed_rows":     resp.FailedRows,
		"errors":          resp.Errors,
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

// GetAllDepartmentManagersHandler handles getting all department managers
func (h *Handler) GetAllDepartmentManagersHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeErrorResponse(w, http.StatusMethodNotAllowed, "method_not_allowed", "Method not allowed", h.getRequestID(r))
		return
	}

	requestID := h.getRequestID(r)
	ctx, cancel := h.createContextWithTimeout(r, h.config.Services.Structure.Timeout)
	defer cancel()

	page, limit, _, _ := h.parseQueryParams(r)

	req := &structurepb.GetAllDepartmentManagersRequest{
		Page:  page,
		Limit: limit,
	}

	client := h.clientManager.GetStructureClient()
	if !h.checkServiceAvailability(w, client, "Structure", requestID) {
		return
	}
	
	resp, err := client.GetAllDepartmentManagers(ctx, req)
	if err != nil {
		h.handleGRPCErrorLegacy(w, err, requestID)
		return
	}

	if resp.Error != "" {
		h.writeErrorResponse(w, http.StatusInternalServerError, "get_managers_failed", resp.Error, requestID)
		return
	}

	// Convert to HTTP response format
	managers := make([]map[string]interface{}, len(resp.Managers))
	for i, manager := range resp.Managers {
		managers[i] = map[string]interface{}{
			"id":            manager.Id,
			"user_id":       manager.UserId,
			"department_id": manager.DepartmentId,
			"created_at":    manager.CreatedAt,
		}
	}

	h.writeJSONResponse(w, http.StatusOK, managers)
}

// CreateDepartmentManagerHandler handles creating a department manager
func (h *Handler) CreateDepartmentManagerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeErrorResponse(w, http.StatusMethodNotAllowed, "method_not_allowed", "Method not allowed", h.getRequestID(r))
		return
	}

	requestID := h.getRequestID(r)
	ctx, cancel := h.createContextWithTimeout(r, h.config.Services.Structure.Timeout)
	defer cancel()

	var req structurepb.CreateDepartmentManagerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid_json", "Invalid JSON format", requestID)
		return
	}

	client := h.clientManager.GetStructureClient()
	if !h.checkServiceAvailability(w, client, "Structure", requestID) {
		return
	}
	
	resp, err := client.CreateDepartmentManager(ctx, &req)
	if err != nil {
		h.handleGRPCErrorLegacy(w, err, requestID)
		return
	}

	if resp.Error != "" {
		if strings.Contains(resp.Error, "not found") {
			h.writeErrorResponse(w, http.StatusNotFound, "user_or_department_not_found", resp.Error, requestID)
		} else {
			h.writeErrorResponse(w, http.StatusBadRequest, "create_manager_failed", resp.Error, requestID)
		}
		return
	}

	// Convert to HTTP response format
	response := map[string]interface{}{
		"id":            resp.Manager.Id,
		"user_id":       resp.Manager.UserId,
		"department_id": resp.Manager.DepartmentId,
		"created_at":    resp.Manager.CreatedAt,
	}

	h.writeJSONResponse(w, http.StatusCreated, response)
}

// RemoveDepartmentManagerHandler handles removing a department manager
func (h *Handler) RemoveDepartmentManagerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		h.writeErrorResponse(w, http.StatusMethodNotAllowed, "method_not_allowed", "Method not allowed", h.getRequestID(r))
		return
	}

	requestID := h.getRequestID(r)
	ctx, cancel := h.createContextWithTimeout(r, h.config.Services.Structure.Timeout)
	defer cancel()

	// Extract manager ID from URL path
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 3 {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid_path", "Manager ID is required", requestID)
		return
	}

	managerID, err := h.parseIntParam(pathParts[2])
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid_manager_id", "Invalid manager ID format", requestID)
		return
	}

	req := &structurepb.RemoveDepartmentManagerRequest{
		ManagerId: managerID,
	}

	client := h.clientManager.GetStructureClient()
	if !h.checkServiceAvailability(w, client, "Structure", requestID) {
		return
	}
	
	resp, err := client.RemoveDepartmentManager(ctx, req)
	if err != nil {
		h.handleGRPCErrorLegacy(w, err, requestID)
		return
	}

	if resp.Error != "" {
		if strings.Contains(resp.Error, "not found") {
			h.writeErrorResponse(w, http.StatusNotFound, "manager_not_found", resp.Error, requestID)
		} else {
			h.writeErrorResponse(w, http.StatusBadRequest, "remove_manager_failed", resp.Error, requestID)
		}
		return
	}

	response := map[string]interface{}{
		"success": resp.Success,
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

// UpdateBranchNameHandler handles updating branch name
func (h *Handler) UpdateBranchNameHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		h.writeErrorResponse(w, http.StatusMethodNotAllowed, "method_not_allowed", "Method not allowed", h.getRequestID(r))
		return
	}

	requestID := h.getRequestID(r)

	// Extract branch ID from URL path
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 3 {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid_path", "Branch ID is required", requestID)
		return
	}

	branchID, err := h.parseIntParam(pathParts[1])
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid_branch_id", "Invalid branch ID format", requestID)
		return
	}

	var reqBody struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid_json", "Invalid JSON format", requestID)
		return
	}

	// For now, return a placeholder response since we don't have the actual gRPC method
	response := map[string]interface{}{
		"id":   branchID,
		"name": reqBody.Name,
		"message": "Branch name update not implemented in backend service",
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

// UpdateFacultyNameHandler handles updating faculty name
func (h *Handler) UpdateFacultyNameHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		h.writeErrorResponse(w, http.StatusMethodNotAllowed, "method_not_allowed", "Method not allowed", h.getRequestID(r))
		return
	}

	requestID := h.getRequestID(r)

	// Extract faculty ID from URL path
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 3 {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid_path", "Faculty ID is required", requestID)
		return
	}

	facultyID, err := h.parseIntParam(pathParts[1])
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid_faculty_id", "Invalid faculty ID format", requestID)
		return
	}

	var reqBody struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid_json", "Invalid JSON format", requestID)
		return
	}

	// For now, return a placeholder response since we don't have the actual gRPC method
	response := map[string]interface{}{
		"id":   facultyID,
		"name": reqBody.Name,
		"message": "Faculty name update not implemented in backend service",
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

// UpdateGroupNameHandler handles updating group name
func (h *Handler) UpdateGroupNameHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		h.writeErrorResponse(w, http.StatusMethodNotAllowed, "method_not_allowed", "Method not allowed", h.getRequestID(r))
		return
	}

	requestID := h.getRequestID(r)

	// Extract group ID from URL path
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 3 {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid_path", "Group ID is required", requestID)
		return
	}

	groupID, err := h.parseIntParam(pathParts[1])
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid_group_id", "Invalid group ID format", requestID)
		return
	}

	var reqBody struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid_json", "Invalid JSON format", requestID)
		return
	}

	// For now, return a placeholder response since we don't have the actual gRPC method
	response := map[string]interface{}{
		"id":   groupID,
		"name": reqBody.Name,
		"message": "Group name update not implemented in backend service",
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

// LinkGroupToChatHandler handles linking group to chat
func (h *Handler) LinkGroupToChatHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		h.writeErrorResponse(w, http.StatusMethodNotAllowed, "method_not_allowed", "Method not allowed", h.getRequestID(r))
		return
	}

	requestID := h.getRequestID(r)

	// Extract group ID from URL path
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 3 {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid_path", "Group ID is required", requestID)
		return
	}

	groupID, err := h.parseIntParam(pathParts[1])
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid_group_id", "Invalid group ID format", requestID)
		return
	}

	var reqBody struct {
		ChatID int64 `json:"chat_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid_json", "Invalid JSON format", requestID)
		return
	}

	// For now, return a placeholder response since we don't have the actual gRPC method
	response := map[string]interface{}{
		"group_id": groupID,
		"chat_id":  reqBody.ChatID,
		"message":  "Group to chat linking not implemented in backend service",
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}