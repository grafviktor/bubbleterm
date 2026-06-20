package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/vt"
)

// teaKeyMsgToKeyPressEvent is an adapter that converts a Bubble Tea key message into a vt.KeyPressEvent.
// When a use presses a key in Bubble Tea, it sends a tea.KeyMsg to the program. This function
// translates that message into a vt.KeyPressEvent, which is used by the terminal emulator to
// handle key presses.
func teaKeyMsgToKeyPressEvent(msg tea.KeyMsg) (vt.KeyPressEvent, bool) {
	mod := vt.KeyMod(0)
	if msg.Alt {
		mod |= vt.ModAlt
	}

	if code, ok := teaCtrlKeys[msg.Type]; ok {
		return vt.KeyPressEvent{Code: code, Mod: mod | vt.ModCtrl}, true
	}
	if code, ok := teaSpecialKeys[msg.Type]; ok {
		return vt.KeyPressEvent{Code: code, Mod: mod}, true
	}

	switch msg.Type { //nolint:exhaustive
	case tea.KeyRunes:
		return vt.KeyPressEvent{
			Code: msg.Runes[0],
			Text: string(msg.Runes),
			Mod:  mod,
		}, true
	case tea.KeyShiftTab:
		return vt.KeyPressEvent{Code: vt.KeyTab, Mod: mod | vt.ModShift}, true
	}

	return vt.KeyPressEvent{}, false
}

var teaCtrlKeys = map[tea.KeyType]rune{
	tea.KeyCtrlA: 'a', tea.KeyCtrlB: 'b', tea.KeyCtrlC: 'c',
	tea.KeyCtrlD: 'd', tea.KeyCtrlE: 'e', tea.KeyCtrlF: 'f',
	tea.KeyCtrlG: 'g', tea.KeyCtrlH: 'h', tea.KeyCtrlI: 'i',
	tea.KeyCtrlJ: 'j', tea.KeyCtrlK: 'k', tea.KeyCtrlL: 'l',
	tea.KeyCtrlM: 'm', tea.KeyCtrlN: 'n', tea.KeyCtrlO: 'o',
	tea.KeyCtrlP: 'p', tea.KeyCtrlQ: 'q', tea.KeyCtrlR: 'r',
	tea.KeyCtrlS: 's', tea.KeyCtrlT: 't', tea.KeyCtrlU: 'u',
	tea.KeyCtrlV: 'v', tea.KeyCtrlW: 'w', tea.KeyCtrlX: 'x',
	tea.KeyCtrlY: 'y', tea.KeyCtrlZ: 'z',
	tea.KeyCtrlBackslash: '\\',
}

var teaSpecialKeys = map[tea.KeyType]rune{
	tea.KeyEnter:     vt.KeyEnter,
	tea.KeyBackspace: vt.KeyBackspace,
	tea.KeyCtrlH:     vt.KeyBackspace,
	tea.KeyTab:       vt.KeyTab,
	tea.KeySpace:     vt.KeySpace,
	tea.KeyEsc:       vt.KeyEscape,

	tea.KeyUp:    vt.KeyUp,
	tea.KeyDown:  vt.KeyDown,
	tea.KeyLeft:  vt.KeyLeft,
	tea.KeyRight: vt.KeyRight,

	tea.KeyHome:   vt.KeyHome,
	tea.KeyEnd:    vt.KeyEnd,
	tea.KeyInsert: vt.KeyInsert,
	tea.KeyDelete: vt.KeyDelete,
	tea.KeyPgUp:   vt.KeyPgUp,
	tea.KeyPgDown: vt.KeyPgDown,

	tea.KeyF1:  vt.KeyF1,
	tea.KeyF2:  vt.KeyF2,
	tea.KeyF3:  vt.KeyF3,
	tea.KeyF4:  vt.KeyF4,
	tea.KeyF5:  vt.KeyF5,
	tea.KeyF6:  vt.KeyF6,
	tea.KeyF7:  vt.KeyF7,
	tea.KeyF8:  vt.KeyF8,
	tea.KeyF9:  vt.KeyF9,
	tea.KeyF10: vt.KeyF10,
	tea.KeyF11: vt.KeyF11,
	tea.KeyF12: vt.KeyF12,
}
