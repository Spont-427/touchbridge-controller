//go:build linux && (amd64 || arm64)

package input

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"syscall"
	"time"
)

const (
	evSyn        = 0x00
	evKey        = 0x01
	evRel        = 0x02
	relX         = 0x00
	relY         = 0x01
	relWheel     = 0x08
	btnLeft      = 0x110
	btnRight     = 0x111
	btnMiddle    = 0x112
	uiSetEVBit   = 0x40045564
	uiSetKeyBit  = 0x40045565
	uiSetRelBit  = 0x40045566
	uiSetAbsBit  = 0x40045567
	uiDevCreate  = 0x5501
	uiDevDestroy = 0x5502
	evAbs        = 0x03
	absX         = 0x00
	absY         = 0x01
	absZ         = 0x02
	absRX        = 0x03
	absRY        = 0x04
	absRZ        = 0x05
	btnSouth     = 0x130
	btnEast      = 0x131
	btnNorth     = 0x133
	btnWest      = 0x134
	btnTL        = 0x136
	btnTR        = 0x137
	btnTL2       = 0x138
	btnTR2       = 0x139
	btnSelect    = 0x13a
	btnStart     = 0x13b
	btnMode      = 0x13c
	btnThumbL    = 0x13d
	btnThumbR    = 0x13e
	btnDpadUp    = 0x220
	btnDpadDown  = 0x221
	btnDpadLeft  = 0x222
	btnDpadRight = 0x223
)

type Linux struct {
	file         *os.File
	gamepadFile  *os.File
	gamepadReady bool
	gamepadErr   string
	heldKeys     map[string]bool
	heldBtns     map[string]bool
	heldGamepad  map[string]bool
}

func New() (Backend, error) {
	f, err := os.OpenFile("/dev/uinput", os.O_WRONLY, 0)
	if err != nil {
		return nil, fmt.Errorf("open /dev/uinput: %w (load the uinput kernel module and grant access; see docs/linux-setup.md)", err)
	}
	l := &Linux{file: f, heldKeys: map[string]bool{}, heldBtns: map[string]bool{}, heldGamepad: map[string]bool{}}
	fail := func(e error) (Backend, error) { _ = f.Close(); return nil, e }
	if err := l.ioctl(uiSetEVBit, evKey); err != nil {
		return fail(fmt.Errorf("enable EV_KEY: %w", err))
	}
	if err := l.ioctl(uiSetEVBit, evRel); err != nil {
		return fail(fmt.Errorf("enable EV_REL: %w", err))
	}
	for _, code := range []int{relX, relY, relWheel} {
		if err := l.ioctl(uiSetRelBit, code); err != nil {
			return fail(fmt.Errorf("enable relative axis: %w", err))
		}
	}
	for _, code := range linuxKeyCodes() {
		if err := l.ioctl(uiSetKeyBit, code); err != nil {
			return fail(fmt.Errorf("enable key code %d: %w", code, err))
		}
	}
	for _, code := range []int{btnLeft, btnRight, btnMiddle} {
		if err := l.ioctl(uiSetKeyBit, code); err != nil {
			return fail(fmt.Errorf("enable mouse button: %w", err))
		}
	}
	dev := make([]byte, 1116) // struct uinput_user_dev on 64-bit Linux.
	copy(dev[:80], []byte("TouchBridge Virtual Input"))
	binary.LittleEndian.PutUint16(dev[80:82], 0x03) // BUS_USB
	binary.LittleEndian.PutUint16(dev[82:84], 0x1D6B)
	binary.LittleEndian.PutUint16(dev[84:86], 0x0104)
	binary.LittleEndian.PutUint16(dev[86:88], 1)
	if _, err := f.Write(dev); err != nil {
		return fail(fmt.Errorf("write uinput device descriptor: %w", err))
	}
	if err := l.ioctl(uiDevCreate, 0); err != nil {
		return fail(fmt.Errorf("create uinput device: %w", err))
	}
	// Gamepad creation is optional: mouse/keyboard remains usable on systems
	// whose uinput policy permits only the primary virtual input device.
	if err := l.createGamepad(); err != nil {
		l.gamepadErr = err.Error()
	}
	return l, nil
}

func (l *Linux) ioctl(req uintptr, value int) error {
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, l.file.Fd(), req, uintptr(value))
	if errno != 0 {
		return errno
	}
	return nil
}

type linuxEvent struct {
	typ   uint16
	code  uint16
	value int32
}

func (l *Linux) writeEvents(events []linuxEvent) error {
	if len(events) == 0 {
		return nil
	}
	const eventSize = 24 // input_event: timeval(16) + type + code + value on 64-bit Linux.
	var storage [eventSize * 3]byte
	if len(events) > 3 {
		return fmt.Errorf("too many input events in one batch: %d", len(events))
	}
	buf := storage[:eventSize*len(events)]
	now := time.Now()
	sec := uint64(now.Unix())
	usec := uint64(now.Nanosecond() / 1000)
	for i, event := range events {
		off := i * eventSize
		binary.LittleEndian.PutUint64(buf[off:off+8], sec)
		binary.LittleEndian.PutUint64(buf[off+8:off+16], usec)
		binary.LittleEndian.PutUint16(buf[off+16:off+18], event.typ)
		binary.LittleEndian.PutUint16(buf[off+18:off+20], event.code)
		binary.LittleEndian.PutUint32(buf[off+20:off+24], uint32(event.value))
	}
	n, err := l.file.Write(buf)
	if err != nil {
		return err
	}
	if n != len(buf) {
		return io.ErrShortWrite
	}
	return nil
}

func (l *Linux) writeGamepadEvents(events []linuxEvent) error {
	if len(events) == 0 || l.gamepadFile == nil {
		return ErrUnsupported
	}
	const eventSize = 24 // Linux input_event layout on amd64/arm64.
	buf := make([]byte, eventSize*len(events))
	now := time.Now()
	sec := uint64(now.Unix())
	usec := uint64(now.Nanosecond() / 1000)
	for i, event := range events {
		off := i * eventSize
		binary.LittleEndian.PutUint64(buf[off:off+8], sec)
		binary.LittleEndian.PutUint64(buf[off+8:off+16], usec)
		binary.LittleEndian.PutUint16(buf[off+16:off+18], event.typ)
		binary.LittleEndian.PutUint16(buf[off+18:off+20], event.code)
		binary.LittleEndian.PutUint32(buf[off+20:off+24], uint32(event.value))
	}
	n, err := l.gamepadFile.Write(buf)
	if err != nil {
		return err
	}
	if n != len(buf) {
		return io.ErrShortWrite
	}
	return nil
}

func (l *Linux) createGamepad() error {
	f, err := os.OpenFile("/dev/uinput", os.O_WRONLY, 0)
	if err != nil {
		return fmt.Errorf("open uinput gamepad: %w", err)
	}
	failed := func(err error) error { _ = f.Close(); return err }
	ioctl := func(req uintptr, value int) error {
		_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, f.Fd(), req, uintptr(value))
		if errno != 0 {
			return errno
		}
		return nil
	}
	for _, typ := range []int{evKey, evAbs} {
		if err := ioctl(uiSetEVBit, typ); err != nil {
			return failed(fmt.Errorf("enable gamepad event type %d: %w", typ, err))
		}
	}
	buttons := []int{btnSouth, btnEast, btnNorth, btnWest, btnTL, btnTR, btnTL2, btnTR2, btnSelect, btnStart, btnMode, btnThumbL, btnThumbR, btnDpadUp, btnDpadDown, btnDpadLeft, btnDpadRight}
	for _, code := range buttons {
		if err := ioctl(uiSetKeyBit, code); err != nil {
			return failed(fmt.Errorf("enable gamepad button %d: %w", code, err))
		}
	}
	axes := []int{absX, absY, absZ, absRX, absRY, absRZ}
	for _, code := range axes {
		if err := ioctl(uiSetAbsBit, code); err != nil {
			return failed(fmt.Errorf("enable gamepad axis %d: %w", code, err))
		}
	}
	// struct uinput_user_dev: name[80], input_id[8], ff_effects_max[4],
	// followed by four ABS_CNT (64) arrays of int32 values.
	dev := make([]byte, 1116)
	copy(dev[:80], []byte("TouchBridge Virtual Gamepad"))
	binary.LittleEndian.PutUint16(dev[80:82], 0x03)   // BUS_USB
	binary.LittleEndian.PutUint16(dev[82:84], 0x1209) // PID VID
	binary.LittleEndian.PutUint16(dev[84:86], 0xB001)
	binary.LittleEndian.PutUint16(dev[86:88], 1)
	for _, code := range []int{absX, absY, absRX, absRY} {
		binary.LittleEndian.PutUint32(dev[92+code*4:96+code*4], 32767)
		binary.LittleEndian.PutUint32(dev[348+code*4:352+code*4], uint32(0xFFFF8001))
	}
	for _, code := range []int{absZ, absRZ} {
		binary.LittleEndian.PutUint32(dev[92+code*4:96+code*4], 255)
		binary.LittleEndian.PutUint32(dev[348+code*4:352+code*4], 0)
	}
	if _, err := f.Write(dev); err != nil {
		return failed(fmt.Errorf("write gamepad descriptor: %w", err))
	}
	if err := ioctl(uiDevCreate, 0); err != nil {
		return failed(fmt.Errorf("create virtual gamepad: %w", err))
	}
	l.gamepadFile = f
	l.gamepadReady = true
	return nil
}

func (l *Linux) GamepadAvailable() bool { return l.gamepadReady }
func (l *Linux) GamepadStatus() string {
	if l.gamepadReady {
		return "Linux uinput virtual gamepad ready"
	}
	if l.gamepadErr != "" {
		return "Linux virtual gamepad unavailable: " + l.gamepadErr
	}
	return "Linux virtual gamepad unavailable"
}

func (l *Linux) SetGamepadState(state GamepadState) error {
	if !l.gamepadReady {
		return ErrUnsupported
	}
	if err := ValidateGamepadState(state); err != nil {
		return err
	}
	axis := []linuxEvent{
		{typ: evAbs, code: absX, value: int32(state.LeftX * 32767)},
		{typ: evAbs, code: absY, value: int32(-state.LeftY * 32767)},
		{typ: evAbs, code: absRX, value: int32(state.RightX * 32767)},
		{typ: evAbs, code: absRY, value: int32(-state.RightY * 32767)},
		{typ: evAbs, code: absZ, value: int32(state.LeftTrigger * 255)},
		{typ: evAbs, code: absRZ, value: int32(state.RightTrigger * 255)},
	}
	buttonCodes := linuxGamepadButtonCodes()
	wanted := make(map[string]bool, len(state.Buttons))
	for _, button := range state.Buttons {
		wanted[button] = true
	}
	events := make([]linuxEvent, 0, len(axis)+len(buttonCodes)+1)
	events = append(events, axis...)
	for button, code := range buttonCodes {
		if wanted[button] == l.heldGamepad[button] {
			continue
		}
		value := int32(0)
		if wanted[button] {
			value = 1
		}
		events = append(events, linuxEvent{typ: evKey, code: code, value: value})
	}
	events = append(events, syncEvent())
	if err := l.writeGamepadEvents(events); err != nil {
		return err
	}
	l.heldGamepad = wanted
	return nil
}

func linuxGamepadButtonCodes() map[string]uint16 {
	return map[string]uint16{
		"A": btnSouth, "B": btnEast, "X": btnWest, "Y": btnNorth,
		"LB": btnTL, "RB": btnTR, "LT": btnTL2, "RT": btnTR2,
		"BACK": btnSelect, "START": btnStart, "GUIDE": btnMode,
		"LS": btnThumbL, "RS": btnThumbR,
		"DPAD_UP": btnDpadUp, "DPAD_DOWN": btnDpadDown, "DPAD_LEFT": btnDpadLeft, "DPAD_RIGHT": btnDpadRight,
	}
}

func (l *Linux) ResetGamepad() error {
	if !l.gamepadReady {
		return nil
	}
	return l.SetGamepadState(GamepadState{})
}

func syncEvent() linuxEvent { return linuxEvent{typ: evSyn} }

func (l *Linux) MoveMouse(dx, dy int) error {
	var events [3]linuxEvent
	n := 0
	if dx != 0 {
		events[n] = linuxEvent{typ: evRel, code: relX, value: int32(clamp(dx, -2000, 2000))}
		n++
	}
	if dy != 0 {
		events[n] = linuxEvent{typ: evRel, code: relY, value: int32(clamp(dy, -2000, 2000))}
		n++
	}
	if n == 0 {
		return nil
	}
	events[n] = syncEvent()
	n++
	return l.writeEvents(events[:n])
}
func (l *Linux) MouseButton(button string, down bool) error {
	code, ok := map[string]uint16{"left": btnLeft, "right": btnRight, "middle": btnMiddle}[button]
	if !ok {
		return fmt.Errorf("invalid mouse button %q", button)
	}
	v := int32(0)
	if down {
		v = 1
	}
	if err := l.writeEvents([]linuxEvent{{typ: evKey, code: code, value: v}, syncEvent()}); err != nil {
		return err
	}
	if down {
		l.heldBtns[button] = true
	} else {
		delete(l.heldBtns, button)
	}
	return nil
}
func (l *Linux) Scroll(amount int) error {
	if amount == 0 {
		return nil
	}
	return l.writeEvents([]linuxEvent{{typ: evRel, code: relWheel, value: int32(clamp(amount, -20, 20))}, syncEvent()})
}
func (l *Linux) Key(key string, down bool) error {
	code, ok := linuxVK[key]
	if !ok {
		return fmt.Errorf("unsupported key %q", key)
	}
	v := int32(0)
	if down {
		v = 1
	}
	if err := l.writeEvents([]linuxEvent{{typ: evKey, code: code, value: v}, syncEvent()}); err != nil {
		return err
	}
	if down {
		l.heldKeys[key] = true
	} else {
		delete(l.heldKeys, key)
	}
	return nil
}
func (l *Linux) ReleaseAll() error {
	var first error
	if err := l.ResetGamepad(); err != nil && first == nil {
		first = err
	}
	for k := range l.heldKeys {
		if e := l.Key(k, false); e != nil && first == nil {
			first = e
		}
	}
	for b := range l.heldBtns {
		if e := l.MouseButton(b, false); e != nil && first == nil {
			first = e
		}
	}
	return first
}
func (l *Linux) Close() error {
	_ = l.ReleaseAll()
	if l.gamepadFile != nil {
		_, _, _ = syscall.Syscall(syscall.SYS_IOCTL, l.gamepadFile.Fd(), uiDevDestroy, 0)
		_ = l.gamepadFile.Close()
	}
	_ = l.ioctl(uiDevDestroy, 0)
	return l.file.Close()
}

func linuxVKTable() map[string]uint16 {
	m := map[string]uint16{"ENTER": 28, "ESC": 1, "SPACE": 57, "TAB": 15, "BACKSPACE": 14, "DELETE": 111, "SHIFT": 42, "CTRL": 29, "ALT": 56,
		"UP": 103, "DOWN": 108, "LEFT": 105, "RIGHT": 106, "HOME": 102, "END": 107, "PAGEUP": 104, "PAGEDOWN": 109,
		"F1": 59, "F2": 60, "F3": 61, "F4": 62, "F5": 63, "F6": 64, "F7": 65, "F8": 66, "F9": 67, "F10": 68, "F11": 87, "F12": 88}
	letters := []uint16{30, 48, 46, 32, 18, 33, 34, 35, 23, 36, 37, 38, 50, 49, 24, 25, 16, 19, 31, 20, 22, 47, 17, 45, 21, 44}
	for i, c := range "ABCDEFGHIJKLMNOPQRSTUVWXYZ" {
		m[string(c)] = letters[i]
	}
	digits := []uint16{11, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	for i, c := range "0123456789" {
		m[string(c)] = digits[i]
	}
	return m
}

var linuxVK = linuxVKTable()

func linuxKeyCodes() []int {
	out := make([]int, 0, len(linuxVK))
	for _, v := range linuxVK {
		out = append(out, int(v))
	}
	return out
}
