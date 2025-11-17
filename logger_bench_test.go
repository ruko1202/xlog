package xlog

import (
	"context"
	"testing"

	"go.uber.org/zap"

	"github.com/ruko1202/xlog/field"
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
		ctx = ContextWithLogger(ctx, zap.NewNop())
		withBenchedLogger(b, func() {
			Info(ctx, "hello world")
		})
	})
}

func BenchmarkAdvance_WithOperation(b *testing.B) {
	ctx := context.Background()

	b.Run("zap", func(b *testing.B) {
		logger := zap.NewNop()
		b.Run("with operation name", func(b *testing.B) {
			withBenchedLogger(b, func() {
				logger.Named("xlog operation").
					Info("hello world")
			})
		})
		b.Run("with operation name and fields", func(b *testing.B) {
			withBenchedLogger(b, func() {
				logger.Named("xlog operation").
					With(zap.String("key", "value")).
					Info("hello world")
			})
		})
	})

	b.Run("xlog", func(b *testing.B) {
		ctx = ContextWithLogger(ctx, zap.NewNop())
		b.Run("with operation name", func(b *testing.B) {
			withBenchedLogger(b, func() {
				ctx = WithOperation(ctx, "xlog operation")
				Info(ctx, "hello world")
			})
		})
		b.Run("with operation name and fields", func(b *testing.B) {
			withBenchedLogger(b, func() {
				ctx = WithOperation(ctx, "xlog operation", field.String("key", "value"))
				Info(ctx, "hello world")
			})
		})
	})
}
func BenchmarkAdvance_WithFields(b *testing.B) {
	ctx := context.Background()

	b.Run("zap", func(b *testing.B) {
		logger := zap.NewNop()
		withBenchedLogger(b, func() {
			logger.
				With(zap.String("key", "value")).
				Info("hello world")
		})
	})

	b.Run("xlog", func(b *testing.B) {
		ctx = ContextWithLogger(ctx, zap.NewNop())
		withBenchedLogger(b, func() {
			ctx = WithFields(ctx, field.String("key", "value"))
			Info(ctx, "hello world")
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
