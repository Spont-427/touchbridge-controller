#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
ARCH="$(uname -m)"
case "$ARCH" in
  x86_64|amd64) BIN="$ROOT/dist/touchbridge-linux-amd64" ;;
  aarch64|arm64) BIN="$ROOT/dist/touchbridge-linux-arm64" ;;
  *) echo "Unsupported Linux architecture: $ARCH" >&2; exit 1 ;;
esac
if [[ ! -f "$BIN" ]]; then
  echo "Missing server binary: $BIN" >&2
  echo "Re-download the complete release archive for this architecture." >&2
  exit 1
fi

CALLER="${SUDO_USER:-${USER:-}}"
if [[ -z "$CALLER" || "$CALLER" == root ]]; then
  echo "Run this installer as your normal desktop user; it will ask for sudo when needed." >&2
  exit 1
fi

SUDO=()
if [[ "$EUID" -ne 0 ]]; then
  command -v sudo >/dev/null || { echo "sudo is required for installation." >&2; exit 1; }
  sudo -v
  SUDO=(sudo)
fi

"${SUDO[@]}" install -Dm755 "$BIN" /usr/local/bin/touchbridge
"${SUDO[@]}" install -Dm644 "$ROOT/packaging/linux/70-touchbridge-uinput.rules" /etc/udev/rules.d/70-touchbridge-uinput.rules
"${SUDO[@]}" install -Dm644 "$ROOT/packaging/linux/uinput.conf" /etc/modules-load.d/touchbridge-uinput.conf
"${SUDO[@]}" install -Dm644 "$ROOT/packaging/linux/touchbridge.desktop" /usr/share/applications/touchbridge.desktop
if ! getent group touchbridge >/dev/null; then
  "${SUDO[@]}" groupadd --system touchbridge
fi
"${SUDO[@]}" usermod -aG touchbridge "$CALLER"
"${SUDO[@]}" modprobe uinput
if command -v udevadm >/dev/null; then
  "${SUDO[@]}" udevadm control --reload-rules
  "${SUDO[@]}" udevadm trigger --subsystem-match=misc || true
  "${SUDO[@]}" udevadm settle || true
fi
if [[ -e /dev/uinput ]] && command -v setfacl >/dev/null; then
  "${SUDO[@]}" setfacl -m "u:${CALLER}:rw" /dev/uinput || true
fi

if [[ ! -e /dev/uinput ]]; then
  echo "TouchBridge installed, but /dev/uinput was not created. Check that the kernel supports uinput." >&2
  exit 1
fi

echo
echo "TouchBridge installed. Starting the server now…"
echo "Later, start it anytime with: touchbridge"
if [[ "$EUID" -eq 0 ]]; then
  # The installer may have just added the user to the group. Start a child
  # with that group now rather than requiring a logout/login on common setups.
  if command -v sudo >/dev/null && command -v sg >/dev/null && \
      sudo -u "$CALLER" -- sg touchbridge -c 'test -w /dev/uinput' >/dev/null 2>&1; then
    exec sudo -u "$CALLER" -- sg touchbridge -c 'exec /usr/local/bin/touchbridge'
  fi
else
  if [[ -w /dev/uinput ]]; then
    exec /usr/local/bin/touchbridge
  fi
  if command -v sg >/dev/null && sg touchbridge -c 'test -w /dev/uinput' >/dev/null 2>&1; then
    exec sg touchbridge -c 'exec /usr/local/bin/touchbridge'
  fi
fi

echo "The server is installed, but this session cannot access /dev/uinput yet."
echo "Log out and back in once, then run: touchbridge"
exit 0
