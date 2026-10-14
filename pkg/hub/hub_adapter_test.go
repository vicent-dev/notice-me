package hub

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func newTestHub() *Hub {
	return &Hub{
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
		quit:       make(chan struct{}),
	}
}

// testClient builds a client whose outbound channel is buffered by size so
// synchronous Notify paths do not block.
func testClient(id, group string, size int) *Client {
	return &Client{ID: id, GroupId: group, send: make(chan []byte, size)}
}

// waitForClose blocks until ch is closed (or times out), draining any buffered
// values first. It establishes a happens-before edge with the hub goroutine.
func waitForClose(t *testing.T, ch chan []byte) {
	t.Helper()

	deadline := time.After(time.Second)
	for {
		select {
		case _, ok := <-ch:
			if !ok {
				return
			}
		case <-deadline:
			t.Fatal("channel was not closed in time")
		}
	}
}

func TestClientsReflectsRegistry(t *testing.T) {
	h := newTestHub()
	c1 := testClient("a", "g", 1)
	c2 := testClient("b", "h", 1)
	h.clients[c1] = true
	h.clients[c2] = true

	if got := len(h.Clients()); got != 2 {
		t.Fatalf("Clients() len = %d, want 2", got)
	}
}

func TestNotifyBroadcast(t *testing.T) {
	h := newTestHub()
	go h.Run()
	defer h.Stop()

	c1 := testClient("a", "g", 1)
	c2 := testClient("b", "h", 1)
	h.RegisterClient(c1)
	h.RegisterClient(c2)

	h.Notify(AllClientId, AllClientGroupId, []byte("blast"))

	for _, c := range []*Client{c1, c2} {
		select {
		case msg := <-c.send:
			if string(msg) != "blast" {
				t.Fatalf("client %s got %q, want %q", c.ID, msg, "blast")
			}
		case <-time.After(time.Second):
			t.Fatalf("client %s did not receive broadcast", c.ID)
		}
	}

	// Unregister before Stop so the hub's shutdown does not close nil Conns.
	h.UnregisterClient(c1)
	waitForClose(t, c1.send)
	h.UnregisterClient(c2)
	waitForClose(t, c2.send)
}

func TestNotifyByClientId(t *testing.T) {
	h := newTestHub()

	target := testClient("target", "g1", 1)
	other := testClient("other", "g2", 1)
	h.clients[target] = true
	h.clients[other] = true

	h.Notify("target", "nobody", []byte("direct"))

	select {
	case msg := <-target.send:
		if string(msg) != "direct" {
			t.Fatalf("target got %q, want %q", msg, "direct")
		}
	case <-time.After(time.Second):
		t.Fatal("target did not receive direct message")
	}

	if len(other.send) != 0 {
		t.Fatal("other client should not receive a targeted message")
	}
}

func TestNotifyByGroupId(t *testing.T) {
	h := newTestHub()

	member := testClient("whoever", "admins", 1)
	outsider := testClient("whoever2", "guests", 1)
	h.clients[member] = true
	h.clients[outsider] = true

	h.Notify("nobody", "admins", []byte("group"))

	select {
	case msg := <-member.send:
		if string(msg) != "group" {
			t.Fatalf("member got %q, want %q", msg, "group")
		}
	case <-time.After(time.Second):
		t.Fatal("group member did not receive message")
	}

	if len(outsider.send) != 0 {
		t.Fatal("outsider should not receive group message")
	}
}

func TestNotifyDeduplicatesClientMatchingIdAndGroup(t *testing.T) {
	h := newTestHub()

	both := testClient("same", "same-group", 4)
	h.clients[both] = true

	h.Notify("same", "same-group", []byte("once"))

	select {
	case msg := <-both.send:
		if string(msg) != "once" {
			t.Fatalf("got %q, want %q", msg, "once")
		}
	case <-time.After(time.Second):
		t.Fatal("client did not receive message")
	}

	if len(both.send) != 0 {
		t.Fatal("client should receive exactly one message")
	}
}

func TestNotifyNoMatch(t *testing.T) {
	h := newTestHub()

	c := testClient("a", "g", 1)
	h.clients[c] = true

	h.Notify("nobody", "nothing", []byte("lost"))

	if len(c.send) != 0 {
		t.Fatal("client should not receive a non-matching message")
	}
}

func TestNotifyErrorTargetsOriginClient(t *testing.T) {
	h := newTestHub()

	origin := testClient("origin", "g", 1)
	bystander := testClient("bystander", "g", 1)
	h.clients[origin] = true
	h.clients[bystander] = true

	if err := h.NotifyError("origin", []byte("boom")); err != nil {
		t.Fatalf("NotifyError: %v", err)
	}

	select {
	case msg := <-origin.send:
		var payload map[string]string
		if err := json.Unmarshal(msg, &payload); err != nil {
			t.Fatalf("NotifyError payload invalid: %v", err)
		}
		if payload["error"] != "boom" {
			t.Fatalf("error = %q, want %q", payload["error"], "boom")
		}
	case <-time.After(time.Second):
		t.Fatal("origin client did not receive error")
	}

	if len(bystander.send) != 0 {
		t.Fatal("bystander should not receive the error")
	}
}

func TestUnregisterClientRemovesAndCloses(t *testing.T) {
	h := newTestHub()
	go h.Run()
	defer h.Stop()

	c := testClient("a", "g", 1)
	h.RegisterClient(c)
	h.UnregisterClient(c)

	// Run closes the client's send channel once it processes the unregister.
	waitForClose(t, c.send)
}

func TestBroadcastEvictsSlowClient(t *testing.T) {
	h := newTestHub()
	go h.Run()

	// Unbuffered send channel is never ready, so the hub evicts this client.
	slow := testClient("slow", "g", 0)
	h.RegisterClient(slow)

	h.Notify(AllClientId, AllClientGroupId, []byte("flood"))

	// Run processes channel cases sequentially: registering the marker can only
	// be accepted after the broadcast case (and its eviction) has completed.
	h.RegisterClient(testClient("marker", "g", 1))

	select {
	case _, ok := <-slow.send:
		if ok {
			t.Fatal("expected the slow client's channel to be closed by eviction")
		}
	case <-time.After(time.Second):
		t.Fatal("slow client was not evicted")
	}
}

func TestClientSendPushesToChannel(t *testing.T) {
	c := testClient("a", "g", 1)
	c.Send([]byte("outbound"))

	select {
	case msg := <-c.send:
		if string(msg) != "outbound" {
			t.Fatalf("got %q, want %q", msg, "outbound")
		}
	default:
		t.Fatal("Send did not push to the channel")
	}
}

func TestClientWriteDeliversToConnection(t *testing.T) {
	h := newTestHub()
	go h.Run()
	defer h.Stop()

	received := make(chan []byte, 1)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := Upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		client := NewClient("id", "group", h, conn, make(chan []byte, 1))
		h.RegisterClient(client)
		go client.Write()

		// Feed the outbound channel; Client.Write relays it to the socket.
		client.send <- []byte("through-socket")

		_ = conn.SetReadDeadline(time.Now().Add(time.Second))
		_, msg, err := conn.ReadMessage()
		if err == nil {
			received <- msg
		}
	}))
	defer srv.Close()

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.Close()

	_ = conn.SetReadDeadline(time.Now().Add(time.Second))
	_, msg, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if string(msg) != "through-socket" {
		t.Fatalf("received %q, want %q", msg, "through-socket")
	}
}

func TestClientReadUnregistersOnDisconnect(t *testing.T) {
	h := newTestHub()
	go h.Run()
	defer h.Stop()

	clientCh := make(chan *Client, 1)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := Upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		client := NewClient("id", "group", h, conn, make(chan []byte, 1))
		h.RegisterClient(client)
		clientCh <- client
		go client.Read()
	}))
	defer srv.Close()

	url := "ws" + strings.TrimPrefix(srv.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}

	client := <-clientCh

	// Send and then close; Read consumes the frame and unregisters on error.
	if err := conn.WriteMessage(websocket.TextMessage, []byte("hi")); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := conn.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}

	// Read's deferred UnregisterClient makes the hub close the send channel.
	waitForClose(t, client.send)
}
