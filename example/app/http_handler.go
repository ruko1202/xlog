package main

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"

	"github.com/ruko1202/xlog"
	"github.com/ruko1202/xlog/xfield"
)

// handleWork - main handler
func handleWork(c echo.Context) error {
	ctx, span := xlog.WithOperationSpan(c.Request().Context(), "handleWork")
	defer span.End()

	xlog.Info(ctx, "Starting request processing...")

	status := "success"
	defer func() {
		xlog.Info(ctx, "Finishing request processing...", xfield.String("status", status))
		// Increment metric at the very end. TraceID will be picked up automatically!
		counter.Add(ctx, 1, metric.WithAttributes(attribute.String("status", status)))
	}()

	err := serviceStep(ctx,
		c.QueryParam("fail") == "true",
		c.QueryParam("user_id"),
	)
	if err != nil {
		status = "error"
		return c.String(http.StatusInternalServerError, "Internal Server Error")
	}

	return c.String(http.StatusOK, "Work Done")
}

func serviceStep(ctx context.Context, fail bool, userID string) error {
	ctx, span := xlog.WithOperationSpan(ctx, "serviceStep",
		xfield.String("user_id", userID),
	)
	defer span.End()

	// Simulate delay (50 to 200 ms)
	time.Sleep(time.Duration(rand.Intn(150)+50) * time.Millisecond)

	span.AddEvent("Starting work execution...")
	xlog.Info(ctx, "Starting work execution...")

	err := dbStep(ctx, fail)
	if err != nil {
		err := fmt.Errorf("dbstep failed: %w", err)

		xlog.Error(ctx, "error during work execution", xfield.Error(err))

		return err
	}

	span.AddEvent("Work completed successfully")
	xlog.Info(ctx, "Work completed successfully")

	return nil
}

func dbStep(ctx context.Context, fail bool) error {
	ctx, span := xlog.WithOperationSpan(ctx, "dbStep")
	defer span.End()

	err := doDbQueryStep(ctx, fail)
	if err != nil {
		err := fmt.Errorf("dbStep: %w", err)
		xlog.Error(ctx, err.Error(), xfield.Error(err))
		return err
	}

	return nil
}

func doDbQueryStep(ctx context.Context, fail bool) error {
	xlog.AddSpanEvent(ctx, "doDbQueryStep")
	xlog.SetSpanAttributes(ctx, attribute.Bool("app.fail", fail))
	time.Sleep(100 * time.Millisecond)

	if fail {
		return fmt.Errorf("database connection refused")
	}

	return nil
}
