package downloader

import (
	"path/filepath"
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

func TestDefaultTrackFilenameSingleAndMultiDisc(t *testing.T) {
	d := Downloader{}
	if got, want := d.trackFilename(client.Song{Number: "1.", Title: "Beyond the Cloud"}, ".flac"), "01 - Beyond the Cloud.flac"; got != want {
		t.Errorf("single disc track filename = %q, want %q", got, want)
	}
	if got, want := d.trackFilename(client.Song{Disk: "1", Number: "1.", Title: "Title"}, ".flac"), "1.01 - Title.flac"; got != want {
		t.Errorf("multi disc track filename = %q, want %q", got, want)
	}
}

func TestDiskFoldersSeparatesCDsAndOmitsDiskPrefix(t *testing.T) {
	d := Downloader{Config: config.Config{DiskFolders: true}}
	song := client.Song{Disk: "2", Number: "1.", Title: "Title"}
	if got, want := d.trackFilename(song, ".flac"), "01 - Title.flac"; got != want {
		t.Errorf("disk folders track filename = %q, want %q", got, want)
	}
	if got, want := d.songOutputDir(filepath.Join("downloads", "Album"), song), filepath.Join("downloads", "Album", "CD2"); got != want {
		t.Errorf("disk folders output dir = %q, want %q", got, want)
	}
}

func TestDefaultAlbumFolderWithoutYear(t *testing.T) {
	d := Downloader{}
	if got, want := d.albumFolderName("Earthlight", ""), "Earthlight"; got != want {
		t.Errorf("album folder = %q, want %q", got, want)
	}
}

func TestCleanInfoTxtRemovesHeaderAndOptionalBlankSecondLine(t *testing.T) {
	withBlankSecondLine := "\xef\xbb\xbfThis file is downloaded from https://downloads.khinsider.com/game-soundtracks/album/test\r\n\r\nName: Test Album\r\nYear: 2009\r\n"
	got := string(cleanInfoTxt([]byte(withBlankSecondLine)))
	want := "Name: Test Album\r\nYear: 2009\r\n"
	if got != want {
		t.Errorf("cleanInfoTxt (blank 2nd line) = %q, want %q", got, want)
	}

	withNonBlankSecondLine := "This file is downloaded from https://downloads.khinsider.com/game-soundtracks/album/test\nName: Test Album\nYear: 2009\n"
	got2 := string(cleanInfoTxt([]byte(withNonBlankSecondLine)))
	want2 := "Name: Test Album\nYear: 2009\n"
	if got2 != want2 {
		t.Errorf("cleanInfoTxt (non-blank 2nd line) = %q, want %q", got2, want2)
	}
}
