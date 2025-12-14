package errors

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// ErrorCode represents a unique error code
type ErrorCode string

const (
	// Validation errors (400)
	ErrCodeValidation       ErrorCode = "VALIDATION_ERROR"
	ErrCodeInvalidPhone     ErrorCode = "INVALID_PHONE"
	ErrCodeMissingField     ErrorCode = "MISSING_FIELD"
	ErrCodeInvalidFormat    ErrorCode = "INVALID_FORMAT"
	ErrCodeInvalidRange     ErrorCode = "INVALID_RANGE"

	// Authentication errors (401)
	ErrCodeUnauthorized     ErrorCode = "UNAUTHORIZED"
	ErrCodeInvalidToken     ErrorCode = "INVALID_TOKEN"
	ErrCodeExpiredToken     ErrorCode = "EXPIRED_TOKEN"
	ErrCodeMissingToken     ErrorCode = "MISSING_TOKEN"
	ErrCodeInvalidCreds     ErrorCode = "INVALID_CREDENTIALS"

	// Authorization errors (403)
	ErrCodeForbidden        ErrorCode = "FORBIDDEN"
	ErrCodeInsufficientPerms ErrorCode = "INSUFFICIENT_PERMISSIONS"
	ErrCodeInvalidRole      ErrorCode = "INVALID_ROLE"

	// Not found errors (404)
	ErrCodeNotFound         ErrorCode = "NOT_FOUND"
	ErrCodeUserNotFound     ErrorCode = "USER_NOT_FOUND"
	ErrCodeResourceNotFound ErrorCode = "RESOURCE_NOT_FOUND"

	// Conflict errors (409)
	ErrCodeConflict         ErrorCode = "CONFLICT"
	ErrCodeAlreadyExists    ErrorCode = "ALREADY_EXISTS"
	ErrCodeCannotDelete     ErrorCode = "CANNOT_DELETE"

	// External service errors (502)
	ErrCodeExternalService  ErrorCode = "EXTERNAL_SERVICE_ERROR"
	ErrCodeServiceUnavailable ErrorCode = "SERVICE_UNAVAILABLE"
	ErrCodeGRPCError        ErrorCode = "GRPC_ERROR"

	// Internal errors (500)
	ErrCodeInternal         ErrorCode = "INTERNAL_ERROR"
	ErrCodeDatabase         ErrorCode = "DATABASE_ERROR"
	ErrCodeTransaction      ErrorCode = "TRANSACTION_ERROR"
)

// AppError represents a structured application error
type AppError struct {
	Code       ErrorCode              `json:"code"`
	Message    string                 `json:"message"`
	Details    map[string]interface{} `json:"details,omitempty"`
	StatusCode int                    `json:"-"`
	Err        error                  `json:"-"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s (%v)", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the underlying error
func (e *AppError) Unwrap() error {
	return e.Err
}

// ErrorResponse represents the JSON error response structure
type ErrorResponse struct {
	Error struct {
		Code    ErrorCode              `json:"code"`
		Message string                 `json:"message"`
		Details map[string]interface{} `json:"details,omitempty"`
	} `json:"error"`
}

// NewAppError creates a new AppError
func NewAppError(code ErrorCode, message string, statusCode int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
		Details:    make(map[string]interface{}),
	}
}

// WithDetails adds details to the error
func (e *AppError) WithDetails(key string, value interface{}) *AppError {
	if e.Details == nil {
		e.Details = make(map[string]interface{})
	}
	e.Details[key] = value
	return e
}

// WithError wraps an underlying error
func (e *AppError) WithError(err error) *AppError {
	e.Err = err
	return e
}

// WriteError writes a structured error response to the HTTP response writer
func WriteError(w http.ResponseWriter, err error, requestID string) {
	var appErr *AppError
	var ok bool

	// Check if it's already an AppError
	if appErr, ok = err.(*AppError); !ok {
		// Convert generic error to AppError
		appErr = NewAppError(ErrCodeInternal, err.Error(), http.StatusInternalServerError)
	}

	// Log the error with context
	LogError(appErr, requestID)

	// Build response
	response := ErrorResponse{}
	response.Error.Code = appErr.Code
	response.Error.Message = appErr.Message
	response.Error.Details = appErr.Details

	// Write response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(appErr.StatusCode)
	json.NewEncoder(w).Encode(response)
}

// LogError logs the error with context
func LogError(err *AppError, requestID string) {
	logMsg := fmt.Sprintf("[ERROR] [%s] Code: %s, Message: %s", requestID, err.Code, err.Message)
	if err.Err != nil {
		logMsg += fmt.Sprintf(", Underlying: %v", err.Err)
	}
	if len(err.Details) > 0 {
		detailsJSON, _ := json.Marshal(err.Details)
		logMsg += fmt.Sprintf(", Details: %s", string(detailsJSON))
	}
	log.Println(logMsg)
}

// Common error constructors

func ValidationError(message string) *AppError {
	return NewAppError(ErrCodeValidation, message, http.StatusBadRequest)
}

func InvalidPhoneError(phone string) *AppError {
	return NewAppError(ErrCodeInvalidPhone, "Invalid phone format", http.StatusBadRequest).
		WithDetails("phone", phone).
		WithDetails("expected", "E.164 format (+7XXXXXXXXXX)")
}

func MissingFieldError(field string) *AppError {
	return NewAppError(ErrCodeMissingField, fmt.Sprintf("Missing required field: %s", field), http.StatusBadRequest).
		WithDetails("field", field)
}

func UnauthorizedError(message string) *AppError {
	return NewAppError(ErrCodeUnauthorized, message, http.StatusUnauthorized)
}

func InvalidTokenError() *AppError {
	return NewAppError(ErrCodeInvalidToken, "Invalid or malformed token", http.StatusUnauthorized)
}

func ExpiredTokenError() *AppError {
	return NewAppError(ErrCodeExpiredToken, "Token has expired", http.StatusUnauthorized)
}

func ForbiddenError(message string) *AppError {
	return NewAppError(ErrCodeForbidden, message, http.StatusForbidden)
}

func InsufficientPermissionsError(resource string) *AppError {
	return NewAppError(ErrCodeInsufficientPerms, "Insufficient permissions for this resource", http.StatusForbidden).
		WithDetails("resource", resource)
}

func NotFoundError(resource string) *AppError {
	return NewAppError(ErrCodeNotFound, fmt.Sprintf("%s not found", resource), http.StatusNotFound).
		WithDetails("resource", resource)
}

func AlreadyExistsError(resource string, identifier string) *AppError {
	return NewAppError(ErrCodeAlreadyExists, fmt.Sprintf("%s already exists", resource), http.StatusConflict).
		WithDetails("resource", resource).
		WithDetails("identifier", identifier)
}

func CannotDeleteError(resource string, reason string) *AppError {
	return NewAppError(ErrCodeCannotDelete, fmt.Sprintf("Cannot delete %s: %s", resource, reason), http.StatusConflict).
		WithDetails("resource", resource).
		WithDetails("reason", reason)
}

func ExternalServiceError(service string, err error) *AppError {
	return NewAppError(ErrCodeExternalService, fmt.Sprintf("%s service error", service), http.StatusBadGateway).
		WithDetails("service", service).
		WithError(err)
}

func GRPCError(service string, method string, err error) *AppError {
	return NewAppError(ErrCodeGRPCError, fmt.Sprintf("gRPC call failed: %s.%s", service, method), http.StatusBadGateway).
		WithDetails("service", service).
		WithDetails("method", method).
		WithError(err)
}

func DatabaseError(operation string, err error) *AppError {
	return NewAppError(ErrCodeDatabase, "Database operation failed", http.StatusInternalServerError).
		WithDetails("operation", operation).
		WithError(err)
}

func InternalError(message string, err error) *AppError {
	return NewAppError(ErrCodeInternal, message, http.StatusInternalServerError).
		WithError(err)
}
