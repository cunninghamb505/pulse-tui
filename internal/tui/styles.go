package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/lucasb-eyer/go-colorful"
)

// Theme is a named, vibrant color palette. Grad lists gradient stops from low
// to high utilization and drives every bar and sparkline.
type Theme struct {
	Name   string
	BG     string
	Dim    string
	Text   string
	Pink   string
	Cyan   string
	Purple string
	Yellow string
	Green  string
	Red    string
	Grad   []string
}

// themes is the cycle the 't' key walks through.
var themes = []Theme{
	{
		Name: "Teal", BG: "#0c1c24", Dim: "#7fa6b8", Text: "#eafcff",
		Pink: "#ff5fbf", Cyan: "#22f5e6", Purple: "#1fc8d6",
		Yellow: "#ffe93b", Green: "#1fff8f", Red: "#ff5577",
		Grad: []string{"#22f5e6", "#1fff8f", "#ffe93b", "#ff9f1c", "#ff5577"},
	},
	{
		Name: "Vapor", BG: "#0d1b2a", Dim: "#7fb3d5", Text: "#f4faff",
		Pink: "#ff5edb", Cyan: "#00e5ff", Purple: "#7c5cff",
		Yellow: "#ffe45e", Green: "#3dffd6", Red: "#ff5577",
		Grad: []string{"#00e5ff", "#3dffd6", "#ffe45e", "#ff8c42", "#ff5edb"},
	},
	{
		Name: "Matrix", BG: "#001100", Dim: "#3f9f5f", Text: "#caffd0",
		Pink: "#7dff5e", Cyan: "#39ff14", Purple: "#00ff9f",
		Yellow: "#ccff33", Green: "#39ff14", Red: "#aaff00",
		Grad: []string{"#003b00", "#19a319", "#39ff14", "#9dff4d", "#ccff33"},
	},
	{
		Name: "Inferno", BG: "#1a0a05", Dim: "#c98a5e", Text: "#fff4e6",
		Pink: "#ff2e88", Cyan: "#ffd24c", Purple: "#ff5e3a",
		Yellow: "#ffe14c", Green: "#ffae42", Red: "#ff1f3d",
		Grad: []string{"#ffe14c", "#ff9f1c", "#ff5e3a", "#ff1f3d", "#c1121f"},
	},
	{
		Name: "Aurora", BG: "#0a1a2f", Dim: "#7fa6c9", Text: "#eafff7",
		Pink: "#ff7ad9", Cyan: "#4cf0ff", Purple: "#9d7bff",
		Yellow: "#e6ff66", Green: "#3dffb0", Red: "#ff6b8a",
		Grad: []string{"#4cf0ff", "#3dffb0", "#9d7bff", "#ff7ad9", "#ff6b8a"},
	},
}

// Active palette colors, rebuilt by applyTheme.
var (
	colBG     lipgloss.Color
	colDim    lipgloss.Color
	colText   lipgloss.Color
	colPink   lipgloss.Color
	colCyan   lipgloss.Color
	colPurple lipgloss.Color
	colYellow lipgloss.Color
	colGreen  lipgloss.Color
	colRed    lipgloss.Color

	gradStops []colorful.Color
)

// Active styles, rebuilt by applyTheme.
var (
	appStyle        lipgloss.Style
	titleStyle      lipgloss.Style
	subtitleStyle   lipgloss.Style
	panelTitleStyle lipgloss.Style
	panelStyle      lipgloss.Style
	labelStyle      lipgloss.Style
	valueStyle      lipgloss.Style
	footerStyle     lipgloss.Style
	keyStyle        lipgloss.Style
	procHeaderStyle lipgloss.Style
	procNameStyle   lipgloss.Style
)

func init() { applyTheme(0) }

// applyTheme makes themes[i] active, wrapping the index, and rebuilds all
// derived colors and styles. Returns the resolved index.
func applyTheme(i int) int {
	n := len(themes)
	i = ((i % n) + n) % n
	t := themes[i]

	colBG = lipgloss.Color(t.BG)
	colDim = lipgloss.Color(t.Dim)
	colText = lipgloss.Color(t.Text)
	colPink = lipgloss.Color(t.Pink)
	colCyan = lipgloss.Color(t.Cyan)
	colPurple = lipgloss.Color(t.Purple)
	colYellow = lipgloss.Color(t.Yellow)
	colGreen = lipgloss.Color(t.Green)
	colRed = lipgloss.Color(t.Red)

	gradStops = gradStops[:0]
	for _, h := range t.Grad {
		gradStops = append(gradStops, hex(h))
	}

	appStyle = lipgloss.NewStyle().Padding(0, 1)
	titleStyle = lipgloss.NewStyle().Foreground(colBG).Background(colPink).Bold(true).Padding(0, 2)
	subtitleStyle = lipgloss.NewStyle().Foreground(colPurple).Italic(true)
	panelTitleStyle = lipgloss.NewStyle().Foreground(colCyan).Bold(true)
	panelStyle = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(colPurple).Padding(0, 1)
	labelStyle = lipgloss.NewStyle().Foreground(colDim)
	valueStyle = lipgloss.NewStyle().Foreground(colText).Bold(true)
	footerStyle = lipgloss.NewStyle().Foreground(colDim)
	keyStyle = lipgloss.NewStyle().Foreground(colYellow).Bold(true)
	procHeaderStyle = lipgloss.NewStyle().Foreground(colPink).Bold(true)
	procNameStyle = lipgloss.NewStyle().Foreground(colText)

	return i
}

func hex(s string) colorful.Color {
	c, _ := colorful.Hex(s)
	return c
}

// gradAt returns the gradient color at position t in [0,1].
func gradAt(t float64) lipgloss.Color {
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}
	if len(gradStops) == 0 {
		return colText
	}
	seg := t * float64(len(gradStops)-1)
	i := int(seg)
	if i >= len(gradStops)-1 {
		return lipgloss.Color(gradStops[len(gradStops)-1].Hex())
	}
	frac := seg - float64(i)
	blended := gradStops[i].BlendHcl(gradStops[i+1], frac).Clamped()
	return lipgloss.Color(blended.Hex())
}

// gradientBar renders a [████░░░░] bar whose filled cells follow the gradient.
func gradientBar(width int, pct float64) string {
	if width < 1 {
		width = 1
	}
	frac := pct / 100.0
	if frac < 0 {
		frac = 0
	}
	if frac > 1 {
		frac = 1
	}
	filled := int(frac * float64(width))

	var b strings.Builder
	for i := 0; i < width; i++ {
		if i < filled {
			pos := float64(i) / (float64(width-1) + 1e-9)
			b.WriteString(lipgloss.NewStyle().Foreground(gradAt(pos)).Render("█"))
		} else {
			b.WriteString(labelStyle.Render("░"))
		}
	}
	return b.String()
}

var sparkRunes = []rune("▁▂▃▄▅▆▇█")

// sparkline renders recent history as colored block characters scaled 0..max.
func sparkline(values []float64, max float64) string {
	if max <= 0 {
		max = 1
	}
	var b strings.Builder
	for _, v := range values {
		t := v / max
		if t < 0 {
			t = 0
		}
		if t > 1 {
			t = 1
		}
		idx := int(t * float64(len(sparkRunes)-1))
		b.WriteString(lipgloss.NewStyle().Foreground(gradAt(t)).Render(string(sparkRunes[idx])))
	}
	return b.String()
}
