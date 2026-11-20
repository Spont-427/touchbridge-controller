package protocol

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/Spont-427/touchbridge-controller/internal/input"
)

const MaxMessageBytes = 4096

var ErrBadMessage = errors.New("invalid input message")

type Message struct {
	Type         string   `json:"type"`
	PIN          string   `json:"pin,omitempty"`
	DeviceToken  string   `json:"device_token,omitempty"`
	DeviceName   string   `json:"device_name,omitempty"`
	Remember     bool     `json:"remember_device,omitempty"`
	DX           int      `json:"dx,omitempty"`
	DY           int      `json:"dy,omitempty"`
	Button       string   `json:"button,omitempty"`
	Down         bool     `json:"down,omitempty"`
	Amount       int      `json:"amount,omitempty"`
	Key          string   `json:"key,omitempty"`
	Timestamp    float64  `json:"timestamp,omitempty"`
	LeftX        float64  `json:"left_x,omitempty"`
	LeftY        float64  `json:"left_y,omitempty"`
	RightX       float64  `json:"right_x,omitempty"`
	RightY       float64  `json:"right_y,omitempty"`
	LeftTrigger  float64  `json:"left_trigger,omitempty"`
	RightTrigger float64  `json:"right_trigger,omitempty"`
	Buttons      []string `json:"buttons,omitempty"`
}

func Decode(raw []byte) (Message, error) {
	var m Message
	if len(raw) == 0 || len(raw) > MaxMessageBytes {
		return m, ErrBadMessage
	}
	if err := json.Unmarshal(raw, &m); err != nil {
		return m, fmt.Errorf("%w: %v", ErrBadMessage, err)
	}
	return m, nil
}

func Apply(m Message, backend input.Backend) error {
	switch m.Type {
	case "mouse_move":
		if m.DX < -500 || m.DX > 500 || m.DY < -500 || m.DY > 500 {
			return ErrBadMessage
		}
		return backend.MoveMouse(m.DX, m.DY)
	case "mouse_button":
		if !input.ValidButton(m.Button) {
			return ErrBadMessage
		}
		return backend.MouseButton(m.Button, m.Down)
	case "scroll":
		if m.Amount < -20 || m.Amount > 20 {
			return ErrBadMessage
		}
		return backend.Scroll(m.Amount)
	case "key":
		if !input.ValidKey(m.Key) {
			return ErrBadMessage
		}
		return backend.Key(m.Key, m.Down)
	case "gamepad_state":
		state := input.GamepadState{
			LeftX: m.LeftX, LeftY: m.LeftY, RightX: m.RightX, RightY: m.RightY,
			LeftTrigger: m.LeftTrigger, RightTrigger: m.RightTrigger, Buttons: m.Buttons,
		}
		if err := input.ValidateGamepadState(state); err != nil {
			return ErrBadMessage
		}
		gamepad, ok := backend.(input.GamepadBackend)
		if !ok || !gamepad.GamepadAvailable() {
			return input.ErrUnsupported
		}
		return gamepad.SetGamepadState(state)
	case "gamepad_reset":
		gamepad, ok := backend.(input.GamepadBackend)
		if !ok || !gamepad.GamepadAvailable() {
			return input.ErrUnsupported
		}
		return gamepad.ResetGamepad()
	case "release_all":
		if gamepad, ok := backend.(input.GamepadBackend); ok && gamepad.GamepadAvailable() {
			_ = gamepad.ResetGamepad()
		}
		return backend.ReleaseAll()
	default:
		return ErrBadMessage
	}
}
