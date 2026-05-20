# pulse

A modern, jazzy terminal system monitor — like `htop`/`top`, but with neon gradient
bars, sparklines, and switchable color themes. Built with
[Bubble Tea](https://github.com/charmbracelet/bubbletea) and
[gopsutil](https://github.com/shirou/gopsutil).

```
 ◢ PULSE ◤   hostname · Windows 11 Pro · up 2d 3h     BAT ↑87%   397 procs
╭─ CPU ───────────────────────╮ ╭─ MEMORY ──────────────────────╮
│ ███████████░░░░░░░░  37.4%   │ │ RAM  ████████████░░░  56.2%    │
│ AMD Ryzen … 3.20GHz   61°C   │ │ SWAP ██████░░░░░░░░░  25.0%    │
│ ▁▂▃▅▇▆▄▃  per-core bars…     │ │ ▂▃▄▅▄▃                         │
╰──────────────────────────────╯ ╰───────────────────────────────╯
╭─ NETWORK ───────────────────╮ ╭─ DISK ────────────────────────╮
│ ▲ UP 121 KB/s  ▁▂▃          │ │ R 2.4 MB/s   W 512 KB/s        │
│ ▼ DN 5.3 MB/s  ▃▅▇          │ │ C: ████████░░░░  40.8%         │
╰──────────────────────────────╯ ╰───────────────────────────────╯
╭─ PROCESSES · sort CPU · 1-12/397 ────────────────────────────────╮
│ ▶ 1234   23.5   800 MiB  chrome.exe                              │
│   …                                                              │
╰──────────────────────────────────────────────────────────────────╯
```

## Features

- Live **CPU** (overall + per-core), **memory/swap**, **network** up/down, and
  **disk** usage + read/write I/O.
- Per-core bars, sparkline history, and a color gradient that tracks load.
- **CPU temperature** and **battery** when the platform reports them.
- Interactive process list: **sort** by CPU/memory, **scroll** with a cursor,
  **filter** by name, and **kill** a process.
- **5 switchable color themes** (`t`) — Synthwave, Vapor, Matrix, Inferno, Aurora.
- Adjustable refresh rate, pause, and an in-app help overlay.
- Preferences (theme, refresh, sort) persist between runs.

## Install

Requires Go 1.21+.

```sh
go install github.com/brand/pulse@latest
```

This drops the `pulse` binary in your `GOBIN` (usually `$(go env GOPATH)/bin`).
Make sure that directory is on your `PATH`, then just run:

```sh
pulse
```

### Build from source

```sh
git clone <repo-url> pulse && cd pulse
go build -o pulse .
./pulse
```

## Keys

| Key            | Action                       |
| -------------- | ---------------------------- |
| `↑` / `↓`      | move selection               |
| `PgUp` / `PgDn`| jump 10 rows                 |
| `g` / `G`      | jump to top / bottom         |
| `c` / `m`      | sort by CPU / memory         |
| `k`            | kill selected process        |
| `/`            | filter by name (`esc` clears)|
| `space`        | pause / resume               |
| `+` / `-`      | faster / slower refresh      |
| `t` / `T`      | next / previous theme        |
| `?`            | toggle help                  |
| `q`            | quit                         |

## Platforms

| OS      | Status                                                    |
| ------- | --------------------------------------------------------- |
| Windows | Primary target. Battery via `GetSystemPowerStatus`.       |
| Linux   | Builds cleanly; battery support TODO.                     |
| macOS   | Builds cleanly; battery support TODO.                     |

CPU temperature depends on what the OS exposes and may be unavailable
(common on Windows desktops); the UI hides it when absent.

## Config

Preferences are stored at `<user config dir>/pulse/config.json`
(e.g. `%AppData%\pulse\config.json` on Windows) and written on quit.
