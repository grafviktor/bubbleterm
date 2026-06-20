package main

import (
	"fmt"
	"os"

	"terminal-x-bubbletea/terminal"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/term"
)

type model struct {
	term *terminal.TermWindow
	init tea.Cmd
}

func initialModel() model {
	width, height, err := term.GetSize(os.Stdout.Fd())
	if err != nil {
		width, height = 80, 24
	}

	tw, cmd := terminal.NewTermWindow(width-10, height-10)

	return model{term: tw, init: cmd}
}

func (m model) Init() tea.Cmd {
	return m.init
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
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

func (m model) View() tea.View {
	if m.term == nil {
		v := tea.NewView("failed to start terminal")
		v.AltScreen = true
		return v
	}

	v := m.term.View()
	v.AltScreen = true
	return v
}

func main() {
	if !term.IsTerminal(os.Stdout.Fd()) {
		fmt.Fprintln(os.Stderr, "this example requires a real terminal")
		os.Exit(1)
	}

	p := tea.NewProgram(initialModel())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
