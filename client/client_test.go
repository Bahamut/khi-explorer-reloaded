package client

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func TestAlbumYear(t *testing.T) {
	if got, want := albumYear("https://downloads.khinsider.com/game-soundtracks/album/earthlight-2006", "Earthlight"), "2006"; got != want {
		t.Errorf("album year = %q, want %q", got, want)
	}
	if got, want := albumYear("https://example.test/album/no-date", "Album Name (1998)"), "1998"; got != want {
		t.Errorf("album year from title = %q, want %q", got, want)
	}
}

func TestExtractAlbumImagesUsesFullSizeImageLinks(t *testing.T) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(`
		<html><body>
			<a href="/images/site-logo.png"><img src="/images/site-logo-small.png"></a>
			<p>Uploaded by: <a href="/forums/member/1">Uploader</a></p>
			<div class="albumImage"><a href="https://images.example/00%20Front.jpg"><img src="https://images.example/thumb-front.jpg"></a></div>
			<div class="albumImage"><a href="/art/01%20Back.PNG"><img src="/art/thumb-back.png"></a></div>
			<div class="albumImage"><a href="/images/render?id=2"><img src="/images/thumb?id=2"></a></div>
			<div class="albumImage"><a href="https://images.example/00%20Front.jpg"><img src="https://images.example/thumb-front-duplicate.jpg"></a></div>
			<a href="/download/album.zip"><img src="/icons/download.png"></a>
		</body></html>`))
	if err != nil {
		t.Fatal(err)
	}

	images := extractAlbumImages(doc, "https://downloads.khinsider.com/game-soundtracks/album/example")
	got := make([]string, len(images))
	for i, image := range images {
		got[i] = image.URL
	}
	want := []string{
		"https://images.example/00%20Front.jpg",
		"https://downloads.khinsider.com/art/01%20Back.PNG",
		"https://downloads.khinsider.com/images/render?id=2",
	}
	if len(got) != len(want) {
		t.Fatalf("got %d images (%v), want %d", len(got), got, len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("image %d = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestParsePublisherAlbums(t *testing.T) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(`
		<div id="pageContent">
			<h2>Top 12 Albums Published by Example</h2>
			<h2>Albums Published by Example Publisher</h2>
			<table><tr><th>Album Name</th></tr>
			<tr><td><a href="/icon">icon</a></td><td><a href="/game-soundtracks/album/first">First Album</a></td><td>PC</td><td>Soundtrack</td><td>2026</td></tr>
			<tr><td><a href="/icon">icon</a></td><td><a href="/game-soundtracks/album/second">Second Album</a></td><td>Switch</td><td>Gamerip</td><td>2025</td></tr></table>
		</div>`))
	if err != nil {
		t.Fatal(err)
	}
	if got, want := publisherName(doc), "Example Publisher"; got != want {
		t.Fatalf("publisher name = %q, want %q", got, want)
	}

	albums := parseAlbums(doc, "#pageContent table tr")
	if got, want := len(albums), 2; got != want {
		t.Fatalf("got %d albums, want %d", got, want)
	}
	if got, want := albums[0], (Album{Title: "First Album", URL: BaseURL + "/game-soundtracks/album/first", Platform: "PC", Type: "Soundtrack", Year: "2026"}); got != want {
		t.Errorf("first album = %#v, want %#v", got, want)
	}
}

func TestPublisherSlug(t *testing.T) {
	if got, want := publisherSlug(" Levo  Lution! "), "levo-lution"; got != want {
		t.Errorf("publisher slug = %q, want %q", got, want)
	}
}
