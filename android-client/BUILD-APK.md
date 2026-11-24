# Build Android APKs

## Requirements

- Android Studio with Android SDK Platform 35 and Android SDK Build Tools 36.0.0.
- JDK 17 (Android Studio's embedded JBR 17 is suitable).
- The project uses Android Gradle Plugin 9.5.2 and the Gradle 9.6.0 wrapper.

## Debug APK (testing)

From the repository root:

```sh
./scripts/build-android-debug.sh
```

The script runs `assembleDebug`, checks that the Gradle `versionName` matches `VERSION`, and copies the APK to `dist/TouchBridge-0.6.0-debug.apk`.

A debug APK is signed with the Android debug key and has WebView debugging enabled. It is useful for testers, but it is not the recommended stable download.

## Stable signed APK

Create a release keystore once and keep it private. Do not put it in the repository, a cloud-synced public folder or a GitHub issue. Keep at least one encrypted offline backup: losing the key means you cannot sign compatible updates for existing stable installations.

For example, create a key outside the repository:

```sh
mkdir -p "$HOME/.keys"
keytool -genkeypair -v \
  -keystore "$HOME/.keys/touchbridge-release.jks" \
  -alias touchbridge \
  -keyalg RSA -keysize 4096 -validity 10000
```

Set the path and alias, then enter the passwords without putting them in shell history:

```sh
export TOUCHBRIDGE_RELEASE_KEYSTORE="$HOME/.keys/touchbridge-release.jks"
export TOUCHBRIDGE_RELEASE_KEY_ALIAS="touchbridge"
read -rsp "Keystore password: " TOUCHBRIDGE_RELEASE_STORE_PASSWORD; echo
export TOUCHBRIDGE_RELEASE_STORE_PASSWORD
read -rsp "Key password: " TOUCHBRIDGE_RELEASE_KEY_PASSWORD; echo
export TOUCHBRIDGE_RELEASE_KEY_PASSWORD
```

Make sure `ANDROID_SDK_ROOT` points to your Android SDK (usually `"$HOME/Android/Sdk"`). Build and sign:

```sh
./scripts/build-android-release.sh
```

The script runs `assembleRelease`, aligns the unsigned APK, signs it using `apksigner`, verifies the signature and writes `dist/TouchBridge-0.6.0.apk` with a SHA-256 checksum. The signing passwords are read from environment variables; the script does not print them.

Upload both `TouchBridge-0.6.0.apk` and `TouchBridge-0.6.0-debug.apk` to the GitHub release if you want to provide a stable build and a separate tester build. The README links point to those exact release asset names.

## Android Studio

1. Open the `android-client` directory as a project.
2. Select JDK 17 under **Settings → Build Tools → Gradle → Gradle JDK**.
3. In **Tools → SDK Manager**, install Android 15 / API 35 and Android SDK Build Tools 36.0.0.
4. Allow Gradle sync to finish.
5. For a test build, choose **Build → Build Bundle(s) / APK(s) → Build APK(s)**. For the stable build, use the signed release script above after creating the keystore.

## Updating from the alpha debug APK

The stable APK is signed with your private release key, not the publicly known Android debug key used for development builds. Android treats those as different signing identities, so the first stable install cannot update an existing alpha debug install in place. Export custom layouts to JSON, then uninstall the alpha app before installing the stable APK. Uninstalling also clears app-local preferences, remembered computers and trusted pairing credentials, so configure preferences and pair the PC again after installation. Keep using the same private release key for every future stable update.

## App architecture

The Android app bundles the shared controller UI under `app/src/main/assets/controller/` and loads it from the fixed local origin `https://touchbridge.appassets.local`. It uses the local network only for discovery and input. After changing `internal/web/web/`, run `scripts/sync-android-assets.sh`; CI verifies byte-for-byte parity.
