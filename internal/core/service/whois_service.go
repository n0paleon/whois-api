package service

import (
	"context"
	"fmt"
	"sync"
	"time"
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

func (w *Whois) SingleLookup(query string, ctx context.Context) (*domain.Whois, error) {
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
				logger.L().Warnf("Failed to update cache data for domain %s with error: %v", query, err)
				return
			}
			if err := w.repository.SaveWhoisData(query, freshData, context.Background()); err == nil {
				logger.L().Info("Cache data updated for domain: ", query)
			}
		})
	}

	return result, nil
}

func (w *Whois) MassLookup(queries []string, ctx context.Context, rateLimit time.Duration) (map[string]*domain.Whois, error) {
	var (
		wg          sync.WaitGroup
		mutex       sync.Mutex
		results     = make(map[string]*domain.Whois)
		rateLimiter = time.Tick(rateLimit)
	)

	// Step 1: Check all queries in cache concurrently
	cacheMisses := make([]string, 0)
	cacheHit := make([]string, 0)

	for _, q := range queries {
		wg.Add(1)
		query := q

		_ = workers.Pool.Submit(func() {
			defer wg.Done()

			result, err := w.repository.GetWhoisData(query, ctx)
			mutex.Lock()
			if err == nil {
				results[query] = result
				cacheHit = append(cacheHit, query)
			} else {
				cacheMisses = append(cacheMisses, query)
			}
			mutex.Unlock()
		})
	}

	wg.Wait()

	// Step 2: Fetch data for cache misses
	for _, q := range cacheMisses {
		wg.Add(1)
		query := q

		_ = workers.Pool.Submit(func() {
			defer wg.Done()

			<-rateLimiter

			result, err := w.adapter.GetWhoisData(query, ctx)
			if err != nil {
				logger.L().Warnf("Failed to fetch data for domain %s: %v", query, err)
				return
			}

			fmt.Println(result.Domain.NameServers)

			err = w.repository.SaveWhoisData(query, result, ctx)
			if err == nil {
				logger.L().Infof("Cache data saved for domain: %s", query)
			} else {
				logger.L().Errorf("Failed to save cache data for domain %s: %v", query, err)
			}

			mutex.Lock()
			results[query] = result
			mutex.Unlock()
		})
	}

	wg.Wait()

	// Step 3: Update old cache entries without waiting
	_ = workers.Pool.Submit(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second)
		defer cancel()

		for _, q := range cacheHit {
			age, err := w.repository.GetCacheAge(q, ctx)
			if err != nil {
				logger.L().Warnf("Failed to fetch cache age for domain %s: %v", q, err)
				continue
			}

			if age > 1*time.Minute {
				logger.L().Warnf("Cache age for domain %s is too old, requesting new data", q)

				<-rateLimiter
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

	return results, nil
}
