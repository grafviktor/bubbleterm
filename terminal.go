package main

import (
	"fmt"
	"os"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/vt"
	"github.com/creack/pty"
)

const (
	headerHeight = 1
	footerHeight = 1
)

// TermWindow embeds an interactive shell inside a Bubble Tea view using
// charmbracelet/x/vt for terminal emulation and creack/pty for the PTY.
type TermWindow struct {
	ptyFile *os.File
	cmd     *exec.Cmd
	emu     *vt.SafeEmulator

	width  int
	height int
	Title  string
	ready  bool
	closed bool
	err    error

	appCursorKeys bool
}

type termOutputMsg []byte
type termClosedMsg struct{ err error }

// NewTermWindow creates a terminal window with the given dimensions.
func NewTermWindow(width, height int) (*TermWindow, tea.Cmd) {
	if width < 10 {
		width = 10
	}
	if height < 3 {
		height = 3
	}

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
		ptyFile: ptmx,
		cmd:     cmd,
		emu:     emu,
		width:   width,
		height:  height,
		ready:   true,
	}

	emu.Emulator.SetCallbacks(vt.Callbacks{
		Title: func(title string) {
			tw.Title = title
		},
		EnableMode: func(mode ansi.Mode) {
			if mode == ansi.ModeCursorKeys {
				tw.appCursorKeys = true
			}
		},
		DisableMode: func(mode ansi.Mode) {
			if mode == ansi.ModeCursorKeys {
				tw.appCursorKeys = false
			}
		},
	})

	// Drain emulator responses (DA1, DA2, etc.) back to the PTY so Write
	// never blocks when the shell queries terminal capabilities.
	go func() {
		buf := make([]byte, 1024)
		for {
			n, err := emu.Read(buf)
			if n > 0 {
				_, _ = ptmx.Write(buf[:n])
			}
			if err != nil {
				return
			}
		}
	}()

	return tw, readPTY(ptmx)
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
		return tw, readPTY(tw.ptyFile)

	case termClosedMsg:
		tw.closed = true
		tw.err = msg.err
		return tw, tea.Quit

	case tea.KeyMsg:
		if tw.closed || tw.ptyFile == nil {
			return tw, nil
		}
		if b := tw.keyToBytes(msg); len(b) > 0 {
			_, _ = tw.ptyFile.Write(b)
		}
		return tw, nil
	}

	return tw, nil
}

func (tw *TermWindow) resize(width, height int) {
	if width < 10 {
		width = 10
	}
	if height < 3 {
		height = 3
	}

	tw.width = width
	tw.height = height
	tw.emu.Resize(width, height)

	if tw.ptyFile != nil {
		_ = pty.Setsize(tw.ptyFile, &pty.Winsize{
			Rows: uint16(height),
			Cols: uint16(width),
		})
	}
}

func (tw *TermWindow) Close() {
	if tw.emu != nil {
		_ = tw.emu.Emulator.Close()
	}
	if tw.cmd != nil && tw.cmd.Process != nil {
		_ = tw.cmd.Process.Kill()
		_ = tw.cmd.Wait()
	}
	if tw.ptyFile != nil {
		_ = tw.ptyFile.Close()
	}
	tw.closed = true
}

func (tw *TermWindow) View() string {
	if tw.err != nil {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Render(
			fmt.Sprintf("terminal error: %v", tw.err),
		)
	}
	if tw.closed {
		return "shell exited"
	}
	return tw.emu.Render()
}

func (tw *TermWindow) keyToBytes(msg tea.KeyMsg) []byte {
	var prefix []byte
	if msg.Alt {
		prefix = []byte{0x1b}
	}

	switch msg.Type { //nolint:exhaustive
	case tea.KeyRunes:
		return append(prefix, []byte(string(msg.Runes))...)
	case tea.KeyEnter:
		return append(prefix, '\r')
	case tea.KeyBackspace, tea.KeyCtrlH:
		return append(prefix, 0x7f)
	case tea.KeyTab:
		return append(prefix, '\t')
	case tea.KeySpace:
		return append(prefix, ' ')
	case tea.KeyEsc:
		return append(prefix, 0x1b)

	case tea.KeyUp:
		return append(prefix, tw.arrowSeq('A')...)
	case tea.KeyDown:
		return append(prefix, tw.arrowSeq('B')...)
	case tea.KeyRight:
		return append(prefix, tw.arrowSeq('C')...)
	case tea.KeyLeft:
		return append(prefix, tw.arrowSeq('D')...)

	case tea.KeyHome:
		if tw.appCursorKeys {
			return append(prefix, []byte("\x1bOH")...)
		}
		return append(prefix, []byte("\x1b[H")...)
	case tea.KeyEnd:
		if tw.appCursorKeys {
			return append(prefix, []byte("\x1bOF")...)
		}
		return append(prefix, []byte("\x1b[F")...)

	case tea.KeyDelete:
		return append(prefix, []byte("\x1b[3~")...)
	case tea.KeyPgUp:
		return append(prefix, []byte("\x1b[5~")...)
	case tea.KeyPgDown:
		return append(prefix, []byte("\x1b[6~")...)
	case tea.KeyInsert:
		return append(prefix, []byte("\x1b[2~")...)

	case tea.KeyShiftTab:
		return append(prefix, []byte("\x1b[Z")...)

	case tea.KeyF1:
		return append(prefix, []byte("\x1bOP")...)
	case tea.KeyF2:
		return append(prefix, []byte("\x1bOQ")...)
	case tea.KeyF3:
		return append(prefix, []byte("\x1bOR")...)
	case tea.KeyF4:
		return append(prefix, []byte("\x1bOS")...)
	case tea.KeyF5:
		return append(prefix, []byte("\x1b[15~")...)
	case tea.KeyF6:
		return append(prefix, []byte("\x1b[17~")...)
	case tea.KeyF7:
		return append(prefix, []byte("\x1b[18~")...)
	case tea.KeyF8:
		return append(prefix, []byte("\x1b[19~")...)
	case tea.KeyF9:
		return append(prefix, []byte("\x1b[20~")...)
	case tea.KeyF10:
		return append(prefix, []byte("\x1b[21~")...)
	case tea.KeyF11:
		return append(prefix, []byte("\x1b[23~")...)
	case tea.KeyF12:
		return append(prefix, []byte("\x1b[24~")...)
	}

	ctrlKeys := map[tea.KeyType]byte{
		tea.KeyCtrlA: 0x01, tea.KeyCtrlB: 0x02, tea.KeyCtrlC: 0x03,
		tea.KeyCtrlD: 0x04, tea.KeyCtrlE: 0x05, tea.KeyCtrlF: 0x06,
		tea.KeyCtrlG: 0x07, tea.KeyCtrlH: 0x08, tea.KeyCtrlK: 0x0b,
		tea.KeyCtrlL: 0x0c, tea.KeyCtrlN: 0x0e, tea.KeyCtrlO: 0x0f,
		tea.KeyCtrlP: 0x10, tea.KeyCtrlQ: 0x11, tea.KeyCtrlR: 0x12,
		tea.KeyCtrlS: 0x13, tea.KeyCtrlT: 0x14, tea.KeyCtrlU: 0x15,
		tea.KeyCtrlV: 0x16, tea.KeyCtrlW: 0x17, tea.KeyCtrlX: 0x18,
		tea.KeyCtrlY: 0x19, tea.KeyCtrlZ: 0x1a,
		tea.KeyCtrlBackslash: 0x1c,
	}
	if b, ok := ctrlKeys[msg.Type]; ok {
		return append(prefix, b)
	}

	return nil
}

func (tw *TermWindow) arrowSeq(dir byte) []byte {
	if tw.appCursorKeys {
		return []byte{0x1b, 'O', dir}
	}
	return []byte{0x1b, '[', dir}
}
