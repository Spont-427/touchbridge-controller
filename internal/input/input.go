package input

import "errors"

var ErrUnsupported = errors.New("input injection is not supported on this platform")

// Backend injects input into the local desktop session.
// Implementations must release all held keyboard, mouse and gamepad state when a client disconnects.
type Backend interface {
	MoveMouse(dx, dy int) error
	MouseButton(button string, down bool) error
	Scroll(amount int) error
	Key(key string, down bool) error
	ReleaseAll() error
	Close() error
}

// ValidKey reports whether key is in the deliberately small, portable key vocabulary.
func ValidKey(key string) bool {
	if len(key) == 1 {
		c := key[0]
		return (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')
	}
	switch key {
	case "ENTER", "ESC", "SPACE", "TAB", "BACKSPACE", "DELETE",
		"SHIFT", "CTRL", "ALT", "UP", "DOWN", "LEFT", "RIGHT",
		"HOME", "END", "PAGEUP", "PAGEDOWN", "F1", "F2", "F3", "F4",
		"F5", "F6", "F7", "F8", "F9", "F10", "F11", "F12":
		return true
	default:
		return false
	}
}

func ValidButton(button string) bool {
	return button == "left" || button == "right" || button == "middle"
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
