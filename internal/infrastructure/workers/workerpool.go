package workers

import (
	"github.com/panjf2000/ants/v2"
	"whois-api/pkg/logger"
)

var Pool *ants.Pool

func InitWorkerPool(size int) error {
	if size <= 0 {
		size = 100
		logger.L().Warn("invalid pool size, using 100 as default workerpool size")
	}

	pool, err := ants.NewPool(size, ants.WithPreAlloc(true), ants.WithNonblocking(true))
	if err != nil {
		logger.L().Error("failed to initialize worker pool", "error", err)
		return err
	}

	Pool = pool

	return nil
}

func CloseWorkerPool() {
	if Pool != nil {
		Pool.Release()
	}
}
