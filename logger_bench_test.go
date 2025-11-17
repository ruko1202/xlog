package xlog

import (
	"context"
	"testing"

	"go.uber.org/zap"
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

func withBenchedLogger(b *testing.B, runBench func()) {
	b.Helper()
	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			runBench()
		}
	})
}
