# remdev

Web-based IDE with a tiling window manager, built as a single Go binary.

## Quick Start

```bash
remdev --root /path/to/project
# Open http://localhost:7000
```

## Features

- **Tiling window manager** — binary-tree splits, Alt+=/- resize, Alt+W close
- **Terminal** — full PTY shell via xterm.js, Alt+N to create
- **Editor** — CodeMirror 5 with syntax highlighting for 12 languages, Alt+B to open files
- **Workspaces** — 4 independent workspace groups, Alt+1-4 to switch
- **WebDAV** — standard DAV server at `/dav/`, mountable by any DAV client
- **Config editor** — edit `remdev.json` in-browser, theme/font changes apply live
- **Single binary** — static builds for Linux/macOS/Windows × amd64/arm64

## Usage

```
remdev --root /path [flags]

Flags:
  --config    path to config file (default: ~/.config/remdev/remdev.json)
  --root      file serving root directory (required)
  --port      listen port (default: 7000)
  --host      listen address (default: 0.0.0.0)
  --option k=v  override config values (repeatable)
```

## Keyboard Shortcuts

| Key | Action |
|---|---|
| Alt+N | New terminal |
| Alt+B | Open file picker |
| Alt+W | Close window |
| Alt+S | Save file |
| Alt+R | Reload file |
| Alt+= | Enlarge window |
| Alt+- | Shrink window |
| Alt+1-4 | Switch workspace |
| Alt+arrows | Navigate focus |

## Build

```bash
make          # all 6 platforms
go build -C src -o dist/remdev .
```

Output: `dist/remdev-{linux,darwin,windows}-{amd64,arm64}[.exe]`

## License

MIT
