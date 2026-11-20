#!/usr/bin/env bash
set -euo pipefail
SUDO=()
if [[ "$EUID" -ne 0 ]]; then sudo -v; SUDO=(sudo); fi
"${SUDO[@]}" rm -f /usr/local/bin/touchbridge
"${SUDO[@]}" rm -f /etc/udev/rules.d/70-touchbridge-uinput.rules
"${SUDO[@]}" rm -f /etc/modules-load.d/touchbridge-uinput.conf
"${SUDO[@]}" rm -f /usr/share/applications/touchbridge.desktop
if command -v udevadm >/dev/null; then "${SUDO[@]}" udevadm control --reload-rules; fi
if getent group touchbridge >/dev/null; then "${SUDO[@]}" groupdel touchbridge || true; fi
echo "TouchBridge files removed. Log out and back in to clear any previous group membership."
