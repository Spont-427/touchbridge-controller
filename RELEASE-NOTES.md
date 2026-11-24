# TouchBridge 0.6.0

TouchBridge 0.6.0 adds racing-style controls, gyro steering, clearer connection diagnostics and a redesigned settings area.

## Highlights

### Triggers and pedals

- Digital LT/RT now send both their button state and full trigger axis pressure so Linux evdev and virtual-gamepad backends receive consistent input.
- Add analog accelerator and brake pedals mapped to RT and LT. Pressure increases toward the bottom of each pedal.

### Steering wheel and gyro

- Add a draggable steering-wheel control that maps rotation to the left stick's horizontal axis and returns to neutral on release.
- Add optional gyro steering for layouts with a steering wheel. The feature supports neutral calibration, low-pass filtering, a deadzone and direction inversion.

### Latency and appearance

- Display rolling median network RTT in the controller. The indicator details include p95 and jitter.
- Add System, Dark, AMOLED black and Light themes. Dark uses a neutral `#0f141b` background rather than pure black.
- Split settings into Appearance, Controls & haptics, Analog sticks and About. About reports app/server versions, connection state, gamepad backend status and the project link.

### Layout editor

- Improve size-aware placement for large controls such as steering wheels and pedals, reducing overlap and keeping them within the visible canvas where possible.

## Downloads

The v0.6.0 release should include:

- **Android stable:** `TouchBridge-0.6.0.apk` — signed with the project's private release key. This is the recommended Android download.
- **Android debug:** `TouchBridge-0.6.0-debug.apk` — debug-signed and intended for testing/troubleshooting only; the workflow attaches it automatically.
- **Windows x64:** attached by the `Release downloads` GitHub Actions workflow.
- **Linux x86_64 / ARM64:** attached by the `Release downloads` GitHub Actions workflow.

Do not publish the currently uploaded `0.5.3-alpha` APK under a `0.6.0` filename. Rebuild after the version bump and verify the embedded version name is `0.6.0` and `versionCode` is 12.

## Upgrade note

The first stable APK uses a private release signing key, which differs from the Android debug key used by previous alpha APKs. Android will not install it over an alpha debug build. Export any custom layouts, then uninstall the previous debug-signed app and install the stable APK. Uninstalling also clears app-local preferences, remembered computers and trusted pairing credentials, so users must configure preferences and pair their PC again. Later stable updates can be installed in place as long as the same release key is retained.

## Validation boundary

The source-level automated checks are recorded in [docs/validation.md](docs/validation.md). They do not replace hands-on validation of the new pedals, steering wheel and gyro on a physical Android phone and the target desktop input backends. Complete the physical-device checklist before marking the GitHub release as stable.
