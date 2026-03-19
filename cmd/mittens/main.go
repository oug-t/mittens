package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"mittens/cmd/internal/tui"
)

func main() {
	p := tea.NewProgram(tui.InitialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error starting Mittens: %v", err)
		os.Exit(1)
	}
}
