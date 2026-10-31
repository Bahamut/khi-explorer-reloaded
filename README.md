# KHI Explorer Reloaded

[![Go Version](https://img.shields.io/github/go-mod/go-version/madLinux7/khi-explorer?style=flat-square)](https://go.dev/)
[![Release](https://img.shields.io/github/v/release/madLinux7/khi-explorer?style=flat-square)](https://github.com/madLinux7/khi-explorer/releases)
[![Build Status](https://img.shields.io/github/actions/workflow/status/madLinux7/khi-explorer/release.yml?style=flat-square)](https://github.com/madLinux7/khi-explorer/actions)
[![Platform](https://img.shields.io/badge/platform-linux%20%7C%20macos%20%7C%20windows-blue?style=flat-square)](https://github.com/madLinux7/khi-explorer)
[![License](https://img.shields.io/github/license/madLinux7/khi-explorer?style=flat-square)](https://github.com/madLinux7/khi-explorer/blob/main/LICENSE)

Standalone browser, player and bulk downloader for [KHInsider](https://downloads.khinsider.com).
It CAN download entire album lists, **including the album art**, separated by folders and even **bulk download a full albums list from any publisher**.
Fully working since October 2026.

<p align="center">
  <img src="https://artifacts.grolmes.com/khi-explorer/demo.gif" alt="Showcase" width="60%" />
</p>

## Features

- **Interactive TUI** — Search for OST, filter tracks, publishers and navigate with keyboard-only controls.
- **Bulk Downloads** — Download individual tracks or entire albums with **D**, or all visible albums with **A**. Publisher bulk downloads are grouped in a publisher folder and album art can be saved in `Images`.
- **Search & Filter** — Find games and tracks quickly with fuzzy search, supporting search by Publisher or by OST.
- **Audio Playback** — Stream playback right from your terminal.
- **Quick Format Switch** — Toggle between FLAC/MP3 inside the app.
- **Quick Images Toggle** — Toggle Album Art downloads.
- **Configurable** — Set your preferred format and download path in `~/.khi_explorer.yaml`
- **Cross-platform** — Use on Linux, macOS and Windows.
## Usage

```
khi-explorer
```

### Navigation

| Key | Action |
|-----|--------|
| `F2` | Switch search mode (OST or publisher) |
| `↑` / `↓` or `k` / `j` | Move selection cursor up/down |
| `←` / `→` | Previous/Next page |
| `Enter` | Select album, search, or play song |
| `Esc` / `Backspace` | Go back to previous view |
| `/` | Enter filter/search mode |
| `Tab` | Toggle download format (FLAC ⇆ MP3) |
| `I` | Toggle album-image downloads (in result and album views) |
| `q` / `Ctrl+C` | Quit |

### Download Controls

| Key | Action |
|-----|--------|
| `D` | Queue highlighted song for download (Album view) |
| `D` | Queue one album for download (Search results view) |
| `A` | Queue all the albums in the list for download (OST album list or Publisher album list) |

### Playback Controls

| Key | Action |
|-----|--------|
| `P` | Pause or resume playback |
| `S` | Stop playback |

### Configuration

Configuration is stored at `~/.khi_explorer.yaml`:

```yaml
format: flac                         # Download format (flac, mp3)
download_path: ~/Downloads/khi_explorer # Windows download destination
download_images: true                # Save full-size album images in Images/
track_filename_format: "{track_number} - {track_name}"
album_folder_format: "{year} - {album_name}"
player: mpv                          # Audio player command
```

`track_filename_format` accepts `{track_number}` and `{track_name}`. `album_folder_format` accepts `{year}` and `{album_name}`. The audio extension is added automatically.

## Installation

### Build from Source

```bash
git clone https://github.com/bahamut/khi-explorer.git
cd khi-explorer
go build -o khi-explorer .
```

## ✨ Acknowledgements ✨

- [KHInsider](https://www.khinsider.com) — For hosting such an extensive collection of video game music
- [madLinux7/khi-explorer](https://github.com/madLinux7/khi-explorer) — This project is a fork of his original khi-explorer mad work!!
- [charmbracelet/bubbletea](https://github.com/charmbracelet/bubbletea) — For the elegant TUI framework
- [charmbracelet/lipgloss](https://github.com/charmbracelet/lipgloss) — For beautiful styling primitives
