package tui

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/cunninghamb505/pulse-tui/internal/system"
)

func fakeStats() system.Stats {
	return system.Stats{
		CPUPercent: 37.4,
		PerCPU:     []float64{12, 88, 45, 5, 99, 30, 60, 20},
		CPUModel:   "Test CPU @ 3.20GHz",
		NumCores:   8,
		MemUsed:    9 << 30, MemTotal: 16 << 30, MemPercent: 56.2,
		SwapUsed: 1 << 30, SwapTotal: 4 << 30, SwapPercent: 25,
		NetUpRate: 124000, NetDownRate: 5_600_000,
		NetUpTotal: 1 << 30, NetDownTotal: 8 << 30,
		DiskReadRate: 2_400_000, DiskWriteRate: 512_000,
		CPUTemp: 61, HasTemp: true,
		Battery:  &system.BatteryInfo{Percent: 87, Charging: true},
		NumProcs: 312,
		Uptime:   50 * time.Hour,
		Hostname: "DESKTOP-TEST", Platform: "Windows 11 Pro",
		Procs: []system.ProcInfo{
			{PID: 1234, Name: "chrome.exe", CPU: 23.5, MemRSS: 800 << 20},
			{PID: 22, Name: "pulse.exe", CPU: 4.1, MemRSS: 12 << 20},
			{PID: 9001, Name: "Code.exe", CPU: 1.2, MemRSS: 450 << 20},
		},
		Disks: []system.DiskInfo{
			{Mount: "C:", FSType: "NTFS", Used: 380 << 30, Total: 931 << 30, Percent: 40.8},
			{Mount: "D:", FSType: "NTFS", Used: 1200 << 30, Total: 1862 << 30, Percent: 64.4},
		},
	}
}

func TestViewRenders(t *testing.T) {
	var m tea.Model = New()
	m, _ = m.Update(tea.WindowSizeMsg{Width: 110, Height: 40})
	m, _ = m.Update(statsMsg(fakeStats()))

	out := m.View()
	for _, want := range []string{"PULSE", "CPU", "MEMORY", "NETWORK", "DISK", "PROCESSES", "chrome.exe", "61°C", "BAT", "87%"} {
		if !strings.Contains(out, want) {
			t.Errorf("rendered view missing %q", want)
		}
	}
	t.Logf("\n%s", out)
}
