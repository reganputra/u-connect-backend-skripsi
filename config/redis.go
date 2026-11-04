package config

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"time"
)

// OpenRedis requires an explicit URL. Configured Redis failure must never silently
// degrade into isolated per-replica delivery or bypass security checks.
func OpenRedis(ctx context.Context, rawURL string) (*redis.Client, error) {
	options, err := redis.ParseURL(rawURL)
	if err != nil {
		return nil, fmt.Errorf("parse Redis URL: invalid configuration")
	}
	options.DialTimeout = 2 * time.Second
	options.ReadTimeout = 2 * time.Second
	options.WriteTimeout = 2 * time.Second
	options.MaxRetries = 1
	client := redis.NewClient(options)
	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("Redis unavailable: %w", err)
	}
	return client, nil
}
