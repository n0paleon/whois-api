package dto

import "whois-api/internal/core/domain"

type SingleDomainQuery struct {
	Domain string `json:"domain"`
}

type MassDomainQuery struct {
	Domain []string `json:"domain"`
}

type MassDomainQueryResponse struct {
	Error      bool          `json:"error"`
	Message    string        `json:"message"`
	DomainName string        `json:"domain_name"`
	WhoisData  *domain.Whois `json:"whois"`
}
