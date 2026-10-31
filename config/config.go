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
	TrackFilenameFormat string `yaml:"track_filename_format"`
	AlbumFolderFormat   string `yaml:"album_folder_format"`
	Player              string `yaml:"player"`
}

func GetConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".khi_explorer.yaml")
}

func LoadConfig() (Config, error) {
	cfg := Config{
		Format:              "flac",
		DownloadPath:        defaultDownloadPath(),
		DownloadImages:      true,
		TrackFilenameFormat: "{track_number} - {track_name}",
		AlbumFolderFormat:   "{year} - {album_name}",
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
	if err == nil && runtime.GOOS == "windows" && samePath(cfg.DownloadPath, legacyDownloadPath()) {
		cfg.DownloadPath = defaultDownloadPath()
		// Persist the one-time migration so later launches do not continue to
		// show the obsolete default path.
		if err := cfg.Save(); err != nil {
			return cfg, err
		}
	}
	return cfg, err
}

func defaultDownloadPath() string {
	home := getHomeDir()
	if runtime.GOOS == "windows" {
		return filepath.Join(home, "Downloads", "khi_explorer")
	}
	return filepath.Join(home, "khi_explorer")
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
