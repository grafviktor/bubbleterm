package main

import (
	"fmt"
	"os"

	"terminal-x-bubbletea/terminal"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/term"
)

type SplitMode string

const (
	SplitModeHorizontal SplitMode = "horizontal"
	SplitModeVertical   SplitMode = "vertical"
)

type model struct {
	focusedTerminal int
	terminals       []*terminal.TermWindow
	init            tea.Cmd
	splitMode       SplitMode
}

func initialModel() model {
	withCommand := terminal.OptionWithCommand("zsh")
	withTitle := terminal.OptionWithTitle("Terminal Window")
	tw1, cmd1 := terminal.NewTermWindow(0, withCommand, withTitle)
	tw2, cmd2 := terminal.NewTermWindow(1, withCommand, withTitle)

	return model{
		focusedTerminal: 1,
		terminals:       []*terminal.TermWindow{tw1, tw2},
		init:            tea.Batch(cmd1, cmd2),
		splitMode:       SplitModeHorizontal,
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
		if msg.String() == "ctrl+w" {
			m.focusedTerminal = (m.focusedTerminal + 1) % len(m.terminals)
			return m, nil
		}
		if msg.String() == "ctrl+n" {
			if m.splitMode == SplitModeHorizontal {
				m.splitMode = SplitModeVertical
			} else {
				m.splitMode = SplitModeHorizontal
			}

			return m.requestResize()
		}
	case tea.WindowSizeMsg:
		return m.requestResize()
	case tea.EnvMsg,
		tea.ColorProfileMsg,
		tea.ModeReportMsg:
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

func (m model) View() tea.View {
	if m.terminals == nil {
		v := tea.NewView("failed to start terminal")
		v.AltScreen = true
		return v
	}

	var views []string
	var cursor tea.Cursor

	for i, tw := range m.terminals {
		tv := tw.View()
		views = append(views, tv.Content)

		if i != m.focusedTerminal || tv.Cursor == nil {
			continue
		}

		cursor = m.getCursor(tv)
	}

	var v tea.View
	if m.splitMode == SplitModeVertical {
		v = tea.NewView(lipgloss.JoinVertical(lipgloss.Top, views...))
	} else {
		v = tea.NewView(lipgloss.JoinHorizontal(lipgloss.Top, views...))
	}

	v.AltScreen = true
	v.Cursor = &cursor
	return v
}

func (m model) requestResize() (tea.Model, tea.Cmd) {
	cmds := make([]tea.Cmd, len(m.terminals))
	for i, tw := range m.terminals {
		width, height := m.getSize()
		updated, cmd := tw.Update(tea.WindowSizeMsg{
			Width:  width,
			Height: height,
		})
		m.terminals[i] = updated
		cmds[i] = cmd
	}
	return m, tea.Batch(cmds...)
}

func (m model) getCursor(terminal tea.View) tea.Cursor {
	c := *terminal.Cursor
	w := 0
	h := 0
	if m.focusedTerminal != 0 {
		w, h = m.terminals[m.focusedTerminal].GetSizeCurrent()
	}

	// TODO: this is a hacky way to get a cursor position. Works only because both terminals have the same width or height.
	switch m.splitMode {
	case SplitModeHorizontal:
		c.X += w
	case SplitModeVertical:
		c.Y += h
	}

	return c
}

func (m model) getSize() (width, height int) {
	width, height, err := term.GetSize(os.Stdout.Fd())
	if err != nil {
		width, height = 80, 24
	}

	nTerminals := len(m.terminals)
	switch m.splitMode {
	case SplitModeHorizontal:
		return width / nTerminals, height
	case SplitModeVertical:
		return width, height / nTerminals
	default:
		return width, height
	}
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
