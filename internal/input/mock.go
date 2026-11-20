package input

import "log"

type Mock struct{ Logger *log.Logger }

func (m *Mock) logf(format string, args ...any) {
	if m.Logger != nil {
		m.Logger.Printf("DRY-RUN: "+format, args...)
	}
}
func (m *Mock) MoveMouse(dx, dy int) error { m.logf("mouse move dx=%d dy=%d", dx, dy); return nil }
func (m *Mock) MouseButton(button string, down bool) error {
	m.logf("mouse %s down=%t", button, down)
	return nil
}
func (m *Mock) Scroll(amount int) error         { m.logf("scroll amount=%d", amount); return nil }
func (m *Mock) Key(key string, down bool) error { m.logf("key %s down=%t", key, down); return nil }
func (m *Mock) ReleaseAll() error               { m.logf("release all held input"); return nil }
func (m *Mock) Close() error                    { return nil }

func (m *Mock) GamepadAvailable() bool { return true }
func (m *Mock) GamepadStatus() string {
	return "Mock virtual gamepad backend is available for testing."
}
func (m *Mock) SetGamepadState(state GamepadState) error {
	m.logf("gamepad state left=(%.2f,%.2f) right=(%.2f,%.2f) buttons=%v", state.LeftX, state.LeftY, state.RightX, state.RightY, state.Buttons)
	return nil
}
func (m *Mock) ResetGamepad() error { m.logf("gamepad reset"); return nil }
