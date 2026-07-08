package msbimport

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
)

// retryCountHook counts conversion log lines that indicate a wasted encode
// (an oversized output that forced a re-encode, or an encode that timed out).
type retryCountHook struct {
	mu        sync.Mutex
	tooLarge  int
	timedOut  int
	fallbacks int
}

func (h *retryCountHook) Levels() []log.Level { return log.AllLevels }

func (h *retryCountHook) Fire(e *log.Entry) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	switch {
	case strings.Contains(e.Message, "output too large"):
		h.tooLarge++
	case strings.Contains(e.Message, "conversion timed out"):
		h.timedOut++
	case strings.Contains(e.Message, "retrying with two-pass frame sequence fallback"):
		h.fallbacks++
	}
	return nil
}

// TestKakaoPackRealConversion runs the real Kakao animated-WebP → WebM pipeline
// over a directory of downloaded stickers and reports, per sticker, how many
// wasted encodes happened and how long conversion took. Set MSB_KAKAO_TEST_DIR
// to the folder holding s*.webp files.
//
//	go test ./pkg/msbimport/ -run TestKakaoPackRealConversion -v -timeout 30m
func TestKakaoPackRealConversion(t *testing.T) {
	dir := os.Getenv("MSB_KAKAO_TEST_DIR")
	if dir == "" {
		t.Skip("set MSB_KAKAO_TEST_DIR to run the real Kakao conversion benchmark")
	}
	InitConvert()

	files, err := filepath.Glob(filepath.Join(dir, "s*.webp"))
	if err != nil || len(files) == 0 {
		t.Fatalf("no s*.webp files in %s", dir)
	}
	sort.Strings(files)

	t.Logf("files: %d", len(files))
	t.Logf("%-10s %-8s %-8s %-9s %-8s %s", "file", "srcKB", "outKB", "encodes", "time", "result")

	var totalRetries, totalTimeouts, failures int
	var totalDur time.Duration
	for _, f := range files {
		// Work on a copy so a rerun starts clean.
		hook := &retryCountHook{}
		log.AddHook(hook)

		status := NewConversionStatus()
		start := time.Now()
		ctx, cancel := context.WithTimeout(context.Background(), 6*time.Minute)
		out, cerr := KakaoAnimatedWebpToWebmContext(ctx, f, status)
		elapsed := time.Since(start)
		cancel()

		// Drop the hook for the next iteration.
		std := log.StandardLogger()
		std.ReplaceHooks(log.LevelHooks{})

		srcKB := fileKB(f)
		outKB := int64(0)
		result := "OK"
		if cerr != nil {
			result = "FAIL: " + cerr.Error()
			failures++
		} else if st, e := os.Stat(out); e == nil {
			outKB = st.Size() / 1024
		}
		// encodes ≈ wasted re-encodes + timeouts + fallbacks + 1 final encode.
		encodes := hook.tooLarge + hook.timedOut + hook.fallbacks + 1
		totalRetries += hook.tooLarge
		totalTimeouts += hook.timedOut
		totalDur += elapsed
		t.Logf("%-10s %-8d %-8d %-9d %-8s %s",
			filepath.Base(f), srcKB, outKB, encodes, elapsed.Round(time.Millisecond), result)
		os.Remove(out)
	}

	t.Logf("---")
	t.Logf("TOTAL: %s, oversize-retries=%d, timeouts=%d, failures=%d/%d",
		totalDur.Round(time.Millisecond), totalRetries, totalTimeouts, failures, len(files))
}

func fileKB(f string) int64 {
	st, err := os.Stat(f)
	if err != nil {
		return 0
	}
	return st.Size() / 1024
}
