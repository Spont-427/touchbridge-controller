package web

import "github.com/Spont-427/touchbridge-controller/internal/input"

// inputOwnership tracks which authenticated sessions currently hold shared
// desktop keys and mouse buttons. A key/button is sent to the OS only for the
// first owner and released only after its final owner lets go.
// All methods are called while Server.inputMu is held.
type inputOwnership struct {
	backend input.Backend
	keys    map[string]map[*controllerSession]struct{}
	buttons map[string]map[*controllerSession]struct{}
}

func newInputOwnership(backend input.Backend) *inputOwnership {
	return &inputOwnership{
		backend: backend,
		keys:    make(map[string]map[*controllerSession]struct{}),
		buttons: make(map[string]map[*controllerSession]struct{}),
	}
}

func (o *inputOwnership) setKey(session *controllerSession, key string, down bool) error {
	if down {
		if session.heldKeys[key] {
			return nil
		}
		owners := o.keys[key]
		if len(owners) == 0 {
			if err := o.backend.Key(key, true); err != nil {
				return err
			}
			owners = make(map[*controllerSession]struct{})
			o.keys[key] = owners
		}
		owners[session] = struct{}{}
		session.heldKeys[key] = true
		return nil
	}

	if !session.heldKeys[key] {
		return nil
	}
	owners := o.keys[key]
	if len(owners) == 1 {
		if err := o.backend.Key(key, false); err != nil {
			return err
		}
		delete(o.keys, key)
	} else {
		delete(owners, session)
	}
	delete(session.heldKeys, key)
	return nil
}

func (o *inputOwnership) setMouseButton(session *controllerSession, button string, down bool) error {
	if down {
		if session.heldButtons[button] {
			return nil
		}
		owners := o.buttons[button]
		if len(owners) == 0 {
			if err := o.backend.MouseButton(button, true); err != nil {
				return err
			}
			owners = make(map[*controllerSession]struct{})
			o.buttons[button] = owners
		}
		owners[session] = struct{}{}
		session.heldButtons[button] = true
		return nil
	}

	if !session.heldButtons[button] {
		return nil
	}
	owners := o.buttons[button]
	if len(owners) == 1 {
		if err := o.backend.MouseButton(button, false); err != nil {
			return err
		}
		delete(o.buttons, button)
	} else {
		delete(owners, session)
	}
	delete(session.heldButtons, button)
	return nil
}

// controllerSession scopes one authenticated phone to its assigned virtual
// gamepad slot and its own holds on shared desktop input devices.
type controllerSession struct {
	backend     input.Backend
	ownership   *inputOwnership
	slot        int
	heldKeys    map[string]bool
	heldButtons map[string]bool
	gamepadUsed bool
}

func newControllerSession(backend input.Backend, slot int, ownership *inputOwnership) *controllerSession {
	return &controllerSession{
		backend:     backend,
		ownership:   ownership,
		slot:        slot,
		heldKeys:    make(map[string]bool),
		heldButtons: make(map[string]bool),
	}
}

func (s *controllerSession) MoveMouse(dx, dy int) error { return s.backend.MoveMouse(dx, dy) }

func (s *controllerSession) MouseButton(button string, down bool) error {
	return s.ownership.setMouseButton(s, button, down)
}

func (s *controllerSession) Scroll(amount int) error { return s.backend.Scroll(amount) }

func (s *controllerSession) Key(key string, down bool) error {
	return s.ownership.setKey(s, key, down)
}

func (s *controllerSession) GamepadAvailable() bool {
	gamepad, ok := s.backend.(input.GamepadBackend)
	return ok && gamepad.GamepadAvailable()
}

func (s *controllerSession) SetGamepadState(state input.GamepadState) error {
	if multi, ok := s.backend.(input.GamepadSlotBackend); ok {
		if err := multi.SetGamepadStateSlot(s.slot, state); err != nil {
			return err
		}
		s.gamepadUsed = true
		return nil
	}
	if s.slot != 0 {
		return input.ErrUnsupported
	}
	gamepad, ok := s.backend.(input.GamepadBackend)
	if !ok || !gamepad.GamepadAvailable() {
		return input.ErrUnsupported
	}
	if err := gamepad.SetGamepadState(state); err != nil {
		return err
	}
	s.gamepadUsed = true
	return nil
}

func (s *controllerSession) ResetGamepad() error {
	var err error
	if multi, ok := s.backend.(input.GamepadSlotBackend); ok {
		err = multi.ResetGamepadSlot(s.slot)
	} else if s.slot == 0 {
		if gamepad, ok := s.backend.(input.GamepadBackend); ok && gamepad.GamepadAvailable() {
			err = gamepad.ResetGamepad()
		}
	}
	if err == nil {
		s.gamepadUsed = false
	}
	return err
}

// ReleaseAll releases only this session's desktop controls and virtual pad.
// Shared keys/buttons remain down while any other controller still owns them.
func (s *controllerSession) ReleaseAll() error {
	var first error
	for key := range s.heldKeys {
		if err := s.ownership.setKey(s, key, false); err != nil && first == nil {
			first = err
		}
	}
	for button := range s.heldButtons {
		if err := s.ownership.setMouseButton(s, button, false); err != nil && first == nil {
			first = err
		}
	}
	if s.gamepadUsed {
		if err := s.ResetGamepad(); err != nil && first == nil {
			first = err
		}
	}
	return first
}

func (s *controllerSession) Close() error { return s.ReleaseAll() }
