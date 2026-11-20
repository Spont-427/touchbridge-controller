//go:build windows && (amd64 || arm64)

package input

import (
	"errors"
	"fmt"
	"sync"
	"syscall"
	"unsafe"
)

// Windows has no public API for creating a virtual gamepad. This provider uses
// ViGEmBus, which exposes a bus device through SetupAPI and accepts the small
// IOCTL protocol encoded in vigem_wire.go. The driver must be installed by the
// user; the TouchBridge server itself stays unelevated.
//
// The provider is deliberately isolated behind gamepadProvider so a second
// backend (for example HIDMaestro) can be added without touching the mouse and
// keyboard path.

var (
	vigemKernel32 = syscall.NewLazyDLL("kernel32.dll")
	vigemSetupAPI = syscall.NewLazyDLL("setupapi.dll")

	procCreateFileW                      = vigemKernel32.NewProc("CreateFileW")
	procDeviceIoControl                  = vigemKernel32.NewProc("DeviceIoControl")
	procCreateEventW                     = vigemKernel32.NewProc("CreateEventW")
	procCloseHandle                      = vigemKernel32.NewProc("CloseHandle")
	procGetOverlappedResult              = vigemKernel32.NewProc("GetOverlappedResult")
	procSetupDiGetClassDevsW             = vigemSetupAPI.NewProc("SetupDiGetClassDevsW")
	procSetupDiEnumDeviceInterfaces      = vigemSetupAPI.NewProc("SetupDiEnumDeviceInterfaces")
	procSetupDiGetDeviceInterfaceDetailW = vigemSetupAPI.NewProc("SetupDiGetDeviceInterfaceDetailW")
	procSetupDiDestroyDeviceInfoList     = vigemSetupAPI.NewProc("SetupDiDestroyDeviceInfoList")
)

const (
	digcfPresent         = 0x00000002
	digcfDeviceInterface = 0x00000010

	genericRead          = 0x80000000
	genericWrite         = 0x40000000
	fileShareRead        = 0x00000001
	fileShareWrite       = 0x00000002
	openExisting         = 3
	fileAttrNormal       = 0x00000080
	fileFlagOverlapped   = 0x40000000
	fileFlagNoBuffering  = 0x20000000
	fileFlagWriteThrough = 0x80000000

	errorInvalidParameter = 87
	errorDevNotExist      = 55
	errorIOPending        = 997
	errorNoMoreItems      = 259

	// The driver supports a bounded number of simultaneous targets. Serial
	// numbers are chosen by the client, so retry the first few slots only.
	vigemMaxSerial   = 32
	xbox360VendorID  = 0x045E
	xbox360ProductID = 0x028E
)

type guid struct {
	data1 uint32
	data2 uint16
	data3 uint16
	data4 [8]byte
}

var vigemBusGUID = guid{
	data1: vigemGUIDData1,
	data2: vigemGUIDData2,
	data3: vigemGUIDData3,
	data4: vigemGUIDData4,
}

type spDeviceInterfaceData struct {
	cbSize             uint32
	interfaceClassGUID guid
	flags              uint32
	reserved           uintptr
}

type spDeviceInterfaceDetailData struct {
	cbSize     uint32
	devicePath [1]uint16
}

type overlapped struct {
	internal     uintptr
	internalHigh uintptr
	offset       uint32
	offsetHigh   uint32
	event        uintptr
}

type vigemGamepad struct {
	mu      sync.Mutex
	handle  syscall.Handle
	event   syscall.Handle
	serial  uint32
	plugged bool
	closed  bool
	lastErr error
}

// newViGEmProvider connects to the ViGEmBus device interface. It does not plug
// a controller until the first state arrives, so starting the server does not
// create a phantom pad (or a device-connect sound) before anyone connects.
func newViGEmProvider() (gamepadProvider, error) {
	event, _, callErr := procCreateEventW.Call(0, 0, 0, 0)
	if event == 0 {
		return nil, fmt.Errorf("create ViGEmBus event: %w", callErr)
	}
	handle, err := openViGEmBus(syscall.Handle(event))
	if err != nil {
		_ = syscall.CloseHandle(syscall.Handle(event))
		return nil, err
	}
	return &vigemGamepad{handle: handle, event: syscall.Handle(event)}, nil
}

func openViGEmBus(event syscall.Handle) (syscall.Handle, error) {
	deviceInfo, _, callErr := procSetupDiGetClassDevsW.Call(
		uintptr(unsafe.Pointer(&vigemBusGUID)), 0, 0,
		digcfPresent|digcfDeviceInterface,
	)
	if deviceInfo == uintptr(syscall.InvalidHandle) || deviceInfo == 0 {
		if callErr == nil || errors.Is(callErr, syscall.Errno(0)) {
			callErr = errors.New("device interface unavailable")
		}
		return 0, fmt.Errorf("ViGEmBus driver not found: %w", callErr)
	}
	defer func() { _, _, _ = procSetupDiDestroyDeviceInfoList.Call(deviceInfo) }()

	var interfaceData spDeviceInterfaceData
	detail := make([]byte, 2048)
	for index := 0; ; index++ {
		interfaceData = spDeviceInterfaceData{cbSize: uint32(unsafe.Sizeof(interfaceData))}
		ok, _, err := procSetupDiEnumDeviceInterfaces.Call(
			deviceInfo, 0, uintptr(unsafe.Pointer(&vigemBusGUID)), uintptr(index),
			uintptr(unsafe.Pointer(&interfaceData)),
		)
		if ok == 0 {
			if errors.Is(err, syscall.Errno(errorNoMoreItems)) {
				break
			}
			return 0, fmt.Errorf("enumerate ViGEmBus interfaces: %w", err)
		}

		*(*uint32)(unsafe.Pointer(&detail[0])) = uint32(unsafe.Sizeof(spDeviceInterfaceDetailData{}))
		var required uint32
		ok, _, err = procSetupDiGetDeviceInterfaceDetailW.Call(
			deviceInfo, uintptr(unsafe.Pointer(&interfaceData)),
			uintptr(unsafe.Pointer(&detail[0])), uintptr(len(detail)),
			uintptr(unsafe.Pointer(&required)), 0,
		)
		if ok == 0 {
			continue
		}

		pathPtr := (*uint16)(unsafe.Pointer(uintptr(unsafe.Pointer(&detail[0])) + 4))
		path := syscall.UTF16ToString(unsafe.Slice(pathPtr, (len(detail)-4)/2))
		pathUTF16, pathErr := syscall.UTF16PtrFromString(path)
		if pathErr != nil {
			continue
		}

		handle, _, err := procCreateFileW.Call(
			uintptr(unsafe.Pointer(pathUTF16)),
			genericRead|genericWrite, fileShareRead|fileShareWrite, 0, openExisting,
			fileAttrNormal|fileFlagOverlapped|fileFlagNoBuffering|fileFlagWriteThrough, 0,
		)
		if handle == uintptr(syscall.InvalidHandle) {
			continue
		}

		connection := &vigemGamepad{handle: syscall.Handle(handle), event: event}
		if err := connection.ioctl(vigemIoctlCheckVersion, vigemCheckVersionRequest()); err != nil {
			_ = syscall.CloseHandle(syscall.Handle(handle))
			continue
		}
		return syscall.Handle(handle), nil
	}
	return 0, errors.New("ViGEmBus interface is present but could not be opened")
}

// ioctl sends one synchronous-looking request over an overlapped handle. The
// handle is opened with FILE_FLAG_OVERLAPPED, so GetOverlappedResult with
// bWait=1 provides the same blocking semantics as the reference client.
func (g *vigemGamepad) ioctl(code uint32, request []byte) error {
	if len(request) == 0 {
		return errors.New("empty ViGEmBus request")
	}
	var transferred uint32
	var async overlapped
	async.event = uintptr(g.event)

	ok, _, callErr := procDeviceIoControl.Call(
		uintptr(g.handle), uintptr(code),
		uintptr(unsafe.Pointer(&request[0])), uintptr(len(request)),
		0, 0,
		uintptr(unsafe.Pointer(&transferred)), uintptr(unsafe.Pointer(&async)),
	)
	if ok == 0 && !errors.Is(callErr, syscall.Errno(errorIOPending)) {
		if callErr == nil || errors.Is(callErr, syscall.Errno(0)) {
			return errors.New("ViGEmBus request was rejected")
		}
		return callErr
	}
	ok, _, callErr = procGetOverlappedResult.Call(
		uintptr(g.handle), uintptr(unsafe.Pointer(&async)),
		uintptr(unsafe.Pointer(&transferred)), 1,
	)
	if ok == 0 {
		if errors.Is(callErr, syscall.Errno(0)) {
			return errors.New("ViGEmBus request failed")
		}
		return callErr
	}
	return nil
}

func (g *vigemGamepad) GamepadAvailable() bool { return true }

func (g *vigemGamepad) GamepadStatus() string {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.plugged && g.lastErr == nil {
		return "Windows ViGEmBus virtual Xbox 360 controller ready"
	}
	if g.lastErr != nil {
		return "Windows ViGEmBus is installed, but the virtual controller failed: " + g.lastErr.Error()
	}
	return "Windows ViGEmBus virtual Xbox 360 controller ready (waiting for the first input)"
}

func (g *vigemGamepad) SetGamepadState(state GamepadState) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.setStateLocked(state)
}

func (g *vigemGamepad) setStateLocked(state GamepadState) error {
	if g.closed {
		return errors.New("ViGEmBus connection is closed")
	}
	if err := g.ensurePluggedLocked(); err != nil {
		g.lastErr = err
		return err
	}
	if err := g.ioctl(vigemIoctlXUSBSubmit, vigemSubmitReport(g.serial, state)); err != nil {
		if errors.Is(err, syscall.Errno(errorDevNotExist)) {
			g.plugged = false
			g.serial = 0
			g.lastErr = err
			return ErrUnsupported
		}
		g.lastErr = err
		return err
	}
	g.lastErr = nil
	return nil
}

func (g *vigemGamepad) ensurePluggedLocked() error {
	if g.plugged {
		return nil
	}
	var lastErr error
	for serial := uint32(1); serial <= vigemMaxSerial; serial++ {
		request := vigemPluginRequest(serial, vigemTargetXbox360Wired, xbox360VendorID, xbox360ProductID)
		if err := g.ioctl(vigemIoctlPluginTarget, request); err != nil {
			lastErr = err
			continue
		}
		g.serial = serial
		g.plugged = true
		// Pre-1.17 drivers do not implement this IOCTL; that is harmless.
		if err := g.ioctl(vigemIoctlWaitReady, vigemWaitReadyRequest(serial)); err != nil &&
			!errors.Is(err, syscall.Errno(errorInvalidParameter)) {
			return err
		}
		return nil
	}
	if lastErr == nil {
		lastErr = errors.New("no free ViGEmBus target slot")
	}
	return lastErr
}

// ResetGamepad neutralises a controller that is already plugged in. It must not
// create one, because ReleaseAll runs on every client disconnect and a pad
// should only appear once the client has actually sent gamepad input.
func (g *vigemGamepad) ResetGamepad() error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed || !g.plugged {
		return nil
	}
	return g.setStateLocked(GamepadState{})
}

func (g *vigemGamepad) Close() error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return nil
	}
	g.closed = true
	var first error
	if g.plugged {
		if err := g.ioctl(vigemIoctlUnplugTarget, vigemUnplugRequest(g.serial)); err != nil {
			first = err
		}
		g.plugged = false
	}
	if g.event != 0 {
		_ = syscall.CloseHandle(g.event)
		g.event = 0
	}
	if g.handle != 0 {
		if err := syscall.CloseHandle(g.handle); err != nil && first == nil {
			first = err
		}
		g.handle = 0
	}
	return first
}
