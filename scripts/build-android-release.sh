#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
VERSION="$(tr -d '\r\n' < "$ROOT/VERSION")"
GRADLE_FILE="$ROOT/android-client/app/build.gradle.kts"
KEYSTORE="${TOUCHBRIDGE_RELEASE_KEYSTORE:-}"
KEY_ALIAS="${TOUCHBRIDGE_RELEASE_KEY_ALIAS:-}"

if [[ -z "$VERSION" ]] || ! grep -Fq "versionName = \"$VERSION\"" "$GRADLE_FILE"; then
  echo "VERSION is empty or does not match android-client/app/build.gradle.kts." >&2
  exit 1
fi
if [[ -z "$KEYSTORE" || ! -f "$KEYSTORE" ]]; then
  echo "Set TOUCHBRIDGE_RELEASE_KEYSTORE to your private release keystore file." >&2
  exit 1
fi
if [[ -z "$KEY_ALIAS" ]]; then
  echo "Set TOUCHBRIDGE_RELEASE_KEY_ALIAS to the alias in your release keystore." >&2
  exit 1
fi
: "${TOUCHBRIDGE_RELEASE_STORE_PASSWORD:?Set TOUCHBRIDGE_RELEASE_STORE_PASSWORD in the environment.}"
: "${TOUCHBRIDGE_RELEASE_KEY_PASSWORD:?Set TOUCHBRIDGE_RELEASE_KEY_PASSWORD in the environment.}"

SDK_ROOT="${ANDROID_SDK_ROOT:-${ANDROID_HOME:-$HOME/Android/Sdk}}"
if [[ -z "${APKSIGNER:-}" ]]; then
  APKSIGNER="$(find "$SDK_ROOT/build-tools" -mindepth 2 -maxdepth 2 -type f -name apksigner -print 2>/dev/null | sort -V | tail -n 1 || true)"
fi
if [[ -z "${ZIPALIGN:-}" ]]; then
  ZIPALIGN="$(find "$SDK_ROOT/build-tools" -mindepth 2 -maxdepth 2 -type f -name zipalign -print 2>/dev/null | sort -V | tail -n 1 || true)"
fi
if [[ -z "${APKSIGNER:-}" || ! -x "$APKSIGNER" ]]; then
  echo "Could not find apksigner under $SDK_ROOT/build-tools. Install Android SDK Build Tools 36.0.0 or set APKSIGNER." >&2
  exit 1
fi
if [[ -z "${ZIPALIGN:-}" || ! -x "$ZIPALIGN" ]]; then
  echo "Could not find zipalign under $SDK_ROOT/build-tools. Install Android SDK Build Tools 36.0.0 or set ZIPALIGN." >&2
  exit 1
fi

cd "$ROOT/android-client"
./gradlew assembleRelease --no-daemon

UNSIGNED_APK="$ROOT/android-client/app/build/outputs/apk/release/app-release-unsigned.apk"
if [[ ! -f "$UNSIGNED_APK" ]]; then
  echo "Gradle completed but the expected unsigned APK was not found: $UNSIGNED_APK" >&2
  exit 1
fi

mkdir -p "$ROOT/dist"
ALIGNED_APK="$ROOT/android-client/app/build/outputs/apk/release/app-release-aligned.apk"
OUTPUT="$ROOT/dist/TouchBridge-${VERSION}.apk"
"$ZIPALIGN" -f -p 4 "$UNSIGNED_APK" "$ALIGNED_APK"
"$APKSIGNER" sign \
  --ks "$KEYSTORE" \
  --ks-key-alias "$KEY_ALIAS" \
  --ks-pass env:TOUCHBRIDGE_RELEASE_STORE_PASSWORD \
  --key-pass env:TOUCHBRIDGE_RELEASE_KEY_PASSWORD \
  --out "$OUTPUT" \
  "$ALIGNED_APK"
"$APKSIGNER" verify --verbose --print-certs "$OUTPUT"
sha256sum "$OUTPUT"
printf '\nSigned stable APK prepared: %s\n' "$OUTPUT"
printf 'Back up the release keystore securely. You must use the same key for every future update.\n'
