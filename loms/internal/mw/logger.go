package mw

import (
	"context"
	"encoding/json"
	myLogger "github.com/kdaadk/route256/pkg/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"time"
)

func Log(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	start := time.Now()

	// Log request
	reqJSON, err := json.Marshal(req)
	if err != nil {
		myLogger.ErrorContext(ctx, "Error marshalling request", zap.Any("request", req), zap.Error(err))

	} else {
		myLogger.InfoContext(ctx, "Request", zap.String("method", info.FullMethod), zap.Any("request", reqJSON))
	}

	resp, err := handler(ctx, req)

	// Log response
	duration := time.Since(start)
	if err != nil {
		myLogger.ErrorContext(ctx, "Response",
			zap.String("duration", duration.String()),
			zap.Error(err),
			zap.String("method", info.FullMethod))
	} else {
		respJSON, _ := json.Marshal(resp) // We ignore marshal errors for response
		myLogger.InfoContext(ctx, "Response",
			zap.String("response", string(respJSON)),
			zap.String("duration", duration.String()),
			zap.String("method", info.FullMethod))
	}

	return resp, err
}
