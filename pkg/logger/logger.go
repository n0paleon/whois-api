package logger

import (
	"github.com/sirupsen/logrus"
	"log"
	"os"
	"whois-api/internal/config"
)

var logCache *logrus.Logger

func NewLogger(cfg *config.Config) (*logrus.Logger, error) {
	logger := logrus.New()

	level, err := logrus.ParseLevel(cfg.Logger.Level)
	if err != nil {
		log.Fatalf("error parsing level: %v", err)
		return nil, err
	}
	logger.SetLevel(level)

	if cfg.Logger.Format == "json" {
		logger.SetFormatter(&logrus.JSONFormatter{
			PrettyPrint: cfg.Logger.PrettyPrint,
		})
	} else {
		logger.SetFormatter(&logrus.TextFormatter{
			ForceColors: true,
		})
	}

	logger.SetReportCaller(cfg.Logger.ReportCaller)
	logger.SetOutput(os.Stdout)

	logCache = logger
	return logCache, nil
}

func L() *logrus.Logger {
	return logCache
}
