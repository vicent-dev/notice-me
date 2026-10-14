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

func TestNewHTTPServerConfiguration(t *testing.T) {
	initialiseMocks()
	s.c.Server.Port = "9999"

	hs := newHTTPServer(s)

	if hs.Addr != ":9999" {
		t.Fatalf("Addr = %q, want %q", hs.Addr, ":9999")
	}
	if hs.Handler == nil {
		t.Fatal("Handler is nil")
	}
	if hs.ReadHeaderTimeout == 0 {
		t.Fatal("ReadHeaderTimeout is not set")
	}
}

func TestGetRepositoryMissingReturnsNil(t *testing.T) {
	initialiseMocks()

	if got := s.getRepository("does-not-exist"); got != nil {
		t.Fatalf("expected nil for unknown repository, got %v", got)
	}
}

func TestServerErrorLogWriterSwallowsOutput(t *testing.T) {
	logger := newServerErrorLog()

	n, err := logger.Writer().Write([]byte("ignored"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != len("ignored") {
		t.Fatalf("Write() = %d bytes, want %d", n, len("ignored"))
	}
}
