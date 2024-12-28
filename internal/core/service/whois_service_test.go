package service

import (
	"context"
	"github.com/stretchr/testify/assert"
	"path/filepath"
	"testing"
	"whois-api/internal/adapters/repository"
	"whois-api/internal/adapters/whoisadapter"
	"whois-api/internal/config"
	"whois-api/internal/infrastructure/database"
	"whois-api/internal/infrastructure/workers"
	"whois-api/pkg/logger"
)

func TestWhoisService(t *testing.T) {
	cfgFile, err := filepath.Abs("../../../config.yaml")
	assert.Nil(t, err)
	assert.FileExists(t, cfgFile, "config file not exists")

	cfg, err := config.LoadConfig(cfgFile)
	assert.Nil(t, err)
	_, err = logger.NewLogger(cfg)
	assert.Nil(t, err)

	adapter := whoisadapter.NewWhoisAdapter()
	redisConn, err := database.NewRedisConn(cfg)
	assert.Nil(t, err)
	assert.Nil(t, workers.InitWorkerPool(100))
	repo := repository.NewWhoisRepository(redisConn)
	whois := NewWhoisService(adapter, repo)

	result, err := whois.CheckWhois("google.com", context.Background())
	assert.Nil(t, err)
	assert.Equal(t, "google.com", result.Domain.Domain)
}

func BenchmarkWhoisService(b *testing.B) {
	cfgFile, err := filepath.Abs("../../../config.yaml")
	assert.Nil(b, err)
	assert.FileExists(b, cfgFile, "config file not exists")

	cfg, err := config.LoadConfig(cfgFile)
	assert.Nil(b, err)
	_, err = logger.NewLogger(cfg)
	assert.Nil(b, err)

	adapter := whoisadapter.NewWhoisAdapter()
	redisConn, err := database.NewRedisConn(cfg)
	assert.Nil(b, err)
	assert.Nil(b, workers.InitWorkerPool(100))
	repo := repository.NewWhoisRepository(redisConn)
	whois := NewWhoisService(adapter, repo)

	for i := 0; i < b.N; i++ {
		result, err := whois.CheckWhois("google.com", context.Background())
		assert.Nil(b, err)
		assert.Equal(b, "google.com", result.Domain.Domain)
	}
}
