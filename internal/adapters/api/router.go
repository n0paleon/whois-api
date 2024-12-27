package api

import (
	"github.com/gofiber/fiber/v2"
	"whois-api/internal/config"
	"whois-api/internal/infrastructure/server"
)

type Router struct {
	app    *fiber.App
	config *config.Config
}

func NewRouter(server *server.Http, cfg *config.Config) *Router {
	return &Router{
		app:    server.App,
		config: cfg,
	}
}

func (r *Router) SetupRoutes() {
	route := r.app.Group("/")

	route.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})
}
