package repository

import (
	"context"
	"errors"
	"github.com/bytedance/sonic"
	"github.com/redis/go-redis/v9"
	"time"
	"whois-api/internal/core/domain"
	"whois-api/internal/core/ports"
	"whois-api/internal/infrastructure/database"
	"whois-api/pkg/logger"
)

type Whois struct {
	conn *database.RedisConn
}

var (
	keyPrefix  = "whois:"
	defaultTTL = 30 * 24 * time.Hour
)

func NewWhoisRepository(conn *database.RedisConn) ports.WhoisRepository {
	return &Whois{conn: conn}
}

func (r *Whois) GetWhoisData(query string, ctx context.Context) (*domain.Whois, error) {
	cacheKey := keyPrefix + query
	data, err := r.conn.Get(ctx, cacheKey).Result()
	if err != nil {
		if !errors.Is(err, redis.Nil) {
			logger.L().Error(err)
		}
		return nil, err
	}

	var whoisData domain.Whois
	if err = sonic.Unmarshal([]byte(data), &whoisData); err != nil {
		logger.L().Errorf("failed to unmarshal: %v", err)
		return nil, err
	}

	return &whoisData, nil
}

func (r *Whois) SaveWhoisData(domain string, whoisData *domain.Whois, ctx context.Context, ttl ...time.Duration) error {
	cacheKey := keyPrefix + domain
	data, err := sonic.Marshal(whoisData)
	if err != nil {
		logger.L().Errorf("failed to marshal: %v", err)
		return err
	}

	var expiration time.Duration
	if len(ttl) > 0 {
		expiration = ttl[0]
	} else {
		expiration = defaultTTL
	}

	return r.conn.Set(ctx, cacheKey, data, expiration).Err()
}
