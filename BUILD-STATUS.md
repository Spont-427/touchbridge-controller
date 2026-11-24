# Build status — 0.6.0

Date: 2026-11-24. Android `versionCode` 12, `versionName` 0.6.0. Desktop server version: 0.6.0.

## Included changes

- Digital LT/RT input sends both the button and full trigger-axis pressure.
- Latency indicator displays rolling median RTT and exposes p95/jitter details.
- Accelerator and brake pedals use analog RT/LT pressure, increasing toward the bottom of the control.
- Steering wheel maps its drag rotation to `left_x` and returns to neutral on release.
- Optional gyro steering supports neutral calibration, filtering, deadzone and inversion and only runs for layouts containing a wheel.
- Theme choices are System, Dark (`#0f141b`), AMOLED black and Light.
- Settings are separated into Appearance, Controls & haptics, Analog sticks and About.
- About displays app/server versions, connection state, gamepad backend status and the project link.
- Size-aware placement improves positioning of large controls and reduces overlap/off-canvas placement.
- Desktop and Android controller assets are synchronized.

## Source checks run in this workspace

- `gofmt -l .` — passed.
- `go test -race ./...` — passed.
- `go vet ./...` — passed.
- Node syntax checks — passed.
- Controller math tests — 18/18 passed.
- Shell syntax — passed.
- DOM binding audit — 107 direct references, no missing IDs.
- Android/web asset parity — 8/8 identical.
- Protocol fuzz smoke test — passed for one second. Web fuzz smoke test timed out and is not marked as passed.

## Toolchain alignment

- Android Gradle Plugin 9.5.2.
- Gradle wrapper 9.6.0.
- JDK 17.
- Android SDK Platform 35.
- Android SDK Build Tools 36.0.0.
- Kotlin uses AGP's built-in Kotlin support; the old `org.jetbrains.kotlin.android` plugin is removed.

## Release artifact status

The prebuilt APK inside the uploaded project is still version `0.5.3-alpha`, Android `versionCode` 11 and debug-signed. It is deliberately excluded from the PR-ready source archive and is not a valid 0.6.0 upload. The release workflow builds and attaches a fresh 0.6.0 debug APK for testers; build the stable signed APK with `scripts/build-android-release.sh` and your private release keystore, then upload it to the same release.

## Remaining validation before stable publication

- Run the Android Gradle build on the release machine using the updated toolchain.
- Install and test the 0.6.0 APK on a physical Android device. Specifically validate gyro permission, neutral calibration, portrait/landscape sensor orientation, steering wheel drag/release, pedals and the new settings flows.
- Test trigger axis/button outputs, pedals and steering on physical Linux and Windows hosts, and in at least one target game or controller tester.
- Run the GitHub Actions checks and vulnerability scan on the PR.
- Build and verify the release-signed APK. Keep at least two encrypted backups of the release keystore.
