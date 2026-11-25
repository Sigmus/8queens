package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Sigmus/8queens/ui"
)

func main() {
	mode := flag.String("mode", "step", "animation mode: step (every attempt) or solutions (complete boards only)")
	speed := flag.Int("speed", 5, "animation speed from 1 (slow) to 8 (fast)")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "eightqueens animates a backtracking search for the eight queens puzzle.\n\n")
		fmt.Fprintf(os.Stderr, "Usage:\n  go run . [-mode step|solutions] [-speed 1-8]\n\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	model, err := ui.New(*mode, *speed)
	if err != nil {
		fmt.Fprintf(os.Stderr, "eightqueens: %v\n", err)
		os.Exit(2)
	}
	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "eightqueens: %v\n", err)
		os.Exit(1)
	}
}
