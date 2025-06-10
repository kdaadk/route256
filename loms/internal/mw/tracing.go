package mw

import (
	"context"
	myLogger "github.com/kdaadk/route256/pkg/logger"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"strings"
	"time"
)

func Tracing(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	start := time.Now()

	tracer := otel.GetTracerProvider().Tracer("loms")
	ctx, span := tracer.Start(ctx, info.FullMethod)
	defer span.End()

	traceID := trace.SpanContextFromContext(ctx).TraceID().String()
	spanID := trace.SpanContextFromContext(ctx).SpanID().String()

	methodName := cleanGRPCMethodName(info.FullMethod)

	resp, err := handler(ctx, req)

	defer func() {
		duration := time.Since(start).Seconds()
		statusCode := status.Code(err).String()
		statusLabel := grpcStatusToLabel(status.Code(err))
		_ = statusLabel

		// Metrics
		//metrics.GRPCRequestDuration.WithLabelValues(methodName, statusLabel).Observe(duration)
		//metrics.GRPCTotalRequests.WithLabelValues(methodName, statusLabel).Inc()

		logTime := time.Now().Format(time.RFC3339)
		logFields := []zap.Field{
			zap.String("service", "cart"),
			zap.String("time", logTime),
			zap.String("method", methodName),
			zap.String("status", statusCode),
			zap.Float64("duration", duration),
			zap.String("trace_id", traceID),
			zap.String("span_id", spanID),
		}

		if err != nil {
			logFields = append(logFields, zap.String("error", err.Error()))
			myLogger.ErrorContext(ctx, "gRPC request failed", logFields...)
		} else {
			myLogger.ErrorContext(ctx, "gRPC request succeeded", logFields...)
		}
	}()

	return resp, err
}

func cleanGRPCMethodName(fullMethod string) string {
	parts := strings.Split(fullMethod, "/")
	if len(parts) >= 3 {
		return parts[2]
	}
	return fullMethod
}

func grpcStatusToLabel(code codes.Code) string {
	switch code {
	case codes.OK:
		return "ok"
	case codes.InvalidArgument:
		return "invalid_argument"
	case codes.DeadlineExceeded:
		return "deadline_exceeded"
	case codes.NotFound:
		return "not_found"
	case codes.AlreadyExists:
		return "already_exists"
	case codes.PermissionDenied:
		return "permission_denied"
	case codes.Internal:
		return "internal"
	case codes.Unavailable:
		return "unavailable"
	default:
		return "unknown"
	}
}
