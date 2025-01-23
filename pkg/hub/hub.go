package hub

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"slices"
)

type HubInterface interface {
	Run()
	RegisterClient(c *Client)
	UnregisterClient(c *Client)
	Notify(clientId, clientGroupId string, body []byte)
	NotifyError(originClientId string, body []byte) error
	Clients() []*Client
}

type Hub struct {
	// Registered clients.
	clients map[*Client]bool

	// Inbound messages from the clients.
	broadcast chan []byte

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client
}

var Upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func (ws *Hub) Clients() []*Client {
	cs := make([]*Client, 0, len(ws.clients))

	for client := range ws.clients {
		cs = append(cs, client)
	}

	return cs
}

func NewHub() HubInterface {
	return &Hub{
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

func (ws *Hub) RegisterClient(c *Client) {
	ws.register <- c
}

func (ws *Hub) UnregisterClient(c *Client) {
	ws.unregister <- c
}

func (ws *Hub) getClientsToNotify(clientId, clientGroupId string) []*Client {
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

func (ws *Hub) Notify(clientId, clientGroupId string, body []byte) {
	// broadcast to all clients
	if clientId == AllClientId || clientGroupId == AllClientGroupId {
		ws.broadcast <- body
		return
	}

	clients := ws.getClientsToNotify(clientId, clientGroupId)

	for _, client := range clients {
		client.Send(body)
	}
}

func (ws *Hub) NotifyError(originClientId string, body []byte) error {
	// broadcast to all clients
	clients := ws.getClientsToNotify(originClientId, "")

	errorBody := make(map[string]string)

	errorBody["error"] = string(body)

	errorBodyBytes, err := json.Marshal(errorBody)

	if err != nil {
		return err
	}

	for _, client := range clients {
		client.Send(errorBodyBytes)
	}

	return nil
}

func (ws *Hub) Run() {
	for {
		select {
		// new Client connected
		case client := <-ws.register:
			ws.clients[client] = true
		// Client disconnected
		case client := <-ws.unregister:
			if _, ok := ws.clients[client]; ok {
				delete(ws.clients, client)
				close(client.send)
			}
		// Send message to clients
		case message := <-ws.broadcast:
			for client := range ws.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(ws.clients, client)
				}
			}
		}
	}
}
