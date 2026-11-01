# KHI Explorer (Reloaded Edition)

[![Go Version](https://img.shields.io/github/go-mod/go-version/Bahamut/khi-explorer-reloaded?style=flat-square)](https://go.dev/)
[![Release](https://img.shields.io/github/v/release/Bahamut/khi-explorer-reloaded?style=flat-square)](https://github.com/Bahamut/khi-explorer-reloaded/releases)
[![Build Status](https://img.shields.io/github/actions/workflow/status/Bahamut/khi-explorer-reloaded/release.yml?style=flat-square)](https://github.com/Bahamut/khi-explorer-reloaded/actions)
[![Platform](https://img.shields.io/badge/platform-linux%20%7C%20macos%20%7C%20windows-blue?style=flat-square)](https://github.com/Bahamut/khi-explorer-reloaded)
[![License](https://img.shields.io/github/license/Bahamut/khi-explorer-reloaded?style=flat-square)](https://github.com/Bahamut/khi-explorer-reloaded/blob/main/LICENSE)

Standalone browser, player and bulk downloader for [KHInsider](https://downloads.khinsider.com).
It can download individual tracks, full albums, **album art (covers & booklets)**, organize multi-disc releases, and **bulk download entire discographies from any publisher**.

## Features

- **Interactive TUI** — Search for soundtracks, filter tracklists, explore publishers, and browse with keyboard-only controls.
- **Smart Publisher Search** — Find publishers and labels even with partial names or single words.
- **Multi-disc Management** — Automatically formats multi-CD albums with `{disk}.{track_number}` (e.g. `1.01 - Title`) or cleanly separates them into subdirectories (`CD1/01 - Title`, `CD2/...`) using the **F** key toggle.
- **Bulk Downloads** — Download individual songs and full albums with **D**, or all albums listed on screen with **A**. Publisher bulk downloads are cleanly organized into dedicated directories.
- **Album Artwork** — Save full-resolution covers, booklets, and discs into an `Images/` folder with toggle **I**.
- **Audio Playback** — Stream songs directly inside the terminal with native pause/resume (**P**) and stop (**S**) controls on both Windows and Unix.
- **Embedded Application Icon** — Native Windows icon embedded directly into the binary and console window.
- **Quick Format Switch** — Toggle between lossless FLAC and MP3 on the fly with **Tab**.
- **Configurable** — Customize templates, default audio format, and download destinations in `~/.khi-reloaded.yaml`.
- **Cross-platform** — Native binaries for Linux, macOS, and Windows.

---

## Usage

```bash
khi-reloaded
```

### Navigation Controls

| Key | Action |
|-----|--------|
| `F2` | Switch search mode (OST ⇆ Publisher) |
| `↑` / `↓` or `k` / `j` | Move selection cursor up / down |
| `←` / `→` | Previous / Next page |
| `Enter` | Select album, search, or play song |
| `Esc` / `Backspace` | Go back to previous view (or clear filter) |
| `/` | Enter search filter mode in track lists |
| `Tab` | Toggle download format (FLAC ⇆ MP3) |
| `I` | Toggle album-image downloads (in result & album views) |
| `F` | Toggle multi-disc folders `CD1`, `CD2`, etc. (in result & album views) |
| `q` / `Ctrl+C` | Quit application |

### Download Controls

| Key | Action |
|-----|--------|
| `D` | Queue highlighted song for download (in Album view) |
| `D` | Queue highlighted album for download (in Search results view) |
| `A` | Queue all albums in current results for bulk download |

### Playback Controls

| Key | Action |
|-----|--------|
| `P` | Pause or resume playback |
| `S` | Stop playback |

---

## Configuration

Configuration is stored at `~/.khi-reloaded.yaml`:

```yaml
format: flac                            # Preferred audio format: flac, mp3
download_path: ~/Downloads/khi-reloaded # Download folder (Windows default)
download_images: true                   # Save full-size album artwork in Images/
disk_folders: false                     # Separate multi-disc albums into CD1, CD2, etc.
track_filename_format: "{disk}.{track_number} - {track_name}"
album_folder_format: "{year} - {album_name}"
player: mpv                             # Audio player command (mpv, ffplay, vlc)
```

- **`track_filename_format`**: supports `{disk}`, `{track_number}`, and `{track_name}`. When an album has only 1 disc or when `disk_folders` is enabled, `{disk}` is automatically omitted to keep filenames clean.
- **`album_folder_format`**: supports `{year}` and `{album_name}`.

---

## Installation

### Pre-built Binaries

Download the latest pre-compiled binary for your operating system from the [Releases](https://github.com/Bahamut/khi-explorer-reloaded/releases) page.

### Build from Source

```bash
git clone https://github.com/Bahamut/khi-explorer-reloaded.git
cd khi-explorer-reloaded
go build -o khi-reloaded .
```

---

## ✨ Acknowledgements ✨

- Fork form original project: [madLinux7/khi-explorer](https://github.com/madLinux7/khi-explorer).
- [charmbracelet/bubbletea](https://github.com/charmbracelet/bubbletea) — For the elegant TUI framework.
- [charmbracelet/lipgloss](https://github.com/charmbracelet/lipgloss) — For beautiful terminal styling primitives.
- [KHInsider](https://www.khinsider.com) — For hosting such an extensive collection of video game music.
