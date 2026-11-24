# Build status — 0.8.0

Date: 2026-11-25. Android `versionCode` 14, `versionName` 0.8.0. Desktop server version: 0.8.0.

## Changes included

- Add an Android in-app update checker for the latest stable GitHub release.
- Select the stable or debug APK asset based on the installed build type.
- Show update status and download progress in Settings → About.
- Before opening Android's package installer, verify APK package name, version name, increasing version code and signing-certificate fingerprint.
- Limit downloaded update APKs to 200 MiB and accept only the expected official repository release-asset URL.
- Improve JSON layout import across Android document providers with inconsistent MIME types and use a FileReader fallback when `Blob.text()` is unavailable.
- Migrate saved and imported `Stick click L/R` labels to `L3/R3 click`.
- Change adaptive stick movement so the entire stick face moves to the initial touch point and remains there until release; remove the former activation-range control.
- Raise the active adaptive stick above neighboring controls and shorten long button labels.
- Add a proper Android launcher icon, adaptive icon definitions and legacy density resources using the existing TouchBridge mark.
- Bump Android and desktop version to `0.8.0` (`versionCode` 14) and the service-worker shell cache to v11.
- Restore executable permissions for the Gradle wrapper and the project's shell scripts.

## Automated checks run in this workspace

- `gofmt -l .` — clean.
- `go test -race ./...` — passed for all Go packages.
- `go vet ./...` — passed.
- Node syntax checks for `app.js`, `theme.js` and `gamepad-core.js` — passed.
- `node --test internal/web/web/gamepad-core.test.cjs` — 17/17 passed.
- `bash -n Install-TouchBridge-Linux.sh scripts/*.sh` — passed.
- Web DOM audit — 110 direct element references, no missing IDs.
- Shared Android/web assets — all eight controller assets are byte-for-byte identical.
- One-second protocol and WebSocket parser fuzz smoke tests — both passed.
- Linux amd64, Linux arm64 and Windows amd64 desktop cross-compilation — passed.
- Android manifest and launcher/adaptive icon XML files — well-formed XML.
- Launcher icon PNG density resources and adaptive foreground were generated from the existing TouchBridge brand mark.

This workspace does not have an Android SDK or Gradle installation, so the Android Gradle build and APK verification were not run here. The pull-request CI Android build must pass before merge. `govulncheck` is part of the repository CI workflow and was not run locally.

## Android updater and icon validation still required

- Install the signed v0.7.0 stable APK and verify that it finds the v0.8.0 stable release.
- Download the v0.8.0 stable APK and verify that the package name, version, version code and signing-certificate checks pass.
- Confirm Android's installer opens and the user can approve the update without losing app data.
- Confirm a mismatched or invalid APK is rejected before the installer opens.
- Test download progress, retry after denied install permission and behavior when GitHub is unavailable or rate-limited.
- Confirm the native TouchBridge icon appears in the launcher and Android's install confirmation on supported Android versions.
- Test JSON import using multiple document providers, including providers that report unusual MIME types.
- Load a layout containing the legacy `Stick click L/R` labels and confirm migration to `L3/R3 click`.
- Test adaptive-stick movement, overlap stacking and reset on release/cancellation on real phones.

## Release artifact status

The uploaded archive included a previously built v0.7.0 debug APK. It is excluded from the PR-ready source archive and must not be renamed or uploaded as v0.8.0. Build `TouchBridge-0.8.0.apk` on the release machine using the same private release keystore that signed v0.7.0. Do not create a new signing key. The release workflow automatically attaches `TouchBridge-0.8.0-debug.apk` and the Windows/Linux packages; upload the signed stable APK manually.

## Stable update compatibility

The in-app updater compares the installed signing-certificate fingerprint with the downloaded APK. Every stable release must be signed using the same private release key and alias. Debug APKs use a separate key and are tester builds; they may need manual replacement if their debug signing key differs. Keep the release keystore and encrypted backups outside the repository.
