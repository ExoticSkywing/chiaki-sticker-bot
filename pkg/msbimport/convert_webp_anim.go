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

	ffArgs := append([]string{}, ffmpegQ...)
	ffArgs = append(ffArgs,
		"-f", "image2pipe", "-vcodec", "png",
		"-framerate", fmt.Sprintf("%g", fps),
		"-i", "pipe:0",
		"-vf", "scale="+scale,
		"-threads", "1", "-pix_fmt", "yuva420p", "-c:v", "libvpx-vp9",
		"-cpu-used", "8", "-lag-in-frames", "0", "-tile-columns", "0", "-tile-rows", "0", "-auto-alt-ref", "0",
		"-minrate", "50k", "-b:v", "350k", "-maxrate", "450k",
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
		return pathOut, fmt.Errorf("WebpToWebmViaPipe: imCmd start: %w", err)
	}
	if err := ffCmd.Start(); err != nil {
		releaseFFmpeg()
		imCmd.Process.Kill()
		return pathOut, fmt.Errorf("WebpToWebmViaPipe: ffCmd start: %w", err)
	}

	imErr := imCmd.Wait()
	pw.Close()
	ffErr := ffCmd.Wait()
	releaseFFmpeg()

	if imErr != nil || ffErr != nil {
		log.Warnln("WebpToWebmViaPipe ERROR ffmpeg:", ffOut.String())
		log.Warnln("WebpToWebmViaPipe: retrying with low-memory frame sequence fallback.")
		os.Remove(pathOut)
		if fallback, err := WebpToWebmViaFrames(f, isCustomEmoji); err == nil {
			return fallback, nil
		} else {
			log.Warnln("WebpToWebmViaPipe fallback ERROR:", err)
		}
		if ffErr != nil {
			return pathOut, ffErr
		}
		return pathOut, imErr
	}
	if st, err := os.Stat(pathOut); err != nil || st.Size() == 0 {
		return pathOut, errors.New("WebpToWebmViaPipe: output empty")
	}
	return pathOut, nil
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

	ffArgs := append([]string{}, ffmpegQ...)
	ffArgs = append(ffArgs,
		"-framerate", fmt.Sprintf("%g", fps),
		"-i", framePattern,
		"-vf", "scale="+scale,
		"-threads", "1", "-pix_fmt", "yuva420p", "-c:v", "libvpx-vp9",
		"-cpu-used", "8", "-lag-in-frames", "0", "-tile-columns", "0", "-tile-rows", "0", "-auto-alt-ref", "0",
		"-minrate", "50k", "-b:v", "350k", "-maxrate", "450k",
		"-to", "00:00:03", "-an", "-y", pathOut,
	)
	releaseFFmpeg := acquireFFmpegSlot()
	out, err := niceCommand(FFMPEG_BIN, ffArgs...).CombinedOutput()
	releaseFFmpeg()
	if err != nil {
		log.Warnln("WebpToWebmViaFrames ffmpeg ERROR:", string(out))
		return pathOut, err
	}
	if st, err := os.Stat(pathOut); err != nil || st.Size() == 0 {
		return pathOut, errors.New("WebpToWebmViaFrames: output empty")
	}
	return pathOut, nil
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
