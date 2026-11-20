#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
DIST="$ROOT/dist"

"$ROOT/scripts/build.sh"

# Windows x64: keep the executable and launcher together so users can start
# the server by double-clicking a single file.
WIN_DIR="$DIST/package-windows-x64"
rm -rf "$WIN_DIR"
mkdir -p "$WIN_DIR"
cp "$DIST/touchbridge-windows-amd64.exe" "$WIN_DIR/touchbridge.exe"
cp "$ROOT/packaging/windows/"* "$WIN_DIR/"
cat > "$WIN_DIR/Start-TouchBridge.bat" <<'BAT'
@echo off
setlocal
cd /d "%~dp0"
reg query "HKLM\SYSTEM\CurrentControlSet\Services\ViGEmBus" >nul 2>&1
if errorlevel 1 (
  echo Note: ViGEmBus is not installed, so gamepad output is unavailable.
  echo Run Setup-TouchBridge-Windows.bat once to install it.
  echo Mouse and keyboard input will still work.
  echo.
)
echo Starting TouchBridge Server...
echo Keep this window open while using TouchBridge.
echo The six-digit pairing PIN will appear below.
echo.
touchbridge.exe
echo.
echo TouchBridge has stopped.
pause
BAT
cat > "$WIN_DIR/README.txt" <<'TXT'
TouchBridge for Windows x64

1. Extract this ZIP file.
2. Optional gamepad setup: double-click Setup-TouchBridge-Windows.bat once.
   It downloads ViGEmBus 1.22.0 from the official release, verifies its
   SHA-256 checksum, installs the driver, and adds private-network firewall
   rules for TCP 8765 and UDP 8766. Windows asks for administrator
   permission, and may ask you to restart afterwards.
   Skip this step if you only need mouse and keyboard input.
3. Double-click Start-TouchBridge.bat.
4. Keep the window open. The six-digit pairing PIN appears there.
5. Connect your Android phone and PC to the same trusted Wi-Fi network.
6. In the Android app, open Computers, choose Find computers, and select this PC.
7. Enter the pairing PIN shown in this window.

TouchBridge itself never asks for administrator rights. Without ViGEmBus,
Start-TouchBridge.bat still runs, but gamepad output is unavailable.
If you skipped the firewall step, allow TouchBridge on Private networks only
when Windows asks.

To stop the server, focus this window and press Ctrl+C.
TXT
(
  cd "$WIN_DIR"
  zip -qr "$DIST/TouchBridge-Windows-x64.zip" .
)

# Linux packages include the existing installer and the correct architecture
# binary in the directory layout expected by scripts/install-linux.sh.
for ARCH in amd64 arm64; do
  case "$ARCH" in
    amd64) LABEL="x86_64" ;;
    arm64) LABEL="ARM64" ;;
  esac
  PKG="$DIST/package-linux-$ARCH"
  rm -rf "$PKG"
  mkdir -p "$PKG/scripts" "$PKG/packaging/linux" "$PKG/dist" "$PKG/docs"
  cp "$ROOT/Install-TouchBridge-Linux.sh" "$PKG/"
  cp "$ROOT/scripts/install-linux.sh" "$ROOT/scripts/uninstall-linux.sh" "$PKG/scripts/"
  cp "$ROOT/packaging/linux/"* "$PKG/packaging/linux/"
  cp "$ROOT/docs/linux-setup.md" "$PKG/docs/"
  cp "$DIST/touchbridge-linux-$ARCH" "$PKG/dist/touchbridge-linux-$ARCH"
  chmod +x "$PKG/Install-TouchBridge-Linux.sh" "$PKG/scripts/"*.sh "$PKG/dist/touchbridge-linux-$ARCH"
  cat > "$PKG/README.txt" <<TXT
TouchBridge for Linux $LABEL

1. Extract this archive.
2. Open a terminal in the extracted TouchBridge folder.
3. Run: ./Install-TouchBridge-Linux.sh
4. Enter your password when asked. The installer configures the Linux virtual-input device and starts the server.
5. The six-digit pairing PIN appears in the terminal.
6. Connect your Android phone and PC to the same trusted Wi-Fi network.
7. In the Android app, open Computers, choose Find computers, and select this PC.
8. Enter the pairing PIN.

The installer requires sudo access and a Linux kernel with /dev/uinput support. See docs/linux-setup.md for details.
TXT
  case "$ARCH" in
    amd64) ARCHIVE="$DIST/TouchBridge-Linux-x86_64.tar.gz" ;;
    arm64) ARCHIVE="$DIST/TouchBridge-Linux-ARM64.tar.gz" ;;
  esac
  tar -czf "$ARCHIVE" -C "$PKG" .
done

printf '\nRelease assets created in %s:\n' "$DIST"
ls -lh "$DIST"/TouchBridge-*.zip "$DIST"/TouchBridge-*.tar.gz
