# Build status — 0.9.0

Date: 2026-11-26. Android `versionCode` 15, `versionName` 0.9.0. Desktop server version: 0.9.0.

## Changes included

- Redesign the portrait layout library with controller previews, a current-layout marker and clearer Play/Edit actions.
- Rework computer discovery and manual pairing instructions, labels, and saved/nearby computer sections.
- Reorganize Settings into clearer groups with updated labels and descriptions.
- Replace portrait-interface text-glyph icons with consistent inline SVGs and align settings icon containers.
- Start newly opened screens at the top; add subtle transitions that respect reduced-motion preferences.
- Recreate the centered phone mark for the Layouts header and Android launcher icon, including adaptive and density-specific resources.
- Add Material 3-inspired light, dark and AMOLED surfaces, larger touch targets, visible focus styles and safe-area spacing.
- Add the design-system reference and update Android UI documentation.
- Keep the landscape controller and layout editor visuals outside the redesign.
- Bump Android and desktop version to `0.9.0` (`versionCode` 15) and the service-worker shell cache to v14.

## Validation

`FUZZ_TIME=1x ./scripts/check.sh` — passed all nine stages. This includes `gofmt`, `go test -race ./...`, `go vet ./...`, JavaScript syntax checks, 17/17 Node tests, shell syntax, 110 DOM references, one-iteration protocol and WebSocket fuzz smoke tests, and byte-for-byte Android/web asset parity. `govulncheck` is not installed locally; CI runs the vulnerability scan. GitHub Actions must also pass the Android Gradle build and CI jobs before merge.

The release author reports that the updated Android APK built successfully and was reviewed in an Android Virtual Device (AVD). This workspace does not have an Android SDK configured, so that build was not independently reproduced here. Generated APKs and Android build directories are excluded from the PR source package.

## Stable release requirements

- Build `TouchBridge-0.9.0.apk` on the release machine with `scripts/build-android-release.sh`.
- Sign it with the exact same private release keystore and alias that signed v0.8.0. Do not create a new key.
- Upload the signed stable APK manually to the GitHub `v0.9.0` release. The release workflow attaches the debug APK and desktop packages.
- Keep the release keystore, passwords and generated APKs outside the source repository.

## Final review

- [ ] PR CI, including Android Gradle build, passes.
- [ ] Confirm the new launcher icon appears in the built APK.
- [ ] Verify light, dark and AMOLED rendering on the supported Android versions.
- [ ] Check layout-card Play/Edit actions and the current-layout indicator.
- [ ] Check saved/nearby discovery and manual pairing copy and flow.
- [ ] Check Settings groups, focus styling, touch targets and safe-area insets.
- [ ] Confirm new screens open at the top and that reduced-motion disables transitions.
- [ ] Confirm the controller and layout editor remain visually unchanged.
- [ ] Test the signed v0.8.0 → v0.9.0 in-place update with app data preserved.
