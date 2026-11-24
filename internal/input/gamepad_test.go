package input

import (
	"math"
	"testing"
)

func TestValidateGamepadState(t *testing.T) {
	valid := GamepadState{
		LeftX: -1, LeftY: 1, RightX: 0.25, RightY: -0.5,
		LeftTrigger: 0, RightTrigger: 1,
		Buttons: []string{"A", "LB", "DPAD_UP"},
	}
	if err := ValidateGamepadState(valid); err != nil {
		t.Fatalf("valid state rejected: %v", err)
	}
	tests := []struct {
		name string
		edit func(*GamepadState)
	}{
		{"axis above range", func(s *GamepadState) { s.LeftX = 1.001 }},
		{"axis below range", func(s *GamepadState) { s.RightY = -1.001 }},
		{"NaN axis", func(s *GamepadState) { s.LeftX = math.NaN() }},
		{"infinite axis", func(s *GamepadState) { s.LeftY = math.Inf(1) }},
		{"negative trigger", func(s *GamepadState) { s.LeftTrigger = -0.01 }},
		{"trigger above range", func(s *GamepadState) { s.RightTrigger = 1.01 }},
		{"unknown button", func(s *GamepadState) { s.Buttons = []string{"MOUSE"} }},
		{"duplicate button", func(s *GamepadState) { s.Buttons = []string{"A", "A"} }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			state := valid
			state.Buttons = append([]string(nil), valid.Buttons...)
			test.edit(&state)
			if err := ValidateGamepadState(state); err == nil {
				t.Fatal("invalid gamepad state was accepted")
			}
		})
	}
}

func TestDigitalTriggersAreAcceptedAsButtons(t *testing.T) {
	state := GamepadState{Buttons: []string{"LT", "RT"}}
	if err := ValidateGamepadState(state); err != nil {
		t.Fatalf("digital trigger buttons rejected: %v", err)
	}
	if !ValidGamepadButton("LT") || !ValidGamepadButton("RT") {
		t.Fatal("LT and RT must be recognised as digital buttons")
	}
}

func TestMockGamepadSlotsAreIndependentAndResettable(t *testing.T) {
	mock := &Mock{}
	stateOne := GamepadState{LeftX: 0.5, Buttons: []string{"A"}}
	stateTwo := GamepadState{LeftX: -0.5, Buttons: []string{"B"}}
	if err := mock.SetGamepadStateSlot(0, stateOne); err != nil {
		t.Fatal(err)
	}
	if err := mock.SetGamepadStateSlot(1, stateTwo); err != nil {
		t.Fatal(err)
	}
	if err := mock.ResetGamepadSlot(1); err != nil {
		t.Fatal(err)
	}
	if got := mock.states[0].LeftX; got != stateOne.LeftX {
		t.Fatalf("player one state changed after player two reset: %v", got)
	}
	if _, ok := mock.states[1]; ok {
		t.Fatal("player two state remained after reset")
	}
	if err := mock.SetGamepadStateSlot(MaxGamepadSlots, stateOne); err == nil {
		t.Fatal("out-of-range gamepad slot was accepted")
	}
}
