//go:build windows && !amd64 && !arm64

package input

// ViGEmBus targets Windows 10/11 x64 and ARM64. Other Windows architectures
// keep the mouse and keyboard backend and report no gamepad provider.
func newViGEmProvider() (gamepadProvider, error) { return nil, ErrUnsupported }
