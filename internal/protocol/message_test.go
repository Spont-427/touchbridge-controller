package protocol

import (
	"testing"

	"github.com/Spont-427/touchbridge-controller/internal/input"
)

type recorder struct{ moves, buttons, scrolls, keys, releases int }

func (r *recorder) MoveMouse(dx, dy int) error         { r.moves++; return nil }
func (r *recorder) MouseButton(b string, d bool) error { r.buttons++; return nil }
func (r *recorder) Scroll(a int) error                 { r.scrolls++; return nil }
func (r *recorder) Key(k string, d bool) error         { r.keys++; return nil }
func (r *recorder) ReleaseAll() error                  { r.releases++; return nil }
func (r *recorder) Close() error                       { return nil }

func TestDecodeAndApply(t *testing.T) {
	m, err := Decode([]byte(`{"type":"mouse_move","dx":10,"dy":-3}`))
	if err != nil {
		t.Fatal(err)
	}
	r := &recorder{}
	if err = Apply(m, r); err != nil {
		t.Fatal(err)
	}
	if r.moves != 1 {
		t.Fatalf("moves=%d", r.moves)
	}
}
func TestRejectsOutOfRangeInput(t *testing.T) {
	tests := []Message{{Type: "mouse_move", DX: 501}, {Type: "mouse_move", DX: -int(^uint(0)>>1) - 1}, {Type: "scroll", Amount: 21}, {Type: "key", Key: "WIN"}, {Type: "mouse_button", Button: "extra"}}
	for _, m := range tests {
		if err := Apply(m, &recorder{}); err == nil {
			t.Errorf("expected rejection for %+v", m)
		}
	}
}
func TestRejectsOversizeMessage(t *testing.T) {
	b := make([]byte, MaxMessageBytes+1)
	if _, err := Decode(b); err == nil {
		t.Fatal("expected size rejection")
	}
}

var _ input.Backend = (*recorder)(nil)

func TestDecodePINAuthentication(t *testing.T) {
	m, err := Decode([]byte(`{"type":"auth","pin":"012345"}`))
	if err != nil {
		t.Fatal(err)
	}
	if m.Type != "auth" || m.PIN != "012345" {
		t.Fatalf("unexpected auth message: %+v", m)
	}
}

func FuzzDecodeNeverPanics(f *testing.F) {
	f.Add([]byte(`{"type":"mouse_move","dx":1,"dy":-1}`))
	f.Add([]byte(`{"type":"gamepad_state","left_x":0.5,"buttons":["A"]}`))
	f.Add([]byte{0xff, 0x00, 0x7f})
	f.Fuzz(func(t *testing.T, raw []byte) {
		if len(raw) > MaxMessageBytes+1 {
			t.Skip()
		}
		_, _ = Decode(raw)
	})
}

func TestGamepadMessagesAreValidatedBeforeInjection(t *testing.T) {
	base := Message{Type: "gamepad_state", LeftX: 0.5, LeftY: -0.25, Buttons: []string{"A", "LB", "DPAD_UP"}}
	backend := &gamepadRecorder{}
	if err := Apply(base, backend); err != nil {
		t.Fatalf("valid gamepad state rejected: %v", err)
	}
	if backend.states != 1 {
		t.Fatalf("gamepad state calls=%d, want 1", backend.states)
	}
	invalid := base
	invalid.LeftX = 1.01
	if err := Apply(invalid, backend); err == nil {
		t.Fatal("out-of-range axis was accepted")
	}
	invalid = base
	invalid.Buttons = []string{"A", "A"}
	if err := Apply(invalid, backend); err == nil {
		t.Fatal("duplicate button was accepted")
	}
}

type gamepadRecorder struct {
	recorder
	states int
}

func (r *gamepadRecorder) GamepadAvailable() bool { return true }
func (r *gamepadRecorder) SetGamepadState(input.GamepadState) error {
	r.states++
	return nil
}
func (r *gamepadRecorder) ResetGamepad() error { return nil }
