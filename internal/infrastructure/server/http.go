package server

import (
	"context"
	"fmt"
	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/healthcheck"
	recover2 "github.com/gofiber/fiber/v2/middleware/recover"
	"sync"
	"time"
	"whois-api/internal/config"
	"whois-api/internal/core/domain"
)

type Http struct {
	App    *fiber.App
	config *config.Config
	ctx    context.Context
	wg     *sync.WaitGroup
}

func NewHttpServer(cfg *config.Config) *Http {
	app := fiber.New(fiber.Config{
		AppName:           cfg.App.Name,
		CaseSensitive:     true,
		EnablePrintRoutes: true,
		JSONEncoder:       sonic.Marshal,
		JSONDecoder:       sonic.Unmarshal,
		StrictRouting:     true,
		WriteTimeout:      10 * time.Second,
		Prefork:           cfg.Service.Http.Prefork,
	})

	serv := Http{
		App:    app,
		config: cfg,
		wg:     &sync.WaitGroup{},
	}
	serv.setDefaultMiddlewares()

	return &serv
}

func (s *Http) Start(addr string, ctx context.Context) error {
	// set the http server context
	s.ctx = ctx

	err := s.App.Listen(addr)
	if err != nil {
		return err
	}

	<-ctx.Done()
	return err
}

func (s *Http) Stop() error {
	s.wg.Wait()
	return s.App.Shutdown()
}

func (s *Http) setDefaultMiddlewares() {
	app := s.App

	app.Use(recover2.New(recover2.Config{
		EnableStackTrace: true,
	}))

	app.Use(func(c *fiber.Ctx) error {
		c.Set("X-Powered-By", fmt.Sprintf("%s %s", s.config.App.Name, s.config.App.Version))
		return c.Next()
	})

	app.Use(healthcheck.New(healthcheck.ConfigDefault))

	app.Use(func(c *fiber.Ctx) error {
		select {
		case <-s.ctx.Done():
			return c.Status(fiber.StatusServiceUnavailable).JSON(domain.APIResponse{
				Error:   true,
				Message: domain.ErrServiceMaintenance.Error(),
			})
		default:
			s.wg.Add(1)
			defer s.wg.Done()
			return c.Next()
		}
	})
}
