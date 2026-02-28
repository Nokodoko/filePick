package main

import (
	tea "charm.land/bubbletea/v2"
)

// keyAction represents an action triggered by a key press.
type keyAction int

const (
	keyNone keyAction = iota
	keyCursorDown
	keyCursorUp
	keyJumpTop
	keyJumpBottom
	keyEnterDir
	keyParentDir
	keyToggleHidden
	keyCycleSort
	keyQuit
)

// resolveKey maps a KeyPressMsg to a keyAction.
func resolveKey(msg tea.KeyPressMsg) keyAction {
	key := msg.Key()

	// Check for ctrl+c
	if key.Code == 'c' && key.Mod == tea.ModCtrl {
		return keyQuit
	}

	// Check special keys
	switch key.Code {
	case tea.KeyDown:
		return keyCursorDown
	case tea.KeyUp:
		return keyCursorUp
	case tea.KeyHome:
		return keyJumpTop
	case tea.KeyEnd:
		return keyJumpBottom
	case tea.KeyEnter:
		return keyEnterDir
	case tea.KeyBackspace:
		return keyParentDir
	case tea.KeyEsc:
		return keyQuit
	}

	// Check character keys
	switch msg.String() {
	case "j":
		return keyCursorDown
	case "k":
		return keyCursorUp
	case "g":
		return keyJumpTop
	case "G":
		return keyJumpBottom
	case "l":
		return keyEnterDir
	case "h":
		return keyParentDir
	case ".":
		return keyToggleHidden
	case "s":
		return keyCycleSort
	case "q":
		return keyQuit
	}

	return keyNone
}
