package valid

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	validDomain   = "turulabs.com"
	subDomain     = "www.turulabs.com"
	invalidDomain = "turulabs"
)

func TestParseRootDomain(t *testing.T) {
	t.Run("Test ParseRootDomain with valid domain", func(t *testing.T) {
		result, err := ParseRootDomain(validDomain)
		assert.Nil(t, err)
		assert.Equal(t, validDomain, result)
	})
	t.Run("Test ParseRootDomain with invalid domain", func(t *testing.T) {
		result, err := ParseRootDomain(invalidDomain)
		assert.NotNil(t, err)
		assert.Equal(t, "", result)
	})
	t.Run("Test ParseRootDomain with valid sub domain", func(t *testing.T) {
		result, err := ParseRootDomain(subDomain)
		assert.Nil(t, err)
		assert.Equal(t, validDomain, result)
	})
}
