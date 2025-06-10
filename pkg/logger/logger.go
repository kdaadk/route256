package logger

import (
	"context"
	"go.uber.org/zap"
	"sync"
)

var (
	globalLogger *MyLogger
	once         = sync.Once{}
)

type MyLogger struct {
	l *zap.SugaredLogger
}

func NewMyLogger(config zap.Config) *MyLogger {
	l, err := config.Build()
	if err != nil {
		panic(err)
	}
	logger := l.Sugar()

	once.Do(func() {
		globalLogger = &MyLogger{logger}
	})

	return globalLogger
}

var LoggerContextKey string

func InfoContext(ctx context.Context, message string, fields ...zap.Field) {
	if l, ok := ctx.Value(LoggerContextKey).(*MyLogger); ok && l != nil {
		l.l.Info(message, fields)
		return
	}

	if globalLogger == nil {
		panic("global logger is nil")
	}
	globalLogger.l.Info(message, fields)
}

func ErrorContext(ctx context.Context, message string, fields ...zap.Field) {
	if l, ok := ctx.Value(LoggerContextKey).(*MyLogger); ok && l != nil {
		l.l.Error(message, fields)
		return
	}

	if globalLogger == nil {
		panic("global logger is nil")
	}
	globalLogger.l.Error(message, fields)
}

func (l *MyLogger) Sync() error {
	return l.l.Sync()
}
