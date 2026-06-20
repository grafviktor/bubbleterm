package main

import (
	"fmt"
	"os"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"
	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/vt"
	"github.com/creack/pty"
)

// TermWindow embeds a shell in Bubble Tea using x/vt for emulation and
// creack/pty for the pseudo-terminal.
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

	case tea.KeyMsg:
		if tw.closed {
			return tw, nil
		}
		if key, ok := teaKeyMsgToKeyPressEvent(msg); ok {
			tw.emu.SendKey(key)
		}
		return tw, nil
	}

	return tw, nil
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

func (tw *TermWindow) View() string {
	if tw.err != nil {
		return fmt.Sprintf("terminal error: %v", tw.err)
	}
	if tw.closed {
		return "shell exited"
	}
	// return tw.emu.Render()
	return tw.renderWithCursor()
}

func (tw *TermWindow) renderWithCursor() string {
	if tw.emu == nil {
		return ""
	}
	if !tw.cursorVisible {
		return tw.emu.Render()
	}

	pos := tw.emu.CursorPosition()
	lines := make([]uv.Line, tw.height)
	for y := range tw.height {
		lines[y] = tw.parseCursor(y, pos.X, pos.Y)
	}
	return uv.Lines(lines).Render()
}

func (tw *TermWindow) parseCursor(y, cursorX, cursorY int) uv.Line {
	emu := tw.emu
	line := uv.NewLine(emu.Width())
	showCursor := y == cursorY

	for x := 0; x < emu.Width(); {
		cell := emu.CellAt(x, y)
		if cell == nil || cell.IsZero() {
			x++
			continue
		}

		c := cell
		if showCursor && x == cursorX {
			c = tw.cursorCell(cell)
		}

		line.Set(x, c)

		w := cell.Width
		if w <= 0 {
			w = 1
		}
		x += w
	}

	return line
}

func (tw *TermWindow) cursorCell(cell *uv.Cell) *uv.Cell {
	var c uv.Cell
	switch {
	case cell == nil || cell.IsZero():
		c = uv.EmptyCell
	case cell.Equal(&uv.EmptyCell):
		c = uv.EmptyCell
	default:
		c = *cell
	}

	if c.Style.Attrs&uv.AttrReverse != 0 {
		c.Style.Attrs &^= uv.AttrReverse
	} else {
		c.Style.Attrs |= uv.AttrReverse
	}

	if c.Content == "" {
		c.Content = " "
		c.Width = 1
	}

	return &c
}
