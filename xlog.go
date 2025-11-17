// Package xlog предоставляет обертку над uber-go/zap для логирования с поддержкой context.Context.
//
// Пакет позволяет сохранять логгер в контексте и автоматически извлекать его при вызове
// функций логирования. Если логгер не найден в контексте, используется глобальный логгер.
package xlog

import (
	"context"

	"go.uber.org/zap"
)

var globalLogger = zap.NewNop()

type ctxKey string

const (
	loggerCtxKey ctxKey = "xLoggerKey"
)

// ContextWithLogger добавляет логгер в контекст и возвращает новый контекст.
//
// Пример:
//
//	logger, _ := zap.NewProduction()
//	ctx := xlog.ContextWithLogger(context.Background(), logger)
func ContextWithLogger(ctx context.Context, logger *zap.Logger) context.Context {
	return context.WithValue(ctx, loggerCtxKey, logger)
}

// FromContext извлекает логгер из контекста.
// Если логгер не найден, возвращает глобальный логгер.
//
// Пример:
//
//	logger := xlog.FromContext(ctx)
//	logger.Info("прямое использование zap логгера")
func FromContext(ctx context.Context) *zap.Logger {
	return fromContext(ctx)
}

// SetGlobalLogger устанавливает глобальный логгер, который используется
// когда логгер не найден в контексте.
//
// По умолчанию используется zap.NewNop() который не производит вывод.
//
// Пример:
//
//	logger, _ := zap.NewDevelopment()
//	xlog.SetGlobalLogger(logger)
func SetGlobalLogger(logger *zap.Logger) {
	globalLogger = logger
}

func fromContext(ctx context.Context) *zap.Logger {
	logger, ok := ctx.Value(loggerCtxKey).(*zap.Logger)
	if !ok {
		return globalLogger
	}

	return logger
}
