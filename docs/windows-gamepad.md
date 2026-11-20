# Windows virtual gamepad

TouchBridge injects mouse and keyboard input through the supported `SendInput`
API. Windows has no public API for creating a virtual gamepad, so controller
output requires a virtual-device driver. TouchBridge uses **ViGEmBus** and
presents an Xbox 360-compatible controller, which games read through XInput.

The integration is optional. Without the driver, TouchBridge still serves the
controller UI and injects mouse and keyboard input; the `capabilities` message
reports `"gamepad": false` and the client stays in mouse/keyboard mode. The
server prints the detected gamepad status in its startup banner.

## Requirements

- Windows 10 or Windows 11, x64 or ARM64.
- [ViGEmBus 1.22.0](https://github.com/nefarius/ViGEmBus/releases/tag/v1.22.0)
  installed once. Installation requires administrator rights and may require a
  restart on first install. The release ZIP's
  `Setup-TouchBridge-Windows.bat` automates this; TouchBridge does not bundle
  the driver.
- The TouchBridge server itself runs **unelevated**. Creating a ViGEmBus target
  does not require administrator rights.

## Automated setup

`Setup-TouchBridge-Windows.bat` is an optional, rerunnable one-time setup:

1. It checks whether ViGEmBus and the TouchBridge firewall rules are already
   present. If so, it exits without changing anything.
2. Otherwise it requests administrator permission once (UAC).
3. It downloads `ViGEmBus_1.22.0_x64_x86_arm64.exe` from the official Nefarius
   GitHub release, rejects the file unless it is exactly 6,278,576 bytes and
   its SHA-256 is
   `89220a7865076b342892f98865f3499fb7c4cfd673159e89d352c360fd014c6a`,
   then runs the installer silently. A verified copy is kept under `%TEMP%` if
   the install fails.
4. It adds inbound firewall rules for TCP 8765 and UDP 8766, limited to the
   **Private** network profile.
5. It reports whether a restart is needed and asks you to run
   `Start-TouchBridge.bat`.

Use `-NoDriver` or `-NoFirewall` to skip either part. The server itself is
still started unprivileged by `Start-TouchBridge.bat`.

## How it behaves

- On startup the server only connects to the ViGEmBus device interface. No
  controller exists yet.
- The virtual Xbox 360 controller is plugged in the first time the paired
  controller sends a `gamepad_state` message, so starting the server does not
  create a phantom pad.
- Complete state snapshots are submitted as `XUSB_SUBMIT_REPORT`; that maps
  directly onto XInput's button, trigger and stick fields.
- Digital LT/RT mode is honoured: when the client sends `LT`/`RT` in the button
  set, the corresponding analog trigger is saturated to its maximum.
- A client disconnect resets all axes, triggers and buttons to neutral but
  keeps the device plugged in, so reconnecting does not churn the device tree.
- Server shutdown unplugs the virtual controller. A hard kill cannot run that
  cleanup; Windows then leaves a stale controller until it is removed from
  **Control Panel → Devices and Printers**.

## Driver status and caveats

ViGEmBus was archived on 2 November 2023 after a trademark settlement with
ViGEM GmbH. The retirement was not technical: the last release, 1.22.0, is
Microsoft-signed, works on Windows 10/11, and is still used by DS4Windows,
x360ce and similar tools.

ViGEmBus 1.22.0 removed the auto-updater entirely. Installs at or below
1.21.442 ship an updater that points at a domain that no longer belongs to the
author; if you must keep one of those, follow the
[end-of-life guidance](https://docs.nefarius.at/projects/ViGEm/End-of-Life/) to
repoint or remove it.

## Uninstall and rollback

1. Stop TouchBridge; it unplugs its virtual controller on exit.
2. Remove ViGEmBus from **Settings → Apps → Installed Apps** if no other
   application uses it.
3. Restart Windows if the uninstaller requests it.

Removing the driver only affects gamepad output. Mouse and keyboard injection
continue to work.

## Not yet verified

This integration and its setup script have been authored and cross-checked
from Linux, but neither has been exercised on a physical Windows machine from
this environment. Before relying on it, verify on real hardware:

- `Setup-TouchBridge-Windows.bat`: UAC flow, checksum verification, Advanced
  Installer silent flags, firewall rule creation, cancellation and rerun.
- XInput slot assignment and detection in **Set up USB game controllers**.
- Rumble behaviour (TouchBridge does not consume force feedback, so games send
  rumble requests that are ignored).
- Sleep/resume, reconnect and server restart with a plugged device.
- Multiple controllers and coexistence with DS4Windows or Steam Input.
- Anti-cheat behaviour in online games.

## Planned second backend

The Windows backend is split behind an internal `gamepadProvider` interface. A
future release can add a HIDMaestro provider, which is actively maintained and
supports more device profiles, without changing the mouse/keyboard path or the
network protocol. HIDMaestro creates devices through an elevated helper or
broker service, so it needs a separate install/elevation design.
