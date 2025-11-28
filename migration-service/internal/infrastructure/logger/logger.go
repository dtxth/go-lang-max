package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"
)

// LogLevel represents the severity of a log entry
type LogLevel string

const (
	DEBUG LogLevel = "DEBUG"
	INFO  LogLevel = "INFO"
	WARN  LogLevel = "WARN"
	ERROR LogLevel = "ERROR"
)

// Logger provides structured JSON logging
type Logger struct {
	output io.Writer
	level  LogLevel
}

// LogEntry represents a structured log entry
type LogEntry struct {
	Timestamp string                 `json:"timestamp"`
	Level     string                 `json:"level"`
	Message   string                 `json:"message"`
	RequestID string                 `json:"request_id,omitempty"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
}

// New creates a new Logger instance
func New(output io.Writer, level LogLevel) *Logger {
	if output == nil {
		output = os.Stdout
	}
	return &Logger{
		output: output,
		level:  level,
	}
}

// NewDefault creates a logger with default settings (stdout, INFO level)
func NewDefault() *Logger {
	return New(os.Stdout, INFO)
}

// log writes a structured log entry
func (l *Logger) log(level LogLevel, ctx context.Context, message string, fields map[string]interface{}) {
	// Skip if log level is below configured level
	if !l.shouldLog(level) {
		return
	}

	entry := LogEntry{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Level:     string(level),
		Message:   message,
		Fields:    fields,
	}

	// Extract request ID from context if available
	if ctx != nil {
		if reqID := getRequestIDFromContext(ctx); reqID != "" {
			entry.RequestID = reqID
		}
	}

	// Marshal to JSON
	data, err := json.Marshal(entry)
	if err != nil {
		// Fallback to simple logging if JSON marshaling fails
		fmt.Fprintf(l.output, "ERROR: Failed to marshal log entry: %v\n", err)
		return
	}

	// Write to output
	fmt.Fprintln(l.output, string(data))
}

// shouldLog checks if the log level should be logged
func (l *Logger) shouldLog(level LogLevel) bool {
	levels := map[LogLevel]int{
		DEBUG: 0,
		INFO:  1,
		WARN:  2,
		ERROR: 3,
	}
	return levels[level] >= levels[l.level]
}

// getRequestIDFromContext extracts request ID from context
func getRequestIDFromContext(ctx context.Context) string {
	type contextKey string
	const RequestIDKey contextKey = "request_id"
	
	if reqID, ok := ctx.Value(RequestIDKey).(string); ok {
		return reqID
	}
	return ""
}

// Debug logs a debug message
func (l *Logger) Debug(ctx context.Context, message string, fields map[string]interface{}) {
	l.log(DEBUG, ctx, message, fields)
}

// Info logs an info message
func (l *Logger) Info(ctx context.Context, message string, fields map[string]interface{}) {
	l.log(INFO, ctx, message, fields)
}

// Warn logs a warning message
func (l *Logger) Warn(ctx context.Context, message string, fields map[string]interface{}) {
	l.log(WARN, ctx, message, fields)
}

// Error logs an error message
func (l *Logger) Error(ctx context.Context, message string, fields map[string]interface{}) {
	l.log(ERROR, ctx, message, fields)
}

// WithFields is a convenience method to log with fields
func (l *Logger) WithFields(fields map[string]interface{}) *LoggerWithFields {
	return &LoggerWithFields{
		logger: l,
		fields: fields,
	}
}

// LoggerWithFields wraps a logger with predefined fields
type LoggerWithFields struct {
	logger *Logger
	fields map[string]interface{}
}

// Debug logs a debug message with predefined fields
func (l *LoggerWithFields) Debug(ctx context.Context, message string) {
	l.logger.Debug(ctx, message, l.fields)
}

// Info logs an info message with predefined fields
func (l *LoggerWithFields) Info(ctx context.Context, message string) {
	l.logger.Info(ctx, message, l.fields)
}

// Warn logs a warning message with predefined fields
func (l *LoggerWithFields) Warn(ctx context.Context, message string) {
	l.logger.Warn(ctx, message, l.fields)
}

// Error logs an error message with predefined fields
func (l *LoggerWithFields) Error(ctx context.Context, message string) {
	l.logger.Error(ctx, message, l.fields)
}
