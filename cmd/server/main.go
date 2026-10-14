package main

import (
	"context"
	"os/signal"
	"syscall"

	"notice-me-server/app"

	"github.com/en-vee/alog"
)

func main() {
	s := app.NewServer()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := s.Run(ctx); err != nil {
		alog.Error("server run error: " + err.Error())
	}
}
