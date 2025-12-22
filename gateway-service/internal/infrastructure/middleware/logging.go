package middleware

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"google.golang.org/grpc/status"
)

// LogLevel represents different log levels
type LogLevel string

const (
	LogLevelDebug LogLevel = "debug"
	LogLevelInfo  LogLevel = "info"
	LogLevelWarn  LogLevel = "warn"
	LogLevelError LogLevel = "error"
)

// LogEntry represents a structured log entry
type LogEntry struct {
	Timestamp   time.Time              `json:"timestamp"`
	Level       LogLevel               `json:"level"`
	Message     string                 `json:"message"`
	RequestID   string                 `json:"request_id,omitempty"`
	TraceID     string                 `json:"trace_id,omitempty"`
	UserID      string                 `json:"user_id,omitempty"`
	Service     string                 `json:"service,omitempty"`
	Method      string                 `json:"method,omitempty"`
	Duration    string                 `json:"duration,omitempty"`
	StatusCode  int                    `json:"status_code,omitempty"`
	Error       string                 `json:"error,omitempty"`
	GRPCCode    string                 `json:"grpc_code,omitempty"`
	HTTPMethod  string                 `json:"http_method,omitempty"`
	HTTPPath    string                 `json:"http_path,omitempty"`
	RemoteAddr  string                 `json:"remote_addr,omitempty"`
	UserAgent   string                 `json:"user_agent,omitempty"`
	Extra       map[string]interface{} `json:"extra,omitempty"`
}

// Logger provides structured logging functionality
type Logger struct {
	level LogLevel
}

// NewLogger creates a new structured logger
func NewLogger(level LogLevel) *Logger {
	return &Logger{level: level}
}

// shouldLog determines if a message should be logged based on level
func (l *Logger) shouldLog(level LogLevel) bool {
	levels := map[LogLevel]int{
		LogLevelDebug: 0,
		LogLevelInfo:  1,
		LogLevelWarn:  2,
		LogLevelError: 3,
	}
	return levels[level] >= levels[l.level]
}

// log writes a structured log entry
func (l *Logger) log(entry LogEntry) {
	if !l.shouldLog(entry.Level) {
		return
	}

	entry.Timestamp = time.Now()
	
	jsonBytes, err := json.Marshal(entry)
	if err != nil {
		log.Printf("Failed to marshal log entry: %v", err)
		return
	}
	
	log.Println(string(jsonBytes))
}

// Debug logs a debug message
func (l *Logger) Debug(ctx context.Context, message string, extra map[string]interface{}) {
	l.log(LogEntry{
		Level:     LogLevelDebug,
		Message:   message,
		RequestID: GetRequestID(ctx),
		TraceID:   GetTraceID(ctx),
		UserID:    GetUserID(ctx),
		Extra:     extra,
	})
}

// Info logs an info message
func (l *Logger) Info(ctx context.Context, message string, extra map[string]interface{}) {
	l.log(LogEntry{
		Level:     LogLevelInfo,
		Message:   message,
		RequestID: GetRequestID(ctx),
		TraceID:   GetTraceID(ctx),
		UserID:    GetUserID(ctx),
		Extra:     extra,
	})
}

// Warn logs a warning message
func (l *Logger) Warn(ctx context.Context, message string, extra map[string]interface{}) {
	l.log(LogEntry{
		Level:     LogLevelWarn,
		Message:   message,
		RequestID: GetRequestID(ctx),
		TraceID:   GetTraceID(ctx),
		UserID:    GetUserID(ctx),
		Extra:     extra,
	})
}

// Error logs an error message
func (l *Logger) Error(ctx context.Context, message string, extra map[string]interface{}) {
	l.log(LogEntry{
		Level:     LogLevelError,
		Message:   message,
		RequestID: GetRequestID(ctx),
		TraceID:   GetTraceID(ctx),
		UserID:    GetUserID(ctx),
		Extra:     extra,
	})
}

// LogHTTPRequest logs an HTTP request
func (l *Logger) LogHTTPRequest(ctx context.Context, r *http.Request) {
	l.Info(ctx, "HTTP request started", map[string]interface{}{
		"http_method": r.Method,
		"http_path":   r.URL.Path,
		"remote_addr": r.RemoteAddr,
		"user_agent":  r.Header.Get("User-Agent"),
		"query":       r.URL.RawQuery,
	})
}

// LogHTTPResponse logs an HTTP response
func (l *Logger) LogHTTPResponse(ctx context.Context, r *http.Request, statusCode int, duration time.Duration) {
	level := LogLevelInfo
	if statusCode >= 400 {
		level = LogLevelWarn
	}
	if statusCode >= 500 {
		level = LogLevelError
	}

	entry := LogEntry{
		Level:      level,
		Message:    "HTTP request completed",
		RequestID:  GetRequestID(ctx),
		TraceID:    GetTraceID(ctx),
		UserID:     GetUserID(ctx),
		StatusCode: statusCode,
		Duration:   duration.String(),
		HTTPMethod: r.Method,
		HTTPPath:   r.URL.Path,
		RemoteAddr: r.RemoteAddr,
		Extra: map[string]interface{}{
			"duration_ms": duration.Milliseconds(),
		},
	}

	l.log(entry)
}

// LogGRPCCall logs a gRPC call
func (l *Logger) LogGRPCCall(ctx context.Context, service, method string, duration time.Duration, err error) {
	level := LogLevelInfo
	message := "gRPC call completed"
	
	entry := LogEntry{
		Level:     level,
		Message:   message,
		RequestID: GetRequestID(ctx),
		TraceID:   GetTraceID(ctx),
		UserID:    GetUserID(ctx),
		Service:   service,
		Method:    method,
		Duration:  duration.String(),
		Extra: map[string]interface{}{
			"duration_ms": duration.Milliseconds(),
		},
	}

	if err != nil {
		level = LogLevelError
		message = "gRPC call failed"
		entry.Level = level
		entry.Message = message
		entry.Error = err.Error()

		// Extract gRPC status code if available
		if st, ok := status.FromError(err); ok {
			entry.GRPCCode = st.Code().String()
		}
	}

	l.log(entry)
}

// LogGRPCError logs a detailed gRPC error
func (l *Logger) LogGRPCError(ctx context.Context, service, method string, err error) {
	entry := LogEntry{
		Level:     LogLevelError,
		Message:   "gRPC error occurred",
		RequestID: GetRequestID(ctx),
		TraceID:   GetTraceID(ctx),
		UserID:    GetUserID(ctx),
		Service:   service,
		Method:    method,
		Error:     err.Error(),
	}

	// Extract detailed gRPC status information
	if st, ok := status.FromError(err); ok {
		entry.GRPCCode = st.Code().String()
		entry.Extra = map[string]interface{}{
			"grpc_message": st.Message(),
			"grpc_details": st.Details(),
		}
	}

	l.log(entry)
}

// LogConnectionError logs connection-specific errors
func (l *Logger) LogConnectionError(ctx context.Context, service string, err error) {
	l.log(LogEntry{
		Level:     LogLevelError,
		Message:   "gRPC connection error",
		RequestID: GetRequestID(ctx),
		TraceID:   GetTraceID(ctx),
		UserID:    GetUserID(ctx),
		Service:   service,
		Error:     err.Error(),
		Extra: map[string]interface{}{
			"error_type": "connection_error",
		},
	})
}

// RequestLoggingMiddleware logs HTTP requests and responses
func RequestLoggingMiddleware(logger *Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Log request start
			logger.LogHTTPRequest(r.Context(), r)

			// Wrap response writer to capture status code
			wrapped := &responseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			start := time.Now()
			
			// Call next handler
			next.ServeHTTP(wrapped, r)

			// Log request completion
			duration := time.Since(start)
			logger.LogHTTPResponse(r.Context(), r, wrapped.statusCode, duration)
		})
	}
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	return rw.ResponseWriter.Write(b)
}