# Windows virtual gamepad status

TouchBridge 0.3.1-alpha provides Windows mouse and keyboard injection through the supported `SendInput` API. Windows virtual gamepad creation remains disabled in this release.

## Why it is not bundled yet

A virtual Xbox/XInput device on Windows is not created by the public `SendInput` API. It requires a separate virtual HID/XUSB device provider and its driver/install lifecycle. We reviewed current choices on 19 November 2026:

- **ViGEmBus**: the [upstream Nefarius repository](https://github.com/nefarius/ViGEmBus) explicitly marks the driver retired and archived (2 November 2023), and its [end-of-life notice](https://docs.nefarius.at/projects/ViGEm/End-of-Life/) says it will receive no further updates. It is not a sensible default for a new project.
- **HIDMaestro**: the [upstream project](https://github.com/hifihedgehog/HIDMaestro) is MIT-licensed and advertises a user-mode virtual-controller approach. It is promising, but it is still a relatively small ecosystem, requires a local driver/certificate installation and an elevated setup step, and has not been independently audited or tested by TouchBridge on physical Windows machines.
- **libvirtualhid**: the [cross-platform library](https://github.com/LizardByte/libvirtualhid) is MIT-licensed, but its [Windows UMDF driver/broker/package](https://github.com/LizardByte/libvirtualhid/blob/master/LICENSES/license-map.md) use a separate source-available license and Windows gamepad creation requires a license. The published driver installer currently targets AMD64, not Windows ARM64.

TouchBridge therefore fails closed: the Gamepad tab reports the missing Windows provider and continues to provide mouse/keyboard control. It will not silently install a retired driver, trust a third-party driver installer, or elevate the complete network-facing server.

## Planned integration requirements

A Windows provider can be added when it meets all of these conditions:

1. Clear redistribution and licensing terms compatible with the intended open-source distribution.
2. Signed and reproducible driver packages with a documented uninstall/rollback path.
3. No administrator elevation for the network server itself; elevation, if needed, is isolated to an explicit installer step.
4. Supported Windows versions and architectures documented by the provider.
5. Physical-device testing for XInput, SDL/GameInput, connection loss, multiple controllers, sleep/resume and uninstall.

Until then, do not assume the Windows Gamepad tab can inject a virtual controller.
