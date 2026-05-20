package tui

import "testing"

func TestThemeCycleWraps(t *testing.T) {
	if got := applyTheme(0); got != 0 {
		t.Fatalf("applyTheme(0) = %d, want 0", got)
	}
	if got := applyTheme(len(themes)); got != 0 {
		t.Fatalf("applyTheme(len) should wrap to 0, got %d", got)
	}
	if got := applyTheme(-1); got != len(themes)-1 {
		t.Fatalf("applyTheme(-1) should wrap to last, got %d", got)
	}
}

func TestEachThemeHasValidColors(t *testing.T) {
	for _, th := range themes {
		applyThemeByValue(th)
		// gradStops must populate so gradAt never falls back.
		if len(gradStops) == 0 {
			t.Fatalf("theme %q produced no gradient stops", th.Name)
		}
		// gradAt across the range must return non-empty colors.
		for _, p := range []float64{0, 0.5, 1} {
			if gradAt(p) == "" {
				t.Fatalf("theme %q gradAt(%.1f) empty", th.Name, p)
			}
		}
	}
}

func TestRegisterThemes(t *testing.T) {
	before := len(themes)
	good := Theme{
		Name: "TestCustom", BG: "#000000", Dim: "#888888", Text: "#ffffff",
		Pink: "#ff00ff", Cyan: "#00ffff", Purple: "#8800ff",
		Yellow: "#ffff00", Green: "#00ff00", Red: "#ff0000",
		Grad: []string{"#00ffff", "#ff0000"},
	}
	badHex := good
	badHex.Name = "BadHex"
	badHex.Cyan = "not-a-color"
	badGrad := good
	badGrad.Name = "BadGrad"
	badGrad.Grad = []string{"#00ffff"} // too few stops

	registerThemes([]Theme{good, badHex, badGrad, good}) // duplicate good ignored
	if got := len(themes) - before; got != 1 {
		t.Fatalf("expected exactly 1 theme registered, got %d", got)
	}
	if !themeExists("TestCustom") {
		t.Error("valid custom theme should be registered")
	}
	if themeExists("BadHex") || themeExists("BadGrad") {
		t.Error("invalid themes should be rejected")
	}
}

// applyThemeByValue applies a literal theme for testing without index math.
func applyThemeByValue(th Theme) {
	for i, t := range themes {
		if t.Name == th.Name {
			applyTheme(i)
			return
		}
	}
}
