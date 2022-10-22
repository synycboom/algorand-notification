package subscriber

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/rs/zerolog/log"
)

// ProcessorFunc represents a processor function that consume message
type ProcessorFunc func(message []byte)

// RedisConfig represents a configuration for Redis subscriber
type RedisConfig struct {
	RedisHost     string
	RedisPassword string
	Channel       string
	Processor     ProcessorFunc
}

// RedisSubscriber handles event subscription
type RedisSubscriber struct {
	conf   *RedisConfig
	rdb    *redis.Client
	pubsub *redis.PubSub
}

// NewRedis creates a new Redis subscriber
func NewRedis(conf *RedisConfig) (*RedisSubscriber, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     conf.RedisHost,
		Password: conf.RedisPassword,
	})

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(10)*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	log.Info().Msg("subscriber: connected to Redis")

	s := &RedisSubscriber{
		conf: conf,
		rdb:  rdb,
	}
	if err := s.subscribe(); err != nil {
		return nil, err
	}

	return s, nil
}

// Close closes a subscription
func (s *RedisSubscriber) Close() error {
	return s.pubsub.Close()
}

func (s *RedisSubscriber) subscribe() error {
	s.pubsub = s.rdb.Subscribe(context.Background(), s.conf.Channel)
	if _, err := s.pubsub.Receive(context.Background()); err != nil {
		return err
	}

	go func() {
		for m := range s.pubsub.Channel() {
			s.conf.Processor([]byte(m.Payload))
		}
	}()

	return nil
}
