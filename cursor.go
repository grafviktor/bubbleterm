package main

import (
	"bytes"
	"io"
	"strings"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/vt"
)

func (tw *TermWindow) renderWithCursor() string {
	if tw.emu == nil {
		return ""
	}
	if !tw.cursorVisible {
		return tw.emu.Render()
	}

	pos := tw.emu.CursorPosition()
	var b strings.Builder
	for y := 0; y < tw.height; y++ {
		if y > 0 {
			b.WriteByte('\n')
		}
		renderTerminalLine(&b, tw.emu, y, pos.X, pos.Y)
	}
	return b.String()
}

func renderTerminalLine(buf io.StringWriter, emu *vt.SafeEmulator, y, cursorX, cursorY int) {
	var pen uv.Style
	var link uv.Link
	var pending bytes.Buffer

	showCursor := y == cursorY

	for x := 0; x < emu.Width(); {
		cell := emu.CellAt(x, y)
		if cell == nil || cell.IsZero() {
			x++
			continue
		}

		c := cell
		if showCursor && x == cursorX {
			c = cursorCell(cell)
		}

		writeTerminalCell(buf, &pen, &link, &pending, c)

		w := cell.Width
		if w <= 0 {
			w = 1
		}
		x += w
	}

	if pending.Len() > 0 {
		_, _ = buf.WriteString(pending.String())
	}
	if link.URL != "" {
		_, _ = buf.WriteString(ansi.ResetHyperlink())
	}
	if !pen.IsZero() {
		_, _ = buf.WriteString(ansi.ResetStyle)
	}
}

func writeTerminalCell(buf io.StringWriter, pen *uv.Style, link *uv.Link, pending *bytes.Buffer, c *uv.Cell) {
	if c.IsZero() {
		return
	}
	if c.Equal(&uv.EmptyCell) {
		if !pen.IsZero() {
			_, _ = buf.WriteString(ansi.ResetStyle)
			*pen = uv.Style{}
		}
		if !link.IsZero() {
			_, _ = buf.WriteString(ansi.ResetHyperlink())
			*link = uv.Link{}
		}
		pending.WriteByte(' ')
		return
	}

	if pending.Len() > 0 {
		_, _ = buf.WriteString(pending.String())
		pending.Reset()
	}

	if c.Style.IsZero() && !pen.IsZero() {
		_, _ = buf.WriteString(ansi.ResetStyle)
		*pen = uv.Style{}
	}
	if !c.Style.Equal(pen) {
		_, _ = buf.WriteString(c.Style.Diff(pen))
		*pen = c.Style
	}

	if c.Link != *link && link.URL != "" {
		_, _ = buf.WriteString(ansi.ResetHyperlink())
		*link = uv.Link{}
	}
	if c.Link != *link {
		_, _ = buf.WriteString(ansi.SetHyperlink(c.Link.URL, c.Link.Params))
		*link = c.Link
	}

	_, _ = buf.WriteString(c.String())
}

func cursorCell(cell *uv.Cell) *uv.Cell {
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
