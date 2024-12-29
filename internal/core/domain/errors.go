package domain

import "errors"

var (
	ErrServiceMaintenance  = errors.New("service is under maintenance")
	ErrDataNotFound        = errors.New("data not found")
	ErrInvalidDomainName   = errors.New("invalid domain name")
	ErrUnsupportedTLD      = errors.New("unsupported TLD")
	ErrInternalServerError = errors.New("internal server error")
)
