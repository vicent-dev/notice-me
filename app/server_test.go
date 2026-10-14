package app

import (
	"context"
	"testing"
	"time"
)

func TestRunGracefulShutdown(t *testing.T) {
	initialiseMocks()

	s.c.Server.Port = "0"
	s.httpServer = newHTTPServer(s)

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan error, 1)
	go func() {
		done <- s.Run(ctx)
	}()

	time.Sleep(100 * time.Millisecond)
	cancel()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("graceful shutdown returned error: %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("Run did not return after context cancellation")
	}
}
