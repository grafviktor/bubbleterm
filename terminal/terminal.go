package terminal

import (
	"fmt"
	"os"
	"os/exec"

	tea "charm.land/bubbletea/v2"
	uv "github.com/charmbracelet/ultraviolet"
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
}

type (
	termOutputMsg []byte
	termClosedMsg struct{ err error }
)

func NewTermWindow(width, height int) (*TermWindow, tea.Cmd) {
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/sh"
	}

	cmd := exec.Command(shell)
	cmd.Env = append(os.Environ(), "TERM=xterm-256color")

	ptmx, err := pty.StartWithSize(cmd, &pty.Winsize{
		Rows: uint16(height),
		Cols: uint16(width),
	})
	if err != nil {
		return &TermWindow{width: width, height: height, closed: true, err: err}, nil
	}

	emu := vt.NewSafeEmulator(width, height)
	tw := &TermWindow{
		pty: ptmx, cmd: cmd, emu: emu,
		width: width, height: height,
		cursorVisible: true,
	}

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
			return termClosedMsg{err: err}
		}
		return termOutputMsg(buf[:n])
	}
}

func (tw *TermWindow) Update(msg tea.Msg) (*TermWindow, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		tw.resize(msg.Width, msg.Height)
		return tw, nil

	case termOutputMsg:
		if tw.closed {
			return tw, nil
		}
		_, _ = tw.emu.Write(msg)
		return tw, readPTY(tw.pty)

	case termClosedMsg:
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

	case tea.MouseMsg:
		if tw.closed {
			return tw, nil
		}
		tw.forwardMouse(msg)
		return tw, nil
	}

	return tw, nil
}

// forwardMouse relays a mouse event to the embedded shell only when the
// emulator reports that mouse tracking is active (e.g. after vim sends ?1002h).
// Otherwise the shell would echo the CSI bytes as visible text.
func (tw *TermWindow) forwardMouse(msg tea.MouseMsg) {
	uvMouse := uv.Mouse{
		X:      msg.Mouse().X,
		Y:      msg.Mouse().Y,
		Button: msg.Mouse().Button,
		Mod:    msg.Mouse().Mod,
	}

	var event uv.MouseEvent
	switch msg.(type) {
	case tea.MouseClickMsg:
		event = uv.MouseClickEvent(uvMouse)
	case tea.MouseReleaseMsg:
		event = uv.MouseReleaseEvent(uvMouse)
	case tea.MouseWheelMsg:
		event = uv.MouseWheelEvent(uvMouse)
	case tea.MouseMotionMsg:
		event = uv.MouseMotionEvent(uvMouse)
	default:
		return
	}

	tw.emu.SendMouse(event)
}

func (tw *TermWindow) resize(width, height int) {
	tw.width, tw.height = width, height
	tw.emu.Resize(width, height)
	_ = pty.Setsize(tw.pty, &pty.Winsize{
		Rows: uint16(height),
		Cols: uint16(width),
	})
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

	// Bubble Tea must capture mouse on the outer terminal so forwardMouse can
	// relay events into the PTY.
	v.MouseMode = tea.MouseModeCellMotion

	return v
}
