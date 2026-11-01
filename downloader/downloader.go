package downloader

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	fhttp "github.com/bogdanfinn/fhttp"

	"github.com/Bahamut/khi-explorer-reloaded/client"
	"github.com/Bahamut/khi-explorer-reloaded/config"
)

type DownloadProgress struct {
	Total     int
	Completed int
	Current   string
}

type Downloader struct {
	Client     *client.Client
	Config     config.Config
	ProgressCB func(DownloadProgress)
}

func NewDownloader(c *client.Client, cfg config.Config) *Downloader {
	return &Downloader{
		Client: c,
		Config: cfg,
	}
}

func (d *Downloader) DownloadSong(song client.Song, albumTitle string) error {
	return d.DownloadSongWithCallback(song, albumTitle, nil)
}

func (d *Downloader) DownloadSongWithCallback(song client.Song, albumTitle string, cb func(DownloadProgress)) error {
	return d.DownloadSongForAlbumWithCallback(song, albumTitle, "", cb)
}

// DownloadSongForAlbumWithCallback saves a song using the configured album
// folder template and the album metadata available in the current view.
func (d *Downloader) DownloadSongForAlbumWithCallback(song client.Song, albumTitle, albumYear string, cb func(DownloadProgress)) error {
	return d.downloadSongWithCallback(song, d.albumFolderName(albumTitle, albumYear), d.Config.DownloadPath, cb)
}

func (d *Downloader) downloadSongWithCallback(song client.Song, albumFolder, downloadPath string, cb func(DownloadProgress)) error {
	downloadURL, err := d.Client.GetDownloadURL(song.URL, d.Config.Format)
	if err != nil {
		return err
	}

	if downloadURL == "" {
		return fmt.Errorf("could not find download URL for %s", song.Title)
	}

	if cb != nil {
		cb(DownloadProgress{Total: 1, Completed: 0, Current: song.Title})
	}

	// Prepare path
	downloadPath = expandHome(downloadPath)
	albumPath := filepath.Join(downloadPath, albumFolder)
	if err := os.MkdirAll(albumPath, 0755); err != nil {
		return err
	}

	// Determine filename
	ext := ".mp3"
	if strings.HasSuffix(strings.ToLower(downloadURL), ".flac") {
		ext = ".flac"
	}

	filename := d.trackFilename(song, ext)
	filePath := filepath.Join(albumPath, filename)

	err = d.downloadFile(downloadURL, filePath)
	if err == nil && cb != nil {
		cb(DownloadProgress{Total: 1, Completed: 1, Current: song.Title})
	}
	return err
}

func (d *Downloader) DownloadAlbum(albumURL string) error {
	return d.DownloadAlbumWithCallback(albumURL, nil)
}

func (d *Downloader) DownloadAlbumWithCallback(albumURL string, cb func(DownloadProgress)) error {
	return d.downloadAlbumWithCallback(albumURL, d.Config.DownloadPath, cb)
}

// DownloadAlbumInSubdirectoryWithCallback downloads an album beneath a named
// folder in the configured destination. Publisher bulk jobs use this to keep
// their albums grouped together.
func (d *Downloader) DownloadAlbumInSubdirectoryWithCallback(albumURL, subdirectory string, cb func(DownloadProgress)) error {
	downloadPath := d.Config.DownloadPath
	if subdirectory != "" {
		downloadPath = filepath.Join(downloadPath, sanitizeFilename(subdirectory))
	}
	return d.downloadAlbumWithCallback(albumURL, downloadPath, cb)
}

func (d *Downloader) downloadAlbumWithCallback(albumURL, downloadPath string, cb func(DownloadProgress)) error {
	details, err := d.Client.GetAlbumDetails(albumURL)
	if err != nil {
		return err
	}

	total := len(details.Songs)
	if d.Config.DownloadImages {
		total += len(details.Images)
	}
	for i, song := range details.Songs {
		if cb != nil {
			cb(DownloadProgress{Total: total, Completed: i, Current: song.Title})
		}
		albumFolder := d.albumFolderName(details.Title, details.Year)
		if err := d.downloadSongWithCallback(song, albumFolder, downloadPath, nil); err != nil {
			// Log error but continue
		}
	}

	if d.Config.DownloadImages {
		// Every album download with image downloads enabled gets an Images
		// directory, even if the source album does not have any images.
		albumPath := filepath.Join(expandHome(downloadPath), d.albumFolderName(details.Title, details.Year))
		imagesPath := filepath.Join(albumPath, "Images")
		if err := os.MkdirAll(imagesPath, 0755); err != nil {
			return err
		}

		usedFilenames := make(map[string]int)
		for i, image := range details.Images {
			filename := imageFilename(image.URL, i+1, usedFilenames)
			if cb != nil {
				cb(DownloadProgress{
					Total:     total,
					Completed: len(details.Songs) + i,
					Current:   fmt.Sprintf("Image %d/%d: %s", i+1, len(details.Images), filename),
				})
			}
			if err := d.downloadFile(image.URL, filepath.Join(imagesPath, filename)); err != nil {
				// Match the existing album behaviour: keep downloading the remaining files.
				continue
			}
		}
	}

	if cb != nil {
		cb(DownloadProgress{Total: total, Completed: total, Current: "Done"})
	}

	return nil
}

func (d *Downloader) trackFilename(song client.Song, extension string) string {
	return sanitizeFilename(applyNameTemplate(d.trackFilenameFormat(), map[string]string{
		"track_number": song.Number,
		"track_name":   song.Title,
	})) + extension
}

func (d *Downloader) albumFolderName(title, year string) string {
	name := sanitizeFilename(applyNameTemplate(d.albumFolderFormat(), map[string]string{
		"year":       year,
		"album_name": title,
	}))
	name = strings.Trim(name, " -_")
	if name == "" {
		return sanitizeFilename(title)
	}
	return name
}

func (d *Downloader) trackFilenameFormat() string {
	if strings.TrimSpace(d.Config.TrackFilenameFormat) == "" {
		return "{track_number} - {track_name}"
	}
	return d.Config.TrackFilenameFormat
}

func (d *Downloader) albumFolderFormat() string {
	if strings.TrimSpace(d.Config.AlbumFolderFormat) == "" {
		return "{year} - {album_name}"
	}
	return d.Config.AlbumFolderFormat
}

func applyNameTemplate(template string, replacements map[string]string) string {
	for placeholder, value := range replacements {
		template = strings.ReplaceAll(template, "{"+placeholder+"}", value)
	}
	return template
}

func imageFilename(imageURL string, sequence int, used map[string]int) string {
	parsed, err := url.Parse(imageURL)
	filename := ""
	if err == nil && strings.Trim(parsed.Path, "/") != "" {
		filename, _ = url.PathUnescape(path.Base(parsed.Path))
	}
	filename = sanitizeFilename(filename)
	if filename == "" || filename == "." {
		filename = fmt.Sprintf("image-%02d", sequence)
	}

	key := strings.ToLower(filename)
	used[key]++
	if used[key] == 1 {
		return filename
	}

	ext := filepath.Ext(filename)
	base := strings.TrimSuffix(filename, ext)
	return fmt.Sprintf("%s-%d%s", base, used[key], ext)
}

func (d *Downloader) downloadFile(url string, path string) error {
	req, err := fhttp.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	d.Client.SetBrowserHeaders(req)
	req.Header.Set("Referer", "https://downloads.khinsider.com/")

	resp, err := d.Client.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func sanitizeFilename(name string) string {
	name = strings.ReplaceAll(name, "/", "-")
	name = strings.ReplaceAll(name, "\\", "-")
	name = strings.ReplaceAll(name, ":", "-")
	name = strings.ReplaceAll(name, "*", "-")
	name = strings.ReplaceAll(name, "?", "-")
	name = strings.ReplaceAll(name, "\"", "-")
	name = strings.ReplaceAll(name, "<", "-")
	name = strings.ReplaceAll(name, ">", "-")
	name = strings.ReplaceAll(name, "|", "-")
	return strings.TrimSpace(name)
}

func expandHome(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, path[2:])
	}
	return path
}
