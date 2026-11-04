package ws

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"github.com/redis/go-redis/v9"
	"time"
)

// TicketStore retains tokens server-side; random handles are consumed atomically
// across replicas. No access token is exposed in the WebSocket URL.
type TicketStore struct{ redis *redis.Client }

func NewTicketStore(client *redis.Client) *TicketStore { return &TicketStore{redis: client} }
func ticketKey(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return "uconnect:ws:ticket:" + hex.EncodeToString(sum[:])
}
func (s *TicketStore) Issue(ctx context.Context, access string) (string, error) {
	if s == nil || s.redis == nil {
		return "", fmt.Errorf("ticket storage unavailable")
	}
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("generate ticket: %w", err)
	}
	raw := base64.RawURLEncoding.EncodeToString(bytes)
	if err := s.redis.Set(ctx, ticketKey(raw), access, 30*time.Second).Err(); err != nil {
		return "", fmt.Errorf("store ticket: %w", err)
	}
	return raw, nil
}
func (s *TicketStore) Consume(ctx context.Context, raw string) (string, error) {
	if s == nil || s.redis == nil || len(raw) != 43 {
		return "", fmt.Errorf("invalid ticket")
	}
	access, err := s.redis.GetDel(ctx, ticketKey(raw)).Result()
	if err != nil {
		return "", fmt.Errorf("consume ticket: %w", err)
	}
	return access, nil
}
