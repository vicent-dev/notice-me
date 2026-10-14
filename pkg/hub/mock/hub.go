package mock

import (
	"notice-me-server/pkg/hub"
)

type NotifyCall struct {
	ClientId      string
	ClientGroupId string
	Body          []byte
}

type ErrorCall struct {
	OriginClientId string
	Body           []byte
}

type Hub struct {
	Notified []NotifyCall
	Errors   []ErrorCall

	RunCalled      bool
	StopCalled     bool
	Registered     []*hub.Client
	Unregistered   []*hub.Client
	NotifyErr      error
	ClientsToServe []*hub.Client
}

func NewHubMock() hub.HubInterface {
	return &Hub{}
}

func (h *Hub) Run() {
	h.RunCalled = true
}

func (h *Hub) Stop() {
	h.StopCalled = true
}

func (h *Hub) RegisterClient(c *hub.Client) {
	h.Registered = append(h.Registered, c)
}

func (h *Hub) UnregisterClient(c *hub.Client) {
	h.Unregistered = append(h.Unregistered, c)
}

func (h *Hub) Notify(clientId, clientGroupId string, body []byte) {
	h.Notified = append(h.Notified, NotifyCall{
		ClientId:      clientId,
		ClientGroupId: clientGroupId,
		Body:          body,
	})
}

func (h *Hub) NotifyError(originClientId string, body []byte) error {
	h.Errors = append(h.Errors, ErrorCall{
		OriginClientId: originClientId,
		Body:           body,
	})

	return h.NotifyErr
}

func (h *Hub) Clients() []*hub.Client {
	return h.ClientsToServe
}
