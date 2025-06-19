package mw

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"net/http"
)

func TracingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := otel.GetTextMapPropagator().Extract(r.Context(), propagation.HeaderCarrier(r.Header))

		tracer := otel.Tracer("product-service")
		ctx, span := tracer.Start(ctx, r.URL.Path)
		defer span.End()

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
