# Build Android APKs

## Requirements

- Android Studio with Android SDK Platform 35 and Android SDK Build Tools 36.0.0.
- JDK 17 (Android Studio's embedded JBR 17 is suitable).
- Android Gradle Plugin 9.5.2 and the Gradle 9.6.0 wrapper.

The Android manifest includes the TouchBridge launcher icon, adaptive icon resources and legacy density-specific icons.

## Debug APK (testing)

From the repository root:

```sh
./scripts/build-android-debug.sh
```

The script runs `assembleDebug`, checks that the Gradle `versionName` matches `VERSION`, and copies the APK to `dist/TouchBridge-0.8.0-debug.apk`.

A debug APK is signed with the Android debug key and may expose WebView debugging. It is for testers and troubleshooting, not the recommended stable download.

## Stable signed APK

**Use the same private release keystore that signed v0.7.0. Do not generate a new key for this update.** Android uses the signing certificate to decide whether an APK can update an installed app. A different key prevents in-place updates and also prevents the in-app updater from accepting the APK.

Keep the keystore outside the repository, protect it with a strong password, and maintain encrypted offline backups. Never commit the keystore or passwords to GitHub.

The release script expects these environment variables:

- `TOUCHBRIDGE_RELEASE_KEYSTORE` — path to the existing `.jks` file.
- `TOUCHBRIDGE_RELEASE_KEY_ALIAS` — `touchbridge` unless you created a different alias.
- `TOUCHBRIDGE_RELEASE_STORE_PASSWORD` — keystore password.
- `TOUCHBRIDGE_RELEASE_KEY_PASSWORD` — private-key password.

Example for Fish shell:

```fish
set -gx TOUCHBRIDGE_RELEASE_KEYSTORE "$HOME/.keys/touchbridge-release.jks"
set -gx TOUCHBRIDGE_RELEASE_KEY_ALIAS "touchbridge"
read -s -P "Keystore password: " TOUCHBRIDGE_RELEASE_STORE_PASSWORD
set -gx TOUCHBRIDGE_RELEASE_STORE_PASSWORD "$TOUCHBRIDGE_RELEASE_STORE_PASSWORD"
echo
read -s -P "Key password: " TOUCHBRIDGE_RELEASE_KEY_PASSWORD
set -gx TOUCHBRIDGE_RELEASE_KEY_PASSWORD "$TOUCHBRIDGE_RELEASE_KEY_PASSWORD"
echo
```

Example for Bash:

```bash
export TOUCHBRIDGE_RELEASE_KEYSTORE="$HOME/.keys/touchbridge-release.jks"
export TOUCHBRIDGE_RELEASE_KEY_ALIAS="touchbridge"
read -rsp "Keystore password: " TOUCHBRIDGE_RELEASE_STORE_PASSWORD; echo
export TOUCHBRIDGE_RELEASE_STORE_PASSWORD
read -rsp "Key password: " TOUCHBRIDGE_RELEASE_KEY_PASSWORD; echo
export TOUCHBRIDGE_RELEASE_KEY_PASSWORD
```

Make sure `ANDROID_SDK_ROOT` points to your Android SDK (usually `"$HOME/Android/Sdk"`). Set JDK 17 as the Gradle JDK and run:

```sh
./scripts/build-android-release.sh
```

The script runs `assembleRelease`, aligns the unsigned APK, signs it with `apksigner`, verifies the signature and writes `dist/TouchBridge-0.8.0.apk` with a SHA-256 checksum.

Upload `TouchBridge-0.8.0.apk` manually to the GitHub release. The release workflow automatically attaches `TouchBridge-0.8.0-debug.apk` and the Windows/Linux packages.

## Android Studio

1. Open the `android-client` directory as a project.
2. Select JDK 17 under **Settings → Build Tools → Gradle → Gradle JDK**.
3. In **Tools → SDK Manager**, install Android 15 / API 35 and Android SDK Build Tools 36.0.0.
4. Allow Gradle Sync to finish.
5. For a test build, choose **Build → Build Bundle(s) / APK(s) → Build APK(s)**. For the stable build, use the signed release script above with the existing release keystore.

## Updating from v0.7.0 stable

The v0.7.0 stable APK was signed with the project's private release key. Build v0.8.0 using that exact same keystore and alias; users can update in place without uninstalling or losing app data. The in-app updater checks that the new APK has the same signing certificate before opening the Android installer. Do not use a debug APK as the stable release.

## App architecture

The Android app bundles the shared controller UI under `app/src/main/assets/controller/` and loads it from the fixed local origin `https://touchbridge.appassets.local`. It uses the local network only for discovery and input. After changing `internal/web/web/`, run `scripts/sync-android-assets.sh`; CI verifies byte-for-byte parity.
