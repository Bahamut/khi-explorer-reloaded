package downloader

import (
	"testing"

	"github.com/Bahamut/khi-explorer-reloaded/client"
	"github.com/Bahamut/khi-explorer-reloaded/config"
)

func TestImageFilenameDecodesAndDisambiguates(t *testing.T) {
	used := make(map[string]int)
	if got, want := imageFilename("https://images.example/00%20Front.jpg", 1, used), "00 Front.jpg"; got != want {
		t.Errorf("first filename = %q, want %q", got, want)
	}
	if got, want := imageFilename("https://images.example/00%20Front.jpg", 2, used), "00 Front-2.jpg"; got != want {
		t.Errorf("duplicate filename = %q, want %q", got, want)
	}
	if got, want := imageFilename("https://images.example/", 3, used), "image-03"; got != want {
		t.Errorf("fallback filename = %q, want %q", got, want)
	}
}

func TestConfiguredNames(t *testing.T) {
	d := Downloader{Config: config.Config{
		TrackFilenameFormat: "{track_name} ({track_number})",
		AlbumFolderFormat:   "{album_name} [{year}]",
	}}
	if got, want := d.trackFilename(client.Song{Number: "03", Title: "Main Theme"}, ".flac"), "Main Theme (03).flac"; got != want {
		t.Errorf("track filename = %q, want %q", got, want)
	}
	if got, want := d.albumFolderName("Earthlight", "2006"), "Earthlight [2006]"; got != want {
		t.Errorf("album folder = %q, want %q", got, want)
	}
}

func TestDefaultAlbumFolderWithoutYear(t *testing.T) {
	d := Downloader{}
	if got, want := d.albumFolderName("Earthlight", ""), "Earthlight"; got != want {
		t.Errorf("album folder = %q, want %q", got, want)
	}
}
