package rabbit

import (
	"testing"
	"time"
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
