package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func newReadyModel() Model {
	var m tea.Model = New()
	m, _ = m.Update(tea.WindowSizeMsg{Width: 110, Height: 40})
	m, _ = m.Update(statsMsg(fakeStats()))
	return m.(Model)
}

func key(s string) tea.KeyMsg {
	switch s {
	case "down":
		return tea.KeyMsg{Type: tea.KeyDown}
	case "up":
		return tea.KeyMsg{Type: tea.KeyUp}
	case "enter":
		return tea.KeyMsg{Type: tea.KeyEnter}
	case "esc":
		return tea.KeyMsg{Type: tea.KeyEsc}
	default:
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
	}
}

func send(m Model, keys ...string) Model {
	var tm tea.Model = m
	for _, k := range keys {
		tm, _ = tm.Update(key(k))
	}
	return tm.(Model)
}

func TestCursorShows(t *testing.T) {
	m := newReadyModel()
	if !strings.Contains(m.View(), "▶") {
		t.Fatal("expected cursor marker in process list")
	}
}

func TestFilterNarrowsList(t *testing.T) {
	m := newReadyModel()
	m = send(m, "/", "C", "o", "d", "e")
	if got := len(m.filteredProcs()); got != 1 {
		t.Fatalf("filter 'Code' should match 1 proc, got %d", got)
	}
	if !strings.Contains(m.View(), "/Code") {
		t.Error("view should show active filter")
	}
	// esc clears the filter.
	m = send(m, "esc")
	if len(m.filteredProcs()) != 3 {
		t.Error("esc should clear filter")
	}
}

func TestKillConfirmAppears(t *testing.T) {
	m := newReadyModel() // cursor on chrome.exe (top CPU)
	m = send(m, "k")
	if !m.confirmKill {
		t.Fatal("k should arm kill confirmation")
	}
	if m.killName != "chrome.exe" {
		t.Errorf("expected kill target chrome.exe, got %q", m.killName)
	}
	if !strings.Contains(m.View(), "Kill") {
		t.Error("footer should show kill prompt")
	}
	// n cancels.
	m = send(m, "n")
	if m.confirmKill {
		t.Error("n should cancel kill")
	}
}

func TestHelpToggles(t *testing.T) {
	m := newReadyModel()
	m = send(m, "?")
	if !m.showHelp || !strings.Contains(m.View(), "keys") {
		t.Fatal("? should open help overlay")
	}
	m = send(m, "x")
	if m.showHelp {
		t.Error("any key should close help")
	}
}

func TestProcessDetailOpens(t *testing.T) {
	m := newReadyModel()
	m = send(m, "enter")
	if !m.showDetail {
		t.Fatal("enter should open the detail overlay")
	}
	if !strings.Contains(m.View(), "to close") {
		t.Error("detail overlay should render")
	}
	m = send(m, "esc")
	if m.showDetail {
		t.Error("esc should close the detail overlay")
	}
}

func TestSortModesAndReverse(t *testing.T) {
	m := newReadyModel()
	m = send(m, "p")
	if m.sort != sortPID {
		t.Fatalf("p should set sort=PID, got %v", m.sort)
	}
	if !strings.Contains(m.View(), "PID↓") {
		t.Error("active PID column should show ↓")
	}
	m = send(m, "r")
	if !m.reverse {
		t.Error("r should toggle reverse")
	}
	if !strings.Contains(m.View(), "PID↑") {
		t.Error("reversed sort should show ↑")
	}

	m = newReadyModel()
	m = send(m, "n")
	if got := m.filteredProcs()[0].Name; got != "chrome.exe" {
		t.Errorf("name sort: first should be chrome.exe, got %q", got)
	}
}

func TestPauseAndRefresh(t *testing.T) {
	m := newReadyModel()
	m = send(m, " ")
	if !m.paused {
		t.Error("space should pause")
	}
	base := m.refresh
	m = send(m, "+")
	if m.refresh >= base {
		t.Error("+ should shorten refresh interval")
	}
	m = send(m, "-", "-")
	if m.refresh <= base {
		t.Error("- should lengthen refresh interval")
	}
}
