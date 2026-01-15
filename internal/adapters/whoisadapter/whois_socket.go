package whoisadapter

import (
	"context"
	"errors"
	"time"

	"github.com/domainr/whois"
)

type socketWhois struct {
	client *whois.Client
}

func newSocket(timeout time.Duration) *socketWhois {
	return &socketWhois{
		client: whois.NewClient(timeout),
	}
}

func (s *socketWhois) fetch(ctx context.Context, query string) ([]byte, error) {
	req, err := whois.NewRequest(query)
	if err != nil {
		return nil, err
	}

	res, err := s.client.FetchContext(ctx, req)
	if err != nil {
		return nil, err
	}

	if res.Body == nil || len(res.Body) < 1 {
		return nil, errors.New("whois server return empty response")
	}

	return res.Body, nil
}
