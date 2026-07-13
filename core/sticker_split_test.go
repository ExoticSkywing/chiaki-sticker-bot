package core

import (
	"strings"
	"testing"

	tele "gopkg.in/telebot.v3"
)

func TestSplitStickerSetBatches(t *testing.T) {
	stickers := make([]*StickerFile, 300)
	for i := range stickers {
		stickers[i] = &StickerFile{}
	}

	batches := splitStickerSetBatches(stickers, true, tele.StickerCustomEmoji)
	if got, want := len(batches), 2; got != want {
		t.Fatalf("batch count = %d, want %d", got, want)
	}
	if got, want := len(batches[0]), 200; got != want {
		t.Fatalf("first batch size = %d, want %d", got, want)
	}
	if got, want := len(batches[1]), 100; got != want {
		t.Fatalf("second batch size = %d, want %d", got, want)
	}
	if batches[0][199] != stickers[199] || batches[1][0] != stickers[200] {
		t.Fatal("stickers were not kept in their original order")
	}
}

func TestSplitStickerSetBatchesOnlySplitsCustomEmojiCreation(t *testing.T) {
	stickers := make([]*StickerFile, 201)
	for _, test := range []struct {
		name      string
		createSet bool
		ssType    string
	}{
		{name: "regular", createSet: true, ssType: tele.StickerRegular},
		{name: "existing set", createSet: false, ssType: tele.StickerCustomEmoji},
	} {
		t.Run(test.name, func(t *testing.T) {
			if got := splitStickerSetBatches(stickers, test.createSet, test.ssType); len(got) != 1 {
				t.Fatalf("batch count = %d, want 1", len(got))
			}
		})
	}
}

func TestStickerSetPartNamePreservesTelegramLimit(t *testing.T) {
	name := strings.Repeat("a", 54) + "_by_test_bot"
	got := stickerSetPartName(name, 2)
	if len(got) > 64 {
		t.Fatalf("part name length = %d, want <= 64: %q", len(got), got)
	}
	if !strings.HasSuffix(got, "_part2_by_test_bot") {
		t.Fatalf("part name = %q, missing part suffix", got)
	}
	if gotFirst := stickerSetPartName(name, 1); gotFirst != name {
		t.Fatalf("first part name = %q, want original %q", gotFirst, name)
	}
}

func TestStickerSetPartTitlePreservesTitleAndLimit(t *testing.T) {
	title := strings.Repeat("表", 130)
	got := stickerSetPartTitle(title, 2, 2)
	if len([]rune(got)) > 128 {
		t.Fatalf("title rune length = %d, want <= 128", len([]rune(got)))
	}
	if !strings.HasSuffix(got, " (2/2)") {
		t.Fatalf("title = %q, missing part suffix", got)
	}
}
