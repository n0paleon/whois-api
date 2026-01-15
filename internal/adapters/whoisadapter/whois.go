package whoisadapter

import (
	"context"
	"time"
	"whois-api/internal/core/domain"
	"whois-api/internal/core/ports"

	"github.com/go-resty/resty/v2"
)

type Whois struct {
	client *resty.Client
	socket *socketWhois
}

var (
	defaultTimeout = 15 * time.Second
)

func NewWhoisAdapter() ports.WhoisAdapter {
	return &Whois{
		client: resty.New(),
		socket: newSocket(defaultTimeout),
	}
}

func (a *Whois) GetWhoisData(query string, ctx context.Context) (*domain.Whois, error) {
	pCtx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	res, err := a.socket.fetch(pCtx, query)
	if err != nil {
		return nil, err
	}

	return parseRawWhois(res)
}

func (a *Whois) GetRawWhoisData(query string, ctx context.Context) (string, error) {
	pCtx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	res, err := a.socket.fetch(pCtx, query)
	if err != nil {
		return "", err
	}

	return string(res), nil
}
