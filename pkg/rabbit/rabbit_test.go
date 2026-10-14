package rabbit

import (
	"testing"
	"time"

	"notice-me-server/pkg/config"
)

func TestStopClosesQuit(t *testing.T) {
	r := NewRabbit(nil, 1, nil).(*Rabbit)

	r.Stop()

	select {
	case <-r.quit:
	case <-time.After(time.Second):
		t.Fatal("quit channel was not closed after Stop")
	}
}

func TestRunConsumersReturnsWithoutQueues(t *testing.T) {
	r := NewRabbit(nil, 1, nil).(*Rabbit)

	done := make(chan struct{})
	go func() {
		r.RunConsumers(nil)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("RunConsumers did not return with no queues")
	}
}

func TestNewRabbitStoresConfig(t *testing.T) {
	qc := []config.QueueConfig{{Name: "q1", Exchange: "e1"}}
	r := NewRabbit(nil, 3, qc)

	if got := r.GetQueuesConfig(); len(got) != 1 || got[0].Name != "q1" {
		t.Fatalf("GetQueuesConfig() = %+v, want [{q1 e1}]", got)
	}
}

func TestSetAndGetConnection(t *testing.T) {
	r := NewRabbit(nil, 1, nil)

	if r.GetConnection() != nil {
		t.Fatal("expected nil connection initially")
	}

	// Setting and retrieving a connection round-trips (nil-safe here).
	r.SetConnection(nil)
	if r.GetConnection() != nil {
		t.Fatal("expected nil connection after SetConnection(nil)")
	}
}
