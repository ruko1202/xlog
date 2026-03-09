package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/ruko1202/xlog"
)

func runWorker(ctx context.Context) {
	i := 0
	for {
		task(ctx, i)
		i++
		time.Sleep(200 * time.Millisecond)
	}
}

func task(ctx context.Context, i int) {
	userID := i
	fail := i%3 == 0

	req, err := http.NewRequest(http.MethodGet, "http://localhost:8080/api/work", nil)
	if err != nil {
		xlog.Error(ctx, "create request failed", zap.Error(err))
		return
	}
	query := req.URL.Query()
	query.Add("user_id", fmt.Sprint(userID))
	if fail {
		query.Add("fail", "true")
	}
	req.URL.RawQuery = query.Encode()

	client := http.DefaultClient
	resp, err := client.Do(req)
	if err != nil {
		xlog.Error(ctx, "do request failed", zap.Error(err))
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		xlog.Error(ctx, "read response body failed", zap.Error(err))
		return
	}
	xlog.Info(ctx, string(body),
		zap.Any("headers", resp.Header),
	)
}
