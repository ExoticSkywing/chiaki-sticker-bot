package msbimport

import (
	"os/exec"
	"strings"

	log "github.com/sirupsen/logrus"
)

// Replaces tgs to gif.
func RlottieToGIF(f string) (string, error) {
	release := acquireLottieGIFSlot()
	defer release()

	bin := "msb_rlottie.py"
	fOut := strings.ReplaceAll(f, ".tgs", ".gif")
	args := []string{f, fOut}
	out, err := exec.Command(bin, args...).CombinedOutput()
	if err != nil {
		log.Errorln("lottieToGIF ERROR!", string(out))
		return "", err
	}
	//Optimize GIF
	exec.Command("gifsicle", "--batch", "-O2", "--lossy=60", fOut).CombinedOutput()
	return fOut, nil
}
