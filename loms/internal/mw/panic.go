package mw

import (
	"context"
	myLogger "github.com/kdaadk/route256/pkg/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func RestoreFromPanic(ctx context.Context, request any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
	defer func() {
		p := recover()
		if p != nil {
			myLogger.ErrorContext(ctx, "Recovered from panic", zap.Any("panic", p), zap.Stack("stack"))
			err = status.Errorf(codes.Internal, "panic: %v", p)
		}
	}()

	return handler(ctx, request)
}
