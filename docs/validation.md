# Validation record — 0.8.0

Date: 2026-11-25. Android target: `versionCode` 14, `versionName` 0.8.0. Desktop server version: 0.8.0.

This record separates automated source checks from Android build and physical-device checks. Passing unit tests and cross-compilation does not prove updater behavior or real-device touch interaction.

## Automated source checks

- `gofmt -l .` — clean.
- `go test -race ./...` — passed for all Go packages.
- `go vet ./...` — passed.
- Node syntax checks for `app.js`, `theme.js` and `gamepad-core.js` — passed.
- `node --test internal/web/web/gamepad-core.test.cjs` — 17 tests passed.
- `bash -n Install-TouchBridge-Linux.sh scripts/*.sh` — passed.
- Web DOM binding audit — 110 direct references, no missing element IDs.
- Shared Android assets — all eight controller assets are byte-for-byte identical to the desktop web UI (`index.html`, `app.css`, `app.js`, `theme.js`, `gamepad-core.js`, `manifest.webmanifest`, `icon.svg`, `sw.js`).
- One-second protocol fuzz smoke test — passed.
- One-second WebSocket parser fuzz smoke test — passed.
- Linux amd64, Linux arm64 and Windows amd64 desktop cross-compilation — passed.
- Android manifest and launcher/adaptive icon XML — parsed successfully.

The Android Gradle build was not run in this preparation workspace because it has no Android SDK or Gradle installation. The pull-request CI Android build must pass before merge. `govulncheck` is run in CI and was not run locally.

## Android update flow

The Android client checks the latest stable GitHub release and chooses the expected asset name for the installed build type. Before opening Android's package installer, it checks the APK package name, version name, increasing version code and signing-certificate fingerprint. The download is size-limited and the user must confirm installation through Android.

These checks have been reviewed in source, but the updater has not yet been exercised end-to-end on a physical Android device. Before publishing stable v0.8.0, update a signed v0.7.0 installation to v0.8.0 and confirm the same signing key is used, app data remains intact and the Android installer completes successfully. Also test invalid APK rejection, denied install permission, download interruption and offline/rate-limited GitHub responses.

## Android launcher icon

The manifest now declares `@mipmap/ic_launcher` and `@mipmap/ic_launcher_round`. The project includes legacy density PNGs, adaptive icon XML and a transparent adaptive foreground generated from the existing TouchBridge brand mark. The resource XML parses and the PNGs are present; confirm the icon visually in the built APK's launcher entry and Android installer on a real device.

## Layout import and adaptive sticks

The native document picker now accepts broad MIME types with preferred JSON/text types, while the web UI validates the selected layout. A FileReader fallback covers WebViews without `Blob.text()`. Legacy `Stick click L/R` labels are migrated to `L3/R3 click` when saved or imported layouts are loaded.

Physical-device checks still required:
- Import/export JSON through multiple Android file providers.
- Load an older layout and verify its stick-click labels migrate.
- Test adaptive-stick touch-down, whole-face movement, overlapping control stacking and reset on release, cancellation or focus loss.

## Release artifact status

The uploaded project included a prebuilt v0.7.0 debug APK. It is excluded from the 0.8.0 PR-ready source archive and must not be renamed or uploaded as an 0.8.0 build. Create `TouchBridge-0.8.0.apk` with the same private release keystore and alias that signed v0.7.0. The release workflow automatically attaches `TouchBridge-0.8.0-debug.apk` and the desktop packages; upload the signed stable APK manually.

## Stable release checklist

- [ ] PR CI, including the Android Gradle build, passes.
- [ ] Confirm the native app icon appears in the launcher and install confirmation.
- [ ] Update an installed signed v0.7.0 APK to signed v0.8.0 without uninstalling or losing app data.
- [ ] Confirm the in-app update rejects an APK with a different signing certificate or package.
- [ ] Test in-app update download progress, Android install permission and user confirmation.
- [ ] Test JSON import with different document providers and legacy label migration.
- [ ] Test adaptive stick motion and release behavior on the target phones.
- [ ] Test the new release on Android versions supported by the project.

Do not mark the release stable until the signed update path and new Android UI behaviors have been exercised on real devices.
