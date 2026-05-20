package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/cunninghamb505/pulse-tui/internal/tui"
)

// version is overridable at build time: -ldflags "-X main.version=1.2.3".
var version = "0.1.0"

func main() {
	showVersion := flag.Bool("version", false, "print version and exit")
	flag.BoolVar(showVersion, "v", false, "print version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Println("pulse", version)
		return
	}

	p := tea.NewProgram(tui.New(), tea.WithAltScreen(), tea.WithMouseCellMotion())
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "pulse:", err)
		os.Exit(1)
	}
}
