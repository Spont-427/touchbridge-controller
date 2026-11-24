# TouchBridge 0.7.0

TouchBridge 0.7.0 adds multiplayer controller sessions and screen-awake support.

## Highlights

### Multiplayer controllers

- Connect up to four phones to one desktop server at the same time.
- Each phone is assigned the next available player slot and shown as `P1`–`P4` in the controller HUD.
- Linux creates up to four `/dev/uinput` gamepads; Windows creates up to four ViGEmBus-backed Xbox 360-compatible controllers.
- A fifth connection receives `controller_busy` while all slots are occupied. Disconnected slots are reused.
- Disconnecting one phone releases only that player's gamepad and held desktop keys/buttons. Shared keys or mouse buttons remain held while another connected player still owns them.

### Keep screen on

- The phone display stays awake by default while the controller screen is open.
- A persistent **Keep screen on** toggle is available under **Settings → Controls & haptics**.
- Android applies `FLAG_KEEP_SCREEN_ON` through the native bridge and clears it when leaving the controller or destroying the activity.

## Downloads

The `v0.7.0` stable release should include:

- **Android stable:** `TouchBridge-0.7.0.apk` — signed with the same private release key used for v0.6.0.
- **Android debug:** `TouchBridge-0.7.0-debug.apk` — debug-signed, attached automatically for testers.
- **Windows x64:** `TouchBridge-Windows-x64.zip` — attached by GitHub Actions.
- **Linux x86_64:** `TouchBridge-Linux-x86_64.tar.gz` — attached by GitHub Actions.
- **Linux ARM64:** `TouchBridge-Linux-ARM64.tar.gz` — attached by GitHub Actions.

Build the stable APK with `scripts/build-android-release.sh` and the existing private release keystore. Do not create a new signing key for this update.

## Validation boundary

Automated tests cover slot assignment, the four-controller limit, slot reuse, per-player gamepad resets, shared keyboard/mouse ownership, screen-awake bridge calls and Android/web asset parity. Physical testing is still required with four phones, four gamepad devices on each desktop OS, and Android sleep/resume behaviour before publication.
