# Validation record — 0.6.0

Date: 2026-11-24. Android target: `versionCode` 12, `versionName` 0.6.0. The uploaded archive contained the 0.6.0 feature source but its prebuilt APK still identified itself as 0.5.3-alpha (`versionCode` 11); that APK is not a 0.6.0 release artifact and must not be uploaded under a 0.6.0 filename.

This record separates automated source checks from physical-device tests. Passing automated checks does not prove stable behaviour on every handset, desktop, game or network.

## 0.6.0 automated checks

- `gofmt -l .` — no unformatted Go source files.
- `go test -race ./...` — passed for all Go packages.
- `go vet ./...` — passed.
- JavaScript syntax checks — `app.js`, `theme.js` and `gamepad-core.js` parse successfully with Node.js.
- `node --test internal/web/web/gamepad-core.test.cjs` — 18 tests passed, including pedal pressure, steering-wheel mapping, gyro deadzone/clamping, wheel/pedal layout validation and large-control placement.
- `bash -n Install-TouchBridge-Linux.sh scripts/*.sh` — passed.
- Web DOM binding audit — 107 direct DOM references, no missing element IDs.
- Shared Android assets — all eight controller assets are byte-for-byte identical to the desktop web UI (`index.html`, `app.css`, `app.js`, `theme.js`, `gamepad-core.js`, `manifest.webmanifest`, `icon.svg`, `sw.js`).
- Protocol fuzz smoke test — `FuzzDecodeNeverPanics` passed with a one-second run.

The web WebSocket fuzz smoke test exceeded the command time limit in this environment and is not recorded as passed. `govulncheck` and the GitHub Actions workflows have not been executed from this preparation workspace; CI is expected to run them on the pull request.

## Feature-level tests present

The controller-math test suite includes checks for:

- Digital/analog trigger handling and existing stick math.
- Accelerator/brake pressure increasing toward the bottom of the pedal.
- Large controls avoiding existing controls where practical.
- Steering-wheel rotation mapping to the left-stick horizontal axis.
- Gyro steering deadzone, direction and full-lock clamp.
- Layout validation for a steering wheel and pedal controls.

These are deterministic logic tests. They do not simulate actual Android touch sampling, phone sensor behaviour, Bluetooth/network scheduling or the OS input stack.

## Not verified in this workspace

- A rebuilt 0.6.0 Android APK. The APK in the uploaded archive is still version 0.5.3-alpha / versionCode 11; rebuild after the version bump.
- A release-signed APK. Generate and back up the private signing keystore, then run `scripts/build-android-release.sh` on the release machine. The release workflow can attach a separate debug APK for testers.
- Physical Android phone testing for wheel drag, gyro permission/calibration, sensor orientation, pedal touch tracking, haptic feel, display cutouts, long sessions and real Wi-Fi.
- Physical Windows and Linux desktop testing for trigger axes/buttons, wheel/pedal input mapping, gamepad enumeration, reconnect and focus handling.
- Testing with real racing games and their particular input mappings, Steam Input and anti-cheat software.
- Windows ViGEmBus and Linux `/dev/uinput` behaviour on a range of machines, including sleep/resume and coexistence with other virtual controllers.
- End-to-end latency, packet loss, multi-hour sessions and unstable networks.
- `govulncheck`, CodeQL and GitHub Actions on the proposed PR.

Complete the physical Android and desktop checklist before publishing the stable release. If a critical input path has not been tested on hardware, publish a pre-release first rather than marking it stable.
