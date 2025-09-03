package dnsadapter

import "github.com/miekg/dns"

var recordTypes = []uint16{
	dns.TypeA,
	dns.TypeAAAA,
	dns.TypeMX,
	dns.TypeNS,
	dns.TypeTXT,
	dns.TypeCNAME,
	dns.TypeSOA,
	dns.TypeSRV,
}

type DnsRecord map[string]any
type DNSRecords []DnsRecord

func parseRR(rr dns.RR) DnsRecord {
	var result DnsRecord
	switch v := rr.(type) {
	case *dns.A:
		result = map[string]any{
			"type": dns.TypeToString[dns.TypeA],
			"ip":   v.A.String(),
			"ttl":  v.Hdr.Ttl,
		}
	case *dns.AAAA:
		result = map[string]any{
			"type": dns.TypeToString[dns.TypeAAAA],
			"ip":   v.AAAA.String(),
			"ttl":  v.Hdr.Ttl,
		}
	case *dns.MX:
		result = map[string]any{
			"type":       dns.TypeToString[dns.TypeMX],
			"host":       v.Mx,
			"preference": v.Preference,
			"ttl":        v.Hdr.Ttl,
		}
	case *dns.NS:
		result = map[string]any{
			"type": dns.TypeToString[dns.TypeNS],
			"host": v.Ns,
			"ttl":  v.Hdr.Ttl,
		}
	case *dns.TXT:
		result = map[string]any{
			"type": dns.TypeToString[dns.TypeTXT],
			"txt":  v.Txt,
			"ttl":  v.Hdr.Ttl,
		}
	case *dns.CNAME:
		result = map[string]any{
			"type":   dns.TypeToString[dns.TypeCNAME],
			"target": v.Target,
			"ttl":    v.Hdr.Ttl,
		}
	case *dns.SOA:
		result = map[string]any{
			"type":    dns.TypeToString[dns.TypeSOA],
			"ns":      v.Ns,
			"mbox":    v.Mbox,
			"serial":  v.Serial,
			"refresh": v.Refresh,
			"retry":   v.Retry,
			"expire":  v.Expire,
			"min_ttl": v.Minttl,
			"ttl":     v.Hdr.Ttl,
		}
	case *dns.SRV:
		result = map[string]any{
			"type":     dns.TypeToString[dns.TypeSRV],
			"target":   v.Target,
			"port":     v.Port,
			"priority": v.Priority,
			"weight":   v.Weight,
			"ttl":      v.Hdr.Ttl,
		}
	default:
		result = map[string]any{
			"value": v.String(),
		}
	}

	return result
}
