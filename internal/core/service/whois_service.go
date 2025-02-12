package service

import (
	"context"
	"errors"
	"sync"
	"time"
	"whois-api/internal/config"
	"whois-api/internal/core/domain"
	"whois-api/internal/core/ports"
	"whois-api/internal/infrastructure/workers"
	"whois-api/pkg/logger"
	"whois-api/pkg/utils"
)

type Whois struct {
	adapter    ports.WhoisAdapter
	repository ports.WhoisRepository
	config     *config.Config
}

func NewWhoisService(adapter ports.WhoisAdapter, repository ports.WhoisRepository, cfg *config.Config) ports.WhoisService {
	return &Whois{
		adapter:    adapter,
		repository: repository,
		config:     cfg,
	}
}

var (
	minDelay    = 10
	maxDelay    = 100
	cacheMaxAge = 14 * 24 * time.Hour
)

func (w *Whois) RawWhoisLookup(query string, ctx context.Context) (string, error) {
	return w.adapter.GetRawWhoisData(query, ctx)
}

func (w *Whois) SingleLookup(query string, ctx context.Context) (*domain.Whois, error) {
	var result *domain.Whois
	var err error

	if w.config.Service.Redis.CacheEnabled {
		result, err = w.repository.GetWhoisData(query, ctx)
		if err == nil && result != nil {
			return result, nil
		}

		result, err = w.adapter.GetWhoisData(query, ctx)
		if err != nil || result.Domain.CreatedAt == nil {
			if errors.Is(err, domain.ErrWhoisParsingError) {
				return nil, errors.New("failed to parse whois response, please use raw whois instead")
			}
			return nil, domain.ErrDataNotFound
		}

		_ = workers.Pool.Submit(func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			_ = w.repository.SaveWhoisData(query, result, ctx, cacheMaxAge)
			logger.L().Info("Cache data successfully saved for domain: ", query)
		})

		return result, nil
	}

	result, err = w.adapter.GetWhoisData(query, ctx)
	if err != nil || result.Domain.CreatedAt == nil {
		if errors.Is(err, domain.ErrWhoisParsingError) {
			return nil, errors.New("failed to parse whois response, please use raw whois instead")
		}
		return nil, domain.ErrDataNotFound
	}

	return result, nil
}

func (w *Whois) MassLookup(queries []string, ctx context.Context) (map[string]*domain.Whois, error) {
	var wg sync.WaitGroup
	var results sync.Map

	cacheMisses := make([]string, 0)
	cacheHit := make([]string, 0)

	if w.config.Service.Redis.CacheEnabled {
		// Step 1: Check all queries in cache concurrently
		for _, q := range queries {
			wg.Add(1)
			query := q

			_ = workers.Pool.Submit(func() {
				defer wg.Done()

				result, err := w.repository.GetWhoisData(query, ctx)
				if err == nil {
					results.Store(query, result)
					cacheHit = append(cacheHit, query)
				} else {
					cacheMisses = append(cacheMisses, query)
					results.Store(query, nil)
				}
			})
		}
		wg.Wait()
	} else {
		cacheMisses = queries
	}

	// Step 2: Fetch data for cache misses with dynamic rate limits
	for _, q := range cacheMisses {
		wg.Add(1)
		query := q

		_ = workers.Pool.Submit(func() {
			defer wg.Done()

			// Dynamic rate limit for each iteration
			rateLimit := time.Duration(utils.RandomInRange(minDelay, maxDelay)) * time.Millisecond
			time.Sleep(rateLimit) // Wait for random duration before sending the request

			result, err := w.adapter.GetWhoisData(query, ctx)
			if err != nil {
				logger.L().Warnf("Failed to fetch data for domain %s: %v", query, err)
				results.Store(query, nil)
				return
			}

			if w.config.Service.Redis.CacheEnabled {
				err = w.repository.SaveWhoisData(query, result, ctx)
				if err == nil {
					logger.L().Infof("Cache data saved for domain: %s", query)
				} else {
					logger.L().Errorf("Failed to save cache data for domain %s: %v", query, err)
				}
			}

			results.Store(query, result)
		})
	}

	wg.Wait()

	// Step 3: Update old cache entries without waiting
	if w.config.Service.Redis.CacheEnabled {
		_ = workers.Pool.Submit(func() {
			ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second)
			defer cancel()

			for _, q := range cacheHit {
				age, err := w.repository.GetCacheAge(q, ctx)
				if err != nil {
					logger.L().Warnf("Failed to fetch cache age for domain %s: %v", q, err)
					continue
				}

				if age > cacheMaxAge {
					logger.L().Warnf("Cache age for domain %s is too old, requesting new data", q)

					rateLimit := time.Duration(utils.RandomInRange(10, 150)) * time.Millisecond
					<-time.After(rateLimit)

					_ = workers.Pool.Submit(func() {
						ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
						defer cancel()

						data, err := w.adapter.GetWhoisData(q, ctx)
						if err != nil {
							logger.L().Warnf("Failed to fetch data for domain %s: %v", q, err)
							return
						}

						err = w.repository.SaveWhoisData(q, data, ctx)
						if err == nil {
							logger.L().Infof("Cache data updated for domain: %s", q)
						} else {
							logger.L().Errorf("Failed to save updated cache data for domain %s: %v", q, err)
						}
					})
				}
			}
		})
	}

	finalResults := make(map[string]*domain.Whois)
	results.Range(func(key, value interface{}) bool {
		if val, ok := value.(*domain.Whois); ok {
			finalResults[key.(string)] = val
		}
		return true
	})

	return finalResults, nil
}
