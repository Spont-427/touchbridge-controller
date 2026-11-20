#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
VERSION="$(tr -d '\r\n' < "$ROOT/VERSION")"
GRADLE_FILE="$ROOT/android-client/app/build.gradle.kts"
APK="$ROOT/android-client/app/build/outputs/apk/debug/app-debug.apk"

if [[ -z "$VERSION" ]]; then
  echo "VERSION is empty; cannot name the APK." >&2
  exit 1
fi

if ! grep -Fq "versionName = \"$VERSION\"" "$GRADLE_FILE"; then
  echo "VERSION ($VERSION) does not match android-client/app/build.gradle.kts." >&2
  exit 1
fi

cd "$ROOT/android-client"
./gradlew assembleDebug --no-daemon

if [[ ! -f "$APK" ]]; then
  echo "Gradle completed but the expected APK was not found: $APK" >&2
  exit 1
fi

mkdir -p "$ROOT/dist"
OUTPUT="$ROOT/dist/TouchBridge-${VERSION}-debug.apk"
cp "$APK" "$OUTPUT"
sha256sum "$OUTPUT"
printf '\nRelease APK prepared: %s\n' "$OUTPUT"
printf 'Attach this file to the GitHub release tagged v%s.\n' "$VERSION"
