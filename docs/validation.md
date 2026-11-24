# Validation record — 0.7.0

Date: 2026-11-24. Android target: `versionCode` 13, `versionName` 0.7.0. Desktop server version: 0.7.0.

This record separates automated source checks from physical-device tests. Passing automated checks does not prove stable behaviour on every handset, desktop, game or network.

## 0.7.0 automated checks

- `gofmt -l .` — no unformatted Go source files.
- `go test -race ./...` — passed for all Go packages.
- `go vet ./...` — passed.
- JavaScript syntax checks — `app.js`, `theme.js` and `gamepad-core.js` parse successfully with Node.js.
- `node --test internal/web/web/gamepad-core.test.cjs` — 18 tests passed.
- `bash -n Install-TouchBridge-Linux.sh scripts/*.sh` — passed.
- Web DOM binding audit — 109 direct references, no missing element IDs.
- Shared Android assets — all eight controller assets are byte-for-byte identical to the desktop web UI (`index.html`, `app.css`, `app.js`, `theme.js`, `gamepad-core.js`, `manifest.webmanifest`, `icon.svg`, `sw.js`).
- Linux amd64, Linux arm64 and Windows amd64 desktop cross-compilation — passed.

The one-second protocol fuzz smoke test passed. The WebSocket parser fuzz smoke test did not finish within the workspace time limit and is not marked as passed. `govulncheck` is expected to run in GitHub Actions. The Android Gradle build was not run in this preparation workspace; build the debug and stable APKs on the Android release machine or rely on the PR CI debug-build job, then verify the stable APK locally.

## Multiplayer regression coverage

The Go tests cover:

- Distinct player slots and correct per-slot gamepad state.
- Four simultaneous authenticated controllers and rejection of the fifth.
- Reuse of a player slot after disconnect.
- Resetting only the disconnecting player's gamepad.
- Shared keyboard and mouse-button ownership: a shared input remains held until its final owning session releases it.
- Independent mock gamepad slots and reset behaviour.

These tests do not create actual Linux uinput devices or Windows ViGEmBus controllers.

## Screen-awake behaviour

The UI test environment verifies DOM references and asset parity, but the `FLAG_KEEP_SCREEN_ON` bridge requires an Android device for end-to-end validation. Confirm the phone stays awake while the controller screen is open by default, turns the flag off when the setting is disabled or the user leaves the controller, and clears the flag when the activity is destroyed.

## Release artifact status

The uploaded project contained a prebuilt debug APK from v0.6.0 (`versionCode` 12). It is excluded from the 0.7.0 PR-ready source archive and must not be renamed or uploaded as a 0.7.0 build. Create the stable `TouchBridge-0.7.0.apk` using the same private release keystore and alias that signed v0.6.0. The release workflow automatically attaches `TouchBridge-0.7.0-debug.apk` and the desktop packages; upload the signed stable APK manually.

## Physical-device validation required

- Four simultaneous Android phones, P1–P4 assignment, fifth-client rejection and freed-slot reuse.
- Four independent gamepads on Linux and Windows, including slot-specific reset on disconnect.
- Two players holding the same keyboard key or mouse button, then disconnecting one player without releasing the other player's hold.
- Player badge, status text, reconnect and idle-timeout behaviour.
- Keep-screen-on default, setting persistence, screen navigation, background/foreground and activity destruction.
- Long-running controller sessions, Wi-Fi interruptions, sleep/resume and actual target games.

Complete this checklist before publishing v0.7.0 as stable. If a core multiplayer path has not been exercised on real machines, publish a pre-release first instead of calling it stable.
