package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"github.com/ruko1202/xlog"
	"github.com/ruko1202/xlog/xfield"
)

// Global objects for the package
var (
	appName           = "sandbox-service"
	appVersion        = "v0.0.1"
	traceCollectorURL = "localhost:4317"

	counter metric.Int64Counter
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(),
		syscall.SIGTERM,
		syscall.SIGINT,
		syscall.SIGQUIT,
	)
	defer stop()

	// 1. Initialize Zap
	logger, _ := zap.NewDevelopment(
		zap.AddCaller(),
		zap.AddCallerSkip(1),
		zap.WithCaller(true),
	)
	defer logger.Sync()
	xlog.ReplaceGlobalLogger(xlog.NewZapAdapter(logger))
	zap.ReplaceGlobals(logger)

	ctx = xlog.ContextWithLogger(ctx, xlog.NewZapAdapter(logger))

	shutdown, err := initOTel(ctx)
	if err != nil {
		xlog.Panic(ctx, "failed to create metric exporter", xfield.Error(err))
	}
	defer shutdown(ctx)
	xlog.ReplaceTracerName(appName)

	meter := otel.Meter(appName)
	initMetrics(ctx, meter)

	e := echo.New()
	echoHandlers(ctx, e)

	go func() {
		logger.Info("Server is running on :8080")
		e.Start(":8080")
	}()
	go runWorker(ctx)

	<-ctx.Done()
	e.Shutdown(ctx)
}

func echoHandlers(_ context.Context, e *echo.Echo) {
	gr := e.Group("")

	// Endpoint for Prometheus scraping
	gr.GET("/metrics", echo.WrapHandler(promhttp.Handler()))

	apiGr := e.Group("api")
	// Middleware for parsing traces from HTTP
	apiGr.Use(otelecho.Middleware(appName))

	// Middleware for injecting TraceID into X-Request-ID response header
	apiGr.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx := c.Request().Context()
			spanCtx := trace.SpanContextFromContext(ctx)

			// If trace was successfully created/received
			if spanCtx.HasTraceID() {
				traceID := spanCtx.TraceID().String()

				// Return to client in response header
				c.Response().Header().Set(echo.HeaderXRequestID, traceID)

				// Store in Echo context (so Echo's standard loggers can pick it up)
				c.Set(echo.HeaderXRequestID, traceID)
			}
			return next(c)
		}
	})

	// Our business logic
	apiGr.GET("/work", handleWork)
}

func initMetrics(ctx context.Context, meter metric.Meter) {
	var err error
	counter, err = meter.Int64Counter("http_requests_total", metric.WithDescription("Total HTTP requests"))
	if err != nil {
		xlog.Panic(ctx, "failed to init counter", xfield.Error(err))
	}
}
