package mw

import (
	"context"
	"encoding/json"
	myLogger "github.com/kdaadk/route256/pkg/logger"
	"github.com/kdaadk/route256/pkg/metrics"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
	statusCode := ""
	if err != nil {
		myLogger.ErrorContext(ctx, "Response",
			zap.String("duration", duration.String()),
			zap.Error(err),
			zap.String("method", info.FullMethod))
		st, _ := status.FromError(err)
		statusCode = grpcToHTTPStatus[st.Code()]
	} else {
		respJSON, _ := json.Marshal(resp) // We ignore marshal errors for response
		myLogger.InfoContext(ctx, "Response",
			zap.String("response", string(respJSON)),
			zap.String("duration", duration.String()),
			zap.String("method", info.FullMethod))
		statusCode = "200"
	}

	metrics.RequestDuration.WithLabelValues(info.FullMethod, statusCode).Observe(duration.Seconds())
	metrics.TotalRequests.WithLabelValues(info.FullMethod, statusCode).Inc()

	return resp, err
}

var grpcToHTTPStatus = map[codes.Code]string{
	codes.OK:                 "200",
	codes.Canceled:           "499",
	codes.Unknown:            "500",
	codes.InvalidArgument:    "400",
	codes.DeadlineExceeded:   "504",
	codes.NotFound:           "404",
	codes.AlreadyExists:      "409",
	codes.PermissionDenied:   "403",
	codes.ResourceExhausted:  "429",
	codes.FailedPrecondition: "412",
	codes.Aborted:            "409",
	codes.OutOfRange:         "400",
	codes.Unimplemented:      "501",
	codes.Internal:           "500",
	codes.Unavailable:        "503",
	codes.DataLoss:           "500",
	codes.Unauthenticated:    "401",
}
