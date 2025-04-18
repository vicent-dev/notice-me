package hub

import (
	"time"

	"github.com/en-vee/alog"
	"github.com/gorilla/websocket"
)

// Client is a middleman between the websocket connection and the WebsocketService.
type Client struct {
	ID string

	GroupId string

	WebsocketService HubInterface

	// The websocket connection.
	Conn *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte
}

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512

	AllClientId      = "*"
	AllClientGroupId = "*"
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

func NewClient(id, groupId string, ws HubInterface, conn *websocket.Conn, send chan []byte) *Client {
	return &Client{
		ID:               id,
		GroupId:          groupId,
		WebsocketService: ws,
		Conn:             conn,
		send:             send,
	}
}

func (c *Client) Read() {
	defer func() {
		c.WebsocketService.UnregisterClient(c)
		c.Conn.Close()
		alog.Info("client disconnected " + c.ID)
	}()
	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error { c.Conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, _, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				alog.Error("Err unexpected close client" + err.Error())
			}
			break
		}
	}
}

func (c *Client) Write() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The WebsocketService closed the channel.
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) Send(message []byte) {
	c.send <- message
}
