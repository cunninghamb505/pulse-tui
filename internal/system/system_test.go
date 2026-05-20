package system

import (
	"testing"
	"time"
)

func TestCollectRealMachine(t *testing.T) {
	c := NewCollector()
	c.Collect() // prime rate baselines
	time.Sleep(300 * time.Millisecond)
	s := c.Collect()

	if s.NumCores < 1 {
		t.Fatalf("expected >=1 core, got %d", s.NumCores)
	}
	if s.MemTotal == 0 {
		t.Fatal("expected non-zero total memory")
	}
	if len(s.PerCPU) == 0 {
		t.Fatal("expected per-core CPU readings")
	}
	if s.NumProcs == 0 {
		t.Fatal("expected at least one process")
	}
	if len(s.Procs) == 0 {
		t.Fatal("expected process rows")
	}
	t.Logf("cores=%d cpu=%.1f%% mem=%d/%d MiB procs=%d up=%s top=%q(%.1f%%)",
		s.NumCores, s.CPUPercent, s.MemUsed>>20, s.MemTotal>>20,
		s.NumProcs, s.Uptime, s.Procs[0].Name, s.Procs[0].CPU)
	t.Logf("diskIO read=%.0f B/s write=%.0f B/s | temp=%.1f°C(ok=%v) | battery=%v | disks=%d",
		s.DiskReadRate, s.DiskWriteRate, s.CPUTemp, s.HasTemp, s.Battery, len(s.Disks))
}
