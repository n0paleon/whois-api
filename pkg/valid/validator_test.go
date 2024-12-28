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

func TestValidateDomain(t *testing.T) {
	t.Run("Test ValidateDomain with valid domain", func(t *testing.T) {
		valid := ValidateDomain(validDomain)
		assert.True(t, valid)
	})
	t.Run("Test ValidateDomain with invalid domain", func(t *testing.T) {
		invalid := ValidateDomain(invalidDomain)
		assert.False(t, invalid)
	})
	t.Run("Test ValidateDomain with valid sub domain", func(t *testing.T) {
		valid := ValidateDomain(subDomain)
		assert.True(t, valid)
	})
}

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
