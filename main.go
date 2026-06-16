package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/term"
)

type model struct {
	term *TermWindow
	init tea.Cmd
}

func initialModel() model {
	width, height, err := term.GetSize(os.Stdout.Fd())
	if err != nil {
		width, height = 80, 24
	}

	tw, cmd := NewTermWindow(width, height)

	return model{term: tw, init: cmd}
}

func (m model) Init() tea.Cmd {
	return m.init
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+q" {
			if m.term != nil {
				m.term.Close()
			}
			return m, tea.Quit
		}
	}

	updated, cmd := m.term.Update(msg)
	m.term = updated
	return m, cmd
}

func (m model) View() string {
	if m.term == nil {
		return "failed to start terminal"
	}

	return m.term.View()
}

func main() {
	if !term.IsTerminal(os.Stdout.Fd()) {
		fmt.Fprintln(os.Stderr, "this example requires a real terminal")
		os.Exit(1)
	}

	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
