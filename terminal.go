package termview

import (
	"context"
	"io"
	"os"
	"os/exec"
	"sync/atomic"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/term"
	"github.com/charmbracelet/x/vt"
	"github.com/charmbracelet/x/xpty"
)

const (
	minHeight     = 5
	minWidth      = 5
	defaultHeight = 24
	defaultWidth  = 80
)

var lastID atomic.Int64

func nextID() int {
	return int(lastID.Add(1))
}

type session struct {
	showCursor atomic.Bool
	closed     atomic.Bool
}

// Model embeds a shell in Bubble Tea using x/vt for emulation and xpty for the pseudo-terminal.
type Model struct {
	pty   xpty.Pty
	cmd   *exec.Cmd
	emu   *vt.SafeEmulator
	state *session

	id            int
	width, height int
	command       string
	commandArgs   []string
	focus         bool
	stdErr        io.Writer
}

func New(opts ...Option) (Model, error) {
	m := Model{id: nextID()}

	for _, opt := range opts {
		opt(&m)
	}

	if m.width == 0 || m.height == 0 {
		w, h := m.getDefaultSize()
		m.width, m.height = w, h
	}

	if m.command == "" {
		m.command = getShellPath()
	}

	cmd := buildCommand(m.command, m.commandArgs...)
	if m.stdErr != nil {
		cmd.Stderr = m.stdErr
	}

	// Init example taken from https://github.com/charmbracelet/freeze/blob/main/pty.go
	pty, err := xpty.NewPty(m.width, m.height)
	if err != nil {
		return m, err
	}

	if err := pty.Start(cmd); err != nil {
		return m, err
	}

	if up, ok := pty.(*xpty.UnixPty); ok {
		_ = up.Slave().Close()
	}

	emu := vt.NewSafeEmulator(m.width, m.height)
	state := &session{}
	state.showCursor.Store(true)
	emu.SetCallbacks(vt.Callbacks{
		CursorVisibility: func(visible bool) {
			state.showCursor.Store(visible)
		},
	})

	m.pty = pty
	m.cmd = cmd
	m.emu = emu
	m.state = state

	return m, nil
}

func (m Model) Init() tea.Cmd {
	// Forward anything the emulator writes to its response pipe — terminal
	// query replies *and* key bytes from SendKey — into the shell PTY.
	go m.terminalViewToPty()

	return m.ptyToTerminalView()
}

func (m Model) terminalViewToPty() {
	buf := make([]byte, 1024)
	for {
		if m.Closed() {
			return
		}

		n, err := m.emu.Read(buf)
		if n > 0 {
			_, _ = m.pty.Write(buf[:n])
		}

		if err != nil {
			return
		}
	}
}

func (m Model) ptyToTerminalView() tea.Cmd {
	return func() tea.Msg {
		buf := make([]byte, 4096)
		n, err := m.pty.Read(buf)
		if n > 0 {
			_, _ = m.emu.Write(buf[:n])
			return OutputMsg{ID: m.id}
		}

		if err != nil {
			if m.cmd != nil {
				_ = xpty.WaitProcess(context.Background(), m.cmd)
			}
			return ClosedMsg{ID: m.id}
		}

		return OutputMsg{ID: m.id}
	}
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.resize(msg.Width, msg.Height)
		if m.Closed() {
			return m, nil
		}
		return m, m.ptyToTerminalView()

	case OutputMsg:
		if m.id != msg.ID {
			return m, nil
		}

		if m.Closed() {
			return m, nil
		}

		return m, m.ptyToTerminalView()

	case ClosedMsg:
		if m.id != msg.ID {
			return m, nil
		}

		m.markClosed()
		return m, nil

	case tea.KeyPressMsg:
		if !m.Focused() {
			return m, nil
		}

		if m.Closed() {
			return m, nil
		}

		m.emu.SendKey(vt.KeyPressEvent(msg))
		return m, nil

	case tea.PasteMsg:
		if !m.Focused() {
			return m, nil
		}

		if m.Closed() {
			return m, nil
		}

		m.emu.Paste(msg.Content)
		return m, nil
	}

	return m, nil
}

func (m *Model) SetWidth(width int) {
	m.width = width
	m.resize(width, m.height)
}

func (m *Model) SetHeight(height int) {
	m.height = height
	m.resize(m.width, height)
}

func (m Model) Width() int {
	return m.width
}

func (m Model) Height() int {
	return m.height
}

func (m *Model) resize(width, height int) {
	if m.emu == nil || m.pty == nil {
		return
	}

	if m.Closed() {
		return
	}

	if width < minWidth {
		width = minWidth
	}

	if height < minHeight {
		height = minHeight
	}

	m.width, m.height = width, height
	m.emu.Resize(width, height)
	_ = m.pty.Resize(width, height)
}

func (m Model) getDefaultSize() (width, height int) {
	var err error
	m.width, m.height, err = term.GetSize(os.Stdout.Fd())
	if err != nil {
		return defaultWidth, defaultHeight
	}
	return m.width, m.height
}

func (m Model) View() string {
	if m.emu == nil {
		return ""
	}

	return m.emu.Render()
}

func (m Model) Focus() Model {
	m.focus = true
	return m
}

func (m Model) Blur() Model {
	m.focus = false
	return m
}

func (m Model) Focused() bool {
	return m.focus
}

func (m Model) Cursor() *tea.Cursor {
	if m.Closed() {
		return nil
	}

	if !m.Focused() {
		return nil
	}

	if m.state == nil || !m.state.showCursor.Load() {
		return nil
	}

	pos := m.emu.CursorPosition()
	return tea.NewCursor(pos.X, pos.Y)
}

func (m Model) ID() int {
	return m.id
}

func (m *Model) Close() {
	if m.Closed() {
		return
	}

	if m.cmd != nil && m.cmd.Process != nil {
		_ = m.cmd.Process.Kill()
		_ = xpty.WaitProcess(context.Background(), m.cmd)
	}

	m.markClosed()

	// Now get rid of the emulator, we're explicitly done with it and no
	// longer need its output.
	if m.emu != nil {
		_ = m.emu.Emulator.Close()
	}
}

func (m *Model) markClosed() {
	if m.Closed() {
		return
	}

	if m.pty != nil {
		_ = m.pty.Close()
	}

	if m.state != nil {
		m.state.closed.Store(true)
	}
}

func (m Model) Closed() bool {
	return m.state != nil && m.state.closed.Load()
}
