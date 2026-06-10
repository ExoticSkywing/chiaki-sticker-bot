package msbimport

import (
	"context"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"time"
)

// Hard ceiling for a single ffmpeg invocation. With a pool size of 1 a hung
// ffmpeg would otherwise block every queued conversion indefinitely.
const ffmpegTimeout = 120 * time.Second

// Telegram rejects video stickers longer than 3s. Sources beyond this can skip
// the first regular encode and go straight to safe mode, while sources at or
// below the limit still get a normal encode so we avoid trimming unnecessarily.
const telegramVideoMaxDuration = 3.0
const telegramVideoMaxDurationArg = "00:00:03"
const telegramVideoSafeDurationArg = "00:00:02.800"

// CPU-heavy encodes (VP9) run niced so the HTTP/health-check goroutine keeps
// getting CPU on the shared single-core VM. `nice` exec-replaces itself with the
// target binary (same PID), so CommandContext timeouts still reach ffmpeg.
const niceLevel = "19"

// heavyConverterSemaphore serializes ffmpeg and rlottie (TGS→GIF) invocations
// against each other. Both are memory-heavy on the 256MB Fly VM, so running them
// concurrently — even though they're different binaries — can OOM the box.
var (
	heavyConverterSemaphore     chan struct{}
	heavyConverterSemaphoreOnce sync.Once
)

func initHeavyConverterSemaphore() {
	heavyConverterSemaphoreOnce.Do(func() {
		concurrency := 1
		if value, err := strconv.Atoi(os.Getenv("MSB_FFMPEG_CONCURRENCY")); err == nil && value > 0 {
			concurrency = value
		}
		heavyConverterSemaphore = make(chan struct{}, concurrency)
	})
}

func niceCommand(bin string, args ...string) *exec.Cmd {
	return exec.Command("nice", append([]string{"-n", niceLevel, bin}, args...)...)
}

func niceCommandContext(ctx context.Context, bin string, args ...string) *exec.Cmd {
	return exec.CommandContext(ctx, "nice", append([]string{"-n", niceLevel, bin}, args...)...)
}

func acquireLottieGIFSlot() func() {
	initHeavyConverterSemaphore()
	heavyConverterSemaphore <- struct{}{}
	return func() {
		<-heavyConverterSemaphore
	}
}

func acquireFFmpegSlot() func() {
	initHeavyConverterSemaphore()
	heavyConverterSemaphore <- struct{}{}
	return func() {
		<-heavyConverterSemaphore
	}
}

func imageMagickResourceArgs() []string {
	memoryLimit := os.Getenv("MSB_IM_MEMORY_LIMIT")
	if memoryLimit == "" {
		memoryLimit = "64MiB"
	}
	mapLimit := os.Getenv("MSB_IM_MAP_LIMIT")
	if mapLimit == "" {
		mapLimit = "128MiB"
	}
	args := []string{}
	if memoryLimit != "0" {
		args = append(args, "-limit", "memory", memoryLimit)
	}
	if mapLimit != "0" {
		args = append(args, "-limit", "map", mapLimit)
	}
	return args
}
