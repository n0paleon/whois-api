package ports

import (
	"context"
	"time"
	"whois-api/internal/core/domain"
)

type WhoisService interface {
	CheckWhois(query string, ctx context.Context) (*domain.Whois, error)
}

type WhoisAdapter interface {
	GetWhoisData(query string, ctx context.Context) (*domain.Whois, error)
}

type WhoisRepository interface {
	GetWhoisData(query string, ctx context.Context) (*domain.Whois, error)
	SaveWhoisData(domain string, whoisData *domain.Whois, ctx context.Context, ttl ...time.Duration) error
}
