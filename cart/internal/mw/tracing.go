package mw

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"net/http"
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
		ctx := otel.GetTextMapPropagator().Extract(r.Context(), propagation.HeaderCarrier(r.Header))

		tracer := otel.Tracer("cart-service")
		ctx, span := tracer.Start(ctx, r.URL.Path)
		defer span.End()

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
