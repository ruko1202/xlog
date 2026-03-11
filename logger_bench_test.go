package xlog

import (
	"context"
	"testing"

	"go.uber.org/zap"

	"github.com/ruko1202/xlog/xfield"
)

func BenchmarkLogger(b *testing.B) {
	ctx := context.Background()

	b.Run("zap", func(b *testing.B) {
		logger := zap.NewNop()
		withBenchedLogger(b, func() {
			logger.Info("hello world")
		})
	})

	b.Run("xlog", func(b *testing.B) {
		ctx = ContextWithLogger(ctx, NewZapAdapter(zap.NewNop()))
		withBenchedLogger(b, func() {
			Info(ctx, "hello world")
		})
	})
}

func BenchmarkAdvance_WithOperation(b *testing.B) {
	ctx := context.Background()

	b.Run("zap", func(b *testing.B) {
		logger := zap.NewNop()
		b.Run("create logger in loop/with operation name", func(b *testing.B) {
			withBenchedLogger(b, func() {
				logger.Named("xlog operation").
					Info("hello world")
			})
		})
		b.Run("create logger in loop/with operation name and fields", func(b *testing.B) {
			withBenchedLogger(b, func() {
				logger.Named("xlog operation").
					With(zap.String("key", "value")).
					Info("hello world")
			})
		})
		b.Run("reuse logger/with operation name", func(b *testing.B) {
			logger := logger.Named("xlog operation")
			withBenchedLogger(b, func() {
				logger.Info("hello world")
			})
		})
		b.Run("reuse logger/with operation name and fields", func(b *testing.B) {
			logger := logger.Named("xlog operation").With(zap.String("key", "value"))
			withBenchedLogger(b, func() {
				logger.Info("hello world")
			})
		})
	})

	b.Run("xlog", func(b *testing.B) {
		ctx = ContextWithLogger(ctx, NewZapAdapter(zap.NewNop()))
		b.Run("create context in loop/with operation name", func(b *testing.B) {
			withBenchedLogger(b, func() {
				ctx := WithOperation(ctx, "xlog operation")
				Info(ctx, "hello world")
			})
		})
		b.Run("create context in loop/with operation name and fields", func(b *testing.B) {
			withBenchedLogger(b, func() {
				ctx := WithOperation(ctx, "xlog operation", xfield.String("key", "value"))
				Info(ctx, "hello world")
			})
		})
		b.Run("reuse context/with operation name", func(b *testing.B) {
			ctx := WithOperation(ctx, "xlog operation")
			withBenchedLogger(b, func() {
				Info(ctx, "hello world")
			})
		})
		b.Run("reuse context/with operation name and fields", func(b *testing.B) {
			ctx := WithOperation(ctx, "xlog operation", xfield.String("key", "value"))
			withBenchedLogger(b, func() {
				Info(ctx, "hello world")
			})
		})
	})
}
func BenchmarkAdvance_WithFields(b *testing.B) {
	ctx := context.Background()

	b.Run("zap", func(b *testing.B) {
		logger := zap.NewNop()
		b.Run("create logger in loop", func(b *testing.B) {
			withBenchedLogger(b, func() {
				logger.
					With(zap.String("key", "value")).
					Info("hello world")
			})
		})
		b.Run("reuse logger", func(b *testing.B) {
			logger := logger.With(zap.String("key", "value"))
			withBenchedLogger(b, func() {
				logger.Info("hello world")
			})
		})
	})

	b.Run("xlog", func(b *testing.B) {
		ctx = ContextWithLogger(ctx, NewZapAdapter(zap.NewNop()))
		b.Run("create context in loop", func(b *testing.B) {
			withBenchedLogger(b, func() {
				ctx := WithFields(ctx, xfield.String("key", "value"))
				Info(ctx, "hello world")
			})
		})
		b.Run("reuse context", func(b *testing.B) {
			ctx := WithFields(ctx, xfield.String("key", "value"))
			withBenchedLogger(b, func() {
				Info(ctx, "hello world")
			})
		})
	})
}

func withBenchedLogger(b *testing.B, runBench func()) {
	b.Helper()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			runBench()
		}
	})
}
