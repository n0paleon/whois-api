package database

import (
	"context"
	"github.com/redis/go-redis/v9"
	"whois-api/internal/config"
)

type RedisConn struct {
	*redis.Client
}

func NewRedisConn(cfg *config.Config) (*RedisConn, error) {
	conn := redis.NewClient(&redis.Options{
		Addr: cfg.Service.Redis.Addr,
		DB:   cfg.Service.Redis.DB,
	})

	_, err := conn.Ping(context.Background()).Result()
	if err != nil {
		return nil, err
	}

	return &RedisConn{conn}, nil
}
