package mw

import (
	"net/http"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

const (
	serviceName = "cart"
)

func TracingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		ctx := otel.GetTextMapPropagator().Extract(r.Context(), propagation.HeaderCarrier(r.Header))

		ctx, span := otel.Tracer(serviceName).Start(
			ctx,
			"HTTP "+r.Method+" "+r.URL.Path,
			trace.WithSpanKind(trace.SpanKindServer),
			trace.WithAttributes(
				attribute.String("http.method", r.Method),
				attribute.String("http.path", r.URL.Path),
				attribute.String("http.host", r.Host),
				attribute.String("http.user_agent", r.UserAgent()),
			),
		)
		defer span.End()

		rec := &responseRecorder{w, http.StatusOK}

		r = r.WithContext(ctx)
		next.ServeHTTP(rec, r)

		duration := time.Since(start)
		span.SetAttributes(
			attribute.Int("http.status_code", rec.status),
			attribute.Float64("http.duration_ms", float64(duration.Milliseconds())),
		)

		if rec.status >= http.StatusBadRequest {
			span.SetStatus(codes.Error, http.StatusText(rec.status))
		}
	})
}

type responseRecorder struct {
	http.ResponseWriter
	status int
}

func (r *responseRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}
