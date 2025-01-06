package handler

import (
	"context"
	"github.com/gofiber/fiber/v2"
	"time"
	"whois-api/internal/adapters/api/dto"
	"whois-api/internal/config"
	"whois-api/internal/core/domain"
	"whois-api/internal/core/ports"
	"whois-api/pkg/utils"
	"whois-api/pkg/valid"
)

type APIV1 struct {
	config       *config.Config
	WhoisService ports.WhoisService
}

func NewAPIV1Handler(config *config.Config, whoisService ports.WhoisService) *APIV1 {
	return &APIV1{
		config:       config,
		WhoisService: whoisService,
	}
}

func (h *APIV1) SingleWhoisCheck(c *fiber.Ctx) error {
	var payload *dto.SingleDomainQuery

	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.APIResponse{
			Error:   true,
			Message: "domain is required",
		})
	}

	query, err := valid.ParseRootDomain(payload.Domain)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.APIResponse{
			Error:   true,
			Message: err.Error(),
		})
	}

	ctx, cancel := context.WithTimeout(c.UserContext(), 30*time.Second)
	defer cancel()

	result, err := h.WhoisService.SingleLookup(query, ctx)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(domain.APIResponse{
			Error:   true,
			Message: err.Error(),
		})
	}

	return c.JSON(domain.APIResponse{
		Error: false,
		Data:  result,
	})
}

func (h *APIV1) MassWhoisCheck(c *fiber.Ctx) error {
	var payload *dto.MassDomainQuery

	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.APIResponse{
			Error:   true,
			Message: "Invalid request body",
		})
	}

	if len(payload.Domain) > 500 {
		return c.Status(fiber.StatusBadRequest).JSON(domain.APIResponse{
			Error:   true,
			Message: "Too many domains. Maximum allowed is 500",
		})
	}

	ctx, cancel := context.WithTimeout(c.UserContext(), 150*time.Second)
	defer cancel()

	var (
		response    []*dto.MassDomainQueryResponse
		domainNames []string
	)
	for _, d := range payload.Domain {
		query, err := valid.ParseRootDomain(d)
		if err != nil {
			response = append(response, &dto.MassDomainQueryResponse{
				Error:      true,
				Message:    err.Error(),
				DomainName: d,
			})
		} else {
			response = append(response, &dto.MassDomainQueryResponse{
				Error:      false,
				DomainName: query,
			})
			domainNames = append(domainNames, query)
		}
	}

	results, err := h.WhoisService.MassLookup(domainNames, ctx, time.Duration(utils.RandomInRange(100, 250))*time.Millisecond)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.APIResponse{
			Error:   true,
			Message: err.Error(),
		})
	}

	for _, res := range response {
		if res.Error {
			continue
		}

		if whoisData, found := results[res.DomainName]; found && whoisData != nil {
			res.WhoisData = *whoisData
		} else {
			res.Error = true
			res.Message = "data not found"
		}
	}

	return c.JSON(domain.APIResponse{
		Error: false,
		Data:  response,
	})
}
