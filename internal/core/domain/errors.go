package domain

import "errors"

var (
	ErrServiceMaintenance  = errors.New("service is under maintenance")
	ErrDataNotFound        = errors.New("data not found")
	ErrInvalidDomainName   = errors.New("invalid domain name")
	ErrUnsupportedTLD      = errors.New("unsupported TLD")
	ErrWhoisServerTimeout  = errors.New("whois server timeout")
	ErrInternalServerError = errors.New("internal server error")
	ErrWhoisParsingError   = errors.New("whois parsing error")
)
