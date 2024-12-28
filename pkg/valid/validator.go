package valid

import (
	"fmt"
	"github.com/asaskevich/govalidator"
	"strings"
)

func ParseRootDomain(domain string) (string, error) {
	parts := strings.Split(domain, ".")
	if len(parts) < 2 {
		return "", fmt.Errorf("invalid domain format: %s", domain)
	}

	// Handle kasus khusus untuk domain dengan format country code (co.uk, com.au, etc)
	knownCCTLDs := map[string]bool{
		"co.uk":  true,
		"com.au": true,
		"co.jp":  true,
		"co.id":  true,
		"com.br": true,
		// Tambahkan country code TLD lainnya sesuai kebutuhan
	}

	// Cek apakah 3 bagian terakhir membentuk known CCTLD
	if len(parts) >= 3 {
		possibleCCTLD := strings.Join(parts[len(parts)-2:], ".")
		if knownCCTLDs[possibleCCTLD] {
			if len(parts) == 3 {
				return domain, nil
			}
			return strings.Join(parts[len(parts)-3:], "."), nil
		}
	}

	// Untuk kasus normal (example.com, akamai.net)
	// Ambil 2 bagian terakhir
	return strings.Join(parts[len(parts)-2:], "."), nil
}

func ValidateDomain(domain string) bool {
	return govalidator.IsDNSName(domain)
}
