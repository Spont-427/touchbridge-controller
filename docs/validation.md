# Validation record — 0.9.0

Date: 2026-11-26. Android target: `versionCode` 15, `versionName` 0.9.0. Desktop server version: 0.9.0.

## Automated source checks

`FUZZ_TIME=1x ./scripts/check.sh` completed successfully through all nine stages.

- `gofmt -l .` — clean.
- `go test -race ./...` — passed for all Go packages.
- `go vet ./...` — passed.
- `node --check` for `app.js`, `theme.js` and `gamepad-core.js` — passed.
- `node --test internal/web/web/gamepad-core.test.cjs` — 17/17 tests passed.
- `bash -n Install-TouchBridge-Linux.sh scripts/*.sh` — passed.
- Web DOM binding audit — 110 direct references, no missing element IDs.
- Protocol and WebSocket fuzz smoke tests — one iteration each, passed.
- Shared Android/web UI — all eight controller assets are byte-for-byte identical.
- `govulncheck` — not installed locally; the repository CI workflow runs the vulnerability scan.

## Android build and AVD review

The release author reports that the updated Android APK built successfully and was reviewed in an Android Virtual Device (AVD). This preparation environment does not have an Android SDK configured, so it could not independently reproduce that APK build or emulator review. The generated APK and `android-client/app/build/` outputs are intentionally excluded from the PR source ZIP.

The APK build and AVD review are author-reported validation, not a claim of physical-device testing. Before publishing stable, check the new launcher icon, all three themes, the layout library, pairing, settings, screen transitions and reduced-motion behavior on the supported Android versions.

## Design scope

The redesign applies to the portrait layout library, computer connection, settings, shared navigation, themes, app mark and Android launcher icon. The immersive landscape controller and layout editor visuals are intentionally left unchanged.

## Stable release artifacts

Build `TouchBridge-0.9.0.apk` after the PR is merged. Sign it with the exact same private release keystore and alias that signed v0.8.0, so existing installations can update in place. Do not create a new signing key or upload a debug APK as the stable release. The release workflow attaches the debug APK and Windows/Linux packages; upload the signed stable APK manually.
