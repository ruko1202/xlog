package xlog

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

func Debug(ctx context.Context, msg string, fields ...zap.Field) {
	logger := fromContext(ctx)
	logger.Debug(msg, fields...)
}

func Debugf(ctx context.Context, template string, args ...any) {
	logger := fromContext(ctx)
	logger.Debug(fmt.Sprintf(template, args...))
}

func Info(ctx context.Context, msg string, fields ...zap.Field) {
	logger := fromContext(ctx)
	logger.Info(msg, fields...)
}

func Infof(ctx context.Context, template string, args ...any) {
	logger := fromContext(ctx)
	logger.Info(fmt.Sprintf(template, args...))
}

func Warn(ctx context.Context, msg string, fields ...zap.Field) {
	logger := fromContext(ctx)
	logger.Warn(msg, fields...)
}

func Warnf(ctx context.Context, template string, args ...any) {
	logger := fromContext(ctx)
	logger.Warn(fmt.Sprintf(template, args...))
}

func Error(ctx context.Context, msg string, fields ...zap.Field) {
	logger := fromContext(ctx)
	logger.Error(msg, fields...)
}

func Errorf(ctx context.Context, template string, args ...any) {
	logger := fromContext(ctx)
	logger.Error(fmt.Sprintf(template, args...))
}

func Fatal(ctx context.Context, msg string, fields ...zap.Field) {
	logger := fromContext(ctx)
	logger.Fatal(msg, fields...)
}

func Fatalf(ctx context.Context, template string, args ...any) {
	logger := fromContext(ctx)
	logger.Fatal(fmt.Sprintf(template, args...))
}

func Panic(ctx context.Context, msg string, fields ...zap.Field) {
	logger := fromContext(ctx)
	logger.Panic(msg, fields...)
}

func Panicf(ctx context.Context, template string, args ...any) {
	logger := fromContext(ctx)
	logger.Panic(fmt.Sprintf(template, args...))
}
