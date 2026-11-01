package config

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Format              string `yaml:"format"`
	DownloadPath        string `yaml:"download_path"`
	DownloadImages      bool   `yaml:"download_images"`
	DiskFolders         bool   `yaml:"disk_folders"`
	TrackFilenameFormat string `yaml:"track_filename_format"`
	AlbumFolderFormat   string `yaml:"album_folder_format"`
	Player              string `yaml:"player"`
}

const (
	DefaultTrackFilenameFormat = "{disk}.{track_number} - {track_name}"
	legacyTrackFilenameFormat  = "{track_number} - {track_name}"
	DefaultAlbumFolderFormat   = "{year} - {album_name}"
)

func GetConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".khi-reloaded.yaml")
}

func LoadConfig() (Config, error) {
	cfg := Config{
		Format:              "flac",
		DownloadPath:        defaultDownloadPath(),
		DownloadImages:      true,
		DiskFolders:         false,
		TrackFilenameFormat: DefaultTrackFilenameFormat,
		AlbumFolderFormat:   DefaultAlbumFolderFormat,
		Player:              "mpv",
	}

	path := GetConfigPath()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return cfg, err
	}

	err = yaml.Unmarshal(data, &cfg)
	if err == nil {
		migrated := false
		if runtime.GOOS == "windows" && samePath(cfg.DownloadPath, legacyDownloadPath()) {
			cfg.DownloadPath = defaultDownloadPath()
			migrated = true
		}
		if strings.TrimSpace(cfg.TrackFilenameFormat) == "" || cfg.TrackFilenameFormat == legacyTrackFilenameFormat {
			cfg.TrackFilenameFormat = DefaultTrackFilenameFormat
			migrated = true
		}
		if strings.TrimSpace(cfg.AlbumFolderFormat) == "" {
			cfg.AlbumFolderFormat = DefaultAlbumFolderFormat
			migrated = true
		}
		if migrated {
			// Persist the one-time migration so later launches do not continue to
			// show obsolete default settings.
			if err := cfg.Save(); err != nil {
				return cfg, err
			}
		}
	}
	return cfg, err
}

func defaultDownloadPath() string {
	home := getHomeDir()
	if runtime.GOOS == "windows" {
		return filepath.Join(home, "Downloads", "khi-reloaded")
	}
	return filepath.Join(home, "khi-reloaded")
}

func legacyDownloadPath() string {
	return filepath.Join(getHomeDir(), "khi_explorer")
}

func samePath(left, right string) bool {
	return strings.EqualFold(filepath.Clean(left), filepath.Clean(right))
}

func (c Config) Save() error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	return os.WriteFile(GetConfigPath(), data, 0644)
}

func getHomeDir() string {
	home, _ := os.UserHomeDir()
	return home
}
