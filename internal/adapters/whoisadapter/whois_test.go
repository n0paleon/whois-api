package whoisadapter

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
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
