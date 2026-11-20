#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
DEST="$ROOT/android-client/app/src/main/assets/controller"
mkdir -p "$DEST"
for asset in index.html app.css app.js theme.js gamepad-core.js manifest.webmanifest icon.svg sw.js; do
  install -m 0644 "$ROOT/internal/web/web/$asset" "$DEST/$asset"
done
printf 'Synced shared controller UI into Android assets.\n'
