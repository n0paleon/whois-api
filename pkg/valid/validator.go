package valid

import (
	"github.com/bombsimon/tld-validator"
	"github.com/weppos/publicsuffix-go/publicsuffix"
	"net/url"
	"strings"
	"unicode"
	"whois-api/internal/core/domain"
)

func ParseRootDomain(q string) (string, error) {
	q = strings.ToLower(strings.TrimSpace(q))
	rootDomain, err := publicsuffix.Domain(q)
	if err == nil {
		if !tld.FromDomainName(rootDomain).IsValid() {
			return "", domain.ErrUnsupportedTLD
		}
		return rootDomain, nil
	}

	parsedURL, err := url.Parse("https://" + q)
	if err != nil {
		return "", domain.ErrInvalidDomainName
	}

	hostParts := strings.Split(parsedURL.Hostname(), ".")
	if len(hostParts) < 2 {
		return "", domain.ErrInvalidDomainName
	}

	rootDomain = strings.Join(hostParts[len(hostParts)-2:], ".")
	if !tld.FromDomainName(rootDomain).IsValid() {
		return "", domain.ErrUnsupportedTLD
	}

	return rootDomain, nil
}

func IsValidDomain(name string) bool {
	if len(name) == 0 || len(name) > 63 {
		return false
	}

	for i, c := range name {
		if !(unicode.IsLetter(c) || unicode.IsDigit(c) || c == '-') {
			return false // Jika ada karakter ilegal
		}
		if (i == 0 || i == len(name)-1) && c == '-' {
			return false // Tidak boleh diawali atau diakhiri dengan "-"
		}
	}
	return true
}
