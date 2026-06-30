package terminal

import (
	"os"
	"os/exec"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/term"
	"github.com/charmbracelet/x/vt"
	"github.com/creack/pty"
)

// TermWindow embeds a shell in Bubble Tea using x/vt for emulation and creack/pty for the pseudo-terminal.
type TermWindow struct {
	pty *os.File
	cmd *exec.Cmd
	emu *vt.SafeEmulator

	id            int
	width, height int
	cursorVisible bool
	closed        bool
	command       string
}

type (
	TermOutputMsg struct{ ID int }
	TermClosedMsg struct{ ID int }
)

func New(id int, opts ...Option) (*TermWindow, tea.Cmd, error) {
	tw := &TermWindow{id: id}

	for _, opt := range opts {
		opt(tw)
	}

	if tw.command == "" {
		tw.command = os.Getenv("SHELL")
		if tw.command == "" {
			tw.command = "/bin/sh"
		}
	}

	if tw.width == 0 || tw.height == 0 {
		w, h := tw.getSizeDefault()
		tw.width, tw.height = w, h
	}

	cmd := exec.Command(tw.command)
	cmd.Env = append(os.Environ(), "TERM=xterm-256color")

	ptmx, err := pty.StartWithSize(cmd, &pty.Winsize{
		Rows: uint16(tw.height),
		Cols: uint16(tw.width),
	})
	if err != nil {
		return &TermWindow{width: tw.width, height: tw.height, closed: true}, nil, err
	}

	emu := vt.NewSafeEmulator(tw.width, tw.height)
	emu.SetCallbacks(vt.Callbacks{
		CursorVisibility: func(visible bool) {
			tw.cursorVisible = visible
		},
	})

	tw.pty = ptmx
	tw.cmd = cmd
	tw.emu = emu
	tw.cursorVisible = true

	// Forward anything the emulator writes to its response pipe — terminal
	// query replies *and* key bytes from SendKey — into the shell PTY.
	go tw.terminalViewToPty()

	return tw, tw.ptyToTerminalView(), nil
}

func (tw *TermWindow) terminalViewToPty() {
	buf := make([]byte, 1024)
	for {
		n, err := tw.emu.Read(buf)
		if n > 0 {
			_, _ = tw.pty.Write(buf[:n])
		}
		if err != nil {
			return
		}
	}
}

func (tw *TermWindow) ptyToTerminalView() tea.Cmd {
	return func() tea.Msg {
		buf := make([]byte, 4096)
		n, err := tw.pty.Read(buf)
		if err != nil {
			return TermClosedMsg{ID: tw.id}
		}
		_, _ = tw.emu.Write(buf[:n])
		return TermOutputMsg{ID: tw.id}
	}
}

func (tw *TermWindow) Update(msg tea.Msg) (*TermWindow, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		tw.resize(msg.Width, msg.Height)
		return tw, nil

	case TermOutputMsg:
		if tw.id != msg.ID {
			return tw, nil
		}
		if tw.closed {
			return tw, nil
		}
		return tw, tw.ptyToTerminalView()

	case tea.KeyPressMsg:
		if tw.closed {
			return tw, nil
		}
		tw.emu.SendKey(vt.KeyPressEvent(msg))
		return tw, nil

	case tea.PasteMsg:
		if tw.closed {
			return tw, nil
		}
		tw.emu.Paste(msg.Content)
		return tw, nil
	}

	return tw, nil
}

func (tw *TermWindow) resize(width, height int) {
	if tw.emu == nil || tw.pty == nil {
		return
	}

	if width < 5 {
		width = 5
	}

	if height < 5 {
		height = 5
	}

	tw.width, tw.height = width, height
	tw.emu.Resize(width, height)
	_ = pty.Setsize(tw.pty, &pty.Winsize{
		Rows: uint16(height),
		Cols: uint16(width),
	})
}

func (tw *TermWindow) getSizeDefault() (width, height int) {
	var err error
	tw.width, tw.height, err = term.GetSize(os.Stdout.Fd())
	if err != nil {
		return 80, 24
	}
	return tw.width, tw.height
}

func (tw *TermWindow) GetSizeCurrent() (width, height int) {
	return tw.width, tw.height
}

func (tw *TermWindow) Close() {
	if tw.emu != nil {
		_ = tw.emu.Emulator.Close()
	}
	if tw.cmd != nil && tw.cmd.Process != nil {
		_ = tw.cmd.Process.Kill()
		_ = tw.cmd.Wait()
	}
	if tw.pty != nil {
		_ = tw.pty.Close()
	}
	tw.closed = true
}

func (tw *TermWindow) View() tea.View {
	if tw.closed {
		return tea.NewView("shell exited")
	}

	v := tea.NewView("")
	if tw.emu == nil {
		v.SetContent("")
		return v
	}

	// It's required to re-calculate width and height with every update.
	content := lipgloss.NewStyle().
		Width(tw.width).
		Height(tw.height).
		Render(tw.emu.Render())
	v.SetContent(content)

	if tw.cursorVisible {
		pos := tw.emu.CursorPosition()
		v.Cursor = tea.NewCursor(pos.X, pos.Y)
	}

	return v
}

func (tw *TermWindow) ID() int {
	return tw.id
}
