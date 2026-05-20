package system

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParsePmset(t *testing.T) {
	cases := []struct {
		name       string
		in         string
		wantNil    bool
		wantPct    float64
		wantCharge bool
	}{
		{
			name:       "discharging",
			in:         "Now drawing from 'Battery Power'\n -InternalBattery-0 (id=123)\t87%; discharging; 4:21 remaining present: true",
			wantPct:    87,
			wantCharge: false,
		},
		{
			name:       "charging on AC",
			in:         "Now drawing from 'AC Power'\n -InternalBattery-0 (id=123)\t54%; charging; 1:02 remaining present: true",
			wantPct:    54,
			wantCharge: true,
		},
		{
			name:       "charged on AC",
			in:         "Now drawing from 'AC Power'\n -InternalBattery-0 (id=123)\t100%; charged; 0:00 remaining present: true",
			wantPct:    100,
			wantCharge: true,
		},
		{
			name:    "no battery (desktop)",
			in:      "Now drawing from 'AC Power'\n",
			wantNil: true,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := parsePmset(c.in)
			if c.wantNil {
				if got != nil {
					t.Fatalf("expected nil, got %+v", got)
				}
				return
			}
			if got == nil {
				t.Fatal("expected battery info, got nil")
			}
			if got.Percent != c.wantPct {
				t.Errorf("percent = %v, want %v", got.Percent, c.wantPct)
			}
			if got.Charging != c.wantCharge {
				t.Errorf("charging = %v, want %v", got.Charging, c.wantCharge)
			}
		})
	}
}

func TestReadBatterySysfs(t *testing.T) {
	base := t.TempDir()
	// A non-battery supply that must be ignored, plus a real battery.
	mustWrite(t, filepath.Join(base, "AC", "online"), "1")
	mustWrite(t, filepath.Join(base, "BAT0", "capacity"), "72\n")
	mustWrite(t, filepath.Join(base, "BAT0", "status"), "Discharging\n")

	got := readBatterySysfs(base)
	if got == nil {
		t.Fatal("expected battery, got nil")
	}
	if got.Percent != 72 {
		t.Errorf("percent = %v, want 72", got.Percent)
	}
	if got.Charging {
		t.Error("expected not charging")
	}

	if readBatterySysfs(filepath.Join(base, "does-not-exist")) != nil {
		t.Error("missing base dir should yield nil")
	}
}

func mustWrite(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
