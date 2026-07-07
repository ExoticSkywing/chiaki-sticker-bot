package msbimport

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
)

const (
	tgsGIFBackendAuto            = "auto"
	tgsGIFBackendLottieConverter = "lottie-converter"
	tgsGIFBackendRlottiePython   = "rlottie-python"
)

// Replaces tgs to gif.
func RlottieToGIF(f string) (string, error) {
	release := acquireLottieGIFSlot()
	defer release()

	backend := strings.ToLower(strings.TrimSpace(os.Getenv("MSB_TGS_GIF_BACKEND")))
	if backend == "" {
		backend = tgsGIFBackendAuto
	}

	switch backend {
	case tgsGIFBackendRlottiePython:
		return rlottiePythonToGIF(f)
	case tgsGIFBackendLottieConverter:
		return lottieConverterToGIF(f)
	case tgsGIFBackendAuto:
		fOut, err := lottieConverterToGIF(f)
		if err == nil {
			return fOut, nil
		}
		log.Warnln("lottie-converter TGS->GIF failed, falling back to rlottie-python:", err)
		return rlottiePythonToGIF(f)
	default:
		return "", fmt.Errorf("unknown MSB_TGS_GIF_BACKEND: %s", backend)
	}
}

func lottieConverterToGIF(f string) (string, error) {
	fOut := strings.ReplaceAll(f, ".tgs", ".gif")
	args := []string{
		"--width", envStringDefault("MSB_TGS_GIF_WIDTH", "512"),
		"--height", envStringDefault("MSB_TGS_GIF_HEIGHT", "512"),
		"--fps", envStringDefault("MSB_TGS_GIF_FPS", "30"),
		"--quality", envStringDefault("MSB_TGS_GIF_QUALITY", "90"),
		"--threads", envStringDefault("MSB_TGS_GIF_THREADS", "2"),
		"--output", fOut,
		f,
	}
	out, err := commandOutputWithTimeout("lottie_to_gif.sh", args...)
	if err != nil {
		log.Warnf("lottie-converter TGS->GIF ERROR:\n%s", string(out))
		return "", err
	}
	return fOut, nil
}

func rlottiePythonToGIF(f string) (string, error) {
	bin := "msb_rlottie.py"
	fOut := strings.ReplaceAll(f, ".tgs", ".gif")
	args := []string{f, fOut}
	out, err := commandOutputWithTimeout(bin, args...)
	if err != nil {
		log.Errorln("lottieToGIF ERROR!", string(out))
		return "", err
	}
	// Optimize GIF.
	commandOutputWithTimeout("gifsicle", "--batch", "-O2", "--lossy=60", fOut)
	return fOut, nil
}

func envStringDefault(name string, defaultValue string) string {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return defaultValue
	}
	if _, err := strconv.Atoi(value); err != nil {
		return defaultValue
	}
	return value
}
