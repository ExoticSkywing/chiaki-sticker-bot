package msbimport

import (
	"os/exec"
	"runtime"
	"strings"

	log "github.com/sirupsen/logrus"
)

var FFMPEG_BIN = "ffmpeg"
var FFPROBE_BIN = "ffprobe"
var BSDTAR_BIN = "bsdtar"
var CONVERT_BIN = "convert"
var IDENTIFY_BIN = "identify"
var CONVERT_ARGS []string
var IDENTIFY_ARGS []string

// ffmpegQ are the standard quiet flags prepended to every ffmpeg call.
// -loglevel error : suppress info/warning messages, only show errors
// -nostats        : suppress the frame=.../fps=.../size=... progress line
var ffmpegQ = []string{"-hide_banner", "-loglevel", "error", "-nostats"}

const (
	FORMAT_TG_REGULAR_STATIC   = "tg_reg_static"
	FORMAT_TG_EMOJI_STATIC     = "tg_emoji_static"
	FORMAT_TG_REGULAR_ANIMATED = "tg_reg_ani"
	FORMAT_TG_EMOJI_ANIMATED   = "tg_emoji_ani"
)

// See: http://en.wikipedia.org/wiki/Binary_prefix
const (
	// Decimal
	KB = 1000
	MB = 1000 * KB
	GB = 1000 * MB
	TB = 1000 * GB
	PB = 1000 * TB

	// Binary
	KiB = 1024
	MiB = 1024 * KiB
	GiB = 1024 * MiB
	TiB = 1024 * GiB
	PiB = 1024 * TiB
)

// Should call before using functions in this package.
// Otherwise, defaults to Linux environment.
// This function also call CheckDeps to check if executables.
func InitConvert() {
	switch runtime.GOOS {
	case "linux":
		CONVERT_BIN = "convert"
	default:
		CONVERT_BIN = "magick"
		IDENTIFY_BIN = "magick"
		CONVERT_ARGS = []string{"convert"}
		IDENTIFY_ARGS = []string{"identify"}
	}
	unfoundBins := CheckDeps()
	if len(unfoundBins) != 0 {
		log.Warning("Following required executables not found!:")
		log.Warnln(strings.Join(unfoundBins, "  "))
		log.Warning("Please install missing executables to your PATH, or some features will not work!")
	}
}

// Check if required dependencies exist and return a string slice
// containing binaries that are not found in PATH.
func CheckDeps() []string {
	unfoundBins := []string{}

	if _, err := exec.LookPath(FFMPEG_BIN); err != nil {
		unfoundBins = append(unfoundBins, FFMPEG_BIN)
	}
	if _, err := exec.LookPath(FFPROBE_BIN); err != nil {
		unfoundBins = append(unfoundBins, FFPROBE_BIN)
	}
	if _, err := exec.LookPath(BSDTAR_BIN); err != nil {
		unfoundBins = append(unfoundBins, BSDTAR_BIN)
	}
	if _, err := exec.LookPath(CONVERT_BIN); err != nil {
		unfoundBins = append(unfoundBins, CONVERT_BIN)
	}
	if _, err := exec.LookPath("gifsicle"); err != nil {
		unfoundBins = append(unfoundBins, "gifsicle")
	}
	return unfoundBins
}
