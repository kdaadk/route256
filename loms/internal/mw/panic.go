package mw

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log/slog"
)

func RestoreFromPanic(ctx context.Context, request any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
	defer func() {
		p := recover()
		if p != nil {
			slog.Warn(p.(string))
			err = status.Errorf(codes.Internal, "panic: %v", p)
		}
	}()

	return handler(ctx, request)
}
