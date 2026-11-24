package input

import (
	"log"
	"sync"
)

type Mock struct {
	Logger *log.Logger
	mu     sync.Mutex
	states map[int]GamepadState
}

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
func (m *Mock) GamepadSlotCount() int  { return MaxGamepadSlots }
func (m *Mock) GamepadStatus() string {
	return "Mock virtual gamepad backend is available for testing (up to 4 players)."
}
func (m *Mock) SetGamepadState(state GamepadState) error {
	return m.SetGamepadStateSlot(0, state)
}
func (m *Mock) SetGamepadStateSlot(slot int, state GamepadState) error {
	if slot < 0 || slot >= MaxGamepadSlots {
		return ErrUnsupported
	}
	if err := ValidateGamepadState(state); err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.states == nil {
		m.states = make(map[int]GamepadState)
	}
	m.states[slot] = state
	m.logf("gamepad slot %d state left=(%.2f,%.2f) right=(%.2f,%.2f) buttons=%v", slot, state.LeftX, state.LeftY, state.RightX, state.RightY, state.Buttons)
	return nil
}
func (m *Mock) ResetGamepad() error { return m.ResetGamepadSlot(0) }
func (m *Mock) ResetGamepadSlot(slot int) error {
	if slot < 0 || slot >= MaxGamepadSlots {
		return ErrUnsupported
	}
	m.mu.Lock()
	delete(m.states, slot)
	m.mu.Unlock()
	m.logf("gamepad slot %d reset", slot)
	return nil
}
