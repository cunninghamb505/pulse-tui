// Package system collects cross-platform machine statistics via gopsutil.
package system

import (
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/net"
	"github.com/shirou/gopsutil/v4/process"
	"github.com/shirou/gopsutil/v4/sensors"
)

// Stats is a single snapshot of the machine's vitals.
type Stats struct {
	CPUPercent float64   // overall CPU busy %
	PerCPU     []float64 // per-core busy %
	CPUModel   string
	NumCores   int

	MemUsed     uint64
	MemTotal    uint64
	MemPercent  float64
	SwapUsed    uint64
	SwapTotal   uint64
	SwapPercent float64

	NetUpRate    float64 // bytes/sec out
	NetDownRate  float64 // bytes/sec in
	NetUpTotal   uint64
	NetDownTotal uint64

	DiskReadRate  float64 // bytes/sec read across all devices
	DiskWriteRate float64 // bytes/sec written across all devices

	CPUTemp float64 // Celsius
	HasTemp bool

	Battery *BatteryInfo // nil if no battery present

	Procs    []ProcInfo
	NumProcs int

	Disks []DiskInfo

	Uptime   time.Duration
	Hostname string
	Platform string
}

// BatteryInfo describes the system battery, when present.
type BatteryInfo struct {
	Percent  float64
	Charging bool
}

// Kill terminates the process with the given PID.
func Kill(pid int32) error {
	p, err := process.NewProcess(pid)
	if err != nil {
		return err
	}
	return p.Kill()
}

// ProcDetail is the richer, on-demand view of a single process.
type ProcDetail struct {
	PID        int32
	Name       string
	Exe        string
	Cmdline    string
	Ppid       int32
	ParentName string
	NumThreads int32
	Username   string
	Status     string
	CreateTime time.Time
	RunTime    time.Duration
	CPU        float64
	MemRSS     uint64
	MemPct     float32
}

// ProcessDetail fetches detailed information for a single PID on demand. It is
// deliberately not collected for every process each tick (too many syscalls).
func ProcessDetail(pid int32) (ProcDetail, error) {
	p, err := process.NewProcess(pid)
	if err != nil {
		return ProcDetail{}, err
	}
	d := ProcDetail{PID: pid}
	d.Name, _ = p.Name()
	d.Exe, _ = p.Exe()
	d.Cmdline, _ = p.Cmdline()
	d.Ppid, _ = p.Ppid()
	if d.Ppid > 0 {
		if parent, err := process.NewProcess(d.Ppid); err == nil {
			d.ParentName, _ = parent.Name()
		}
	}
	d.NumThreads, _ = p.NumThreads()
	d.Username, _ = p.Username()
	if st, err := p.Status(); err == nil {
		d.Status = strings.Join(st, ",")
	}
	if ms, err := p.CreateTime(); err == nil && ms > 0 {
		d.CreateTime = time.UnixMilli(ms)
		d.RunTime = time.Since(d.CreateTime)
	}
	if cp, err := p.CPUPercent(); err == nil {
		d.CPU = cp
	}
	if mi, err := p.MemoryInfo(); err == nil && mi != nil {
		d.MemRSS = mi.RSS
	}
	d.MemPct, _ = p.MemoryPercent()
	return d, nil
}

// DiskInfo is usage for one mounted filesystem.
type DiskInfo struct {
	Mount   string
	FSType  string
	Used    uint64
	Total   uint64
	Percent float64
}

// ProcInfo is a single process row.
type ProcInfo struct {
	PID    int32
	Name   string
	CPU    float64 // % of one core's worth, summed across cores
	MemRSS uint64
	MemPct float32
}

// Collector holds the previous snapshot's state so it can compute rates.
type Collector struct {
	prevNetTime   time.Time
	prevBytesSent uint64
	prevBytesRecv uint64

	prevProcCPU  map[int32]float64
	prevProcTime time.Time

	prevDiskTime  time.Time
	prevDiskRead  uint64
	prevDiskWrite uint64

	numCores int
	cpuModel string
	platform string
}

// NewCollector builds a collector and primes the values that never change.
func NewCollector() *Collector {
	c := &Collector{
		prevProcCPU: make(map[int32]float64),
		numCores:    runtime.NumCPU(),
	}
	if info, err := cpu.Info(); err == nil && len(info) > 0 {
		c.cpuModel = info[0].ModelName
	}
	if p, _, _, err := host.PlatformInformation(); err == nil {
		c.platform = p
	}
	// Prime the CPU percent baseline so the first real reading isn't garbage.
	_, _ = cpu.Percent(0, false)
	_, _ = cpu.Percent(0, true)
	return c
}

// Collect gathers a fresh snapshot. Rates are measured against the previous call.
func (c *Collector) Collect() Stats {
	now := time.Now()
	s := Stats{
		NumCores: c.numCores,
		CPUModel: c.cpuModel,
		Platform: c.platform,
	}

	if total, err := cpu.Percent(0, false); err == nil && len(total) > 0 {
		s.CPUPercent = total[0]
	}
	if per, err := cpu.Percent(0, true); err == nil {
		s.PerCPU = per
	}

	if vm, err := mem.VirtualMemory(); err == nil {
		s.MemUsed = vm.Used
		s.MemTotal = vm.Total
		s.MemPercent = vm.UsedPercent
	}
	if sw, err := mem.SwapMemory(); err == nil {
		s.SwapUsed = sw.Used
		s.SwapTotal = sw.Total
		s.SwapPercent = sw.UsedPercent
	}

	c.collectNet(&s, now)
	c.collectProcs(&s, now)
	collectDisks(&s)
	c.collectDiskIO(&s, now)
	collectTemp(&s)
	s.Battery = readBattery()

	if up, err := host.Uptime(); err == nil {
		s.Uptime = time.Duration(up) * time.Second
	}
	if hn, err := host.Info(); err == nil {
		s.Hostname = hn.Hostname
	}

	return s
}

func collectDisks(s *Stats) {
	parts, err := disk.Partitions(false)
	if err != nil {
		return
	}
	for _, p := range parts {
		u, err := disk.Usage(p.Mountpoint)
		if err != nil || u.Total == 0 {
			continue
		}
		s.Disks = append(s.Disks, DiskInfo{
			Mount:   p.Mountpoint,
			FSType:  p.Fstype,
			Used:    u.Used,
			Total:   u.Total,
			Percent: u.UsedPercent,
		})
	}
}

func (c *Collector) collectDiskIO(s *Stats, now time.Time) {
	counters, err := disk.IOCounters()
	if err != nil || len(counters) == 0 {
		return
	}
	var read, write uint64
	for _, io := range counters {
		read += io.ReadBytes
		write += io.WriteBytes
	}
	if !c.prevDiskTime.IsZero() {
		dt := now.Sub(c.prevDiskTime).Seconds()
		if dt > 0 {
			if read >= c.prevDiskRead {
				s.DiskReadRate = float64(read-c.prevDiskRead) / dt
			}
			if write >= c.prevDiskWrite {
				s.DiskWriteRate = float64(write-c.prevDiskWrite) / dt
			}
		}
	}
	c.prevDiskTime = now
	c.prevDiskRead = read
	c.prevDiskWrite = write
}

// collectTemp picks a representative CPU temperature when the platform exposes
// one. Many Windows machines report nothing here, in which case HasTemp stays
// false and the UI omits it.
func collectTemp(s *Stats) {
	temps, err := sensors.SensorsTemperatures()
	if err != nil || len(temps) == 0 {
		return
	}
	best := 0.0
	for _, t := range temps {
		if t.Temperature <= 0 {
			continue
		}
		k := strings.ToLower(t.SensorKey)
		if strings.Contains(k, "cpu") || strings.Contains(k, "core") ||
			strings.Contains(k, "k10temp") || strings.Contains(k, "package") {
			if t.Temperature > best {
				best = t.Temperature
			}
		}
	}
	if best == 0 { // no obvious CPU sensor — fall back to the hottest reading
		for _, t := range temps {
			if t.Temperature > best {
				best = t.Temperature
			}
		}
	}
	if best > 0 {
		s.CPUTemp = best
		s.HasTemp = true
	}
}

func (c *Collector) collectNet(s *Stats, now time.Time) {
	counters, err := net.IOCounters(false)
	if err != nil || len(counters) == 0 {
		return
	}
	sent := counters[0].BytesSent
	recv := counters[0].BytesRecv
	s.NetUpTotal = sent
	s.NetDownTotal = recv

	if !c.prevNetTime.IsZero() {
		dt := now.Sub(c.prevNetTime).Seconds()
		if dt > 0 {
			if sent >= c.prevBytesSent {
				s.NetUpRate = float64(sent-c.prevBytesSent) / dt
			}
			if recv >= c.prevBytesRecv {
				s.NetDownRate = float64(recv-c.prevBytesRecv) / dt
			}
		}
	}
	c.prevNetTime = now
	c.prevBytesSent = sent
	c.prevBytesRecv = recv
}

func (c *Collector) collectProcs(s *Stats, now time.Time) {
	procs, err := process.Processes()
	if err != nil {
		return
	}
	s.NumProcs = len(procs)

	dt := 0.0
	if !c.prevProcTime.IsZero() {
		dt = now.Sub(c.prevProcTime).Seconds()
	}

	nextCPU := make(map[int32]float64, len(procs))
	rows := make([]ProcInfo, 0, len(procs))

	for _, p := range procs {
		times, err := p.Times()
		if err != nil {
			continue
		}
		busy := times.User + times.System
		nextCPU[p.Pid] = busy

		cpuPct := 0.0
		if dt > 0 {
			if prev, ok := c.prevProcCPU[p.Pid]; ok && busy >= prev {
				cpuPct = (busy - prev) / dt * 100.0
			}
		}

		name, _ := p.Name()
		row := ProcInfo{PID: p.Pid, Name: name, CPU: cpuPct}
		if mi, err := p.MemoryInfo(); err == nil && mi != nil {
			row.MemRSS = mi.RSS
		}
		if mp, err := p.MemoryPercent(); err == nil {
			row.MemPct = mp
		}
		rows = append(rows, row)
	}

	c.prevProcCPU = nextCPU
	c.prevProcTime = now

	sort.Slice(rows, func(i, j int) bool {
		if rows[i].CPU != rows[j].CPU {
			return rows[i].CPU > rows[j].CPU
		}
		return rows[i].MemRSS > rows[j].MemRSS
	})
	s.Procs = rows
}
