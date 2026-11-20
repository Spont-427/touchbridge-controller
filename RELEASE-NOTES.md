# TouchBridge 0.5.3-alpha

This alpha adds optional Windows virtual gamepad output and fixes a narrow-screen Android pairing layout issue.

## Highlights

### Windows virtual gamepad

- Add an optional Xbox 360-compatible virtual controller through ViGEmBus 1.22.0. Games that use XInput can detect it when the driver is installed.
- Create the virtual controller on first gamepad input, reset it on disconnect and unplug it when the server shuts down.
- Keep the desktop server unprivileged. Only the optional driver and firewall setup requires administrator approval.
- Add `Setup-TouchBridge-Windows.bat` to download the pinned ViGEmBus installer, verify its SHA-256 checksum, install it and add private-network firewall rules for TCP 8765 and UDP 8766.
- Without ViGEmBus, Windows mouse and keyboard input continue to work and the server reports that gamepad output is unavailable.

### Android pairing fix

- Fix the PIN pairing row on narrow screens. The Pair device and Disconnect buttons now occupy separate grid columns, preventing the disabled Disconnect button from overlapping or intercepting taps on Pair device.

## Downloads

- **Android:** `TouchBridge-0.5.3-alpha-debug.apk` — compiled on the developer's established release machine, debug build for testing (not a Play Store release). Its signing certificate matches the previous alpha APK. The APK is attached separately to the GitHub release; generated APKs are not committed to the source repository.
- **Windows x64:** desktop package is attached by the GitHub `Release downloads` workflow.
- **Linux x86_64 / ARM64:** desktop packages are attached by the GitHub `Release downloads` workflow.

## Platform support

- **Android:** Android 6.0 or newer; `versionCode` 11, `versionName` 0.5.3-alpha.
- **Windows x64:** mouse and keyboard out of the box; Xbox 360-compatible virtual gamepad when ViGEmBus 1.22.0 is installed.
- **Linux x86_64 / ARM64:** mouse, keyboard and generic virtual gamepad through `/dev/uinput`, subject to kernel support and permissions.

## Validation and limitations

The pairing CSS fix is present in both the desktop web UI and the Android-bundled UI; the asset copies are checked for byte-for-byte parity. Existing automated Go and controller UI tests remain part of CI. The new release APK must be built from this version on the developer's established Android build machine. Physical Android hardware, physical Windows/Linux input injection, real games, anti-cheat compatibility, long sessions and unstable networks still require testing.

Windows virtual gamepad output depends on the third-party ViGEmBus driver. ViGEmBus is not bundled with TouchBridge. The default desktop HTTP/WebSocket transport is unencrypted; use a trusted private network or configure TLS, and never expose the server directly to the public internet.
