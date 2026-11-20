#!/usr/bin/env bash
# Reproducible local quality gate. This script never rewrites source files.
set -euo pipefail
ROOT="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"
FUZZ_TIME="${FUZZ_TIME:-1s}"

echo '[1/9] Go formatting'
UNFORMATTED="$(gofmt -l .)"
if [[ -n "$UNFORMATTED" ]]; then
  printf 'Go files need gofmt:\n%s\n' "$UNFORMATTED" >&2
  exit 1
fi

echo '[2/9] Go race-enabled tests'
go test -race ./...

echo '[3/9] Go static analysis'
go vet ./...

echo '[4/9] JavaScript syntax and controller math tests'
node --check internal/web/web/app.js
node --check internal/web/web/theme.js
node --check internal/web/web/gamepad-core.js
node --test internal/web/web/gamepad-core.test.cjs

echo '[5/9] Shell syntax'
bash -n Install-TouchBridge-Linux.sh scripts/*.sh

echo '[6/9] Web DOM bindings'
python3 - <<'PY'
import re
from pathlib import Path
html = Path('internal/web/web/index.html').read_text()
js = Path('internal/web/web/app.js').read_text()
ids = set(re.findall(r'\bid="([A-Za-z0-9_-]+)"', html))
refs = set(re.findall(r"\$\('([A-Za-z0-9_-]+)'\)", js))
missing = sorted(refs - ids)
if missing:
    raise SystemExit(f'JavaScript references missing HTML IDs: {missing}')
print(f'Validated {len(refs)} direct DOM ID references.')
PY

echo '[7/9] Go fuzz smoke tests'
GOMAXPROCS=1 go test ./internal/protocol -run='^$' -fuzz=FuzzDecodeNeverPanics -fuzztime="$FUZZ_TIME" -parallel=1
GOMAXPROCS=1 go test ./internal/web -run='^$' -fuzz=FuzzReadMessageNeverPanics -fuzztime="$FUZZ_TIME" -parallel=1

echo '[8/9] Android bundled UI matches the desktop web UI'
for asset in index.html app.css app.js theme.js gamepad-core.js manifest.webmanifest icon.svg sw.js; do
  cmp -s "internal/web/web/$asset" "android-client/app/src/main/assets/controller/$asset" || {
    echo "Android asset out of sync: $asset. Run scripts/sync-android-assets.sh" >&2
    exit 1
  }
done
echo '[9/9] Known-vulnerability scan (optional locally; CI requires it)'
if command -v govulncheck >/dev/null 2>&1; then
  govulncheck ./...
else
  echo 'govulncheck is not installed; GitHub Actions runs this check in its vulnerability-scan job.'
fi

echo 'All available local checks passed.'
