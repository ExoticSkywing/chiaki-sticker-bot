package msbimport

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
)

type webmRateControl struct {
	minrate string
	bitrate string
	maxrate string
}

var kakaoWebmRateControls = []webmRateControl{
	{minrate: "80k", bitrate: "650k", maxrate: "760k"},
	{minrate: "70k", bitrate: "560k", maxrate: "680k"},
	{minrate: "60k", bitrate: "480k", maxrate: "580k"},
	{minrate: "50k", bitrate: "400k", maxrate: "500k"},
	{minrate: "50k", bitrate: "350k", maxrate: "450k"},
	{minrate: "30k", bitrate: "260k", maxrate: "360k"},
	{minrate: "20k", bitrate: "180k", maxrate: "280k"},
}

// WebpToWebmViaPipe converts an animated WebP to webm by streaming PNG frames
// from ImageMagick directly into ffmpeg via pipe, avoiding large intermediate
// files and reducing peak memory usage.
func WebpToWebmViaPipe(f string, isCustomEmoji bool) (string, error) {
	pathOut := f + ".webm"

	fps := webpFPS(f)
	log.Debugf("WebpToWebmViaPipe: %s fps=%.2f", f, fps)

	scale := "512:512:force_original_aspect_ratio=decrease"
	if isCustomEmoji {
		scale = "100:100:force_original_aspect_ratio=decrease"
	}

	var lastErr error
	for _, rc := range kakaoWebmRateControls {
		err := webpToWebmViaPipeOnce(f, pathOut, scale, fps, rc)
		if err != nil {
			lastErr = err
			log.Warnln("WebpToWebmViaPipe: retrying with low-memory frame sequence fallback.")
			os.Remove(pathOut)
			if fallback, fallbackErr := WebpToWebmViaFrames(f, isCustomEmoji); fallbackErr == nil {
				return fallback, nil
			} else {
				log.Warnln("WebpToWebmViaPipe fallback ERROR:", fallbackErr)
			}
			return pathOut, err
		}
		st, err := os.Stat(pathOut)
		if err != nil || st.Size() == 0 {
			lastErr = errors.New("WebpToWebmViaPipe: output empty")
			os.Remove(pathOut)
			continue
		}
		if st.Size() <= 255*KiB {
			return pathOut, nil
		}
		lastErr = fmt.Errorf("WebpToWebmViaPipe: output too large: %d bytes", st.Size())
		log.Warnf("WebpToWebmViaPipe: output too large at %s, retrying lower bitrate: %d bytes", rc.bitrate, st.Size())
		os.Remove(pathOut)
	}
	if lastErr != nil {
		return pathOut, lastErr
	}
	return pathOut, errors.New("WebpToWebmViaPipe: no encode attempts")
}

func webpToWebmViaPipeOnce(f string, pathOut string, scale string, fps float64, rc webmRateControl) error {
	ffArgs := append([]string{}, ffmpegQ...)
	ffArgs = append(ffArgs,
		"-f", "image2pipe", "-vcodec", "png",
		"-framerate", fmt.Sprintf("%g", fps),
		"-i", "pipe:0",
		"-vf", "scale="+scale,
		"-threads", "1", "-pix_fmt", "yuva420p", "-c:v", "libvpx-vp9",
		"-cpu-used", "5", "-lag-in-frames", "0", "-tile-columns", "0", "-tile-rows", "0", "-auto-alt-ref", "0",
		"-minrate", rc.minrate, "-b:v", rc.bitrate, "-maxrate", rc.maxrate,
		"-to", "00:00:03", "-an", "-y", pathOut,
	)

	imArgs := append([]string{}, CONVERT_ARGS...)
	imArgs = append(imArgs, imageMagickResourceArgs()...)
	imArgs = append(imArgs, "WEBP:"+f, "-coalesce", "png:-")

	imCmd := exec.Command(CONVERT_BIN, imArgs...)
	ffCmd := niceCommand(FFMPEG_BIN, ffArgs...)

	pr, pw := io.Pipe()
	imCmd.Stdout = pw
	var ffOut bytes.Buffer
	ffCmd.Stdin = pr
	ffCmd.Stderr = &ffOut

	releaseFFmpeg := acquireFFmpegSlot()
	if err := imCmd.Start(); err != nil {
		releaseFFmpeg()
		return fmt.Errorf("WebpToWebmViaPipe: imCmd start: %w", err)
	}
	if err := ffCmd.Start(); err != nil {
		releaseFFmpeg()
		imCmd.Process.Kill()
		return fmt.Errorf("WebpToWebmViaPipe: ffCmd start: %w", err)
	}

	imErr := imCmd.Wait()
	pw.Close()
	ffErr := ffCmd.Wait()
	releaseFFmpeg()

	if imErr != nil || ffErr != nil {
		log.Warnln("WebpToWebmViaPipe ERROR ffmpeg:", ffOut.String())
		if ffErr != nil {
			return ffErr
		}
		return imErr
	}
	return nil
}

// WebpToWebmViaFrames trades temporary disk writes for a lower memory peak:
// ImageMagick exits before ffmpeg starts, so the two large processes do not
// overlap in RSS on 256MB deployments.
func WebpToWebmViaFrames(f string, isCustomEmoji bool) (string, error) {
	pathOut := f + ".webm"
	fps := webpFPS(f)
	frameDir, err := os.MkdirTemp(filepath.Dir(f), filepath.Base(f)+".frames-*")
	if err != nil {
		return pathOut, err
	}
	defer os.RemoveAll(frameDir)

	framePattern := filepath.Join(frameDir, "frame-%05d.png")
	size := "512x512>"
	scale := "512:512:force_original_aspect_ratio=decrease"
	if isCustomEmoji {
		size = "100x100>"
		scale = "100:100:force_original_aspect_ratio=decrease"
	}

	imArgs := append([]string{}, CONVERT_ARGS...)
	imArgs = append(imArgs, imageMagickResourceArgs()...)
	imArgs = append(imArgs, "WEBP:"+f, "-coalesce", "-resize", size, framePattern)
	imOut, err := exec.Command(CONVERT_BIN, imArgs...).CombinedOutput()
	if err != nil {
		log.Warnln("WebpToWebmViaFrames ImageMagick ERROR:", string(imOut))
		return pathOut, err
	}
	frames, err := filepath.Glob(filepath.Join(frameDir, "frame-*.png"))
	if err != nil || len(frames) == 0 {
		return pathOut, errors.New("WebpToWebmViaFrames: no frames produced")
	}

	var lastErr error
	for _, rc := range kakaoWebmRateControls {
		ffArgs := append([]string{}, ffmpegQ...)
		ffArgs = append(ffArgs,
			"-framerate", fmt.Sprintf("%g", fps),
			"-i", framePattern,
			"-vf", "scale="+scale,
			"-threads", "1", "-pix_fmt", "yuva420p", "-c:v", "libvpx-vp9",
			"-cpu-used", "5", "-lag-in-frames", "0", "-tile-columns", "0", "-tile-rows", "0", "-auto-alt-ref", "0",
			"-minrate", rc.minrate, "-b:v", rc.bitrate, "-maxrate", rc.maxrate,
			"-to", "00:00:03", "-an", "-y", pathOut,
		)
		releaseFFmpeg := acquireFFmpegSlot()
		out, err := niceCommand(FFMPEG_BIN, ffArgs...).CombinedOutput()
		releaseFFmpeg()
		if err != nil {
			log.Warnln("WebpToWebmViaFrames ffmpeg ERROR:", string(out))
			return pathOut, err
		}
		st, err := os.Stat(pathOut)
		if err != nil || st.Size() == 0 {
			lastErr = errors.New("WebpToWebmViaFrames: output empty")
			os.Remove(pathOut)
			continue
		}
		if st.Size() <= 255*KiB {
			return pathOut, nil
		}
		lastErr = fmt.Errorf("WebpToWebmViaFrames: output too large: %d bytes", st.Size())
		log.Warnf("WebpToWebmViaFrames: output too large at %s, retrying lower bitrate: %d bytes", rc.bitrate, st.Size())
		os.Remove(pathOut)
	}
	if lastErr != nil {
		return pathOut, lastErr
	}
	return pathOut, errors.New("WebpToWebmViaFrames: no encode attempts")
}

// webpFPS returns the playback FPS of an animated WebP by reading the first
// frame's delay (in centiseconds) via identify. Falls back to 25 if unknown.
func webpFPS(f string) float64 {
	out, err := exec.Command(IDENTIFY_BIN,
		append(IDENTIFY_ARGS, "-format", "%T\n", "WEBP:"+f)...,
	).Output()
	if err != nil || len(out) == 0 {
		return 25
	}
	first := strings.SplitN(strings.TrimSpace(string(out)), "\n", 2)[0]
	delay, err := strconv.ParseFloat(strings.TrimSpace(first), 64)
	if err != nil || delay <= 0 {
		return 25
	}
	return 100.0 / delay
}
