package main

import (
	"fmt"
	"os"

	"terminal-x-bubbletea/terminal"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/term"
)

type model struct {
	focusedTerminal int
	terminals       []*terminal.TermWindow
	init            tea.Cmd
}

func initialModel() model {
	withCommand := terminal.OptionWithCommand("zsh")
	withTitle := terminal.OptionWithTitle("Terminal Window")
	tw1, cmd1 := terminal.NewTermWindow(1, withCommand, withTitle)
	tw2, cmd2 := terminal.NewTermWindow(2, withCommand, withTitle)

	return model{
		focusedTerminal: 0,
		terminals:       []*terminal.TermWindow{tw1, tw2},
		init:            tea.Batch(cmd1, cmd2),
	}
}

func (m model) Init() tea.Cmd {
	return m.init
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if msg.String() == "ctrl+q" {
			if m.terminals != nil {
				for _, tw := range m.terminals {
					tw.Close()
				}
			}
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		for i, terminal := range m.terminals {
			terminal.Resize(msg.Width/len(m.terminals), msg.Height)
			m.terminals[i] = terminal
		}
		return m, nil
	case tea.EnvMsg,
		tea.ColorProfileMsg,
		tea.ModeReportMsg:
		// terminal.TermOutputMsg:
		cmds := make([]tea.Cmd, len(m.terminals))
		for i, tw := range m.terminals {
			updated, cmd := tw.Update(msg)
			m.terminals[i] = updated
			cmds[i] = cmd
		}
		return m, tea.Batch(cmds...)

	case terminal.TermOutputMsg:
		for i, tw := range m.terminals {
			if tw.ID() == msg.ID {
				updated, cmd := tw.Update(msg)
				m.terminals[i] = updated
				return m, cmd
			}
		}
		return m, nil

	case terminal.TermClosedMsg:
		for i, tw := range m.terminals {
			if tw.ID() != msg.ID {
				continue
			}
			updated, cmd := tw.Update(msg)
			m.terminals[i] = updated
			if cmd != nil {
				return m, cmd
			}
		}
		return m, nil
	}

	for i, tw := range m.terminals {
		if i == m.focusedTerminal {
			updated, cmd := tw.Update(msg)
			m.terminals[i] = updated
			return m, cmd
		}
	}

	return m, nil
}

func (m model) terminalView() []string {
	var views []string
	for i := range m.terminals {
		views = append(views, m.terminals[i].View().Content)
	}
	return views
}

func (m model) View() tea.View {
	if m.terminals == nil {
		v := tea.NewView("failed to start terminal")
		v.AltScreen = true
		return v
	}

	v := tea.NewView(lipgloss.JoinHorizontal(lipgloss.Top, m.terminalView()...))
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
