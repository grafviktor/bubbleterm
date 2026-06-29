package manager

import (
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

type Manager struct {
	focusedTerminal int
	terminals       []*terminal.TermWindow
	init            tea.Cmd
	splitMode       SplitMode
}

func New() Manager {
	withTitle := terminal.OptionWithTitle("Terminal Window")
	tw1, cmd1 := terminal.New(0, withTitle)
	tw2, cmd2 := terminal.New(1, withTitle)

	return Manager{
		focusedTerminal: 0,
		terminals:       []*terminal.TermWindow{tw1, tw2},
		init:            tea.Batch(cmd1, cmd2),
		splitMode:       SplitModeHorizontal,
	}
}

func (m Manager) Init() tea.Cmd {
	return m.init
}

func (m Manager) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

func (m Manager) View() tea.View {
	if m.terminals == nil {
		v := tea.NewView("failed to start terminal")
		v.AltScreen = true
		return v
	}

	var views []string
	var cursor tea.Cursor

	for i, tw := range m.terminals {
		tv := tw.View()
		if i == m.focusedTerminal {
			views = append(views, focusedStyle.Render(tv.Content))
		} else {
			views = append(views, paneStyle.Render(tv.Content))
		}

		if i != m.focusedTerminal || tv.Cursor == nil {
			continue
		}

		cursor = m.getCursor(tv)
	}

	var v tea.View
	if m.splitMode == SplitModeVertical {
		v = tea.NewView(lipgloss.JoinVertical(lipgloss.Left, views...))
	} else {
		v = tea.NewView(lipgloss.JoinHorizontal(lipgloss.Top, views...))
	}

	v.AltScreen = true
	v.Cursor = &cursor
	return v
}

func (m Manager) requestResize() (tea.Model, tea.Cmd) {
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

func (m Manager) getCursor(terminal tea.View) tea.Cursor {
	// This is a hacky way to get a cursor position. Works only because because we have only 2 terminals and they have the same width or height.
	c := *terminal.Cursor
	// Why 1? Because for the first pane we count only left border
	horizontalBorderOffset := 1
	// Why 1? Because for the top pane we count only top border
	verticalBorderOffset := 1
	w, h := 0, 0
	if m.focusedTerminal != 0 {
		w, h = m.terminals[m.focusedTerminal].GetSizeCurrent()
		if m.splitMode == SplitModeHorizontal {
			// Why 3? Because for the second pane in horizontal split we count left and right borders of the first pane and left border of the second pane
			horizontalBorderOffset = 3
		} else {
			// Why 3? Because for the second pane in vertical split we count top and bottom borders of the first pane and top border of the second pane
			verticalBorderOffset = 3
		}
	}

	switch m.splitMode {
	case SplitModeHorizontal:
		c.X += w + horizontalBorderOffset
		c.Y += verticalBorderOffset
	case SplitModeVertical:
		c.X += horizontalBorderOffset
		c.Y += h + verticalBorderOffset
	}

	return c
}

func (m Manager) getSize() (width, height int) {
	width, height, err := term.GetSize(os.Stdout.Fd())
	if err != nil {
		width, height = 80, 24
	}

	var w, h int
	borderSize := 2
	nTerminals := len(m.terminals)
	switch m.splitMode {
	case SplitModeHorizontal:
		w, h = width/nTerminals, height
	case SplitModeVertical:
		w, h = width, height/nTerminals
	default:
		w, h = width, height
	}

	return w - borderSize, h - borderSize
}
