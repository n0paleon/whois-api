package whoisadapter

import (
	"context"
	"github.com/go-resty/resty/v2"
	"github.com/likexian/whois"
	whoisparser "github.com/likexian/whois-parser"
	"time"
	"whois-api/internal/core/domain"
	"whois-api/internal/core/ports"
	"whois-api/pkg/logger"
)

type Whois struct {
	client *resty.Client
}

var (
	defaultTimeout = 30 * time.Second
	alternativeAPI = "https://check-host.net/ip-info/whois"
)

func NewWhoisAdapter() ports.WhoisAdapter {
	return &Whois{
		client: resty.New(),
	}
}

func (a *Whois) GetWhoisData(query string, ctx context.Context) (*domain.Whois, error) {
	pCtx, pCancel := context.WithTimeout(ctx, 2000*time.Millisecond)
	defer pCancel()

	var parsedResult *domain.Whois
	result, err := a.primaryWhoisCheck(query, pCtx)
	if err != nil {
		result, err = a.secondaryWhoisCheck(query, ctx)
		if err != nil {
			return nil, err
		}

		parsedResult, err = a.parseRawWhois(result)
		if err != nil {
			return nil, err
		}
	} else {
		parsedResult, err = a.parseRawWhois(result)
		if err != nil {
			result, err = a.secondaryWhoisCheck(query, ctx)
			if err != nil {
				return nil, err
			}

			parsedResult, err = a.parseRawWhois(result)
			if err != nil {
				return nil, err
			}
		}
	}

	return parsedResult, nil
}

func (a *Whois) GetRawWhoisData(query string, ctx context.Context) (string, error) {
	client := whois.NewClient()
	deadline, ok := ctx.Deadline()
	if !ok {
		deadline = time.Now().Add(defaultTimeout)
	}

	timeout := time.Until(deadline)
	client.SetTimeout(timeout)
	client.SetDisableStats(true)

	result, err := client.Whois(query)
	if err != nil || len(result) == 0 {
		logger.L().Warn("failed to get raw whois data", err)
		return "", err
	}

	return result, nil
}

func (a *Whois) primaryWhoisCheck(query string, ctx context.Context) ([]byte, error) {
	client := whois.NewClient()
	deadline, ok := ctx.Deadline()
	if !ok {
		deadline = time.Now().Add(defaultTimeout)
	}
	timeout := time.Until(deadline)
	client.SetTimeout(timeout)

	server, _, _ := GetWhoisServer(query)
	result, err := client.Whois(query, server)
	if err != nil {
		return nil, domain.ErrWhoisServerTimeout
	}

	return []byte(result), nil
}

// TODO: add implementation of whois cli
func (a *Whois) secondaryWhoisCheck(query string, ctx context.Context) ([]byte, error) {
	// if runtime.GOOS == "linux" && isWhoisCLIAvailable() {
	//	return a.secondaryWhoisWithCLI(query, ctx)
	// }

	return a.secondaryWhoisWithAPI(query, ctx)
}

func (a *Whois) secondaryWhoisWithAPI(query string, ctx context.Context) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	resp, err := a.client.R().
		SetContext(ctx).
		SetHeader("Content-Type", "multipart/form-data").
		SetMultipartFormData(map[string]string{
			"host": query,
		}).
		Post(alternativeAPI)

	if err != nil {
		return nil, domain.ErrInternalServerError
	}

	if resp.StatusCode() != 200 {
		return nil, domain.ErrDataNotFound
	}

	return resp.Body(), nil
}

func (a *Whois) parseRawWhois(data []byte) (*domain.Whois, error) {
	parsedResult, err := whoisparser.Parse(string(data))
	if err != nil {
		return nil, domain.ErrWhoisParsingError
	}

	whoisDomain := new(domain.Whois)
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

	return whoisDomain, nil
}
