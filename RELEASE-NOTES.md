# TouchBridge: Controller 0.5.2-alpha

This release turns the Android client into a more complete controller app: a dynamic layout library with a manager, faster layout building, a cleaner fullscreen play surface and reliable button haptics.

Interactive testing was performed in an Android emulator. Physical phones, physical desktops and real games still need validation; see [docs/validation.md](docs/validation.md).

## Highlights

### Layout library and manager

- **Settings → Manage layouts** lists every installed layout with its type, control count and active state.
- Create named custom layouts from the Xbox-style base.
- Duplicate any preset or custom layout, including its current controls.
- Rename custom layouts.
- Delete custom layouts after an inline confirmation. Presets have no delete action and the delete handler rejects preset IDs.
- The home library is rendered from the registry, so custom layouts appear immediately and can be played or edited like presets.

### Layout editor

- The add menu is grouped into **Single buttons** and **Button sets**.
- Place a complete D-pad, face-button diamond, bumper pair or trigger pair in one step.
- Sets are positioned in free space and can be moved individually afterward.
- Existing tools remain: control list, drag positioning, grid, 2% snap, smart placement, label/action/size/opacity, reset and JSON import/export.

### Controller and haptics

- Removed the small on-screen back chip. The Android system Back gesture leaves the controller and restores portrait mode.
- Button haptics now fire immediately on pointer-down using the platform predefined click waveform. The previous 9 ms one-shot was often below the perceptible threshold, making WebView's long-press haptic the only felt feedback.
- The **Button haptics** setting gates all feedback. With it off, no vibration request reaches Android.
- Sticks do not vibrate, and long-press context menus are suppressed on the controller surface.
- Adaptive stick origin rendering and neutral-on-release behavior are unchanged.

### Appearance

- System, Light and Dark themes, persisted on the device.
- Native status and navigation bar icons follow the selected theme.
- The dark theme remains the restrained AMOLED controller look; the light theme covers the library, manager, settings, controller and editor surfaces.

## Platform support

- **Android:** debug APK built from `android-client/`. `minSdk` 23, target/compile SDK 35, versionCode 10.
- **Windows x64:** mouse and keyboard through `SendInput`. Virtual gamepad output remains disabled.
- **Linux x64 / ARM64:** mouse, keyboard and generic evdev gamepad through `/dev/uinput` where permitted.
- **Desktop server:** version 0.5.2-alpha.

## Validation summary

Automated checks passed for Go tests, JavaScript syntax, DOM bindings, CSS balance and Android asset parity. The Android debug APK built and installed on an API 37 emulator.

Emulator tests covered layout creation, duplication, renaming, deletion, preset protection, active-layout fallback, persistence across force-stop, adding a D-pad set, themes, system Back exit and haptics. Haptics were confirmed through the system vibrator service: immediate `Prebaked=CLICK` with the setting on, no request with it off.

Not validated: physical Android hardware, physical Windows/Linux input injection, real games, unstable networks, release signing and remote CI. The full `scripts/check.sh` could not run because Node is broken in the build environment.

## Security note

The default desktop transport remains unencrypted HTTP/WebSocket. Android encrypts trusted tokens at rest with Android Keystore AES-GCM and the server stores only token hashes, but a LAN observer can capture traffic without TLS. Use trusted private Wi-Fi or configure TLS. Never expose the server to the public internet.

## Upgrade notes

- The Android package remains `io.github.touchbridge.mobile`; versionCode increases from 9 to 10.
- Existing local layouts and preferences are retained. The service-worker cache is bumped to v7 so the updated UI replaces cached assets.
- Windows virtual gamepad and API 37 local-network permission work are still future items.
