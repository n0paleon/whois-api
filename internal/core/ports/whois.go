package ports

import (
	"context"
	"time"
	"whois-api/internal/core/domain"
)

type WhoisService interface {
	SingleLookup(query string, ctx context.Context) (*domain.Whois, error)
	MassLookup(query []string, ctx context.Context, rateLimit time.Duration) (map[string]*domain.Whois, error)
}

type WhoisAdapter interface {
	GetWhoisData(query string, ctx context.Context) (*domain.Whois, error)
}

type WhoisRepository interface {
	GetWhoisData(query string, ctx context.Context) (*domain.Whois, error)
	SaveWhoisData(domain string, whoisData *domain.Whois, ctx context.Context, ttl ...time.Duration) error
	SetCacheAge(query string, ctx context.Context, ttl ...time.Duration) error
	GetCacheAge(query string, ctx context.Context) (time.Duration, error)
}
