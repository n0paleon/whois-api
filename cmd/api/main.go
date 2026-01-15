package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
	"time"
	"whois-api/internal/adapters/api"
	"whois-api/internal/adapters/api/handler"
	"whois-api/internal/adapters/dnsadapter"
	"whois-api/internal/adapters/repository"
	"whois-api/internal/adapters/whoisadapter"
	"whois-api/internal/config"
	"whois-api/internal/core/ports"
	"whois-api/internal/core/service"
	"whois-api/internal/infrastructure/database"
	"whois-api/internal/infrastructure/server"
	"whois-api/internal/infrastructure/workers"
	"whois-api/pkg/logger"

	"github.com/sirupsen/logrus"
	"go.uber.org/fx"
)

func main() {
	// load config
	cfgFile, err := filepath.Abs("./config.yaml")
	if err != nil {
		log.Fatal(err)
	}
	cfg, err := config.LoadConfig(cfgFile)
	if err != nil {
		log.Fatal(err)
	}

	// init logger
	logrs, err := logger.NewLogger(cfg)
	if err != nil {
		log.Fatal(err)
	}

	// init worker pool
	if err := workers.InitWorkerPool(cfg.Workerpool.Size); err != nil {
		logrs.Fatal(err)
	}

	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// check if http debug monitoring is enabled
	if cfg.App.Debug {
		_ = workers.Pool.Submit(func() {
			server.NewHttpDebugging()
		})
	}

	rootCtx, rootCtxCancel := context.WithCancel(context.Background())
	defer rootCtxCancel()

	app := fx.New(
		fx.Provide(func() *config.Config { return cfg }),
		fx.Provide(func() context.Context { return rootCtx }),
		fx.Provide(func() *logrus.Logger { return logrs }),
		fx.Provide(database.NewRedisConn),
		fx.Provide(repository.NewWhoisRepository),
		fx.Provide(func() *dnsadapter.DNS {
			return dnsadapter.NewDNS(cfg.Service.DNSResolver)
		}),
		// provide handler
		fx.Provide(
			handler.NewAPIV1Handler,
			fx.Annotate(
				service.NewWhoisService,
				fx.As(new(ports.WhoisService)),
			),
			fx.Annotate(
				whoisadapter.NewWhoisAdapter,
				fx.As(new(ports.WhoisAdapter)),
			),
		),
		fx.Provide(server.NewHttpServer),
		fx.Provide(api.NewRouter),
		fx.Invoke(func(r *api.Router) {
			r.SetupRoutes()
		}),
		fx.Invoke(func(lc fx.Lifecycle, s *server.Http) {
			lc.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					return workers.Pool.Submit(func() {
						_ = s.Start(net.JoinHostPort(cfg.Service.Http.Host, strconv.Itoa(cfg.Service.Http.Port)), ctx)
					})
				},
				OnStop: func(ctx context.Context) error {
					rootCtxCancel()
					workers.CloseWorkerPool()
					logger.L().Info("received shutdown signal, waiting all process completed before shutdown")
					return s.Stop()
				},
			})
		}),
	)

	if err := app.Start(rootCtx); err != nil {
		logrs.Fatal(err)
	}

	// Wait for shutdown signal
	<-sigChan

	// Give shutdown timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := app.Stop(shutdownCtx); err != nil {
		logrs.Errorf("error during shutdown: %v", err)
	}
}
