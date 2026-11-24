# TouchBridge

TouchBridge turns Android phones into customizable wireless controllers for a Windows or Linux PC over your local network.

[GitHub repository](https://github.com/Spont-427/touchbridge-controller) · [All releases](https://github.com/Spont-427/touchbridge-controller/releases) · [Report a bug](https://github.com/Spont-427/touchbridge-controller/issues)

**Version:** `0.7.0` · **Release:** preparing the next stable release

## Screenshots

**Controller layout**

<p align="center">
  <img src="docs/screenshots/controller-landscape.jpg" alt="TouchBridge controller layout" width="100%">
</p>

**Android app**

<table>
  <tr>
    <th align="center">Layout library</th>
    <th align="center">Settings</th>
  </tr>
  <tr>
    <td align="center"><img src="docs/screenshots/layout-library.jpg" alt="TouchBridge layout library" width="260"></td>
    <td align="center"><img src="docs/screenshots/settings.jpg" alt="TouchBridge settings" width="260"></td>
  </tr>
</table>

## Download

You need the Android app on each phone and the desktop server on your PC. Download both from the release assets. You do not need to clone this repository or compile the project.

### Android app

<p>
  <a href="https://github.com/Spont-427/touchbridge-controller/releases/download/v0.7.0/TouchBridge-0.7.0.apk"><img src="internal/web/web/icon.svg" alt="TouchBridge app icon" width="32" height="32" align="middle"> <strong>Download TouchBridge 0.7.0 for Android</strong></a>
</p>

The stable APK is signed with the project's private release key and is the recommended download for normal use.

[Download the separate debug APK (for testing)](https://github.com/Spont-427/touchbridge-controller/releases/download/v0.7.0/TouchBridge-0.7.0-debug.apk)

The debug APK is debug-signed and may expose WebView debugging. Use it for testing and troubleshooting, not as the normal stable installation. The `Release downloads` workflow builds and attaches it automatically. Only install APKs from the official TouchBridge Releases page.

The stable download link becomes active when the `v0.7.0` release and its matching APK are published.

### Desktop server

Choose the package for your PC:

| PC | Download | Start it |
|---|---|---|
| Windows x64 (Intel/AMD) | [TouchBridge for Windows (.zip)](https://github.com/Spont-427/touchbridge-controller/releases/download/v0.7.0/TouchBridge-Windows-x64.zip) | Extract the ZIP and double-click `Start-TouchBridge.bat`. Run `Setup-TouchBridge-Windows.bat` once for virtual gamepad support. |
| Linux x86_64 | [TouchBridge for Linux x86_64 (.tar.gz)](https://github.com/Spont-427/touchbridge-controller/releases/download/v0.7.0/TouchBridge-Linux-x86_64.tar.gz) | Extract the archive and run `./Install-TouchBridge-Linux.sh` in a terminal. |
| Linux ARM64 | [TouchBridge for Linux ARM64 (.tar.gz)](https://github.com/Spont-427/touchbridge-controller/releases/download/v0.7.0/TouchBridge-Linux-ARM64.tar.gz) | Extract the archive and run `./Install-TouchBridge-Linux.sh` in a terminal. |

The `Release downloads` GitHub Actions workflow builds and attaches the Windows and Linux packages plus the Android debug APK when a GitHub Release is published. Build the stable signed APK on your release machine with `scripts/build-android-release.sh`, then upload `TouchBridge-0.7.0.apk` to the same release.

### Windows virtual gamepad

Windows mouse and keyboard input work without extra software. Xbox 360-compatible virtual gamepad output requires the optional ViGEmBus 1.22.0 driver:

1. Extract the Windows ZIP.
2. Double-click `Setup-TouchBridge-Windows.bat` and approve the Windows administrator prompt. The script downloads the pinned official driver, verifies its SHA-256 checksum and configures private-network firewall rules.
3. Start TouchBridge with `Start-TouchBridge.bat`.
4. Keep the server window open. It displays the six-digit pairing PIN.

The TouchBridge server itself runs without administrator rights. See [Windows gamepad setup](docs/windows-gamepad.md).

### Linux virtual gamepad

Linux exposes up to four generic virtual gamepads through `/dev/uinput`. The installer configures the required device permissions. See [Linux setup](docs/linux-setup.md).

## Connect your phones to your PC

1. Start the desktop server on your PC.
2. Connect the PC and all phones to the same trusted Wi-Fi network.
3. Open TouchBridge on each phone and go to **Computers**.
4. Tap **Find computers** and select the PC. You can enter its address manually if it is not found.
5. Enter the six-digit pairing PIN shown by the desktop server on each phone.

Up to **four phones** can control the same PC simultaneously. Each phone receives a player slot (`P1`–`P4`) and its own virtual gamepad. A fifth connection is rejected while all four slots are occupied. When a phone disconnects, its slot is released for the next player; the other players keep their gamepad and held input state.

## Features

- Up to four simultaneous phone controllers with player badges and independent virtual gamepad slots.
- Per-player cleanup: disconnecting one phone releases only its gamepad and its owned keyboard/mouse holds.
- Customizable controller layouts with two analog sticks, a D-pad, buttons and triggers.
- Analog accelerator and brake pedals mapped to RT and LT.
- Draggable steering wheel mapped to the left stick's horizontal axis, with automatic recentering.
- Optional gyro steering with neutral calibration, filtering, deadzone and inversion.
- Digital and analog trigger modes; digital LT/RT send both the button and full trigger pressure.
- Network latency indicator with RTT, p95 and jitter details.
- Keeps the phone screen awake while the controller is open, enabled by default and configurable under **Settings → Controls & haptics**.
- Layout creation, editing, duplication, import and export, with size-aware placement for large controls.
- System, Dark, AMOLED black and Light themes.
- Settings for Appearance, Controls & haptics, Analog sticks and About.
- PIN pairing, trusted reconnection and automatic PC discovery.

## Platform support

- **Android:** Android 6.0 or newer.
- **Linux:** mouse, keyboard and up to four generic virtual gamepads through `/dev/uinput`, subject to kernel support and permissions.
- **Windows:** mouse and keyboard out of the box; up to four Xbox 360-compatible virtual gamepads through ViGEmBus 1.22.0 when installed.

## Build from source

Most users should use the downloads above. These steps are for developers.

### Android

Requirements: JDK 17, Android SDK Platform 35 and Android SDK Build Tools 36.0.0. The project uses Android Gradle Plugin 9.5.2 and the Gradle 9.6.0 wrapper.

Build a debug APK for testing:

```sh
./scripts/build-android-debug.sh
```

Build a stable signed APK using the same private release keystore as earlier stable releases. See [Android APK build instructions](android-client/BUILD-APK.md):

```sh
./scripts/build-android-release.sh
```

The scripts write APKs to `dist/`. Do not commit generated APKs or signing keys.

### Desktop server

Requirement: Go 1.23 or newer.

```sh
go build -o touchbridge ./cmd/touchbridge
./touchbridge
```

Build the Windows and Linux packages locally:

```sh
./scripts/package-release.sh
```

The packages are created in `dist/`.

## Security

The default HTTP/WebSocket connection is not encrypted. Use TouchBridge only on a trusted private network unless you configure TLS. Do not expose its ports directly to the public internet. Windows virtual gamepad output uses the third-party ViGEmBus driver; see the [security notes](SECURITY.md).

## Documentation

- [Linux setup](docs/linux-setup.md)
- [Windows gamepad](docs/windows-gamepad.md)
- [Controller UI](docs/controller-ui.md)
- [Network protocol](docs/protocol.md)
- [Stability and validation](docs/stability.md)
- [Security notes](SECURITY.md)
- [Changelog](CHANGELOG.md)
- [Release notes](RELEASE-NOTES.md)

## License

TouchBridge is licensed under the [MIT License](LICENSE).
