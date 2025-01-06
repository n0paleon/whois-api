package whoisadapter

import (
	"bytes"
	"context"
	"errors"
	whois2 "github.com/domainr/whois"
	"github.com/likexian/whois"
	whoisparser "github.com/likexian/whois-parser"
	"io"
	"mime/multipart"
	"net/http"
	"os/exec"
	"runtime"
	"sync"
	"time"
	"whois-api/internal/core/domain"
	"whois-api/internal/core/ports"
)

type Whois struct {
}

var (
	defaultTimeout = 30 * time.Second
	alternativeAPI = "https://check-host.net/ip-info/whois"
)

func NewWhoisAdapter() ports.WhoisAdapter {
	return &Whois{}
}

func (a *Whois) GetWhoisData(query string, ctx context.Context) (*domain.Whois, error) {
	pCtx, pCancel := context.WithTimeout(ctx, 3*time.Second)
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

func (a *Whois) primaryWhoisCheck(query string, ctx context.Context) ([]byte, error) {
	var (
		wg       sync.WaitGroup
		rawWhois []byte
		errChan  = make(chan error, 1)
	)
	wg.Add(1)
	go func() {
		defer wg.Done()

		whoisServer, _, err := whois2.Server(query)
		if err != nil {
			errChan <- err
			return
		}

		client := whois.NewClient()
		result, err := client.Whois(query, whoisServer)
		if err != nil {
			errChan <- err
			return
		}

		if len(result) > 0 {
			rawWhois = []byte(result)
			return
		}
	}()

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		select {
		case err := <-errChan:
			return nil, err
		default:
			return rawWhois, nil
		}
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (a *Whois) secondaryWhoisCheck(query string, ctx context.Context) ([]byte, error) {
	if runtime.GOOS == "linux" && isWhoisCLIAvailable() {
		return a.secondaryWhoisWithCLI(query, ctx)
	}
	return a.secondaryWhoisWithAPI(query, ctx)
}

func (a *Whois) secondaryWhoisWithAPI(query string, ctx context.Context) ([]byte, error) {
	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)

	_ = writer.WriteField("host", query)
	if err := writer.Close(); err != nil {
		return nil, err
	}

	client := &http.Client{Timeout: defaultTimeout}
	req, err := http.NewRequest("POST", alternativeAPI, payload)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func (a *Whois) secondaryWhoisWithCLI(query string, ctx context.Context) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "whois", query)
	output, err := cmd.Output()
	if err != nil {
		return nil, errors.New("failed to execute whois-cli")
	}
	return output, nil
}

func isWhoisCLIAvailable() bool {
	_, err := exec.LookPath("whois")
	return err == nil
}

func (a *Whois) parseRawWhois(data []byte) (*domain.Whois, error) {
	parsedResult, err := whoisparser.Parse(string(data))
	if err != nil {
		return nil, err
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
