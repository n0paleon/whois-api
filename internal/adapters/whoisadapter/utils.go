package whoisadapter

import (
	"fmt"
	"github.com/zonedb/zonedb"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
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
