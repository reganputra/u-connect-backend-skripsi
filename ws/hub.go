package ws

import (
	"sync"

	fws "github.com/gofiber/contrib/websocket"
)

// OutgoingMsg is the JSON envelope sent to a WebSocket client.
type OutgoingMsg struct {
	Type    string `json:"type"`             // "message" | "error" | "read"
	Message string `json:"message,omitempty"` // used for "error" type
	Data    any    `json:"data,omitempty"`    // used for "message" / "read" type
}

// Client represents a connected WebSocket session.
type Client struct {
	UserID uint
	Conn   *fws.Conn
	Send   chan []byte
	Done   chan struct{}
}

// Hub manages all active WebSocket connections.
type Hub struct {
	mu         sync.RWMutex
	clients    map[uint]*Client // keyed by userID (one connection per user)
	Register   chan *Client
	Unregister chan *Client
	Deliver    chan delivery // direct message delivery
}

type delivery struct {
	ReceiverID uint
	Payload    []byte
}

// NewHub creates a new Hub instance (not started yet).
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[uint]*Client),
		Register:   make(chan *Client, 16),
		Unregister: make(chan *Client, 16),
		Deliver:    make(chan delivery, 256),
	}
}

// Run starts the hub event loop. Call this in a goroutine.
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.mu.Lock()
			// If the user already has an open connection, close the old one gracefully.
			if old, ok := h.clients[client.UserID]; ok {
				close(old.Done)
			}
			h.clients[client.UserID] = client
			h.mu.Unlock()

		case client := <-h.Unregister:
			h.mu.Lock()
			if current, ok := h.clients[client.UserID]; ok && current == client {
				delete(h.clients, client.UserID)
			}
			h.mu.Unlock()

		case d := <-h.Deliver:
			h.mu.RLock()
			client, ok := h.clients[d.ReceiverID]
			h.mu.RUnlock()
			if ok {
				select {
				case client.Send <- d.Payload:
				default:
					// Client send channel full — drop to avoid blocking hub goroutine
				}
			}
		}
	}
}

// SendToUser enqueues a payload for direct delivery to a connected user.
// Returns false if the user is not currently connected.
func (h *Hub) SendToUser(userID uint, payload []byte) bool {
	h.mu.RLock()
	client, ok := h.clients[userID]
	h.mu.RUnlock()
	if !ok {
		return false
	}
	select {
	case client.Send <- payload:
		return true
	default:
		return false
	}
}

// IsOnline reports whether a user has an active WebSocket connection.
func (h *Hub) IsOnline(userID uint) bool {
	h.mu.RLock()
	_, ok := h.clients[userID]
	h.mu.RUnlock()
	return ok
}
