package tracing

import (
	"context"
	"fmt"
	"time"

	"gateway-service/internal/infrastructure/middleware"
)

// Span represents a trace span
type Span struct {
	TraceID     string                 `json:"trace_id"`
	SpanID      string                 `json:"span_id"`
	ParentID    string                 `json:"parent_id,omitempty"`
	Operation   string                 `json:"operation"`
	Service     string                 `json:"service"`
	Method      string                 `json:"method,omitempty"`
	StartTime   time.Time              `json:"start_time"`
	EndTime     time.Time              `json:"end_time,omitempty"`
	Duration    time.Duration          `json:"duration,omitempty"`
	Tags        map[string]interface{} `json:"tags,omitempty"`
	Logs        []LogEntry             `json:"logs,omitempty"`
	Error       bool                   `json:"error"`
	ErrorMsg    string                 `json:"error_msg,omitempty"`
}

// LogEntry represents a log entry within a span
type LogEntry struct {
	Timestamp time.Time              `json:"timestamp"`
	Fields    map[string]interface{} `json:"fields"`
}

// Tracer manages distributed tracing
type Tracer struct {
	serviceName string
	spans       map[string]*Span
}

// NewTracer creates a new tracer instance
func NewTracer(serviceName string) *Tracer {
	return &Tracer{
		serviceName: serviceName,
		spans:       make(map[string]*Span),
	}
}

// StartSpan creates a new trace span
func (t *Tracer) StartSpan(ctx context.Context, operation string) (*Span, context.Context) {
	traceID := middleware.GetTraceID(ctx)
	spanID := middleware.GenerateRequestID() // Reuse the ID generation function
	
	span := &Span{
		TraceID:   traceID,
		SpanID:    spanID,
		Operation: operation,
		Service:   t.serviceName,
		StartTime: time.Now(),
		Tags:      make(map[string]interface{}),
		Logs:      make([]LogEntry, 0),
	}

	// Store span for later retrieval
	t.spans[spanID] = span

	// Add span to context
	spanCtx := context.WithValue(ctx, "current_span", span)
	
	return span, spanCtx
}

// StartChildSpan creates a child span from the current context
func (t *Tracer) StartChildSpan(ctx context.Context, operation string) (*Span, context.Context) {
	parentSpan := t.GetSpanFromContext(ctx)
	
	span, spanCtx := t.StartSpan(ctx, operation)
	
	if parentSpan != nil {
		span.ParentID = parentSpan.SpanID
	}
	
	return span, spanCtx
}

// GetSpanFromContext retrieves the current span from context
func (t *Tracer) GetSpanFromContext(ctx context.Context) *Span {
	if span, ok := ctx.Value("current_span").(*Span); ok {
		return span
	}
	return nil
}

// FinishSpan completes a trace span
func (t *Tracer) FinishSpan(span *Span, err error) {
	if span == nil {
		return
	}

	span.EndTime = time.Now()
	span.Duration = span.EndTime.Sub(span.StartTime)
	
	if err != nil {
		span.Error = true
		span.ErrorMsg = err.Error()
		span.SetTag("error", true)
		span.SetTag("error.message", err.Error())
	}

	// In a real implementation, you would send this span to a tracing backend
	// like Jaeger, Zipkin, or OpenTelemetry collector
	t.logSpan(span)
}

// SetTag adds a tag to a span
func (s *Span) SetTag(key string, value interface{}) {
	if s.Tags == nil {
		s.Tags = make(map[string]interface{})
	}
	s.Tags[key] = value
}

// LogFields adds a log entry to a span
func (s *Span) LogFields(fields map[string]interface{}) {
	entry := LogEntry{
		Timestamp: time.Now(),
		Fields:    fields,
	}
	s.Logs = append(s.Logs, entry)
}

// logSpan outputs the span information (in a real implementation, this would send to a tracing backend)
func (t *Tracer) logSpan(span *Span) {
	// For now, we'll just log the span as JSON for debugging
	// In production, this would be sent to Jaeger, Zipkin, etc.
	fmt.Printf("TRACE: %+v\n", span)
}

// StartGRPCSpan creates a span for gRPC calls
func (t *Tracer) StartGRPCSpan(ctx context.Context, service, method string) (*Span, context.Context) {
	operation := fmt.Sprintf("grpc.%s.%s", service, method)
	span, spanCtx := t.StartChildSpan(ctx, operation)
	
	span.Service = service
	span.Method = method
	span.SetTag("component", "grpc-client")
	span.SetTag("grpc.service", service)
	span.SetTag("grpc.method", method)
	
	return span, spanCtx
}

// StartHTTPSpan creates a span for HTTP requests
func (t *Tracer) StartHTTPSpan(ctx context.Context, method, path string) (*Span, context.Context) {
	operation := fmt.Sprintf("http.%s %s", method, path)
	span, spanCtx := t.StartSpan(ctx, operation)
	
	span.SetTag("component", "http-server")
	span.SetTag("http.method", method)
	span.SetTag("http.path", path)
	span.SetTag("span.kind", "server")
	
	return span, spanCtx
}

// InjectIntoGRPCMetadata injects tracing information into gRPC metadata
func (t *Tracer) InjectIntoGRPCMetadata(ctx context.Context) context.Context {
	span := t.GetSpanFromContext(ctx)
	if span == nil {
		return ctx
	}

	// The context propagation is already handled by middleware.PropagateContextToGRPC
	// but we can add additional tracing-specific metadata here
	return middleware.PropagateContextToGRPC(ctx)
}