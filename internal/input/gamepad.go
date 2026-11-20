package input

import (
	"errors"
	"math"
)

// GamepadState is a normalized Xbox-style controller snapshot. Stick axes use
// -1..1, triggers use 0..1, and Buttons contains the currently held controls.
type GamepadState struct {
	LeftX        float64  `json:"left_x"`
	LeftY        float64  `json:"left_y"`
	RightX       float64  `json:"right_x"`
	RightY       float64  `json:"right_y"`
	LeftTrigger  float64  `json:"left_trigger"`
	RightTrigger float64  `json:"right_trigger"`
	Buttons      []string `json:"buttons"`
}

// GamepadBackend is optional: a platform can still support mouse/keyboard when
// no suitable virtual gamepad interface is available.
type GamepadBackend interface {
	GamepadAvailable() bool
	SetGamepadState(GamepadState) error
	ResetGamepad() error
}

// GamepadStatusProvider returns a human-readable explanation for the
// capability response. It must not include local paths, secrets or PINs.
type GamepadStatusProvider interface {
	GamepadStatus() string
}

var ErrInvalidGamepadState = errors.New("invalid gamepad state")

var gamepadButtons = map[string]bool{
	"A": true, "B": true, "X": true, "Y": true,
	"LB": true, "RB": true, "LT": true, "RT": true, "BACK": true, "START": true,
	"GUIDE": true, "LS": true, "RS": true,
	"DPAD_UP": true, "DPAD_DOWN": true, "DPAD_LEFT": true, "DPAD_RIGHT": true,
}

func ValidGamepadButton(button string) bool { return gamepadButtons[button] }

func ValidateGamepadState(s GamepadState) error {
	axes := []float64{s.LeftX, s.LeftY, s.RightX, s.RightY}
	for _, axis := range axes {
		if math.IsNaN(axis) || math.IsInf(axis, 0) || axis < -1 || axis > 1 {
			return ErrInvalidGamepadState
		}
	}
	triggers := []float64{s.LeftTrigger, s.RightTrigger}
	for _, trigger := range triggers {
		if math.IsNaN(trigger) || math.IsInf(trigger, 0) || trigger < 0 || trigger > 1 {
			return ErrInvalidGamepadState
		}
	}
	if len(s.Buttons) > len(gamepadButtons) {
		return ErrInvalidGamepadState
	}
	seen := make(map[string]bool, len(s.Buttons))
	for _, button := range s.Buttons {
		if !ValidGamepadButton(button) || seen[button] {
			return ErrInvalidGamepadState
		}
		seen[button] = true
	}
	return nil
}
