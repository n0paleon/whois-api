package ports

import (
	"context"
	"whois-api/internal/core/domain"
)

type WhoisService interface {
	Whois(query string, ctx context.Context) (*domain.Whois, error)
}
