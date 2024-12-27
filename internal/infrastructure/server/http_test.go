package server

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"whois-api/internal/config"
)

func TestApiHttpServer(t *testing.T) {
	cfgFile, err := filepath.Abs("../../../config.yaml")
	assert.NoError(t, err)
	assert.NotEmpty(t, cfgFile)
	assert.FileExists(t, cfgFile)

	cfg, err := config.LoadConfig(cfgFile)
	assert.NoError(t, err)
	assert.NotNil(t, cfg)

	var server *Http
	t.Run("Init Fiber HTTP server", func(t *testing.T) {
		server = NewHttpServer(cfg)
		assert.NotNil(t, server)
	})

	t.Run("Start HTTP server", func(t *testing.T) {
		go func() {
			err = server.Start(fmt.Sprintf("%s:%d", cfg.Service.Http.Host, cfg.Service.Http.Port), context.Background())
			assert.NoError(t, err)
		}()
	})

	t.Run("Check HTTP health check endpoint", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/livez", nil)
		assert.NoError(t, err)
		resp, err := server.App.Test(req)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("Shutdown HTTP server", func(t *testing.T) {
		err = server.Stop()
		assert.NoError(t, err)
	})
}

func BenchmarkApiHttpServer(b *testing.B) {
	cfgFile, err := filepath.Abs("../../../config.yaml")
	assert.NoError(b, err)
	assert.NotEmpty(b, cfgFile)
	assert.FileExists(b, cfgFile)

	cfg, err := config.LoadConfig(cfgFile)
	assert.NoError(b, err)
	assert.NotNil(b, cfg)

	server := NewHttpServer(cfg)
	assert.NotNil(b, server)

	b.Run("Benchmark health check endpoint", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			req := httptest.NewRequest("GET", "/livez", nil)
			resp, err := server.App.Test(req)
			assert.NoError(b, err)
			assert.NotNil(b, resp)
			assert.Equal(b, http.StatusOK, resp.StatusCode)
		}
	})
}
