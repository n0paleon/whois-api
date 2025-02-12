package api

import (
	"github.com/gofiber/fiber/v2"
	"whois-api/internal/adapters/api/handler"
	"whois-api/internal/config"
	"whois-api/internal/infrastructure/server"
)

type Router struct {
	app        *fiber.App
	config     *config.Config
	apiHandler *handler.APIV1
}

func NewRouter(server *server.Http, cfg *config.Config, apiHandler *handler.APIV1) *Router {
	return &Router{
		app:        server.App,
		config:     cfg,
		apiHandler: apiHandler,
	}
}

func (r *Router) SetupRoutes() {
	route := r.app.Group("/v1/")

	route.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})
	route.Post("/whois/lookup", r.apiHandler.SingleWhoisCheck)
	route.Post("/whois/lookup/raw", r.apiHandler.RawWhoisCheck)
	route.Post("/whois/mass-lookup", r.apiHandler.MassWhoisCheck)

	route.Get("/available-tlds", r.apiHandler.GetTLDList)
}
