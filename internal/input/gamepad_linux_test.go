//go:build linux && (amd64 || arm64)

package input

import "testing"

func TestLinuxDigitalTriggerButtonCodes(t *testing.T) {
	codes := linuxGamepadButtonCodes()
	if got, ok := codes["LT"]; !ok || got != btnTL2 {
		t.Fatalf("LT digital mapping = %v, want BTN_TL2 (%v)", got, btnTL2)
	}
	if got, ok := codes["RT"]; !ok || got != btnTR2 {
		t.Fatalf("RT digital mapping = %v, want BTN_TR2 (%v)", got, btnTR2)
	}
}
