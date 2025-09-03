package dnsadapter

import (
	"context"
	"github.com/miekg/dns"
	"sync"
)

type DNS struct {
	resolver string
}

func NewDNS(resolver string) *DNS {
	if resolver == "" {
		resolver = "8.8.8.8:53" // default Google DNS
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

// Query: ambil record berdasarkan type
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

		go func(qtype uint16) {
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
		}(qtype)
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
