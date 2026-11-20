//go:build windows

package input

import (
	"encoding/binary"
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

type Windows struct {
	heldKeys map[string]bool
	heldBtns map[string]bool
}

func New() (Backend, error) {
	if err := user32.Load(); err != nil {
		return nil, fmt.Errorf("load user32.dll: %w", err)
	}
	return &Windows{heldKeys: map[string]bool{}, heldBtns: map[string]bool{}}, nil
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
func (w *Windows) Close() error { return w.ReleaseAll() }

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

// Windows gamepad injection is intentionally unavailable in this build.
// The retired ViGEmBus driver is not bundled. A maintained third-party
// UMDF-based provider is being evaluated, but it requires elevation for
// virtual-device creation; we will not silently elevate the entire desktop
// server or ship an unverified driver integration.
func (w *Windows) GamepadAvailable() bool { return false }
func (w *Windows) GamepadStatus() string {
	return "Windows virtual gamepad is not enabled yet. Mouse and keyboard work normally; no third-party driver is installed or required."
}
func (w *Windows) SetGamepadState(GamepadState) error { return ErrUnsupported }
func (w *Windows) ResetGamepad() error                { return nil }
