package handler

import (
	"context"
	"github.com/gofiber/fiber/v2"
	"time"
	"whois-api/internal/adapters/api/dto"
	"whois-api/internal/config"
	"whois-api/internal/core/domain"
	"whois-api/internal/core/ports"
	"whois-api/pkg/valid"
)

type APIV1 struct {
	config       *config.Config
	WhoisService ports.WhoisService
}

func NewAPIV1Handler(config *config.Config, WhoisService ports.WhoisService) *APIV1 {
	return &APIV1{
		config:       config,
		WhoisService: WhoisService,
	}
}

func (h *APIV1) Whois(c *fiber.Ctx) error {
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

	result, err := h.WhoisService.CheckWhois(query, ctx)
	if err != nil {
		return c.JSON(domain.APIResponse{
			Error:   true,
			Message: err.Error(),
		})
	}

	return c.JSON(domain.APIResponse{
		Error: false,
		Data:  result,
	})
}
