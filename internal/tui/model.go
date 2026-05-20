// Package tui implements the Bubble Tea interface for pulse.
package tui

import (
	"fmt"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/brand/pulse/internal/system"
)

const (
	defaultRefresh = 1500 * time.Millisecond
	minRefresh     = 250 * time.Millisecond
	maxRefresh     = 10 * time.Second
	refreshStep    = 250 * time.Millisecond
	historyLen     = 60 // sparkline sample count
)

type sortMode int

const (
	sortCPU sortMode = iota
	sortMem
)

func (s sortMode) String() string {
	if s == sortMem {
		return "MEM"
	}
	return "CPU"
}

type statsMsg system.Stats
type tickMsg time.Time

// Model is the root Bubble Tea model.
type Model struct {
	collector *system.Collector
	stats     system.Stats
	ready     bool

	width  int
	height int

	sort       sortMode
	cursor     int // index into the filtered process list
	procOffset int // top of the visible process window
	theme      int
	refresh    time.Duration
	paused     bool

	filter    string
	filtering bool // capturing filter text

	confirmKill bool
	killPID     int32
	killName    string
	statusMsg   string

	showHelp bool

	cpuHist  []float64
	memHist  []float64
	upHist   []float64
	downHist []float64
}

// New builds the initial model, restoring saved preferences.
func New() Model {
	cfg := loadConfig()
	m := Model{collector: system.NewCollector()}

	m.theme = applyTheme(cfg.Theme)

	r := time.Duration(cfg.RefreshMs) * time.Millisecond
	if r < minRefresh {
		r = minRefresh
	}
	if r > maxRefresh {
		r = maxRefresh
	}
	m.refresh = r

	if cfg.Sort == "mem" {
		m.sort = sortMem
	}
	return m
}

func (m Model) Init() tea.Cmd {
	return collectCmd(m.collector)
}

func collectCmd(c *system.Collector) tea.Cmd {
	return func() tea.Msg {
		return statsMsg(c.Collect())
	}
}

func scheduleTick(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)

	case statsMsg:
		m.stats = system.Stats(msg)
		m.ready = true
		m.sortProcs()
		m.clampCursor()
		m.pushHistory()
		return m, scheduleTick(m.refresh)

	case tickMsg:
		if m.paused {
			return m, scheduleTick(m.refresh)
		}
		return m, collectCmd(m.collector)
	}

	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	if key == "ctrl+c" {
		m.saveConfig()
		return m, tea.Quit
	}

	if m.showHelp {
		if key == "q" {
			m.saveConfig()
			return m, tea.Quit
		}
		m.showHelp = false
		return m, nil
	}

	if m.confirmKill {
		switch key {
		case "y", "Y", "enter":
			if err := system.Kill(m.killPID); err != nil {
				m.statusMsg = fmt.Sprintf("✗ kill %s (%d) failed: %v", m.killName, m.killPID, err)
			} else {
				m.statusMsg = fmt.Sprintf("✓ killed %s (%d)", m.killName, m.killPID)
			}
			m.confirmKill = false
		case "n", "N", "esc":
			m.confirmKill = false
			m.statusMsg = "kill cancelled"
		}
		return m, nil
	}

	if m.filtering {
		switch msg.Type {
		case tea.KeyEnter:
			m.filtering = false
		case tea.KeyEsc:
			m.filtering = false
			m.filter = ""
		case tea.KeyBackspace:
			if r := []rune(m.filter); len(r) > 0 {
				m.filter = string(r[:len(r)-1])
			}
		case tea.KeySpace:
			m.filter += " "
		case tea.KeyRunes:
			m.filter += string(msg.Runes)
		}
		m.clampCursor()
		return m, nil
	}

	// Normal mode. Status messages are transient — clear on the next keypress.
	m.statusMsg = ""
	switch key {
	case "q":
		m.saveConfig()
		return m, tea.Quit
	case "?":
		m.showHelp = true
	case "c":
		m.sort = sortCPU
		m.sortProcs()
	case "m":
		m.sort = sortMem
		m.sortProcs()
	case "up":
		m.moveCursor(-1)
	case "down":
		m.moveCursor(1)
	case "pgup":
		m.moveCursor(-10)
	case "pgdown":
		m.moveCursor(10)
	case "home", "g":
		m.cursor = 0
		m.procOffset = 0
	case "end", "G":
		m.moveCursor(len(m.filteredProcs()))
	case "k":
		procs := m.filteredProcs()
		if m.cursor >= 0 && m.cursor < len(procs) {
			m.killPID = procs[m.cursor].PID
			m.killName = procs[m.cursor].Name
			m.confirmKill = true
		}
	case "/":
		m.filtering = true
	case "esc":
		m.filter = ""
		m.clampCursor()
	case "t":
		m.theme = applyTheme(m.theme + 1)
	case "T":
		m.theme = applyTheme(m.theme - 1)
	case " ":
		m.paused = !m.paused
	case "+", "=":
		m.refresh -= refreshStep
		if m.refresh < minRefresh {
			m.refresh = minRefresh
		}
	case "-", "_":
		m.refresh += refreshStep
		if m.refresh > maxRefresh {
			m.refresh = maxRefresh
		}
	}
	return m, nil
}

func (m Model) filteredProcs() []system.ProcInfo {
	if m.filter == "" {
		return m.stats.Procs
	}
	f := strings.ToLower(m.filter)
	out := make([]system.ProcInfo, 0, len(m.stats.Procs))
	for _, p := range m.stats.Procs {
		if strings.Contains(strings.ToLower(p.Name), f) {
			out = append(out, p)
		}
	}
	return out
}

func (m *Model) moveCursor(delta int) {
	n := len(m.filteredProcs())
	if n == 0 {
		m.cursor, m.procOffset = 0, 0
		return
	}
	m.cursor += delta
	if m.cursor < 0 {
		m.cursor = 0
	}
	if m.cursor >= n {
		m.cursor = n - 1
	}
	m.followCursor()
}

// followCursor scrolls the visible window so the cursor stays in view.
func (m *Model) followCursor() {
	visible := m.procRows()
	if m.cursor < m.procOffset {
		m.procOffset = m.cursor
	}
	if m.cursor > m.procOffset+visible-1 {
		m.procOffset = m.cursor - visible + 1
	}
	if m.procOffset < 0 {
		m.procOffset = 0
	}
}

func (m *Model) clampCursor() {
	n := len(m.filteredProcs())
	if m.cursor >= n {
		m.cursor = n - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
	m.followCursor()
}

func (m *Model) sortProcs() {
	p := m.stats.Procs
	switch m.sort {
	case sortMem:
		sort.Slice(p, func(i, j int) bool { return p[i].MemRSS > p[j].MemRSS })
	default:
		sort.Slice(p, func(i, j int) bool {
			if p[i].CPU != p[j].CPU {
				return p[i].CPU > p[j].CPU
			}
			return p[i].MemRSS > p[j].MemRSS
		})
	}
}

func (m *Model) pushHistory() {
	m.cpuHist = appendCapped(m.cpuHist, m.stats.CPUPercent)
	m.memHist = appendCapped(m.memHist, m.stats.MemPercent)
	m.upHist = appendCapped(m.upHist, m.stats.NetUpRate)
	m.downHist = appendCapped(m.downHist, m.stats.NetDownRate)
}

func appendCapped(s []float64, v float64) []float64 {
	s = append(s, v)
	if len(s) > historyLen {
		s = s[len(s)-historyLen:]
	}
	return s
}
