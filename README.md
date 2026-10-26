# khi-explorer

[![Go Version](https://img.shields.io/github/go-mod/go-version/madLinux7/khi-explorer?style=flat-square)](https://go.dev/)
[![Release](https://img.shields.io/github/v/release/madLinux7/khi-explorer?style=flat-square)](https://github.com/madLinux7/khi-explorer/releases)
[![Build Status](https://img.shields.io/github/actions/workflow/status/madLinux7/khi-explorer/release.yml?style=flat-square)](https://github.com/madLinux7/khi-explorer/actions)
[![Platform](https://img.shields.io/badge/platform-linux%20%7C%20macos%20%7C%20windows-blue?style=flat-square)](https://github.com/madLinux7/khi-explorer)
[![License](https://img.shields.io/github/license/madLinux7/khi-explorer?style=flat-square)](https://github.com/madLinux7/khi-explorer/blob/main/LICENSE)

A polished standalone browser, player and bulk downloader for [KHInsider](https://downloads.khinsider.com).

![Showcase](https://artifacts.grolmes.com/khi-explorer/demo.gif)

YES, you can download entire albums with this, including the album art, separated by folders and even bulk download a full albums list.

## Features

- **Interactive TUI** — Search for OST, filter tracks and navigate with keyboard-only controls
- **Audio Playback** — Stream playback right from your terminal
- **Bulk Downloads** — Download individual tracks or entire albums by pressing **D** on the entry
- **Search & Filter** — Find games and tracks quickly with fuzzy search
- **Quick Format Switch** — Toggle FLAC/MP3 inside the app by pressing TAB
- **Configurable** — Set your preferred format and download path in `~/.khi_explorer.yaml`
- **Cross-platform** — Use on Linux, macOS and Windows

## Usage

```
khi-explorer
```

### TUI Navigation

| Key | Action |
|-----|--------|
| `F2` | Change search mode: Toggle Between OST search or Publisher |
| `↑` / `↓` or `k` / `j` | Move selection cursor up/down |
| `←` / `→` | Previous/Next page |
| `Enter` | Select album, search, or play song |
| `Esc` / `Backspace` | Go back to previous view |
| `/` | Enter filter/search mode |
| `Tab` | Toggle download format (FLAC ⇆ MP3) |
| `q` / `Ctrl+C` | Quit |

### Playback Controls

| Key | Action |
|-----|--------|
| `P` | Pause or resume playback |
| `S` | Stop playback |

### Download Controls

| Key | Action |
|-----|--------|
| `D` | Queue highlighted song for download (Album view) |
| `D` | Queue one album for download (Search results view) |
| `A` | Queue all the albums in the list for download (OST album list or Publisher album list) |
| `I` | Toggle Album art downloading ON/OFF |

### Configuration

Configuration is stored at `~/.khi_explorer.yaml`:

```yaml
format: flac                         # Download format (flac, mp3)
download_path: ~/khi_explorer        # Download destination (default windows folder is the Downloads folder)
player: mpv                          # Audio player command
```

## Installation

### Build from Source

```bash
git clone https://github.com/bahamut/khi-explorer.git
cd khi-explorer
go build -o khi-explorer .
```

## ✨ Acknowledgements ✨

- [KHInsider](https://www.khinsider.com) — For hosting such an extensive collection of video game music
- [charmbracelet/bubbletea](https://github.com/charmbracelet/bubbletea) — For the elegant TUI framework
- [charmbracelet/lipgloss](https://github.com/charmbracelet/lipgloss) — For beautiful styling primitives
- "This project is based on the original work from [madLinux7](https://github.com/madLinux7/khi-explorer)."
