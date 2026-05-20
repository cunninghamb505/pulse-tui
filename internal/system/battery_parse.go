package system

import (
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var batteryPctRe = regexp.MustCompile(`(\d+)%`)

// parsePmset extracts battery state from `pmset -g batt` output (macOS).
func parsePmset(s string) *BatteryInfo {
	m := batteryPctRe.FindStringSubmatch(s)
	if m == nil {
		return nil
	}
	pct, err := strconv.Atoi(m[1])
	if err != nil {
		return nil
	}
	return &BatteryInfo{
		Percent:  float64(pct),
		Charging: strings.Contains(s, "AC Power"),
	}
}

// readBatterySysfs reads the first battery under a Linux power-supply directory.
// Parameterized on base so it can be tested with a fixture directory.
func readBatterySysfs(base string) *BatteryInfo {
	entries, err := os.ReadDir(base)
	if err != nil {
		return nil
	}
	for _, e := range entries {
		if !strings.HasPrefix(strings.ToUpper(e.Name()), "BAT") {
			continue
		}
		dir := filepath.Join(base, e.Name())
		capacity, err := os.ReadFile(filepath.Join(dir, "capacity"))
		if err != nil {
			continue
		}
		pct, err := strconv.Atoi(strings.TrimSpace(string(capacity)))
		if err != nil {
			continue
		}
		status, _ := os.ReadFile(filepath.Join(dir, "status"))
		st := strings.TrimSpace(string(status))
		return &BatteryInfo{
			Percent:  float64(pct),
			Charging: st == "Charging" || st == "Full",
		}
	}
	return nil
}
