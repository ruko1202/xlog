package xlog

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContextWithLogger_Validation(t *testing.T) {
	t.Run("uses global logger when logger is nil", func(t *testing.T) {
		ctx := context.Background()

		// Should not panic, but use global logger
		assert.NotPanics(t, func() {
			newCtx := ContextWithLogger(ctx, nil)

			// Verify logger was added from global
			logger := newCtx.Value(loggerCtxKey).(Logger)
			assert.NotNil(t, logger)
			assert.Equal(t, GlobalLogger(), logger)
		})
	})

	t.Run("panics when context is nil", func(t *testing.T) {
		logger, _ := initTestLogger(t)

		assert.Panics(t, func() {
			ContextWithLogger(nil, logger) //nolint:staticcheck // Intentionally testing nil context behavior
		}, "should panic when context is nil")
	})
}
