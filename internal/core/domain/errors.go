package domain

import "errors"

var (
	ErrServiceMaintenance = errors.New("service is under maintenance")
	ErrDataNotFound       = errors.New("data not found")
)
