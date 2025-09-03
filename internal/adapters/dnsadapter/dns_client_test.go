package dnsadapter

import (
	"context"
	"github.com/miekg/dns"
	"testing"
	"time"
	"whois-api/internal/infrastructure/workers"
)

func init() {
	_ = workers.InitWorkerPool(1500)
}

func TestDNS_Query_ARecord(t *testing.T) {
	dnsClient := NewDNS("8.8.8.8:53")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	records, err := dnsClient.Query(ctx, "example.com", dns.TypeA)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	if len(records) == 0 {
		t.Errorf("expected at least one A record for example.com, got none")
	}
}

func TestDNS_QueryAll(t *testing.T) {
	dnsClient := NewDNS("8.8.8.8:53")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := dnsClient.QueryAll(ctx, "example.com")
	if err != nil {
		t.Fatalf("QueryAll failed: %v", err)
	}

	if len(result) == 0 {
		t.Errorf("expected records, got empty result")
	}

	validRecord := false
	for _, record := range result {
		if record["type"] == "A" {
			validRecord = true
			break
		}
	}
	if !validRecord {
		t.Errorf("expected A record to be valid")
	}
}

func BenchmarkDNS_QueryAll_Cloudflare(b *testing.B) {
	dnsClient := NewDNS("1.1.1.1:53")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, _ = dnsClient.QueryAll(ctx, "example.com")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := dnsClient.QueryAll(ctx, "example.com")
		if err != nil {
			b.Fatalf("QueryAll failed: %v", err)
		}
	}
}

func BenchmarkDNS_QueryAll_Google(b *testing.B) {
	dnsClient := NewDNS("8.8.8.8:53")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, _ = dnsClient.QueryAll(ctx, "example.com")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := dnsClient.QueryAll(ctx, "example.com")
		if err != nil {
			b.Fatalf("QueryAll failed: %v", err)
		}
	}
}
