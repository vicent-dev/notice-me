package websocket

import (
	"github.com/gorilla/websocket"
	"slices"
)

type Hub struct {
	// Registered clients.
	clients map[*Client]bool

	// Inbound messages from the clients.
	Broadcast chan []byte

	// Register requests from the clients.
	Register chan *Client

	// Unregister requests from clients.
	unregister chan *Client
}

var Upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func NewHub() *Hub {
	return &Hub{
		Broadcast:  make(chan []byte),
		Register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

func (ws *Hub) GetClientsToNotify(clientId, clientGroupId string) []*Client {
	var clients []*Client

	for c := range ws.clients {
		if c.GroupId == clientGroupId && c.GroupId != "" && !slices.Contains(clients, c) {
			clients = append(clients, c)
		}

		if c.ID == clientId && c.ID != "" && !slices.Contains(clients, c) {
			clients = append(clients, c)
		}
	}

	return clients
}

func (ws *Hub) Run() {
	for {
		select {
		// new Client connected
		case client := <-ws.Register:
			ws.clients[client] = true
		// Client disconnected
		case client := <-ws.unregister:
			if _, ok := ws.clients[client]; ok {
				delete(ws.clients, client)
				close(client.Send)
			}
		// Send message to clients
		case message := <-ws.Broadcast:
			for client := range ws.clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(ws.clients, client)
				}
			}
		}
	}
}
