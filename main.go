package main

import (
	"fmt"
	"os"

	"terminal-x-bubbletea/manager"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/term"
)

func main() {
	if !term.IsTerminal(os.Stdout.Fd()) {
		fmt.Fprintln(os.Stderr, "this example requires a real terminal")
		os.Exit(1)
	}

	p := tea.NewProgram(manager.New())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
