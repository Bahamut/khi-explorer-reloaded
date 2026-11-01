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

func TestExtractAlbumYearFromPageMetadata(t *testing.T) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(`
		<div id="pageContent">
			<h2>Ragnarok Battle Offline Opening Theme - Sajin no Kanata he</h2>
			<p>Platforms: <a href="/game-soundtracks/windows">Windows</a><br/>
			Year: <b>2006</b><br/>
			Catalog Number: <b>FBCD-011</b></p>
		</div>`))
	if err != nil {
		t.Fatal(err)
	}

	got := extractAlbumYear(doc, "https://downloads.khinsider.com/game-soundtracks/album/ragnarok-battle-offline-opening", "Ragnarok Battle Offline Opening Theme - Sajin no Kanata he")
	if want := "2006"; got != want {
		t.Errorf("extractAlbumYear = %q, want %q", got, want)
	}
}

func TestParseAlbumDetailsSingleAndMultiDisc(t *testing.T) {
	singleDoc, err := goquery.NewDocumentFromReader(strings.NewReader(`
		<div id="pageContent">
			<h2>Single Disc OST</h2>
			<p>Year: <b>2006</b></p>
			<table id="songlist">
				<tr id="songlist_header"><th></th><th>#</th><th colspan="2">Song Name</th><th>MP3</th><th>FLAC</th></tr>
				<tr><td></td><td>1.</td><td class="clickable-row"><a href="/song1">Beyond the Cloud</a></td><td class="clickable-row">4:50</td><td>9 MB</td><td>30 MB</td></tr>
				<tr><td></td><td>10.</td><td class="clickable-row"><a href="/song10">Finale</a></td><td class="clickable-row">3:10</td><td>6 MB</td><td>20 MB</td></tr>
			</table>
		</div>`))
	if err != nil {
		t.Fatal(err)
	}

	single := parseAlbumDetails(singleDoc, "https://downloads.khinsider.com/game-soundtracks/album/single-ost")
	if single.Year != "2006" {
		t.Errorf("single.Year = %q, want 2006", single.Year)
	}
	if len(single.Songs) != 2 {
		t.Fatalf("len(single.Songs) = %d, want 2", len(single.Songs))
	}
	if single.Songs[0].Disk != "" || single.Songs[0].Number != "01" || single.Songs[0].Title != "Beyond the Cloud" {
		t.Errorf("single.Songs[0] = %+v", single.Songs[0])
	}
	if single.Songs[1].Disk != "" || single.Songs[1].Number != "10" {
		t.Errorf("single.Songs[1] = %+v", single.Songs[1])
	}

	multiDoc, err := goquery.NewDocumentFromReader(strings.NewReader(`
		<div id="pageContent">
			<h2>RAGNAROK Online Complete Soundtrack</h2>
			<p>Year: <b>2009</b></p>
			<table id="songlist">
				<tr id="songlist_header"><th></th><th>CD</th><th>#</th><th colspan="2">Song Name</th><th>MP3</th><th>FLAC</th></tr>
				<tr><td></td><td>1</td><td>1.</td><td class="clickable-row"><a href="/s1">Title</a></td><td class="clickable-row">1:16</td><td>2.82 MB</td><td>6.29 MB</td></tr>
				<tr><td></td><td>2</td><td>12.</td><td class="clickable-row"><a href="/s2">Monk Zonk</a></td><td class="clickable-row">2:40</td><td>5.00 MB</td><td>14.00 MB</td></tr>
			</table>
		</div>`))
	if err != nil {
		t.Fatal(err)
	}

	multi := parseAlbumDetails(multiDoc, "https://downloads.khinsider.com/game-soundtracks/album/ragnarok-online-complete-soundtrack")
	if multi.Year != "2009" {
		t.Errorf("multi.Year = %q, want 2009", multi.Year)
	}
	if len(multi.Songs) != 2 {
		t.Fatalf("len(multi.Songs) = %d, want 2", len(multi.Songs))
	}
	if multi.Songs[0].Disk != "1" || multi.Songs[0].Number != "01" || multi.Songs[0].Title != "Title" {
		t.Errorf("multi.Songs[0] = %+v", multi.Songs[0])
	}
	if multi.Songs[1].Disk != "2" || multi.Songs[1].Number != "12" || multi.Songs[1].Title != "Monk Zonk" {
		t.Errorf("multi.Songs[1] = %+v", multi.Songs[1])
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

func TestPublisherQueryWordsAndMatching(t *testing.T) {
	words := publisherQueryWords("Levo Lution")
	if len(words) != 2 || words[0] != "levo" || words[1] != "lution" {
		t.Fatalf("publisherQueryWords = %v, want [levo lution]", words)
	}
	if !publisherMatchesAnyWord("Levo Lution", []string{"levo"}) {
		t.Errorf("expected Levo Lution to match 'levo'")
	}
	if !publisherMatchesAnyWord("Levo Lution", []string{"lution"}) {
		t.Errorf("expected Levo Lution to match 'lution'")
	}
	if !publisherMatchesAnyWord("Square Enix", []string{"square", "falcom"}) {
		t.Errorf("expected Square Enix to match any word in [square falcom]")
	}
	if publisherMatchesAnyWord("Capcom", []string{"levo", "lution"}) {
		t.Errorf("did not expect Capcom to match [levo lution]")
	}
}
