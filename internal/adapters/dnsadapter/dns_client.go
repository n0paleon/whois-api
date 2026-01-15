package dnsadapter

import (
	"context"
	"sync"
	"whois-api/internal/infrastructure/workers"

	"github.com/miekg/dns"
)

type DNS struct {
	resolver string
}

func NewDNS(resolver string) *DNS {
	if resolver == "" {
		resolver = "1.1.1.1:53" // default Google DNS
	}
	return &DNS{resolver: resolver}
}

type DNSRecord struct {
	Type  string `json:"type"`
	Value string `json:"value"`
	TTL   uint32 `json:"ttl"`
	Class string `json:"class"`
	Raw   string `json:"raw"`
}

// Query record berdasarkan type
func (a *DNS) Query(ctx context.Context, domain string, queryType uint16) ([]dns.RR, error) {
	m := new(dns.Msg)
	m.SetQuestion(dns.Fqdn(domain), queryType)

	c := new(dns.Client)
	in, _, err := c.ExchangeContext(ctx, m, a.resolver)
	if err != nil {
		return nil, err
	}

	return in.Answer, nil
}

func (a *DNS) QueryAll(ctx context.Context, domain string) (DNSRecords, error) {
	var results DNSRecords
	var mu sync.Mutex
	var wg sync.WaitGroup

	errCh := make(chan error, len(recordTypes))

	for _, qtype := range recordTypes {
		wg.Add(1)

		_ = workers.Pool.Submit(func() {
			defer wg.Done()

			ans, err := a.Query(ctx, domain, qtype)
			if err != nil {
				errCh <- err
				return
			}

			mu.Lock()
			for _, rr := range ans {
				results = append(results, parseRR(rr))
			}
			mu.Unlock()
		})
	}

	wg.Wait()
	close(errCh)

	var firstErr error
	for err := range errCh {
		if firstErr == nil {
			firstErr = err
		}
	}

	return results, firstErr
}
