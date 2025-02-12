package whoisadapter

import (
	"context"
	"github.com/likexian/whois"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
	"whois-api/pkg/valid"
)

var (
	validDomain   = "turulabs.com"
	subDomain     = "www.turulabs.com"
	invalidDomain = "wwwturulabscom"
)

func TestGetWhoisData(t *testing.T) {
	whois := NewWhoisAdapter()
	ctx := context.Background()

	t.Run("Test GetWhoisData with valid domain", func(t *testing.T) {
		data, err := whois.GetWhoisData(validDomain, ctx)
		assert.Nil(t, err)
		assert.NotNil(t, data)
		assert.Equal(t, validDomain, data.Domain.Domain)
	})
	t.Run("Test GetWhoisData with invalid domain", func(t *testing.T) {
		data, err := whois.GetWhoisData(invalidDomain, ctx)
		assert.NotNil(t, err)
		assert.Nil(t, data)
	})
	t.Run("Test GetWhoisData with valid sub domain", func(t *testing.T) {
		rootDomain, err := valid.ParseRootDomain(subDomain)
		assert.Nil(t, err)
		assert.Equal(t, rootDomain, validDomain)

		data, err := whois.GetWhoisData(rootDomain, ctx)
		assert.Nil(t, err)
		assert.NotNil(t, data)
	})
}

func BenchmarkGetWhois(b *testing.B) {
	whois := NewWhoisAdapter()
	ctx := context.Background()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			result, err := whois.GetWhoisData(validDomain, ctx)
			if err != nil {
				b.Fatalf("unexpected error: %v", err)
			}
			if result == nil || result.Domain.Domain != validDomain {
				b.Fatalf("unexpected result: %v", result)
			}
		}
	})

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := whois.GetWhoisData(invalidDomain, ctx)
			if err == nil {
				b.Fatalf("whois return nil error for invalid domain but expecting not nil error")
			}
		}
	})
}

func TestLikexianWhois(t *testing.T) {
	domainName := "turulabs.com"
	timeout := 2000 * time.Millisecond

	client := whois.NewClient()
	client.SetTimeout(timeout)
	client.SetDisableReferral(false)

	server, _, _ := GetWhoisServer(domainName)
	result, err := client.Whois(domainName, server)
	assert.Nil(t, err)
	assert.NotNil(t, result)
}

func TestGetAvailableTLDs(t *testing.T) {
	list := GetAvailableTLDs()
	assert.NotNil(t, list)

	t.Log(list)
	t.Logf("total %d TLDs", len(list))
}
