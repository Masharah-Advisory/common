package redis

import (
	"context"
	"log"

	"github.com/go-redis/redis/v8"
)

type Config struct {
	RedisAddr string
	RedisPass string
	RedisDB   int
}

func NewClient(cfg *Config) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPass,
		DB:       cfg.RedisDB,
	})

	// Test the connection
	ctx := context.Background()
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		log.Printf("Warning: Failed to connect to Redis: %v", err)
		return rdb // Return client anyway, as Redis might not be critical for basic functionality
	}

	log.Println("Redis connected successfully")
	return rdb
}
