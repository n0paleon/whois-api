package workers

import (
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"testing"
	"whois-api/internal/config"
	"whois-api/pkg/logger"
)

func TestWorkerPool(t *testing.T) {
	dir, err := os.Getwd()
	assert.NoError(t, err)

	filename := filepath.Join(dir, "../../../config.yaml")
	cfg, err := config.LoadConfig(filename)
	assert.NoError(t, err)
	assert.NotNil(t, cfg)

	logrs, err := logger.NewLogger(cfg)
	assert.NoError(t, err)
	assert.NotNil(t, logrs)

	size := 100

	err = InitWorkerPool(size)
	assert.NoError(t, err)

	submitTask := Pool.Submit(func() {
		t.Log("workerpool initialized successfully")
	})
	assert.NoError(t, submitTask)
}
