# TouchBridge Android client

This is the Android client for **TouchBridge 0.9.0**. Its controller interface is bundled inside the APK under `app/src/main/assets/controller/`, so the app can show the home screen, select profiles, edit layouts, preview controls and save preferences without a running desktop server.

## Features

- Redesigned Material 3-inspired portrait layout library with controller previews, current-layout status and clear Play/Edit actions.
- Reworked computer discovery and manual pairing screens, and reorganized settings.
- Light, dark and AMOLED themes, consistent SVG icons, accessible focus states and safe-area spacing.
- Local controller UI with offline profile selection and persistent layout editing.
- Create, edit, duplicate, import and export custom layouts.
- Native Android document picker import and save-dialog export for JSON layouts, with compatibility for more document providers.
- Legacy `Stick click L/R` labels are migrated to `L3/R3 click` when saved or imported layouts are loaded.
- Adaptive sticks move the complete stick face to the first touch point; fixed-centre mode is also available.
- Multiplayer support for up to four phones, each with its own player badge and virtual gamepad slot.
- Keep the display awake while the controller screen is open; the setting is enabled by default and can be changed in Buttons & feedback.
- Analog pedals, steering wheel, optional gyro steering and digital/analog trigger modes.
- Live latency display with RTT, p95 and jitter details.
- System, Dark, AMOLED black and Light themes.
- Check for the latest stable GitHub release from Settings → About. The app verifies the downloaded APK's package, version and signing certificate before opening Android's installer.
- A native TouchBridge launcher icon is included in the APK.
- Trusted device tokens are encrypted at rest using Android Keystore AES-GCM.

## Build

Open this directory in Android Studio with Android SDK Platform 35, Build Tools 36.0.0 and JDK 17 installed. The project uses AGP 9.5.2 and the Gradle 9.6.0 wrapper.

For a debug APK, run `../../scripts/build-android-debug.sh` from the repository root. For a stable signed APK, follow [BUILD-APK.md](BUILD-APK.md).

The shared web UI is copied from `../../internal/web/web/`. After editing those files, run `../../scripts/sync-android-assets.sh` from the repository root. `scripts/check.sh` verifies that the Android copy is up to date.

The version is `0.9.0` (`versionCode` 15); `minSdk` is 23 and the project targets SDK 35. Android 17 (API 37) requires the `ACCESS_LOCAL_NETWORK` runtime permission for apps targeting API 37 or higher; do not raise the target SDK until this permission and its rationale/denial flow are implemented. Apps targeting 35 still use the legacy `INTERNET`-permission behaviour described by Android's [local network permission guide](https://developer.android.com/privacy-and-security/local-network-permission).

## Network permissions

The app requires `INTERNET`, `ACCESS_NETWORK_STATE` and `ACCESS_WIFI_STATE` for LAN discovery and control, plus `REQUEST_INSTALL_PACKAGES` for the user-confirmed in-app updater. It uses cleartext HTTP/WebSocket because the default desktop server uses HTTP. Trusted server credentials are encrypted locally, but the default network transport does not encrypt the token in transit; use trusted Wi-Fi or configure TLS on the desktop server.
