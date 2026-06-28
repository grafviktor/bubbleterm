package terminal

import (
	"fmt"
	"os"
	"os/exec"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/term"
	"github.com/charmbracelet/x/vt"
	"github.com/creack/pty"
)

// TermWindow embeds a shell in Bubble Tea using x/vt for emulation and creack/pty for the pseudo-terminal.
type TermWindow struct {
	pty *os.File
	cmd *exec.Cmd
	emu *vt.SafeEmulator

	width, height int
	Title         string
	cursorVisible bool
	closed        bool
	err           error
	command       string
}

type (
	TermOutputMsg []byte
	TermClosedMsg struct{ err error }
)

type Option func(*TermWindow)

var OptionWithCommand = func(cmd string) Option {
	return func(tw *TermWindow) {
		tw.command = cmd
	}
}

var OptionWithSize = func(width, height int) Option {
	return func(tw *TermWindow) {
		tw.width = width
		tw.height = height
	}
}

var OptionWithTitle = func(title string) Option {
	return func(tw *TermWindow) {
		tw.Title = title
	}
}

func NewTermWindow(opts ...Option) (*TermWindow, tea.Cmd) {
	tw := &TermWindow{}

	for _, opt := range opts {
		opt(tw)
	}

	if tw.command == "" {
		tw.command = os.Getenv("SHELL")
		if tw.command == "" {
			tw.command = "/bin/sh"
		}
	}

	cmd := exec.Command(tw.command)
	cmd.Env = append(os.Environ(), "TERM=xterm-256color")

	ptmx, err := pty.StartWithSize(cmd, &pty.Winsize{
		Rows: uint16(tw.height),
		Cols: uint16(tw.width),
	})
	if err != nil {
		return &TermWindow{width: tw.width, height: tw.height, closed: true, err: err}, nil
	}

	emu := vt.NewSafeEmulator(tw.width, tw.height)

	tw.pty = ptmx
	tw.cmd = cmd
	tw.emu = emu
	tw.cursorVisible = true

	// Forward anything the emulator writes to its response pipe — terminal
	// query replies *and* key bytes from SendKey — into the shell PTY.
	go tw.drainEmulator()

	return tw, readPTY(ptmx)
}

func (tw *TermWindow) drainEmulator() {
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

func readPTY(f *os.File) tea.Cmd {
	return func() tea.Msg {
		buf := make([]byte, 4096)
		n, err := f.Read(buf)
		if err != nil {
			return TermClosedMsg{err: err}
		}
		return TermOutputMsg(buf[:n])
	}
}

func (tw *TermWindow) Update(msg tea.Msg) (*TermWindow, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		tw.Resize(msg.Width, msg.Height)
		return tw, nil

	case TermOutputMsg:
		if tw.closed {
			return tw, nil
		}
		_, _ = tw.emu.Write(msg)
		return tw, readPTY(tw.pty)

	case TermClosedMsg:
		tw.closed = true
		tw.err = msg.err
		return tw, tea.Quit

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

func (tw *TermWindow) Resize(width, height int) {
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
	if tw.err != nil {
		return tea.NewView(fmt.Sprintf("terminal error: %v", tw.err))
	}
	if tw.closed {
		return tea.NewView("shell exited")
	}

	var v tea.View
	if tw.emu == nil {
		v.SetContent("")
		return v
	}

	v.SetContent(tw.emu.Render())

	if tw.cursorVisible {
		pos := tw.emu.CursorPosition()
		v.Cursor = tea.NewCursor(pos.X, pos.Y)
	}

	if tw.Title != "" {
		v.WindowTitle = tw.Title
	}

	return v
}
