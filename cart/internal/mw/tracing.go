package mw

import (
	myLogger "github.com/kdaadk/route256/pkg/logger"
	"github.com/kdaadk/route256/pkg/metrics"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"log/slog"
	"net/http"
	"strconv"
	"time"
)

type TracingTransport struct {
	transport http.RoundTripper
	tracer    trace.Tracer
}

func (t *TracingTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	ctx, span := t.tracer.Start(
		r.Context(),
		"HTTP "+r.Method,
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.String("http.method", r.Method),
			attribute.String("http.url", r.URL.String()),
		),
	)
	defer span.End()

	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(r.Header))

	resp, err := t.transport.RoundTrip(r.WithContext(ctx))
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	span.SetAttributes(
		attribute.Int("http.status_code", resp.StatusCode),
		attribute.String("http.response_content_type", resp.Header.Get("Content-Type")),
	)

	return resp, nil
}

func NewTracingTransport(transport http.RoundTripper, tracer trace.Tracer) *TracingTransport {
	return &TracingTransport{
		transport: transport,
		tracer:    tracer,
	}
}

func TracingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Start timer and tracing
		start := time.Now()
		ctx, span := otel.Tracer("").Start(r.Context(), r.URL.Path)
		defer span.End()

		// Get trace context
		traceID := trace.SpanContextFromContext(ctx).TraceID().String()
		spanID := trace.SpanContextFromContext(ctx).SpanID().String()

		// Create response recorder
		recorder := &statusRecorder{
			ResponseWriter: w,
			status:         http.StatusOK,
		}

		// Deferred logging and metrics
		defer func() {
			duration := time.Since(start).Seconds()
			statusCode := strconv.Itoa(recorder.status)

			myLogger.InfoContext(ctx, "write metrics", zap.Float64("duration", duration), zap.String("status_code", statusCode))
			// Record metrics
			metrics.RequestDuration.WithLabelValues(r.URL.Path, statusCode).Observe(duration)
			metrics.TotalRequests.WithLabelValues(r.URL.Path, statusCode).Inc()

			// Prepare log fields
			logFields := []any{
				"service", "cart",
				"time", time.Now().Format(time.RFC3339),
				"method", r.Method,
				"path", r.URL.Path,
				"status", recorder.status,
				"duration", duration,
				"trace_id", traceID,
				"span_id", spanID,
			}

			// Add error message if present
			if recorder.errMessage != "" {
				logFields = append(logFields, "error", recorder.errMessage)
			}

			// Log based on status code
			if recorder.status >= 400 {
				slog.Error("HTTP request failed", logFields...)
			} else {
				slog.Info("HTTP request succeeded", logFields...)
			}
		}()

		// Call the next handler with the new context
		next.ServeHTTP(recorder, r.WithContext(ctx))
	})
}

// statusRecorder remains the same as your original implementation
type statusRecorder struct {
	http.ResponseWriter
	status     int
	errMessage string
}

func (r *statusRecorder) WriteHeader(statusCode int) {
	r.status = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}
