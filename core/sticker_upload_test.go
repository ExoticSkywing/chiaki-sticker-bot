package core

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"testing"

	tele "gopkg.in/telebot.v3"
)

func TestUploadedMediaSavePathPreservesImageType(t *testing.T) {
	workDir := t.TempDir()

	tests := []struct {
		name string
		msg  *tele.Message
		want string
	}{
		{
			name: "telegram photo",
			msg:  &tele.Message{Photo: &tele.Photo{}},
			want: ".jpg",
		},
		{
			name: "PNG document filename",
			msg:  &tele.Message{Document: &tele.Document{FileName: "sticker.PNG", MIME: "image/png"}},
			want: ".PNG",
		},
		{
			name: "PNG document MIME fallback",
			msg:  &tele.Message{Document: &tele.Document{MIME: "image/png"}},
			want: ".png",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := uploadedMediaSavePath(tt.msg, workDir, "upload")
			if want := filepath.Join(workDir, "upload") + tt.want; got != want {
				t.Fatalf("uploadedMediaSavePath() = %q, want %q", got, want)
			}
		})
	}
}

func TestStickerSourceFilesSkipsArchiveMetadata(t *testing.T) {
	dir := t.TempDir()
	png := filepath.Join(dir, "sticker.png")
	metadata := filepath.Join(dir, "._sticker.png")
	unsupported := filepath.Join(dir, ".DS_Store")

	pngBytes, err := base64.StdEncoding.DecodeString("iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVQIHWP4z8DwHwAF/gL+Zl5eAAAAAElFTkSuQmCC")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(png, pngBytes, 0o600); err != nil {
		t.Fatal(err)
	}
	for _, file := range []string{metadata, unsupported} {
		if err := os.WriteFile(file, []byte("metadata"), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	got := stickerSourceFiles([]string{metadata, unsupported, png})
	if len(got) != 1 || got[0] != png {
		t.Fatalf("stickerSourceFiles() = %v, want [%s]", got, png)
	}
}
