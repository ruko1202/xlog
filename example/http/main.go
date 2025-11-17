package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/ruko1202/xlog"
)

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}

	xlog.ReplaceGlobal(logger)
	ctx := xlog.ContextWithLogger(context.Background(), logger)

	runApp(ctx)
}

func runApp(ctx context.Context) {
	xlog.Info(ctx, "start app on port 8080...")
	http.Handle("/example", http.HandlerFunc(handleRequest))
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		xlog.Fatalf(ctx, "ListenAndServe: %v", err)
	}
	xlog.Info(ctx, "app stopped")
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx = xlog.WithOperation(ctx, "handleRequest",
		zap.String("request_id", uuid.NewString()),
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path),
		zap.String("query", r.URL.RawQuery),
	)

	xlog.Info(ctx, "request processing started")

	if err := processRequest(ctx, r); err != nil {
		xlog.Error(ctx, "request processing error", zap.Error(err))
		http.Error(w, "Internal Server Error", 500)
		return
	}

	xlog.Info(ctx, "request successfully processed")
}

func processRequest(ctx context.Context, r *http.Request) error {
	ctx = xlog.WithOperation(ctx, "processRequest")

	userID := r.URL.Query().Get("userId")
	if userID == "" {
		return fmt.Errorf("missing `userId`")
	}

	ctx = xlog.WithFields(ctx, zap.String("userId", userID))
	xlog.Info(ctx, "request processed")

	return nil
}
