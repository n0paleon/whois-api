package service

import (
	"context"
	whois2 "github.com/domainr/whois"
	"github.com/likexian/whois"
	whoisparser "github.com/likexian/whois-parser"
	"sync"
	"time"
	"whois-api/internal/core/domain"
	"whois-api/internal/core/ports"
	"whois-api/internal/infrastructure/workers"
	"whois-api/pkg/logger"
)

type Whois struct {
}

func NewWhoisService() ports.WhoisService {
	return &Whois{}
}

func (w *Whois) Whois(query string, ctx context.Context) (*domain.Whois, error) {
	var (
		wg          sync.WaitGroup
		whoisDomain = new(domain.Whois)
		errMsg      = make(chan error, 1)
	)

	wg.Add(1)
	_ = workers.Pool.Submit(func() {
		defer wg.Done()

		server, _, err := whois2.Server(query)
		if err != nil {
			logger.L().Warn(err)
			errMsg <- err
			return
		}

		client := whois.NewClient()
		client.SetTimeout(5 * time.Second)
		client.SetDisableReferral(false)
		result, err := client.Whois(query, server)
		if err != nil {
			logger.L().Warnf("domain %s error: %v", query, err)
			errMsg <- domain.ErrDataNotFound
			return
		}

		parsedResult, err := whoisparser.Parse(result)
		if err != nil {
			logger.L().Errorf("failed to parse whois data: %+v", err)
			errMsg <- domain.ErrDataNotFound
			return
		}

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

		return
	})

	done := make(chan struct{})
	_ = workers.Pool.Submit(func() {
		wg.Wait()
		close(done)
	})

	select {
	case <-done:
		select {
		case err := <-errMsg:
			return nil, err
		default:
			return whoisDomain, nil
		}
	case <-ctx.Done():
		logger.L().Warnf("context deadline exceed for whois service!")
		return nil, ctx.Err()
	}
}
