# TouchBridge Android client

This is the Android client for **TouchBridge 0.5.2-alpha**. Its controller interface is bundled inside the APK under `app/src/main/assets/controller/`, so the app can show the home screen, select profiles, edit layouts, preview controls and save preferences without a running desktop server.

## Features

- Local controller UI with offline profile selection and persistent layout editing. The editor provides a selectable control list, grid and snap-to-grid, smart placement, size and opacity controls, plus one-step D-pad, face-button, bumper and trigger sets.
- Settings → Manage layouts creates named custom layouts and duplicates any installed layout. Custom layouts can be renamed or deleted; presets cannot be deleted.
- Dedicated landscape, immersive-fullscreen controller and layout studio screens; the system Back gesture returns to the portrait home screen.
- Light, dark or system theme, persisted on the device and reflected in the native status and navigation bar icons.
- Automatic IPv4 UDP discovery of Windows/Linux TouchBridge servers and manual address fallback.
- First pairing with a six-digit PIN; subsequent connections use a device token issued by the server.
- Trusted device tokens encrypted at rest using Android Keystore AES-GCM. The JS bridge exposes tokens to the bundled local page only when needed for reconnection.
- Native document picker import and save-dialog export for JSON layouts.
- Optional mouse/keyboard tools; gamepad is the primary controller experience.
- Settings for theme, button haptics, analog/digital LT/RT triggers, stick centring, deadzone and response curve.
- Immediate button haptics using the platform click waveform, fired on pointer-down and gated by the Settings toggle. Sticks never vibrate.
- No on-screen controller back chip; the Android system Back gesture leaves the controller and restores portrait mode.
- Layout manager with create, duplicate, rename and inline-confirmed delete. Presets are protected from deletion.
- Editor add-menu groups for single buttons and complete D-pad, face-button, bumper and trigger sets.

## Build

Open this directory in Android Studio with Android SDK 35 and JDK 17 installed. Sync Gradle and choose **Build > Build APK(s)**. The debug APK is generated under `app/build/outputs/apk/debug/`.

The shared web UI is copied from `../../internal/web/web/`. After editing those files, run `../../scripts/sync-android-assets.sh` from the repository root. `scripts/check.sh` verifies that the Android copy is up to date.

The version is `0.5.2-alpha` (`versionCode` 10); `minSdk` is 23 and the project targets SDK 35. Android 17 (API 37) requires the `ACCESS_LOCAL_NETWORK` runtime permission for apps targeting API 37 or higher; do not raise the target SDK until this permission and its rationale/denial flow are implemented. Apps targeting 35 still use the legacy `INTERNET`-permission behaviour described by Android's [local network permission guide](https://developer.android.com/privacy-and-security/local-network-permission).

## Network permissions

The app requires `INTERNET`, `ACCESS_NETWORK_STATE` and `ACCESS_WIFI_STATE` for LAN discovery and control. It uses cleartext HTTP/WebSocket because the default desktop server uses HTTP. Trusted server credentials are encrypted locally, but the default network transport does not encrypt the token in transit; use trusted Wi-Fi or configure TLS on the desktop server.
