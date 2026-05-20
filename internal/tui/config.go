package tui

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// config is the persisted user preferences, written on quit.
type config struct {
	Theme     int     `json:"theme"`
	RefreshMs int     `json:"refresh_ms"`
	Sort      string  `json:"sort"`
	Reverse   bool    `json:"reverse"`
	Themes    []Theme `json:"themes,omitempty"` // user-defined themes
}

func configPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "pulse", "config.json"), nil
}

func loadConfig() config {
	c := config{
		Theme:     0,
		RefreshMs: int(defaultRefresh / time.Millisecond),
		Sort:      "cpu",
	}
	p, err := configPath()
	if err != nil {
		return c
	}
	data, err := os.ReadFile(p)
	if err != nil {
		return c
	}
	_ = json.Unmarshal(data, &c)
	return c
}

// save persists the current preferences. Errors are ignored — a missing config
// just means defaults next time.
func (m Model) saveConfig() {
	p, err := configPath()
	if err != nil {
		return
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return
	}
	// Start from the existing config so user-defined themes are preserved.
	c := loadConfig()
	c.Theme = m.theme
	c.RefreshMs = int(m.refresh / time.Millisecond)
	c.Sort = m.sort.configString()
	c.Reverse = m.reverse
	if data, err := json.MarshalIndent(c, "", "  "); err == nil {
		_ = os.WriteFile(p, data, 0o644)
	}
}
