package ws

import (
	"context"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

func testRedis(t *testing.T) *redis.Client {
	t.Helper()
	raw := os.Getenv("TEST_REDIS_URL")
	if raw == "" {
		t.Skip("set TEST_REDIS_URL for Redis integration tests")
	}
	options, err := redis.ParseURL(raw)
	if err != nil {
		t.Fatal(err)
	}
	client := redis.NewClient(options)
	t.Cleanup(func() { _ = client.Close() })
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		t.Fatal(err)
	}
	return client
}

// Separate Redis clients and hubs model two independent application replicas.
func TestRedisCrossReplicaDelivery(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	publisher, err := NewRedisHub(ctx, testRedis(t))
	if err != nil {
		t.Fatal(err)
	}
	defer publisher.Close()
	receiver, err := NewRedisHub(ctx, testRedis(t))
	if err != nil {
		t.Fatal(err)
	}
	defer receiver.Close()
	client := &Client{UserID: 918273, Send: make(chan []byte, 2), Done: make(chan struct{})}
	receiver.mu.Lock()
	receiver.clients[client.UserID] = client
	receiver.mu.Unlock()
	if !publisher.SendToUser(client.UserID, []byte("cross-replica")) {
		t.Fatal("publication failed")
	}
	select {
	case got := <-client.Send:
		if string(got) != "cross-replica" {
			t.Fatalf("unexpected payload: %q", got)
		}
	case <-ctx.Done():
		t.Fatal("remote hub did not receive payload")
	}
	select {
	case <-client.Send:
		t.Fatal("duplicate local delivery")
	case <-time.After(100 * time.Millisecond):
	}
}

func TestRedisTicketSingleUse(t *testing.T) {
	client := testRedis(t)
	store := NewTicketStore(client)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	ticket, err := store.Issue(ctx, "server-side-token")
	if err != nil {
		t.Fatal(err)
	}
	if ticket == "server-side-token" || len(ticket) != 43 {
		t.Fatal("ticket must be an opaque random handle")
	}
	ttl, err := client.TTL(ctx, ticketKey(ticket)).Result()
	if err != nil || ttl <= 0 || ttl > 30*time.Second {
		t.Fatalf("unexpected TTL %s: %v", ttl, err)
	}
	var successes atomic.Int32
	var wg sync.WaitGroup
	for range 12 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			got, err := NewTicketStore(client).Consume(ctx, ticket)
			if err == nil {
				if got != "server-side-token" {
					t.Error("unexpected token")
				}
				successes.Add(1)
			}
		}()
	}
	wg.Wait()
	if successes.Load() != 1 {
		t.Fatalf("expected exactly one consumption, got %d", successes.Load())
	}
	for _, tc := range []struct{ name, raw string }{{"missing", ""}, {"malformed", "nope"}, {"replay", ticket}} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := store.Consume(ctx, tc.raw); err == nil {
				t.Fatal("expected rejection")
			}
		})
	}
	expired, err := store.Issue(ctx, "expired")
	if err != nil {
		t.Fatal(err)
	}
	if err := client.PExpire(ctx, ticketKey(expired), time.Millisecond).Err(); err != nil {
		t.Fatal(err)
	}
	time.Sleep(10 * time.Millisecond)
	if _, err := store.Consume(ctx, expired); err == nil {
		t.Fatal("expired ticket accepted")
	}
}

func TestLocalHubClose(t *testing.T) {
	for _, tc := range []struct {
		name   string
		buffer int
	}{{"buffered", 1}, {"full", 0}} {
		t.Run(tc.name, func(t *testing.T) {
			hub := NewHub()
			client := &Client{UserID: 1, Send: make(chan []byte, tc.buffer), Done: make(chan struct{})}
			hub.clients[1] = client
			if got := hub.SendToUser(1, []byte("hello")); got != (tc.buffer > 0) {
				t.Fatalf("accepted=%v", got)
			}
			hub.Close()
			hub.Close()
			if hub.SendToUser(1, nil) {
				t.Fatal("closed hub accepted delivery")
			}
		})
	}
}
