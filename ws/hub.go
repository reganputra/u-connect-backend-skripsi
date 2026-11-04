package ws

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	fws "github.com/gofiber/contrib/websocket"
	"github.com/redis/go-redis/v9"
)

const deliveryChannel = "uconnect:ws:delivery:v1"

type OutgoingMsg struct {
	Type    string `json:"type"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
}

type Client struct {
	UserID uint
	Conn   *fws.Conn
	Send   chan []byte
	Done   chan struct{}
}

type delivery struct {
	UserID  uint   `json:"user_id"`
	Payload []byte `json:"payload"`
}

// Hub keeps sockets local; Redis broadcasts delivery requests to every replica.
// Redis Pub/Sub is transient, not a durable queue. REST/DB remains the source of truth.
type Hub struct {
	mu           sync.RWMutex
	clients      map[uint]*Client
	Register     chan *Client
	Unregister   chan *Client
	quit         chan struct{}
	once         sync.Once
	redis        *redis.Client
	subscription *redis.PubSub
	cancel       context.CancelFunc
	workers      sync.WaitGroup
}

// NewHub explicitly selects local-only (single replica) mode.
func NewHub() *Hub {
	return &Hub{clients: make(map[uint]*Client), Register: make(chan *Client, 16), Unregister: make(chan *Client, 16), quit: make(chan struct{})}
}

// NewRedisHub waits for the subscription acknowledgement before allowing sends.
func NewRedisHub(ctx context.Context, client *redis.Client) (*Hub, error) {
	if client == nil {
		return nil, fmt.Errorf("Redis hub requires a client")
	}
	h := NewHub()
	h.redis = client
	h.subscription = client.Subscribe(ctx, deliveryChannel)
	if _, err := h.subscription.Receive(ctx); err != nil {
		_ = h.subscription.Close()
		return nil, fmt.Errorf("subscribe WebSocket delivery: %w", err)
	}
	workerCtx, cancel := context.WithCancel(context.Background())
	h.cancel = cancel
	h.workers.Add(1)
	go h.receive(workerCtx)
	return h, nil
}

func (h *Hub) receive(ctx context.Context) {
	defer h.workers.Done()
	// go-redis reconnects and re-subscribes automatically; Close unblocks this channel.
	ch := h.subscription.Channel()
	for {
		select {
		case <-ctx.Done():
			return
		case message, ok := <-ch:
			if !ok {
				return
			}
			var d delivery
			if err := json.Unmarshal([]byte(message.Payload), &d); err != nil {
				slog.Warn("invalid Redis delivery payload", "error", err)
				continue
			}
			h.sendLocal(d.UserID, d.Payload)
		}
	}
}

func (h *Hub) Run() {
	for {
		select {
		case <-h.quit:
			return
		case client := <-h.Register:
			h.mu.Lock()
			select {
			case <-h.quit:
				close(client.Done)
				h.mu.Unlock()
				continue
			default:
			}
			if old := h.clients[client.UserID]; old != nil && old != client {
				close(old.Done)
				if old.Conn != nil {
					_ = old.Conn.Close()
				}
			}
			h.clients[client.UserID] = client
			h.mu.Unlock()
		case client := <-h.Unregister:
			h.mu.Lock()
			if h.clients[client.UserID] == client {
				delete(h.clients, client.UserID)
				close(client.Done)
			}
			h.mu.Unlock()
		}
	}
}

func (h *Hub) Close() {
	h.once.Do(func() {
		close(h.quit)
		if h.cancel != nil {
			h.cancel()
		}
		if h.subscription != nil {
			_ = h.subscription.Close()
		}
		h.workers.Wait()
		h.mu.Lock()
		defer h.mu.Unlock()
		for _, client := range h.clients {
			close(client.Done)
			if client.Conn != nil {
				_ = client.Conn.Close()
			}
		}
		h.clients = make(map[uint]*Client)
	})
}

// SendToUser returns whether delivery was accepted, NOT whether a remote user
// received it. Redis subscriber counts are not per-user presence acknowledgements.
func (h *Hub) SendToUser(userID uint, payload []byte) bool {
	select {
	case <-h.quit:
		return false
	default:
	}
	if h.redis == nil {
		return h.sendLocal(userID, payload)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	data, err := json.Marshal(delivery{UserID: userID, Payload: payload})
	if err == nil {
		err = h.redis.Publish(ctx, deliveryChannel, data).Err()
	}
	if err != nil {
		slog.Error("publish WebSocket delivery", "user_id", userID, "error", err)
		return false
	}
	return true
}

func (h *Hub) sendLocal(userID uint, payload []byte) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	client := h.clients[userID]
	if client == nil {
		return false
	}
	select {
	case <-client.Done:
		return false
	case client.Send <- payload:
		return true
	default:
		return false
	}
}

// IsOnline reports local presence only; it must not gate Redis publication.
func (h *Hub) IsOnline(userID uint) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.clients[userID] != nil
}
