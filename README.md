# pulse

A modern, jazzy terminal system monitor — like `htop`/`top`, but with neon gradient
bars, sparklines, and switchable color themes. Built with
[Bubble Tea](https://github.com/charmbracelet/bubbletea) and
[gopsutil](https://github.com/shirou/gopsutil).

![pulse running in a terminal](docs/screenshot.png)

## Features

- Live **CPU** (overall + per-core), **memory/swap**, **network** up/down, and
  **disk** usage + read/write I/O.
- Per-core bars, sparkline history, and a color gradient that tracks load.
- **CPU temperature** and **battery** when the platform reports them.
- Interactive process list: **sort** by CPU/memory, **scroll** with a cursor,
  **filter** by name, and **kill** a process.
- **5 switchable color themes** (`t`) — Teal (default), Vapor, Matrix, Inferno, Aurora.
- Adjustable refresh rate, pause, and an in-app help overlay.
- Preferences (theme, refresh, sort) persist between runs.

## Install

### Prebuilt binaries

Download the binary for your platform from the
[latest release](https://github.com/cunninghamb505/pulse-tui/releases/latest),
then put it somewhere on your `PATH` (rename it to `pulse` if you like).

On Windows (PowerShell):

```powershell
# after downloading pulse-windows-amd64.exe
Move-Item .\pulse-windows-amd64.exe "$env:USERPROFILE\bin\pulse.exe"
```

On macOS / Linux:

```sh
chmod +x pulse-* && sudo mv pulse-* /usr/local/bin/pulse
```

### With Go (1.21+)

```sh
go install github.com/cunninghamb505/pulse-tui/cmd/pulse@latest
```

This drops the `pulse` binary in your `GOBIN` (usually `$(go env GOPATH)/bin`).
Make sure that directory is on your `PATH`, then just run:

```sh
pulse
```

### Build from source

```sh
git clone https://github.com/cunninghamb505/pulse-tui && cd pulse-tui
go build -o pulse ./cmd/pulse
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

| OS      | Battery source                                  | Notes                          |
| ------- | ----------------------------------------------- | ------------------------------ |
| Windows | `GetSystemPowerStatus`                          | Primary / most-tested target.  |
| Linux   | sysfs (`/sys/class/power_supply`)               | Newer — please report issues.  |
| macOS   | `pmset -g batt`                                 | Newer — please report issues.  |

CPU temperature depends on what the OS exposes and may be unavailable
(common on Windows desktops and on cgo-free macOS builds); the UI hides it
when absent. Battery is hidden on machines without one.

## Config

Preferences are stored at `<user config dir>/pulse/config.json`
(e.g. `%AppData%\pulse\config.json` on Windows) and written on quit.
