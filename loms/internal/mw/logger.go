package mw

import (
	"context"
	"encoding/json"
	"google.golang.org/grpc"
	"log/slog"
	"time"
)

func Log(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	start := time.Now()

	// Log request
	reqJSON, err := json.Marshal(req)
	if err != nil {
		slog.ErrorContext(ctx, "failed to marshal request",
			slog.String("method", info.FullMethod),
			slog.Any("error", err),
		)
	} else {
		slog.InfoContext(ctx, "incoming request",
			slog.String("method", info.FullMethod),
			slog.String("request", string(reqJSON)),
		)
	}

	resp, err := handler(ctx, req)

	// Log response
	duration := time.Since(start)
	if err != nil {
		slog.ErrorContext(ctx, "request failed",
			slog.String("method", info.FullMethod),
			slog.String("duration", duration.String()),
			slog.Any("error", err),
		)
	} else {
		respJSON, _ := json.Marshal(resp) // We ignore marshal errors for response
		slog.InfoContext(ctx, "request completed",
			slog.String("method", info.FullMethod),
			slog.String("duration", duration.String()),
			slog.String("response", string(respJSON)),
		)
	}

	return resp, err
}
