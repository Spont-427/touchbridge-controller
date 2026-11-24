//go:build windows

package input

import (
	"encoding/binary"
	"errors"
	"fmt"
	"runtime"
	"syscall"
	"unsafe"
)

var (
	user32        = syscall.NewLazyDLL("user32.dll")
	sendInputProc = user32.NewProc("SendInput")
)

type winInput struct {
	Type uint32
	_    [4]byte // INPUT union is aligned to 8 bytes on 64-bit Windows.
	Data [32]byte
}

// gamepadProvider is the optional virtual-gamepad transport. Keeping it behind
// an interface lets a second Windows backend be added without changing the
// mouse and keyboard path or the protocol layer.
type gamepadProvider interface {
	GamepadAvailable() bool
	GamepadStatus() string
	SetGamepadState(GamepadState) error
	ResetGamepad() error
	Close() error
}

type Windows struct {
	heldKeys   map[string]bool
	heldBtns   map[string]bool
	gamepads   []gamepadProvider
	gamepadErr error
}

func New() (Backend, error) {
	if err := user32.Load(); err != nil {
		return nil, fmt.Errorf("load user32.dll: %w", err)
	}
	w := &Windows{heldKeys: map[string]bool{}, heldBtns: map[string]bool{}}
	gamepad, err := newViGEmProvider()
	if err != nil {
		// A missing driver must not disable mouse and keyboard control.
		w.gamepadErr = err
	} else {
		w.gamepads = []gamepadProvider{gamepad}
	}
	return w, nil
}

func (w *Windows) send(in *winInput) error {
	r, _, callErr := sendInputProc.Call(1, uintptr(unsafe.Pointer(in)), unsafe.Sizeof(*in))
	runtime.KeepAlive(in)
	if r != 1 {
		if callErr != syscall.Errno(0) {
			return fmt.Errorf("SendInput failed: %w", callErr)
		}
		return fmt.Errorf("SendInput did not inject the event")
	}
	return nil
}

func (w *Windows) mouse(dx, dy int32, data, flags uint32) error {
	var in winInput
	in.Type = 0 // INPUT_MOUSE
	binary.LittleEndian.PutUint32(in.Data[0:4], uint32(dx))
	binary.LittleEndian.PutUint32(in.Data[4:8], uint32(dy))
	binary.LittleEndian.PutUint32(in.Data[8:12], data)
	binary.LittleEndian.PutUint32(in.Data[12:16], flags)
	return w.send(&in)
}

func (w *Windows) keyboard(vk uint16, down bool) error {
	var in winInput
	in.Type = 1 // INPUT_KEYBOARD
	binary.LittleEndian.PutUint16(in.Data[0:2], vk)
	flags := uint32(0)
	if vk == 0x21 || vk == 0x22 || vk == 0x23 || vk == 0x24 || vk == 0x25 || vk == 0x26 || vk == 0x27 || vk == 0x28 || vk == 0x2E {
		flags |= 0x0001 // KEYEVENTF_EXTENDEDKEY
	}
	if !down {
		flags |= 0x0002 // KEYEVENTF_KEYUP
	}
	binary.LittleEndian.PutUint32(in.Data[4:8], flags)
	return w.send(&in)
}

func (w *Windows) MoveMouse(dx, dy int) error {
	// MOUSEEVENTF_MOVE_NOCOALESCE prevents Windows from merging successive
	// relative motion messages, which otherwise makes a high-rate touchpad feel
	// uneven on some desktops.
	return w.mouse(int32(clamp(dx, -2000, 2000)), int32(clamp(dy, -2000, 2000)), 0, 0x0001|0x2000)
}
func (w *Windows) MouseButton(button string, down bool) error {
	flags := map[string][2]uint32{"left": {0x0002, 0x0004}, "right": {0x0008, 0x0010}, "middle": {0x0020, 0x0040}}
	pair, ok := flags[button]
	if !ok {
		return fmt.Errorf("invalid mouse button %q", button)
	}
	flag := pair[1]
	if down {
		flag = pair[0]
	}
	if err := w.mouse(0, 0, 0, flag); err != nil {
		return err
	}
	if down {
		w.heldBtns[button] = true
	} else {
		delete(w.heldBtns, button)
	}
	return nil
}
func (w *Windows) Scroll(amount int) error {
	if amount == 0 {
		return nil
	}
	return w.mouse(0, 0, uint32(int32(clamp(amount, -20, 20)*120)), 0x0800)
}
func (w *Windows) Key(key string, down bool) error {
	vk, ok := windowsVK[key]
	if !ok {
		return fmt.Errorf("unsupported key %q", key)
	}
	if err := w.keyboard(vk, down); err != nil {
		return err
	}
	if down {
		w.heldKeys[key] = true
	} else {
		delete(w.heldKeys, key)
	}
	return nil
}
func (w *Windows) ReleaseAll() error {
	var first error
	for slot := range w.gamepads {
		if err := w.ResetGamepadSlot(slot); err != nil && first == nil {
			first = err
		}
	}
	for key := range w.heldKeys {
		if err := w.Key(key, false); err != nil && first == nil {
			first = err
		}
	}
	for btn := range w.heldBtns {
		if err := w.MouseButton(btn, false); err != nil && first == nil {
			first = err
		}
	}
	return first
}
func (w *Windows) Close() error {
	first := w.ReleaseAll()
	for _, gamepad := range w.gamepads {
		if gamepad != nil {
			if err := gamepad.Close(); err != nil && first == nil {
				first = err
			}
		}
	}
	w.gamepads = nil
	return first
}

var windowsVK = func() map[string]uint16 {
	m := map[string]uint16{"ENTER": 0x0D, "ESC": 0x1B, "SPACE": 0x20, "TAB": 0x09, "BACKSPACE": 0x08, "DELETE": 0x2E,
		"SHIFT": 0x10, "CTRL": 0x11, "ALT": 0x12, "UP": 0x26, "DOWN": 0x28, "LEFT": 0x25, "RIGHT": 0x27,
		"HOME": 0x24, "END": 0x23, "PAGEUP": 0x21, "PAGEDOWN": 0x22,
		"F1": 0x70, "F2": 0x71, "F3": 0x72, "F4": 0x73, "F5": 0x74, "F6": 0x75, "F7": 0x76, "F8": 0x77,
		"F9": 0x78, "F10": 0x79, "F11": 0x7A, "F12": 0x7B}
	for c := 'A'; c <= 'Z'; c++ {
		m[string(c)] = uint16(c)
	}
	for c := '0'; c <= '9'; c++ {
		m[string(c)] = uint16(c)
	}
	return m
}()

// Gamepad injection requires a virtual-device driver because Windows exposes
// no public API for it. ViGEmBus is used when installed; without it, the
// capability response advertises no gamepad and mouse/keyboard keep working.
func (w *Windows) GamepadAvailable() bool {
	return len(w.gamepads) > 0 && w.gamepads[0] != nil && w.gamepads[0].GamepadAvailable()
}

func (w *Windows) GamepadSlotCount() int { return MaxGamepadSlots }

func (w *Windows) GamepadStatus() string {
	if len(w.gamepads) > 0 && w.gamepads[0] != nil {
		return w.gamepads[0].GamepadStatus() + " (up to 4 players)"
	}
	if w.gamepadErr != nil && !errors.Is(w.gamepadErr, ErrUnsupported) {
		return "Windows virtual gamepad unavailable (" + w.gamepadErr.Error() + "). Install ViGEmBus 1.22.0 to enable an Xbox 360-compatible controller; mouse and keyboard work without it."
	}
	return "Windows virtual gamepad unavailable: install ViGEmBus 1.22.0 to enable an Xbox 360-compatible controller. Mouse and keyboard work without it."
}

func (w *Windows) gamepadSlot(slot int) (gamepadProvider, error) {
	if slot < 0 || slot >= MaxGamepadSlots {
		return nil, ErrUnsupported
	}
	for len(w.gamepads) <= slot {
		w.gamepads = append(w.gamepads, nil)
	}
	if w.gamepads[slot] != nil {
		return w.gamepads[slot], nil
	}
	gamepad, err := newViGEmProvider()
	if err != nil {
		w.gamepadErr = err
		return nil, err
	}
	w.gamepads[slot] = gamepad
	return gamepad, nil
}

func (w *Windows) SetGamepadState(state GamepadState) error {
	return w.SetGamepadStateSlot(0, state)
}

func (w *Windows) SetGamepadStateSlot(slot int, state GamepadState) error {
	gamepad, err := w.gamepadSlot(slot)
	if err != nil {
		return err
	}
	return gamepad.SetGamepadState(state)
}

func (w *Windows) ResetGamepad() error { return w.ResetGamepadSlot(0) }

func (w *Windows) ResetGamepadSlot(slot int) error {
	if slot < 0 || slot >= len(w.gamepads) || w.gamepads[slot] == nil {
		return nil
	}
	return w.gamepads[slot].ResetGamepad()
}
