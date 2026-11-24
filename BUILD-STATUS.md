# Build status — 0.7.0

Date: 2026-11-24. Android `versionCode` 13, `versionName` 0.7.0. Desktop server version: 0.7.0.

## Included changes

- Support up to four authenticated phone controllers, each assigned a player slot and dedicated virtual gamepad.
- Reuse player slots after disconnect and reject a fifth simultaneous controller with `controller_busy` and the configured limit.
- Scope gamepad resets to the disconnecting player's slot.
- Track ownership of shared keyboard keys and mouse buttons so one player's disconnect cannot release an input another player is still holding.
- Add P1–P4 player badges and player-aware connection status in the controller UI.
- Keep the Android display awake by default while the controller is open, with a persistent setting under Controls & haptics.
- Set and clear `FLAG_KEEP_SCREEN_ON` through the Android bridge and clear the flag when the activity is destroyed.
- Update protocol, stability and controller documentation for multiplayer and screen-awake behaviour.
- Bump Android and desktop version to `0.7.0` (`versionCode` 13) and service-worker shell cache to v10.

## Automated checks run in this workspace

- `gofmt -l .` — passed.
- `go test -race ./...` — passed across all Go packages, including multiplayer slot, disconnect and shared keyboard/mouse ownership tests.
- `go vet ./...` — passed.
- Node syntax checks — passed.
- `node --test internal/web/web/gamepad-core.test.cjs` — 18/18 passed.
- `bash -n Install-TouchBridge-Linux.sh scripts/*.sh` — passed.
- Web DOM binding audit — 109 direct references, no missing IDs.
- Android/desktop asset parity — all eight shared assets identical.
- Cross-compilation — Linux amd64, Linux arm64 and Windows amd64 builds passed.

The protocol fuzz smoke test passed for one second. The WebSocket parser fuzz test did not finish within the workspace time limit, so it is not marked as passed. `govulncheck` remains a CI check. This preparation environment did not build the Android APK; the Android Gradle build should run on the release machine or in GitHub Actions.

## Release artifact status

The uploaded project contained a previously built `0.6.0` debug APK. It is not a 0.7.0 artifact and is excluded from the PR-ready source ZIP. Build the new signed stable APK on the release machine with `scripts/build-android-release.sh`, using the exact same private release keystore that signed v0.6.0. Do not create a new signing key. GitHub Actions attaches the new debug APK and Windows/Linux packages when the v0.7.0 release is published; upload the stable signed APK manually.

## Physical-device validation before stable publication

- Connect four Android phones at once and verify each receives P1–P4; confirm a fifth is rejected and a freed slot is reused.
- Test separate gamepad state on four Linux `/dev/uinput` devices and four Windows ViGEmBus providers.
- Have two phones hold the same keyboard key and mouse button; disconnect one and confirm the input remains held until the other phone releases it.
- Check the player badge/status, reconnect and idle-timeout cleanup.
- Verify Keep screen on defaults to enabled only on the controller screen, can be disabled in Settings, and clears after leaving the controller or closing the activity.
- Test sleep/resume, screen rotation, Android background/foreground transitions and long controller sessions.

Complete this checklist before marking v0.7.0 stable. Automated tests do not replace testing with real phones and desktop backends.
