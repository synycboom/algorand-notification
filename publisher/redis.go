package publisher

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/rs/zerolog/log"
)

// RedisConfig represents a configuration for Redis publisher
type RedisConfig struct {
	RedisHost     string
	RedisPassword string
	Channel       string
}

// RedisPublisher handles event publishing
type RedisPublisher struct {
	conf RedisConfig
	rdb  *redis.Client
}

// NewRedis creates a new Redis publisher
func NewRedis(conf RedisConfig) (*RedisPublisher, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     conf.RedisHost,
		Password: conf.RedisPassword,
	})

  ctx, cancel := context.WithTimeout(context.Background(), time.Duration(10) * time.Second)
  defer cancel()

  if err := rdb.Ping(ctx).Err(); err != nil {
    return nil, err
  }

  log.Info().Msg("publisher: connected to Redis")

	return &RedisPublisher{
		conf: conf,
		rdb:  rdb,
	}, nil
}

// Publish send an event
func (p *RedisPublisher) Publish(ctx context.Context, message []byte) error {
  if err := p.rdb.Publish(ctx, p.conf.Channel, message).Err(); err != nil {
    return err
  }

  return nil
}

