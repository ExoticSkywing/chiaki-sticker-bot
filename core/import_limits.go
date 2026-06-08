package core

import (
	"context"
	"os"
	"strconv"
	"sync"
)

var (
	importSemaphore     chan struct{}
	importSemaphoreOnce sync.Once
)

func acquireImportSlot(ctx context.Context) (func(), error) {
	importSemaphoreOnce.Do(func() {
		concurrency := 1
		if value, err := strconv.Atoi(os.Getenv("MSB_IMPORT_CONCURRENCY")); err == nil && value > 0 {
			concurrency = value
		}
		importSemaphore = make(chan struct{}, concurrency)
	})

	select {
	case importSemaphore <- struct{}{}:
		return func() {
			<-importSemaphore
		}, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
