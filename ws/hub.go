package ws

import (
	"log"
	"sync"

	fws "github.com/gofiber/contrib/websocket"
)

// OutgoingMsg is the JSON envelope sent to a WebSocket client.
type OutgoingMsg struct {
	Type    string `json:"type"`              // "message" | "notification" | "error"
	Message string `json:"message,omitempty"` // used for "error" type
	Data    any    `json:"data,omitempty"`    // used for "message" / "notification" type
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
	quit       chan struct{} // closed by Close() to stop Run()
}

// NewHub creates a new Hub instance (not started yet).
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[uint]*Client),
		Register:   make(chan *Client, 16),
		Unregister: make(chan *Client, 16),
		quit:       make(chan struct{}),
	}
}

// Run starts the hub event loop. Call this in a goroutine.
func (h *Hub) Run() {
	log.Println("[WS/HUB] INFO  hub started")
	for {
		select {
		case <-h.quit:
			log.Println("[WS/HUB] INFO  hub stopped")
			return
		case client := <-h.Register:
			h.mu.Lock()
			if old, ok := h.clients[client.UserID]; ok {
				// A second tab / device opened — drop the stale connection.
				log.Printf("[WS/HUB] WARN  replacing existing connection — userID: %d", client.UserID)
				close(old.Done)
			}
			h.clients[client.UserID] = client
			total := len(h.clients)
			h.mu.Unlock()
			log.Printf("[WS/HUB] INFO  registered   — userID: %d, online: %d", client.UserID, total)

		case client := <-h.Unregister:
			h.mu.Lock()
			if current, ok := h.clients[client.UserID]; ok && current == client {
				delete(h.clients, client.UserID)
				total := len(h.clients)
				h.mu.Unlock()
				log.Printf("[WS/HUB] INFO  unregistered — userID: %d, online: %d", client.UserID, total)
			} else {
				h.mu.Unlock() // stale unregister (already replaced), nothing to do
			}
		}
	}
}

// Close signals the hub to stop its Run loop and disconnects all clients.
func (h *Hub) Close() {
	close(h.quit)
	h.mu.Lock()
	defer h.mu.Unlock()
	for _, client := range h.clients {
		close(client.Done)
	}
	h.clients = make(map[uint]*Client)
	log.Println("[WS/HUB] INFO  all clients disconnected")
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
