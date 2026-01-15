package whoisadapter

import (
	"fmt"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"whois-api/internal/core/domain"

	whoisparser "github.com/likexian/whois-parser"
	"github.com/zonedb/zonedb"
)

var (
	IANA     = "whois.iana.org"
	tldCache sync.Map
	initOnce sync.Once
)

type TLD struct {
	RootTLD          string   `json:"root_tld"`
	SubTLD           []string `json:"sub_tld"`
	RegistryOperator string   `json:"registry_operator"`
	InfoURL          string   `json:"info_url"`
	Tags             []string `json:"tags"`
}

func GetAvailableTLDs() []*TLD {
	initOnce.Do(func() {
		list := zonedb.TLDs

		var counter atomic.Int64
		for _, tld := range list {
			if tld.WhoisServer() != "" && len(tld.NameServers) != 0 {
				data := &TLD{
					RootTLD:          tld.Domain,
					RegistryOperator: tld.RegistryOperator,
					InfoURL:          tld.InfoURL,
				}
				for _, subdomain := range tld.Subdomains {
					data.SubTLD = append(data.SubTLD, subdomain.Domain)
				}
				data.Tags = strings.Split(tld.Tags.String(), " ")

				tldCache.Store(tld.Domain, data)
				counter.Add(1)
			}
		}
	})

	var results []*TLD
	tldCache.Range(func(k, v interface{}) bool {
		results = append(results, v.(*TLD))
		return true
	})

	return results
}

// GetWhoisServer function source: https://github.com/domainr/whois/blob/v0.1.0/whois.go#L22
func GetWhoisServer(query string) (string, string, error) {
	// Queries on TLDs always against IANA
	if strings.Index(query, ".") < 0 {
		return IANA, "", nil
	}
	z := zonedb.PublicZone(query)
	if z == nil {
		return "", "", fmt.Errorf("no public zone found for %s", query)
	}

	// Try whois URL first (these are relatively rare)
	wu := z.WhoisURL()
	if wu != "" {
		u, err := url.Parse(wu)
		if err == nil && u.Host != "" {
			return u.Host, wu, nil
		}
	}

	// Then try host (more common)
	h := z.WhoisServer()
	if h != "" {
		return h, "", nil
	}

	return "", "", fmt.Errorf("no whois server found for %s", query)
}

func parseRawWhois(data []byte) (*domain.Whois, error) {
	parsedResult, err := whoisparser.Parse(string(data))
	if err != nil {
		dataLower := strings.ToLower(string(data))
		if strings.Contains(dataLower, "no entries found") {
			return nil, domain.ErrDataNotFound
		}
		return nil, domain.ErrWhoisParsingError
	}

	whoisDomain := new(domain.Whois)
	if parsedResult.Domain != nil {
		whoisDomain.Domain = &domain.WhoisDomain{
			ID:          parsedResult.Domain.ID,
			Domain:      parsedResult.Domain.Domain,
			Punycode:    parsedResult.Domain.Punycode,
			Name:        parsedResult.Domain.Name,
			Extension:   parsedResult.Domain.Extension,
			WhoisServer: parsedResult.Domain.WhoisServer,
			Status:      parsedResult.Domain.Status,
			NameServers: parsedResult.Domain.NameServers,
			DNSSec:      parsedResult.Domain.DNSSec,
			CreatedAt:   parsedResult.Domain.CreatedDateInTime,
			UpdatedAt:   parsedResult.Domain.UpdatedDateInTime,
			ExpiresAt:   parsedResult.Domain.ExpirationDateInTime,
		}
	}
	if parsedResult.Registrar != nil {
		whoisDomain.Registrar = &domain.WhoisContact{
			ID:           parsedResult.Registrar.ID,
			Name:         parsedResult.Registrar.Name,
			Organization: parsedResult.Registrar.Organization,
			Street:       parsedResult.Registrar.Street,
			City:         parsedResult.Registrar.City,
			Province:     parsedResult.Registrar.Province,
			PostalCode:   parsedResult.Registrar.PostalCode,
			Country:      parsedResult.Registrar.Country,
			Phone:        parsedResult.Registrar.Phone,
			Fax:          parsedResult.Registrar.Fax,
			Email:        parsedResult.Registrar.Email,
			ReferralURL:  parsedResult.Registrar.ReferralURL,
		}
	}
	if parsedResult.Registrant != nil {
		whoisDomain.Registrant = &domain.WhoisContact{
			ID:           parsedResult.Registrant.ID,
			Name:         parsedResult.Registrant.Name,
			Organization: parsedResult.Registrant.Organization,
			Street:       parsedResult.Registrant.Street,
			City:         parsedResult.Registrant.City,
			Province:     parsedResult.Registrant.Province,
			PostalCode:   parsedResult.Registrant.PostalCode,
			Country:      parsedResult.Registrant.Country,
			Phone:        parsedResult.Registrant.Phone,
			Fax:          parsedResult.Registrant.Fax,
			Email:        parsedResult.Registrant.Email,
			ReferralURL:  parsedResult.Registrant.ReferralURL,
		}
	}

	return whoisDomain, nil
}
