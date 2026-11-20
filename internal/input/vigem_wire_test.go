package input

import (
	"encoding/binary"
	"testing"
)

func TestViGEmRequestsHaveExpectedShape(t *testing.T) {
	if got := len(vigemCheckVersionRequest()); got != 8 {
		t.Fatalf("check version request length = %d, want 8", got)
	}
	if got := len(vigemPluginRequest(1, vigemTargetXbox360Wired, 0x045E, 0x028E)); got != 16 {
		t.Fatalf("plugin request length = %d, want 16", got)
	}
	if got := len(vigemUnplugRequest(1)); got != 8 {
		t.Fatalf("unplug request length = %d, want 8", got)
	}
	if got := len(vigemWaitReadyRequest(1)); got != 8 {
		t.Fatalf("wait-ready request length = %d, want 8", got)
	}
	if got := len(vigemSubmitReport(1, GamepadState{})); got != 20 {
		t.Fatalf("submit report length = %d, want 20", got)
	}
}

func TestViGEmCheckVersionCarriesCommonVersion(t *testing.T) {
	request := vigemCheckVersionRequest()
	if got := binary.LittleEndian.Uint32(request[0:4]); got != vigemCheckVersionSize {
		t.Fatalf("Size field = %d, want %d", got, vigemCheckVersionSize)
	}
	if got := binary.LittleEndian.Uint32(request[4:8]); got != vigemCommonVersion {
		t.Fatalf("Version field = %d, want %d", got, vigemCommonVersion)
	}
}

func TestViGEmPluginRequestLayout(t *testing.T) {
	request := vigemPluginRequest(7, vigemTargetXbox360Wired, 0x045E, 0x028E)
	if got := binary.LittleEndian.Uint32(request[0:4]); got != vigemPluginSize {
		t.Fatalf("Size = %d, want %d", got, vigemPluginSize)
	}
	if got := binary.LittleEndian.Uint32(request[4:8]); got != 7 {
		t.Fatalf("SerialNo = %d, want 7", got)
	}
	if got := binary.LittleEndian.Uint32(request[8:12]); got != vigemTargetXbox360Wired {
		t.Fatalf("TargetType = %d, want %d", got, vigemTargetXbox360Wired)
	}
	if got := binary.LittleEndian.Uint16(request[12:14]); got != 0x045E {
		t.Fatalf("VendorId = %#x, want 0x045E", got)
	}
	if got := binary.LittleEndian.Uint16(request[14:16]); got != 0x028E {
		t.Fatalf("ProductId = %#x, want 0x028E", got)
	}
}

func TestXUSBButtonMapping(t *testing.T) {
	buttons := []string{
		"A", "B", "X", "Y", "LB", "RB", "BACK", "START", "GUIDE", "LS", "RS",
		"DPAD_UP", "DPAD_DOWN", "DPAD_LEFT", "DPAD_RIGHT",
	}
	// Bit 0x0800 is reserved in XInput; every defined bit is set here.
	want := uint16(0xF7FF)
	if got, _, _ := xusbReport(GamepadState{Buttons: buttons}); got != want {
		t.Fatalf("button mask = %#04x, want %#04x", got, want)
	}
	for _, button := range buttons {
		mask, _, _ := xusbReport(GamepadState{Buttons: []string{button}})
		if mask == 0 {
			t.Fatalf("%s produced an empty button mask", button)
		}
	}
}

func TestXUSBDigitalTriggersSaturateAnalogAxes(t *testing.T) {
	_, left, right := xusbReport(GamepadState{Buttons: []string{"LT", "RT"}})
	if left != 255 || right != 255 {
		t.Fatalf("digital triggers = (%d, %d), want (255, 255)", left, right)
	}
	_, left, right = xusbReport(GamepadState{LeftTrigger: 0.5, RightTrigger: 0})
	if left != 128 || right != 0 {
		t.Fatalf("analog triggers = (%d, %d), want (128, 0)", left, right)
	}
}

func TestViGEmAxisConversion(t *testing.T) {
	cases := []struct {
		in   float64
		want int16
	}{
		{0, 0},
		{1, 32767},
		{-1, -32767},
		{0.5, 16384},
		{-0.5, -16384},
		{2, 32767},
		{-2, -32767},
	}
	for _, test := range cases {
		if got := vigemAxis(test.in); got != test.want {
			t.Fatalf("vigemAxis(%v) = %d, want %d", test.in, got, test.want)
		}
	}
}

func TestViGEmTriggerConversion(t *testing.T) {
	cases := []struct {
		in   float64
		want byte
	}{
		{0, 0},
		{1, 255},
		{0.5, 128},
		{-1, 0},
		{2, 255},
	}
	for _, test := range cases {
		if got := vigemTrigger(test.in); got != test.want {
			t.Fatalf("vigemTrigger(%v) = %d, want %d", test.in, got, test.want)
		}
	}
}

func TestViGEmSubmitReportLayout(t *testing.T) {
	state := GamepadState{
		LeftX: 1, LeftY: -1, RightX: 0.5, RightY: 0,
		LeftTrigger: 1, RightTrigger: 0, Buttons: []string{"A", "DPAD_UP"},
	}
	request := vigemSubmitReport(3, state)
	if got := binary.LittleEndian.Uint32(request[0:4]); got != vigemSubmitSize {
		t.Fatalf("Size = %d, want %d", got, vigemSubmitSize)
	}
	if got := binary.LittleEndian.Uint32(request[4:8]); got != 3 {
		t.Fatalf("SerialNo = %d, want 3", got)
	}
	if got := binary.LittleEndian.Uint16(request[8:10]); got != xusbA|xusbDpadUp {
		t.Fatalf("buttons = %#04x, want %#04x", got, xusbA|xusbDpadUp)
	}
	if request[10] != 255 || request[11] != 0 {
		t.Fatalf("triggers = (%d, %d), want (255, 0)", request[10], request[11])
	}
	if got := int16(binary.LittleEndian.Uint16(request[12:14])); got != 32767 {
		t.Fatalf("LeftX = %d, want 32767", got)
	}
	if got := int16(binary.LittleEndian.Uint16(request[14:16])); got != -32767 {
		t.Fatalf("LeftY = %d, want -32767", got)
	}
	if got := int16(binary.LittleEndian.Uint16(request[16:18])); got != 16384 {
		t.Fatalf("RightX = %d, want 16384", got)
	}
	if got := int16(binary.LittleEndian.Uint16(request[18:20])); got != 0 {
		t.Fatalf("RightY = %d, want 0", got)
	}
}

func TestViGEmGUIDMatchesDriverInterface(t *testing.T) {
	if vigemGUIDData1 != 0x96E42B22 || vigemGUIDData2 != 0xF5E9 || vigemGUIDData3 != 0x42F8 {
		t.Fatal("ViGEmBus device interface GUID changed")
	}
	want := [8]byte{0xB0, 0x43, 0xED, 0x0F, 0x93, 0x2F, 0x01, 0x4F}
	if vigemGUIDData4 != want {
		t.Fatalf("GUID tail = %v, want %v", vigemGUIDData4, want)
	}
}
