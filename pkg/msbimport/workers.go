package msbimport

import (
	"context"
	"os"
	"strconv"

	"github.com/panjf2000/ants/v2"
	log "github.com/sirupsen/logrus"
)

// Workers pool for converting webm
var wpConvertWebm, _ = ants.NewPoolWithFunc(webmWorkerConcurrency(), wConvertWebm)

func webmWorkerConcurrency() int {
	concurrency := 1
	if value, err := strconv.Atoi(os.Getenv("MSB_WEBM_WORKER_CONCURRENCY")); err == nil && value > 0 {
		concurrency = value
	}
	return concurrency
}

// Accepts *LineFile
func wConvertWebm(i interface{}) {
	lf := i.(*LineFile)
	defer lf.Wg.Done()
	log.Debugln("Converting in pool for:", lf)

	var err error
	ctx := lf.Ctx
	if ctx == nil {
		ctx = context.Background()
	}
	select {
	case <-ctx.Done():
		lf.CError = ctx.Err()
		return
	default:
	}
	lf.ConvertedFile, err = FFToWebmTGVideoContextWithStatus(ctx, lf.OriginalFile, lf.ConvertToEmoji, lf.Status)
	if err != nil {
		lf.CError = err
	}
	log.Debugln("convert OK: ", lf.ConvertedFile)
}
