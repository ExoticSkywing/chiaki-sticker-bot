package core

import (
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
