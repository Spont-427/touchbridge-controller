package input

import (
	"encoding/binary"
	"math"
)

// ViGEmBus user-mode protocol. The bus is a kernel driver, but every request is
// a plain DeviceIoControl with a small little-endian struct, so the encoding is
// kept free of platform build tags and covered by tests on every platform.
//
// These values mirror include/ViGEm/km/BusShared.h in ViGEmClient 1.22.0:
//
//	CTL_CODE(FILE_DEVICE_BUS_EXTENDER=0x2A, function, METHOD_BUFFERED, access)
//	PLUGIN(0x801, write), UNPLUG(0x802, write), CHECK_VERSION(0x803, write),
//	WAIT_READY(0x804, write), XUSB_SUBMIT_REPORT(0xA02, write).
const (
	vigemIoctlPluginTarget = 0x2AA004
	vigemIoctlUnplugTarget = 0x2AA008
	vigemIoctlCheckVersion = 0x2AA00C
	vigemIoctlWaitReady    = 0x2AA010
	vigemIoctlXUSBSubmit   = 0x2AA808

	vigemTargetXbox360Wired = 0
	// VIGEM_COMMON_VERSION is the ABI version the driver compares.
	vigemCommonVersion = 1

	vigemCheckVersionSize = 8
	vigemPluginSize       = 16
	vigemUnplugSize       = 8
	vigemSubmitSize       = 20

	// XInput button bits from XUSB_BUTTON / XINPUT_GAMEPAD.
	xusbDpadUp        = 0x0001
	xusbDpadDown      = 0x0002
	xusbDpadLeft      = 0x0004
	xusbDpadRight     = 0x0008
	xusbStart         = 0x0010
	xusbBack          = 0x0020
	xusbLeftThumb     = 0x0040
	xusbRightThumb    = 0x0080
	xusbLeftShoulder  = 0x0100
	xusbRightShoulder = 0x0200
	xusbGuide         = 0x0400
	xusbA             = 0x1000
	xusbB             = 0x2000
	xusbX             = 0x4000
	xusbY             = 0x8000
)

// The ViGEmBus device interface GUID {96E42B22-F5E9-42F8-B043-ED0F932F014F}.
const (
	vigemGUIDData1 = 0x96E42B22
	vigemGUIDData2 = 0xF5E9
	vigemGUIDData3 = 0x42F8
)

var vigemGUIDData4 = [8]byte{0xB0, 0x43, 0xED, 0x0F, 0x93, 0x2F, 0x01, 0x4F}

func vigemCheckVersionRequest() []byte {
	request := make([]byte, vigemCheckVersionSize)
	binary.LittleEndian.PutUint32(request[0:4], vigemCheckVersionSize)
	binary.LittleEndian.PutUint32(request[4:8], vigemCommonVersion)
	return request
}

func vigemPluginRequest(serial uint32, targetType uint32, vendorID, productID uint16) []byte {
	request := make([]byte, vigemPluginSize)
	binary.LittleEndian.PutUint32(request[0:4], vigemPluginSize)
	binary.LittleEndian.PutUint32(request[4:8], serial)
	binary.LittleEndian.PutUint32(request[8:12], targetType)
	binary.LittleEndian.PutUint16(request[12:14], vendorID)
	binary.LittleEndian.PutUint16(request[14:16], productID)
	return request
}

func vigemUnplugRequest(serial uint32) []byte {
	request := make([]byte, vigemUnplugSize)
	binary.LittleEndian.PutUint32(request[0:4], vigemUnplugSize)
	binary.LittleEndian.PutUint32(request[4:8], serial)
	return request
}

func vigemWaitReadyRequest(serial uint32) []byte {
	request := make([]byte, vigemUnplugSize)
	binary.LittleEndian.PutUint32(request[0:4], vigemUnplugSize)
	binary.LittleEndian.PutUint32(request[4:8], serial)
	return request
}

// vigemSubmitReport encodes an XUSB_SUBMIT_REPORT: Size, SerialNo, XUSB_REPORT.
// XUSB_REPORT is 12 bytes: buttons (u16), two u8 triggers, four i16 sticks.
func vigemSubmitReport(serial uint32, state GamepadState) []byte {
	buttons, leftTrigger, rightTrigger := xusbReport(state)
	request := make([]byte, vigemSubmitSize)
	binary.LittleEndian.PutUint32(request[0:4], vigemSubmitSize)
	binary.LittleEndian.PutUint32(request[4:8], serial)
	binary.LittleEndian.PutUint16(request[8:10], buttons)
	request[10] = leftTrigger
	request[11] = rightTrigger
	binary.LittleEndian.PutUint16(request[12:14], uint16(vigemAxis(state.LeftX)))
	binary.LittleEndian.PutUint16(request[14:16], uint16(vigemAxis(state.LeftY)))
	binary.LittleEndian.PutUint16(request[16:18], uint16(vigemAxis(state.RightX)))
	binary.LittleEndian.PutUint16(request[18:20], uint16(vigemAxis(state.RightY)))
	return request
}

// xusbReport maps a normalized snapshot onto an XUSB_REPORT. LT/RT normally
// arrive as analog pressure; when the client's digital trigger mode is active
// they arrive in Buttons instead, and XInput has no separate digital trigger
// bit, so they saturate the analog trigger.
func xusbReport(state GamepadState) (buttons uint16, leftTrigger, rightTrigger byte) {
	leftTrigger = vigemTrigger(state.LeftTrigger)
	rightTrigger = vigemTrigger(state.RightTrigger)
	for _, button := range state.Buttons {
		switch button {
		case "LT":
			leftTrigger = 255
		case "RT":
			rightTrigger = 255
		case "A":
			buttons |= xusbA
		case "B":
			buttons |= xusbB
		case "X":
			buttons |= xusbX
		case "Y":
			buttons |= xusbY
		case "LB":
			buttons |= xusbLeftShoulder
		case "RB":
			buttons |= xusbRightShoulder
		case "BACK":
			buttons |= xusbBack
		case "START":
			buttons |= xusbStart
		case "GUIDE":
			buttons |= xusbGuide
		case "LS":
			buttons |= xusbLeftThumb
		case "RS":
			buttons |= xusbRightThumb
		case "DPAD_UP":
			buttons |= xusbDpadUp
		case "DPAD_DOWN":
			buttons |= xusbDpadDown
		case "DPAD_LEFT":
			buttons |= xusbDpadLeft
		case "DPAD_RIGHT":
			buttons |= xusbDpadRight
		}
	}
	return buttons, leftTrigger, rightTrigger
}

// vigemAxis maps a normalized [-1, 1] axis to XInput's i16 range. The symmetric
// bound avoids the -32768 asymmetry that would make one direction slightly
// stronger and keeps zero exactly centred.
func vigemAxis(value float64) int16 {
	if math.IsNaN(value) {
		return 0
	}
	scaled := int(math.Round(clampFloat(value, -1, 1) * 32767))
	if scaled < -32767 {
		scaled = -32767
	}
	if scaled > 32767 {
		scaled = 32767
	}
	return int16(scaled)
}

func vigemTrigger(value float64) byte {
	if math.IsNaN(value) {
		return 0
	}
	scaled := int(math.Round(clampFloat(value, 0, 1) * 255))
	if scaled < 0 {
		return 0
	}
	if scaled > 255 {
		return 255
	}
	return byte(scaled)
}

func clampFloat(value, low, high float64) float64 {
	if value < low {
		return low
	}
	if value > high {
		return high
	}
	return value
}
