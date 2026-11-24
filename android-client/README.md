# TouchBridge Android client

This is the Android client for **TouchBridge 0.6.0**. Its controller interface is bundled inside the APK under `app/src/main/assets/controller/`, so the app can show the home screen, select profiles, edit layouts, preview controls and save preferences without a running desktop server.

## Features

- Local controller UI with offline profile selection and persistent layout editing. The editor provides a selectable control list, grid and snap-to-grid, smart placement, size and opacity controls, plus one-step D-pad, face-button, bumper and trigger sets.
- Settings → Manage layouts creates named custom layouts and duplicates any installed layout. Custom layouts can be renamed or deleted; presets cannot be deleted.
- Dedicated landscape, immersive-fullscreen controller and layout studio screens; the system Back gesture returns to the portrait home screen.
- System, Dark, AMOLED black and Light themes, persisted on the device and reflected in the native status and navigation bar icons.
- Automatic IPv4 UDP discovery of Windows/Linux TouchBridge servers and manual address fallback.
- First pairing with a six-digit PIN; subsequent connections use a device token issued by the server.
- Trusted device tokens encrypted at rest using Android Keystore AES-GCM. The JS bridge exposes tokens to the bundled local page only when needed for reconnection.
- Native document picker import and save-dialog export for JSON layouts.
- Optional mouse/keyboard tools; gamepad is the primary controller experience.
- Analog accelerator and brake pedals mapped to RT/LT, plus a draggable steering wheel mapped to the left stick horizontal axis.
- Optional gyro steering for wheel layouts, with neutral calibration, filtering, deadzone and direction inversion.
- Digital triggers send both button state and full trigger pressure; the controller shows RTT with p95/jitter details.
- Reorganized settings for Appearance, Controls & haptics, Analog sticks and About.
- Immediate button haptics using the platform click waveform, fired on pointer-down and gated by the Settings toggle. Sticks never vibrate.
- No on-screen controller back chip; the Android system Back gesture leaves the controller and restores portrait mode.
- Layout manager with create, duplicate, rename and inline-confirmed delete. Presets are protected from deletion.
- Editor add-menu groups for single buttons and complete D-pad, face-button, bumper and trigger sets.

## Build

Open this directory in Android Studio with Android SDK Platform 35, Build Tools 36.0.0 and JDK 17 installed. The project uses AGP 9.5.2 and the Gradle 9.6.0 wrapper. For a debug APK, run `../../scripts/build-android-debug.sh` from the repository root. For a stable signed APK, follow [BUILD-APK.md](BUILD-APK.md).

The shared web UI is copied from `../../internal/web/web/`. After editing those files, run `../../scripts/sync-android-assets.sh` from the repository root. `scripts/check.sh` verifies that the Android copy is up to date.

The version is `0.6.0` (`versionCode` 12); `minSdk` is 23 and the project targets SDK 35. Android 17 (API 37) requires the `ACCESS_LOCAL_NETWORK` runtime permission for apps targeting API 37 or higher; do not raise the target SDK until this permission and its rationale/denial flow are implemented. Apps targeting 35 still use the legacy `INTERNET`-permission behaviour described by Android's [local network permission guide](https://developer.android.com/privacy-and-security/local-network-permission).

## Network permissions

The app requires `INTERNET`, `ACCESS_NETWORK_STATE` and `ACCESS_WIFI_STATE` for LAN discovery and control. It uses cleartext HTTP/WebSocket because the default desktop server uses HTTP. Trusted server credentials are encrypted locally, but the default network transport does not encrypt the token in transit; use trusted Wi-Fi or configure TLS on the desktop server.
