# TouchBridge 0.8.0

TouchBridge 0.8.0 adds in-app Android updates and controller usability refinements.

## Highlights

### In-app updates

- Check the latest stable GitHub release from **Settings → About**.
- See update status and download progress.
- Before opening Android's installer, verify the APK package name, version, version code and signing certificate.
- Installation remains under the user's control; TouchBridge does not silently install updates.
- Stable updates must be signed with the same private release key. Debug APKs are separate tester builds.

### Controller and layout improvements

- Make JSON layout import work with a wider range of Android document providers.
- Use a `FileReader` fallback when `Blob.text()` is unavailable.
- Migrate old `Stick click L/R` labels to `L3/R3 click` in saved and imported layouts.
- In adaptive mode, move the whole stick face to the first touch point and keep it there until release.
- Keep the active adaptive stick above nearby controls.
- Shorten long button labels to fit the control.
- Add a native Android launcher icon using the TouchBridge brand mark.

## Downloads

The `v0.8.0` stable release should include:

- **Android stable:** `TouchBridge-0.8.0.apk` — signed with the same private release key used for v0.7.0.
- **Android debug:** `TouchBridge-0.8.0-debug.apk` — debug-signed and attached automatically for testers.
- **Windows x64:** `TouchBridge-Windows-x64.zip` — attached by GitHub Actions.
- **Linux x86_64:** `TouchBridge-Linux-x86_64.tar.gz` — attached by GitHub Actions.
- **Linux ARM64:** `TouchBridge-Linux-ARM64.tar.gz` — attached by GitHub Actions.

Build the stable APK with `scripts/build-android-release.sh` and the existing private release keystore. Do not create a new signing key.

## Validation boundary

Automated source checks cover Go, JavaScript, shell scripts, DOM references, shared Android/web assets and desktop cross-compilation. The Android Gradle build and update installation flow must pass CI and be tested on a real Android device before publication. In particular, test updating from the signed v0.7.0 stable APK to v0.8.0 with the same release key.
