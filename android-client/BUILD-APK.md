# Build the Android APK

## Requirements

- Android Studio with Android SDK Platform 35 and Build Tools 35.0.0.
- JDK 17 (Android Studio's embedded JBR 17 is suitable).
- Gradle 8.9 for the AGP 8.7.3 build configured in CI.

## Android Studio

1. Extract the TouchBridge source archive.
2. Open the `android-client` directory as the project (not the outer release directory).
3. Select JDK 17 under **Settings → Build Tools → Gradle → Gradle JDK**.
4. In **Tools → SDK Manager → SDK Platforms**, install **Android 15 / API 35**. Install Android SDK Build-Tools 35.0.0 from SDK Tools.
5. Sync the project. CI pins Gradle 8.9 for reproducible builds.
6. Select **Build → Build Bundle(s) / APK(s) → Build APK(s)**.
7. The debug APK is written to `app/build/outputs/apk/debug/app-debug.apk`.

## Command line

From the repository root, with JDK 17 and Android SDK 35 configured:

```sh
./scripts/build-android-release.sh
```

The script runs the Gradle wrapper and copies the APK to `dist/TouchBridge-0.5.3-alpha-debug.apk`. Attach that file to the GitHub release tagged `v0.5.3-alpha`.

**Keep the same signing key for updates:** the alpha APKs are debug-signed. Build on the same development machine used for the previous alpha APK so Android Studio/Gradle reuses that machine's `~/.android/debug.keystore`. Do not delete or regenerate that keystore, and never commit it to Git. The `0.5.3-alpha` APK built for this release has a signing certificate matching the previously published `0.5.2-alpha` APK. The APK is distributed separately as a GitHub Release asset; generated APKs are intentionally excluded from the source repository. Building on a different machine or an ephemeral CI runner will normally use a different debug key, and Android will reject it as an in-place update; users would have to uninstall the old app first (which removes its local app data).

## App architecture

The Android app loads its bundled UI from `app/src/main/assets/controller/` using the fixed local origin `https://touchbridge.appassets.local`. It does not load UI pages from the PC. Only discovery and input use the local network. The native bridge handles LAN UDP discovery, encrypted trusted-server storage and orientation/fullscreen changes.

After changing the shared web UI in `internal/web/web/`, run `scripts/sync-android-assets.sh` from the repository root. The quality gate verifies that the two copies match.

## Install

```sh
adb install -r app/build/outputs/apk/debug/app-debug.apk
```
