package service

import (
	"context"
	"whois-api/internal/core/domain"
	"whois-api/internal/core/ports"
	"whois-api/internal/infrastructure/workers"
	"whois-api/pkg/logger"
)

type Whois struct {
	adapter    ports.WhoisAdapter
	repository ports.WhoisRepository
}

func NewWhoisService(adapter ports.WhoisAdapter, repository ports.WhoisRepository) ports.WhoisService {
	return &Whois{
		adapter:    adapter,
		repository: repository,
	}
}

func (w *Whois) CheckWhois(query string, ctx context.Context) (*domain.Whois, error) {
	result, err := w.repository.GetWhoisData(query, ctx)
	if err != nil {
		result, err = w.adapter.GetWhoisData(query, ctx)
		if err != nil {
			return nil, domain.ErrDataNotFound
		}

		_ = workers.Pool.Submit(func() {
			if err := w.repository.SaveWhoisData(query, result, context.Background()); err == nil {
				logger.L().Info("Cache data successfully saved for domain: ", query)
			}
		})
	} else {
		_ = workers.Pool.Submit(func() {
			freshData, err := w.adapter.GetWhoisData(query, ctx)
			if err != nil {
				return
			}
			if err := w.repository.SaveWhoisData(query, freshData, context.Background()); err == nil {
				logger.L().Info("Cache data updated for domain: ", query)
			}
		})
	}

	return result, nil
}
