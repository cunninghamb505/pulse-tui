package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) View() string {
	if !m.ready || m.width < 20 {
		return appStyle.Render(subtitleStyle.Render("starting pulse…"))
	}

	width := m.width - 2 // appStyle horizontal padding
	if width < 30 {
		width = 30
	}

	if m.showHelp {
		return appStyle.Render(m.renderHelp(width))
	}
	if m.showDetail {
		return appStyle.Render(m.renderDetail(width))
	}

	header := m.renderHeader(width)

	gap := 1
	leftW := (width - gap) / 2
	rightW := width - gap - leftW

	cpuBody := m.renderCPUBody(leftW - 4)
	memBody := m.renderMemBody(rightW - 4)

	rowH := lipgloss.Height(cpuBody)
	if h := lipgloss.Height(memBody); h > rowH {
		rowH = h
	}

	cpuPanel := panel("◤ CPU", leftW, rowH, cpuBody)
	memPanel := panel("◤ MEMORY", rightW, rowH, memBody)
	topRow := lipgloss.JoinHorizontal(lipgloss.Top, cpuPanel, strings.Repeat(" ", gap), memPanel)

	netBody := m.renderNetBody(leftW - 4)
	diskBody := m.renderDiskBody(rightW - 4)
	midH := lipgloss.Height(netBody)
	if h := lipgloss.Height(diskBody); h > midH {
		midH = h
	}
	netPanel := panel("◤ NETWORK", leftW, midH, netBody)
	diskPanel := panel("◤ DISK", rightW, midH, diskBody)
	midRow := lipgloss.JoinHorizontal(lipgloss.Top, netPanel, strings.Repeat(" ", gap), diskPanel)

	footer := m.renderFooter(width)
	procPanel := m.renderProcPanel(width, m.procRows())

	sections := []string{header, topRow, midRow}
	if len(m.stats.GPUs) > 0 {
		sections = append(sections, panel("◤ GPU", width, 0, m.renderGPUBody(width-4)))
	}
	sections = append(sections, procPanel, footer)

	view := lipgloss.JoinVertical(lipgloss.Left, sections...)
	return appStyle.Render(view)
}

// procRows analytically computes how many process rows fit below the other
// panels. Both Update (cursor scrolling) and View (slicing) call this so the
// visible window stays in sync. It mirrors the line counts of the body
// renderers and biases slightly tall to avoid overflowing the screen.
func (m Model) procRows() int {
	if m.height < 6 {
		return 3
	}
	width := m.width - 2
	if width < 30 {
		width = 30
	}
	leftW := (width - 1) / 2

	inner := leftW - 4
	if inner < 10 {
		inner = 10
	}
	per := len(m.stats.PerCPU)
	cols := inner / 20
	if cols < 1 {
		cols = 1
	}
	if per > 0 && cols > per {
		cols = per
	}
	coreRows := 0
	if per > 0 {
		coreRows = (per + cols - 1) / cols
	}
	cpuLines := 1 // bar
	if m.stats.CPUModel != "" || m.stats.HasTemp {
		cpuLines++ // info line (model / temp)
	}
	cpuLines += 2 + coreRows // sparkline + blank + cores

	memLines := 5 // RAM, bar, used, sparkline, blank
	if m.stats.SwapTotal > 0 {
		memLines += 3
	} else {
		memLines += 2
	}

	topBody := cpuLines
	if memLines > topBody {
		topBody = memLines
	}
	topRowH := topBody + 3 // title + 2 border

	diskLines := 1 + 1 // io line + "no disks"
	if n := len(m.stats.Disks); n > 0 {
		diskLines = 1 + 2*n // io line + per-disk
	}
	midBody := 3 // network rows
	if diskLines > midBody {
		midBody = diskLines
	}
	midRowH := midBody + 3

	gpuRowH := 0
	if n := len(m.stats.GPUs); n > 0 {
		gpuRowH = 2*n + 3 // (name + bar) per GPU, + title + 2 border
	}

	fixed := 1 + topRowH + midRowH + gpuRowH + 1 // header + rows + footer
	rows := m.height - fixed - 4                 // proc chrome: 2 border + title + column header
	if rows < 3 {
		rows = 3
	}
	return rows
}

// panel wraps body in a titled rounded box of the given total width.
// If height > 0 the content area is fixed to that many rows.
func panel(title string, totalWidth, height int, body string) string {
	inner := totalWidth - 4
	if inner < 1 {
		inner = 1
	}
	content := panelTitleStyle.Render(title) + "\n" + body
	s := panelStyle.Width(inner)
	if height > 0 {
		s = s.Height(height + 1) // +1 for the title row
	}
	return s.Render(content)
}

func (m Model) renderHeader(width int) string {
	title := titleStyle.Render("◢ PULSE ◤")
	host := m.stats.Hostname
	plat := m.stats.Platform
	meta := fmt.Sprintf("%s · %s · up %s", host, plat, humanDuration(m.stats.Uptime))
	sub := subtitleStyle.Render(meta)

	left := title + "  " + sub

	right := ""
	if bat := m.stats.Battery; bat != nil {
		col := colGreen
		switch {
		case bat.Percent <= 20:
			col = colRed
		case bat.Percent <= 50:
			col = colYellow
		}
		mark := ""
		if bat.Charging {
			mark = "↑"
		}
		right += lipgloss.NewStyle().Foreground(col).Bold(true).
			Render(fmt.Sprintf("BAT %s%.0f%%", mark, bat.Percent)) + "   "
	}
	right += labelStyle.Render(fmt.Sprintf("%d procs", m.stats.NumProcs))

	pad := width - lipgloss.Width(left) - lipgloss.Width(right)
	if pad < 1 {
		pad = 1
	}
	return left + strings.Repeat(" ", pad) + right
}

func (m Model) renderCPUBody(inner int) string {
	if inner < 10 {
		inner = 10
	}
	var b strings.Builder

	barW := inner - 9
	if barW < 4 {
		barW = 4
	}
	b.WriteString(fmt.Sprintf("%s %s\n",
		gradientBar(barW, m.stats.CPUPercent),
		valueStyle.Render(fmt.Sprintf("%5.1f%%", m.stats.CPUPercent))))

	if m.stats.CPUModel != "" || m.stats.HasTemp {
		left := truncate(m.stats.CPUModel, inner-8)
		line := labelStyle.Render(left)
		if m.stats.HasTemp {
			temp := fmt.Sprintf("%.0f°C", m.stats.CPUTemp)
			styled := lipgloss.NewStyle().Foreground(gradAt((m.stats.CPUTemp - 30) / 70)).
				Bold(true).Render(temp)
			pad := inner - lipgloss.Width(left) - lipgloss.Width(temp)
			if pad < 1 {
				pad = 1
			}
			line += strings.Repeat(" ", pad) + styled
		}
		b.WriteString(line + "\n")
	}
	b.WriteString(sparkline(m.cpuHist, 100) + "\n\n")

	// Per-core mini bars laid out in columns.
	per := m.stats.PerCPU
	colW := 20
	cols := inner / colW
	if cols < 1 {
		cols = 1
	}
	if cols > len(per) && len(per) > 0 {
		cols = len(per)
	}
	miniBarW := inner/cols - 11
	if miniBarW < 3 {
		miniBarW = 3
	}
	for i := 0; i < len(per); i += cols {
		var row strings.Builder
		for j := i; j < i+cols && j < len(per); j++ {
			cell := fmt.Sprintf("%s%s%s",
				labelStyle.Render(fmt.Sprintf("c%02d ", j)),
				gradientBar(miniBarW, per[j]),
				labelStyle.Render(fmt.Sprintf(" %3.0f", per[j])),
			)
			row.WriteString(cell + " ")
		}
		b.WriteString(strings.TrimRight(row.String(), " ") + "\n")
	}
	return strings.TrimRight(b.String(), "\n")
}

func (m Model) renderMemBody(inner int) string {
	if inner < 10 {
		inner = 10
	}
	barW := inner - 9
	if barW < 4 {
		barW = 4
	}
	var b strings.Builder

	b.WriteString(labelStyle.Render("RAM ") + "\n")
	b.WriteString(fmt.Sprintf("%s %s\n",
		gradientBar(barW, m.stats.MemPercent),
		valueStyle.Render(fmt.Sprintf("%5.1f%%", m.stats.MemPercent))))
	b.WriteString(labelStyle.Render(fmt.Sprintf("%s / %s used",
		humanBytes(m.stats.MemUsed), humanBytes(m.stats.MemTotal))) + "\n")
	b.WriteString(sparkline(m.memHist, 100) + "\n\n")

	b.WriteString(labelStyle.Render("SWAP") + "\n")
	if m.stats.SwapTotal > 0 {
		b.WriteString(fmt.Sprintf("%s %s\n",
			gradientBar(barW, m.stats.SwapPercent),
			valueStyle.Render(fmt.Sprintf("%5.1f%%", m.stats.SwapPercent))))
		b.WriteString(labelStyle.Render(fmt.Sprintf("%s / %s used",
			humanBytes(m.stats.SwapUsed), humanBytes(m.stats.SwapTotal))))
	} else {
		b.WriteString(labelStyle.Render("none"))
	}
	return b.String()
}

func (m Model) renderNetBody(inner int) string {
	half := inner/2 - 1
	if half < 12 {
		half = 12
	}
	sparkW := half - 12
	if sparkW < 4 {
		sparkW = 4
	}

	upMax := maxOf(m.upHist)
	downMax := maxOf(m.downHist)

	up := fmt.Sprintf("%s %s %s",
		lipgloss.NewStyle().Foreground(colPink).Render("▲ UP"),
		valueStyle.Render(rightPad(humanRate(m.stats.NetUpRate), 11)),
		sparkline(lastN(m.upHist, sparkW), upMax))
	down := fmt.Sprintf("%s %s %s",
		lipgloss.NewStyle().Foreground(colCyan).Render("▼ DN"),
		valueStyle.Render(rightPad(humanRate(m.stats.NetDownRate), 11)),
		sparkline(lastN(m.downHist, sparkW), downMax))

	totals := labelStyle.Render(fmt.Sprintf("session ▲ %s   ▼ %s",
		humanBytes(m.stats.NetUpTotal), humanBytes(m.stats.NetDownTotal)))

	return up + "\n" + down + "\n" + totals
}

func (m Model) renderDiskBody(inner int) string {
	if inner < 10 {
		inner = 10
	}
	barW := inner - 9
	if barW < 4 {
		barW = 4
	}
	var b strings.Builder

	b.WriteString(fmt.Sprintf("%s %s   %s %s\n",
		lipgloss.NewStyle().Foreground(colCyan).Render("R"),
		valueStyle.Render(humanRate(m.stats.DiskReadRate)),
		lipgloss.NewStyle().Foreground(colPink).Render("W"),
		valueStyle.Render(humanRate(m.stats.DiskWriteRate))))

	if len(m.stats.Disks) == 0 {
		b.WriteString(labelStyle.Render("no disks"))
		return b.String()
	}
	for i, d := range m.stats.Disks {
		b.WriteString(fmt.Sprintf("%s %s\n",
			labelStyle.Render(rightPad(d.Mount, 6)),
			labelStyle.Render(fmt.Sprintf("%s / %s",
				humanBytes(d.Used), humanBytes(d.Total)))))
		b.WriteString(fmt.Sprintf("%s %s",
			gradientBar(barW, d.Percent),
			valueStyle.Render(fmt.Sprintf("%5.1f%%", d.Percent))))
		if i < len(m.stats.Disks)-1 {
			b.WriteString("\n")
		}
	}
	return b.String()
}

func (m Model) renderGPUBody(inner int) string {
	if inner < 10 {
		inner = 10
	}
	var b strings.Builder
	for i, g := range m.stats.GPUs {
		temp := lipgloss.NewStyle().Foreground(gradAt((g.TempC - 30) / 70)).Bold(true).
			Render(fmt.Sprintf("%.0f°C", g.TempC))
		tempW := lipgloss.Width(temp)
		name := truncate(g.Name, inner-tempW-3)
		pad := inner - lipgloss.Width(name) - tempW - 2
		if pad < 1 {
			pad = 1
		}
		b.WriteString(labelStyle.Render(name) + strings.Repeat(" ", pad) + temp + "\n")

		pct := fmt.Sprintf("%5.1f%%", g.UtilPct)
		vram := fmt.Sprintf("VRAM %s / %s", humanBytes(g.MemUsed), humanBytes(g.MemTotal))
		gpuBarW := inner - len(pct) - lipgloss.Width(vram) - 6
		if gpuBarW < 6 {
			gpuBarW = 6
		}
		b.WriteString(fmt.Sprintf("%s %s   %s",
			gradientBar(gpuBarW, g.UtilPct),
			valueStyle.Render(pct),
			labelStyle.Render(vram)))
		if i < len(m.stats.GPUs)-1 {
			b.WriteString("\n")
		}
	}
	return b.String()
}

// renderProcPanel builds the titled process box showing `visible` rows of the
// filtered, sorted process list, highlighting the cursor row.
func (m Model) renderProcPanel(totalWidth, visible int) string {
	inner := totalWidth - 4
	const prefixW = 2 // "▶ " / "  "
	pidW, cpuW, memW := 7, 7, 10
	nameW := inner - prefixW - pidW - cpuW - memW - 3
	if nameW < 6 {
		nameW = 6
	}

	procs := m.filteredProcs()
	total := len(procs)
	offset := m.procOffset
	if offset > total-visible {
		offset = total - visible
	}
	if offset < 0 {
		offset = 0
	}
	end := offset + visible
	if end > total {
		end = total
	}

	arrow := "↓"
	if m.reverse {
		arrow = "↑"
	}
	pidHdr, cpuHdr, memHdr, nameHdr := "PID", "CPU%", "MEM", "NAME"
	switch m.sort {
	case sortPID:
		pidHdr += arrow
	case sortCPU:
		cpuHdr += arrow
	case sortMem:
		memHdr += arrow
	case sortName:
		nameHdr += arrow
	}
	colHeader := procHeaderStyle.Render(fmt.Sprintf("%-*s%-*s %*s %*s  %-*s",
		prefixW, "", pidW, pidHdr, cpuW, cpuHdr, memW, memHdr, nameW, nameHdr))

	cursorStyle := lipgloss.NewStyle().Foreground(colBG).Background(colCyan).Bold(true).Width(inner)

	var b strings.Builder
	b.WriteString(colHeader + "\n")
	for i := offset; i < end; i++ {
		p := procs[i]
		if i == m.cursor {
			plain := fmt.Sprintf("▶ %-*d %*.1f %*s  %s",
				pidW, p.PID, cpuW, p.CPU, memW, humanBytes(p.MemRSS), truncate(p.Name, nameW))
			b.WriteString(cursorStyle.Render(plain) + "\n")
			continue
		}
		cpuStr := lipgloss.NewStyle().Foreground(gradAt(p.CPU / 100)).
			Render(fmt.Sprintf("%*.1f", cpuW, p.CPU))
		row := fmt.Sprintf("  %s %s %s  %s",
			labelStyle.Render(fmt.Sprintf("%-*d", pidW, p.PID)),
			cpuStr,
			valueStyle.Render(fmt.Sprintf("%*s", memW, humanBytes(p.MemRSS))),
			procNameStyle.Render(truncate(p.Name, nameW)),
		)
		b.WriteString(row + "\n")
	}
	body := strings.TrimRight(b.String(), "\n")

	title := fmt.Sprintf("◤ PROCESSES  ·  sort %s  ·  %d-%d/%d",
		m.sort.String(), min(offset+1, total), end, total)
	if m.filter != "" || m.filtering {
		title += "  ·  /" + m.filter
		if m.filtering {
			title += "▏"
		}
	}
	return panel(title, totalWidth, visible+1, body)
}

func (m Model) renderFooter(width int) string {
	if m.confirmKill {
		prompt := lipgloss.NewStyle().Foreground(colBG).Background(colRed).Bold(true).
			Render(fmt.Sprintf(" Kill %s (PID %d)? ", truncate(m.killName, 30), m.killPID))
		yn := keyStyle.Render("y") + footerStyle.Render(" confirm   ") +
			keyStyle.Render("n") + footerStyle.Render(" cancel")
		return prompt + "  " + yn
	}

	var left string
	switch {
	case m.filtering:
		left = keyStyle.Render("/") + valueStyle.Render(m.filter) +
			lipgloss.NewStyle().Foreground(colCyan).Render("▏") +
			footerStyle.Render("   "+"enter apply · esc clear")
	case m.statusMsg != "":
		left = lipgloss.NewStyle().Foreground(colYellow).Bold(true).Render(m.statusMsg)
	default:
		keys := []struct{ k, d string }{
			{"↑↓", "select"},
			{"enter", "info"},
			{"k", "kill"},
			{"/", "filter"},
			{"?", "help"},
			{"q", "quit"},
		}
		parts := make([]string, 0, len(keys))
		for _, kv := range keys {
			parts = append(parts, keyStyle.Render(kv.k)+" "+footerStyle.Render(kv.d))
		}
		left = strings.Join(parts, footerStyle.Render(" · "))
	}

	var segs []string
	if m.paused {
		segs = append(segs, lipgloss.NewStyle().Foreground(colRed).Bold(true).Render("⏸ PAUSED"))
	}
	segs = append(segs, footerStyle.Render(fmt.Sprintf("%gs", m.refresh.Seconds())))
	segs = append(segs, lipgloss.NewStyle().Foreground(colPink).Bold(true).
		Render("◆ "+themes[m.theme%len(themes)].Name))
	right := strings.Join(segs, footerStyle.Render("  ·  "))

	pad := width - lipgloss.Width(left) - lipgloss.Width(right)
	if pad < 1 {
		pad = 1
	}
	return left + strings.Repeat(" ", pad) + right
}

func (m Model) renderDetail(width int) string {
	inner := width - 6
	if inner > 88 {
		inner = 88
	}
	if inner < 24 {
		inner = 24
	}
	labelW := 12
	valW := inner - labelW
	if valW < 8 {
		valW = 8
	}

	var b strings.Builder
	b.WriteString(titleStyle.Render(" "+truncate(m.detail.Name, inner-4)+" ") + "\n\n")

	if m.detailErr != nil {
		b.WriteString(lipgloss.NewStyle().Foreground(colRed).Width(inner).
			Render("could not read process: " + m.detailErr.Error()))
	} else {
		d := m.detail
		started := "—"
		if !d.CreateTime.IsZero() {
			started = d.CreateTime.Format("2006-01-02 15:04:05")
		}
		rows := [][2]string{
			{"PID", fmt.Sprintf("%d", d.PID)},
			{"Parent", strings.TrimSpace(fmt.Sprintf("%d %s", d.Ppid, d.ParentName))},
			{"User", d.Username},
			{"Status", d.Status},
			{"Threads", fmt.Sprintf("%d", d.NumThreads)},
			{"CPU", fmt.Sprintf("%.1f%%", d.CPU)},
			{"Memory", fmt.Sprintf("%s (%.1f%%)", humanBytes(d.MemRSS), d.MemPct)},
			{"Started", started},
			{"Run time", humanDuration(d.RunTime)},
			{"Exe", d.Exe},
			{"Command", d.Cmdline},
		}
		lbl := keyStyle.Width(labelW)
		val := valueStyle.Width(valW)
		for _, r := range rows {
			v := r[1]
			if strings.TrimSpace(v) == "" {
				v = "—"
			}
			b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, lbl.Render(r[0]), val.Render(v)) + "\n")
		}
	}
	b.WriteString("\n" + subtitleStyle.Render("esc / enter to close"))

	box := panelStyle.Width(inner).Render(strings.TrimRight(b.String(), "\n"))
	h := m.height
	if h < 1 {
		h = lipgloss.Height(box)
	}
	return lipgloss.Place(width, h, lipgloss.Center, lipgloss.Center, box)
}

func (m Model) renderHelp(width int) string {
	rows := [][2]string{
		{"↑ / ↓", "move selection"},
		{"PgUp / PgDn", "jump 10 rows"},
		{"g / G", "jump to top / bottom"},
		{"enter", "process details"},
		{"c / m / n / p", "sort by CPU / mem / name / PID"},
		{"r", "reverse sort order"},
		{"k", "kill selected process"},
		{"/", "filter by name"},
		{"esc", "clear filter / close"},
		{"space", "pause / resume"},
		{"+ / -", "faster / slower refresh"},
		{"t / T", "next / previous theme"},
		{"?", "toggle this help"},
		{"q", "quit"},
	}
	var b strings.Builder
	b.WriteString(titleStyle.Render(" PULSE — keys ") + "\n\n")
	for _, r := range rows {
		b.WriteString(keyStyle.Render(fmt.Sprintf("%-14s", r[0])) +
			footerStyle.Render(r[1]) + "\n")
	}
	b.WriteString("\n" + subtitleStyle.Render("press any key to close"))
	box := panelStyle.Render(strings.TrimRight(b.String(), "\n"))

	h := m.height
	if h < 1 {
		h = lipgloss.Height(box)
	}
	return lipgloss.Place(width, h, lipgloss.Center, lipgloss.Center, box)
}

// ---- formatting helpers ----

func humanBytes(b uint64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := uint64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(b)/float64(div), "KMGTPE"[exp])
}

func humanRate(bps float64) string {
	units := []string{"B/s", "KB/s", "MB/s", "GB/s"}
	i := 0
	for bps >= 1024 && i < len(units)-1 {
		bps /= 1024
		i++
	}
	return fmt.Sprintf("%.1f %s", bps, units[i])
}

func humanDuration(d time.Duration) string {
	if d <= 0 {
		return "—"
	}
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	mins := int(d.Minutes()) % 60
	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm", days, hours, mins)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, mins)
	}
	return fmt.Sprintf("%dm", mins)
}

func truncate(s string, w int) string {
	if w < 1 {
		return ""
	}
	if lipgloss.Width(s) <= w {
		return s
	}
	if w == 1 {
		return "…"
	}
	r := []rune(s)
	if len(r) > w-1 {
		r = r[:w-1]
	}
	return string(r) + "…"
}

func rightPad(s string, w int) string {
	if lipgloss.Width(s) >= w {
		return s
	}
	return s + strings.Repeat(" ", w-lipgloss.Width(s))
}

func maxOf(vs []float64) float64 {
	m := 1.0
	for _, v := range vs {
		if v > m {
			m = v
		}
	}
	return m
}

func lastN(vs []float64, n int) []float64 {
	if n <= 0 || len(vs) <= n {
		return vs
	}
	return vs[len(vs)-n:]
}
