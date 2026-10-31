package client

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"unicode"

	"github.com/PuerkitoBio/goquery"
	fhttp "github.com/bogdanfinn/fhttp"
	tls_client "github.com/bogdanfinn/tls-client"
	"github.com/bogdanfinn/tls-client/profiles"
)

const BaseURL = "https://downloads.khinsider.com"

type Client struct {
	UserAgent string
	HTTP      tls_client.HttpClient
}

type Album struct {
	Title    string
	URL      string
	Platform string
	Type     string
	Year     string
}

type Publisher struct {
	Name string
	URL  string
}

type PublisherDetails struct {
	Name   string
	Albums []Album
}

type Song struct {
	Number   string
	Title    string
	URL      string // Page URL
	Duration string
	MP3Size  string
	FLACSize string
}

type AlbumDetails struct {
	Title  string
	Year   string
	Songs  []Song
	Images []AlbumImage
}

// AlbumImage is a full-size image linked from an album page.
type AlbumImage struct {
	URL string
}

const DefaultUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36 Edg/124.0.0.0"

var yearPattern = regexp.MustCompile(`(?:^|[^0-9])([12][0-9]{3})(?:[^0-9]|$)`)

func NewClient() *Client {
	options := []tls_client.HttpClientOption{
		tls_client.WithClientProfile(profiles.Chrome_120),
		tls_client.WithRandomTLSExtensionOrder(),
	}

	client, _ := tls_client.NewHttpClient(tls_client.NewNoopLogger(), options...)

	return &Client{
		UserAgent: DefaultUserAgent,
		HTTP:      client,
	}
}

func (c *Client) SetBrowserHeaders(req *fhttp.Request) {
	req.Header.Set("User-Agent", c.UserAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Sec-Ch-Ua", `"Not_A Brand";v="8", "Chromium";v="110", "Microsoft Edge";v="110"`)
	req.Header.Set("Sec-Ch-Ua-Mobile", "?0")
	req.Header.Set("Sec-Ch-Ua-Platform", `"Windows"`)
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "none")
	req.Header.Set("Sec-Fetch-User", "?1")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
}

func (c *Client) doRequest(url string) (*goquery.Document, error) {
	req, err := fhttp.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	c.SetBrowserHeaders(req)
	req.Header.Set("Referer", BaseURL)

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("status code error: %d %s", resp.StatusCode, resp.Status)
	}

	return goquery.NewDocumentFromReader(resp.Body)
}

func (c *Client) Search(query string) ([]Album, error) {
	url := fmt.Sprintf("%s/search?search=%s", BaseURL, strings.ReplaceAll(query, " ", "+"))
	doc, err := c.doRequest(url)
	if err != nil {
		return nil, err
	}

	return parseAlbums(doc, "table.albumList tr"), nil
}

// SearchPublishers resolves a publisher name to its KHInsider publisher page.
// KHInsider does not expose a publisher-search endpoint, so a valid result is
// presented as a selectable publisher before its albums are loaded.
func (c *Client) SearchPublishers(query string) ([]Publisher, error) {
	slug := publisherSlug(query)
	if slug == "" {
		return nil, nil
	}

	publisherURL := fmt.Sprintf("%s/game-soundtracks/publisher/%s", BaseURL, url.PathEscape(slug))
	doc, err := c.doRequest(publisherURL)
	if err != nil {
		return nil, err
	}

	name := publisherName(doc)
	if name == "" {
		return nil, nil
	}
	return []Publisher{{Name: name, URL: publisherURL}}, nil
}

// GetPublisherAlbums retrieves every page of a publisher's album list.
func (c *Client) GetPublisherAlbums(publisherURL string) (PublisherDetails, error) {
	doc, err := c.doRequest(publisherURL)
	if err != nil {
		return PublisherDetails{}, err
	}

	name := publisherName(doc)
	if name == "" {
		return PublisherDetails{}, fmt.Errorf("publisher not found")
	}

	details := PublisherDetails{Name: name}
	seen := make(map[string]struct{})
	appendAlbums := func(albums []Album) {
		for _, album := range albums {
			if _, exists := seen[album.URL]; exists {
				continue
			}
			seen[album.URL] = struct{}{}
			details.Albums = append(details.Albums, album)
		}
	}
	appendAlbums(parseAlbumRows(publisherAlbumRows(doc)))

	for _, pageURL := range publisherPageURLs(doc, publisherURL) {
		pageDoc, err := c.doRequest(pageURL)
		if err != nil {
			return PublisherDetails{}, err
		}
		appendAlbums(parseAlbumRows(publisherAlbumRows(pageDoc)))
	}

	return details, nil
}

func parseAlbums(doc *goquery.Document, selector string) []Album {
	return parseAlbumRows(doc.Find(selector))
}

func parseAlbumRows(rows *goquery.Selection) []Album {
	seen := make(map[string]struct{})
	var albums []Album

	rows.Each(func(_ int, row *goquery.Selection) {
		tds := row.Find("td")
		if tds.Length() == 0 {
			return
		}

		titleIndex := -1
		var titleLink *goquery.Selection
		tds.EachWithBreak(func(i int, td *goquery.Selection) bool {
			var link *goquery.Selection
			td.Find("a[href*='/game-soundtracks/album/']").EachWithBreak(func(_ int, candidate *goquery.Selection) bool {
				if strings.TrimSpace(candidate.Text()) == "" {
					return true
				}
				link = candidate
				return false
			})
			if link == nil {
				return true
			}
			titleIndex = i
			titleLink = link
			return false
		})
		if titleIndex < 0 || titleLink == nil {
			return
		}

		album := Album{
			Title: strings.TrimSpace(titleLink.Text()),
			URL:   absoluteURL(titleLink.AttrOr("href", "")),
		}
		if album.Title == "" || album.URL == "" {
			return
		}
		if titleIndex+1 < tds.Length() {
			album.Platform = cleanText(tds.Eq(titleIndex + 1).Text())
		}
		if titleIndex+2 < tds.Length() {
			album.Type = cleanText(tds.Eq(titleIndex + 2).Text())
		}
		if titleIndex+3 < tds.Length() {
			album.Year = cleanText(tds.Eq(titleIndex + 3).Text())
		}
		if _, exists := seen[album.URL]; exists {
			return
		}

		seen[album.URL] = struct{}{}
		albums = append(albums, album)
	})

	return albums
}

func publisherAlbumRows(doc *goquery.Document) *goquery.Selection {
	var rows *goquery.Selection
	doc.Find("#pageContent table").EachWithBreak(func(_ int, table *goquery.Selection) bool {
		if strings.Contains(cleanText(table.Find("tr").First().Text()), "Album Name") {
			rows = table.Find("tr")
			return false
		}
		return true
	})
	if rows == nil {
		return doc.Find(".publisher-album-table tr")
	}
	return rows
}

func publisherName(doc *goquery.Document) string {
	const prefix = "Albums Published by "
	name := ""
	doc.Find("#pageContent h2").EachWithBreak(func(_ int, s *goquery.Selection) bool {
		text := strings.TrimSpace(s.Text())
		if strings.HasPrefix(text, prefix) {
			name = strings.TrimSpace(strings.TrimPrefix(text, prefix))
			return false
		}
		return true
	})
	return name
}

func publisherPageURLs(doc *goquery.Document, publisherURL string) []string {
	base, err := url.Parse(publisherURL)
	if err != nil {
		return nil
	}

	pages := make(map[int]*url.URL)
	doc.Find("#pageContent a[href]").Each(func(_ int, s *goquery.Selection) {
		href := s.AttrOr("href", "")
		ref, err := url.Parse(href)
		if err != nil {
			return
		}
		page := base.ResolveReference(ref)
		pageNumber := page.Query().Get("page")
		if pageNumber == "" {
			return
		}
		var number int
		if _, err := fmt.Sscanf(pageNumber, "%d", &number); err != nil || number < 2 {
			return
		}
		pages[number] = page
	})

	maxPage := 1
	for page := range pages {
		if page > maxPage {
			maxPage = page
		}
	}
	if maxPage == 1 {
		return nil
	}

	template := pages[maxPage]
	urls := make([]string, 0, maxPage-1)
	for page := 2; page <= maxPage; page++ {
		if pageURL, ok := pages[page]; ok {
			urls = append(urls, pageURL.String())
			continue
		}
		pageURL := *template
		query := pageURL.Query()
		query.Set("page", fmt.Sprint(page))
		pageURL.RawQuery = query.Encode()
		urls = append(urls, pageURL.String())
	}
	return urls
}

func publisherSlug(query string) string {
	var slug strings.Builder
	separator := false
	for _, r := range strings.ToLower(strings.TrimSpace(query)) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			slug.WriteRune(r)
			separator = false
			continue
		}
		if slug.Len() > 0 && !separator {
			slug.WriteByte('-')
			separator = true
		}
	}
	return strings.Trim(slug.String(), "-")
}

func absoluteURL(href string) string {
	if href == "" {
		return ""
	}
	if strings.HasPrefix(href, "http") {
		return href
	}
	return BaseURL + href
}

func cleanText(text string) string {
	return strings.TrimSpace(strings.Join(strings.Fields(text), " "))
}

func (c *Client) GetAlbumDetails(albumURL string) (AlbumDetails, error) {
	doc, err := c.doRequest(albumURL)
	if err != nil {
		return AlbumDetails{}, err
	}

	details := AlbumDetails{
		Title:  doc.Find("h2").First().Text(),
		Year:   albumYear(albumURL, doc.Find("h2").First().Text()),
		Images: extractAlbumImages(doc, albumURL),
	}

	// Find column indices dynamically using colspan tracking
	tdIndexMap := make(map[string]int)
	currentTdIdx := 0
	doc.Find("table#songlist tr#songlist_header th").Each(func(i int, s *goquery.Selection) {
		text := strings.ToLower(strings.TrimSpace(s.Text()))
		colspan := 1
		if cStr, exists := s.Attr("colspan"); exists {
			fmt.Sscanf(cStr, "%d", &colspan)
		}

		if strings.Contains(text, "song name") {
			if colspan > 1 {
				tdIndexMap["duration"] = currentTdIdx + 1
			}
		} else if strings.Contains(text, "mp3") {
			tdIndexMap["mp3"] = currentTdIdx
		} else if strings.Contains(text, "flac") {
			tdIndexMap["flac"] = currentTdIdx
		}

		currentTdIdx += colspan
	})

	doc.Find("table#songlist tr").Each(func(i int, s *goquery.Selection) {
		if s.AttrOr("id", "") == "songlist_header" || s.Find("td").Length() < 4 {
			return
		}

		tds := s.Find("td")
		titleLink := s.Find("td.clickable-row a").First()
		if titleLink.Length() == 0 {
			return
		}

		song := Song{
			Number: strings.TrimSpace(tds.Eq(1).Text()),
			Title:  titleLink.Text(),
			URL:    titleLink.AttrOr("href", ""),
		}

		if idx, ok := tdIndexMap["duration"]; ok && idx < tds.Length() {
			song.Duration = cleanIconText(tds.Eq(idx).Text())
		}
		if idx, ok := tdIndexMap["mp3"]; ok && idx < tds.Length() {
			song.MP3Size = cleanIconText(tds.Eq(idx).Text())
		}
		if idx, ok := tdIndexMap["flac"]; ok && idx < tds.Length() {
			song.FLACSize = cleanIconText(tds.Eq(idx).Text())
		}

		if !strings.HasPrefix(song.URL, "http") {
			song.URL = BaseURL + song.URL
		}

		details.Songs = append(details.Songs, song)
	})

	return details, nil
}

func albumYear(albumURL, title string) string {
	for _, value := range []string{albumURL, title} {
		match := yearPattern.FindStringSubmatch(value)
		if len(match) > 1 {
			return match[1]
		}
	}
	return ""
}

// extractAlbumImages finds the full-size image links used by KHInsider for an
// album. The page places the thumbnails after the uploader information in
// .albumImage elements; each element's anchor points to the original-resolution
// file. Targeting that class avoids collecting navigation and site icons.
func extractAlbumImages(doc *goquery.Document, albumURL string) []AlbumImage {
	seen := make(map[string]struct{})
	var images []AlbumImage

	doc.Find(".albumImage > a[href]").Each(func(_ int, s *goquery.Selection) {
		href := strings.TrimSpace(s.AttrOr("href", ""))
		imageURL, ok := resolveImageURL(albumURL, href)
		if !ok {
			return
		}
		if _, exists := seen[imageURL]; exists {
			return
		}

		seen[imageURL] = struct{}{}
		images = append(images, AlbumImage{URL: imageURL})
	})

	return images
}

func resolveImageURL(albumURL, href string) (string, bool) {
	parsed, err := url.Parse(href)
	if err != nil || parsed.Scheme == "" && parsed.Host == "" && parsed.Path == "" {
		return "", false
	}

	base, err := url.Parse(albumURL)
	if err != nil {
		return "", false
	}
	resolved := base.ResolveReference(parsed)

	switch strings.ToLower(resolved.Scheme) {
	case "http", "https":
		return resolved.String(), true
	default:
		return "", false
	}
}

func cleanIconText(text string) string {
	text = strings.TrimSpace(text)
	// Khinsider uses these Material icon names in cells
	text = strings.TrimPrefix(text, "get_app")
	text = strings.TrimSuffix(text, "get_app")
	text = strings.TrimPrefix(text, "playlist_add")
	text = strings.TrimSuffix(text, "playlist_add")
	text = strings.TrimPrefix(text, "play_arrow")
	text = strings.TrimSuffix(text, "play_arrow")
	return strings.TrimSpace(text)
}

func (c *Client) GetDownloadURL(songURL string, format string) (string, error) {
	doc, err := c.doRequest(songURL)
	if err != nil {
		return "", err
	}

	var downloadURL string
	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		href := s.AttrOr("href", "")
		if format == "flac" && strings.HasSuffix(strings.ToLower(href), ".flac") {
			downloadURL = href
		} else if format == "mp3" && strings.HasSuffix(strings.ToLower(href), ".mp3") && downloadURL == "" {
			downloadURL = href
		}
	})

	if downloadURL == "" {
		downloadURL = doc.Find("audio#audio").AttrOr("src", "")
	}

	if downloadURL != "" && !strings.HasPrefix(downloadURL, "http") {
		downloadURL = BaseURL + downloadURL
	}

	return downloadURL, nil
}
