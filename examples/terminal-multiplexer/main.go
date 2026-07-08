package main

import (
	"log"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/term"
)

func main() {
	if !term.IsTerminal(os.Stdout.Fd()) {
		log.Fatal("the application requires a real terminal")
	}

	p := tea.NewProgram(New(2))

	if _, err := p.Run(); err != nil {
		log.Fatalf("failed to run program: %v", err)
	}
}
