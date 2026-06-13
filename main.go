package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/term"
)

type model struct {
	term *TermWindow
	init tea.Cmd
}

const borderPadding = 2 // lipgloss border consumes 2 columns

func initialModel() model {
	width, height, err := term.GetSize(os.Stdout.Fd())
	if err != nil {
		width, height = 80, 24
	}

	innerWidth := width - borderPadding
	innerHeight := height - headerHeight - footerHeight - borderPadding
	tw, cmd := NewTermWindow(innerWidth, innerHeight)

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

	case tea.WindowSizeMsg:
		if w, h, err := term.GetSize(os.Stdout.Fd()); err == nil {
			msg.Width, msg.Height = w, h
		}
		innerWidth := msg.Width - borderPadding
		innerHeight := msg.Height - headerHeight - footerHeight - borderPadding
		resized := tea.WindowSizeMsg{Width: innerWidth, Height: innerHeight}
		if m.term != nil {
			updated, cmd := m.term.Update(resized)
			m.term = updated
			return m, cmd
		}
	}

	if m.term == nil {
		return m, nil
	}

	updated, cmd := m.term.Update(msg)
	m.term = updated
	return m, cmd
}

func (m model) View() string {
	if m.term == nil {
		return "failed to start terminal"
	}

	title := m.term.Title
	if title == "" {
		title = "shell"
	}
	header := headerStyle.Render(fmt.Sprintf(" embedded shell · %s ", title))
	footer := footerStyle.Render(" ctrl+q quit · type commands in the shell below ")
	body := terminalStyle.Width(m.term.width).Height(m.term.height).Render(m.term.View())

	return lipgloss.JoinVertical(lipgloss.Left, header, body, footer)
}

var (
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("15")).
			Background(lipgloss.Color("4")).
			Padding(0, 1)

	footerStyle = lipgloss.NewStyle().
			Faint(true).
			Foreground(lipgloss.Color("8")).
			Padding(0, 1)

	terminalStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("8"))
)

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
