# TouchBridge

TouchBridge turns an Android phone into a customizable controller for a Windows or Linux PC over your local network.

[GitHub repository](https://github.com/Spont-427/touchbridge-controller) · [All releases](https://github.com/Spont-427/touchbridge-controller/releases)

**Version prepared:** `0.6.0` · **Release status:** preparing the first stable release

## Screenshots

**Controller layout**

<p align="center">
  <img src="docs/screenshots/controller-landscape.jpg" alt="TouchBridge controller with sticks, D-pad and buttons" width="100%">
</p>

**Layout library**

<p align="center">
  <img src="docs/screenshots/layout-library.jpg" alt="TouchBridge Android layout library" width="420">
</p>

## Download

You need the Android app on your phone and the desktop server on your PC. Download both from the release assets. You do not need to clone this repository or compile the project.

### Android

For normal use, install the **signed stable APK**:

[**Download TouchBridge 0.6.0 for Android**](https://github.com/Spont-427/touchbridge-controller/releases/download/v0.6.0/TouchBridge-0.6.0.apk)

A separate debug build is provided for testers and troubleshooting:

[Download the TouchBridge 0.6.0 debug APK](https://github.com/Spont-427/touchbridge-controller/releases/download/v0.6.0/TouchBridge-0.6.0-debug.apk)

The debug APK is debug-signed and may expose WebView debugging. It is intended for testers and troubleshooting, not normal stable use. GitHub Actions creates it on a temporary runner, so a later debug build may have a different signing key and may require uninstalling the earlier test build. Only install APKs downloaded from the official TouchBridge Releases page. These versioned links become available when the `v0.6.0` release and matching APK assets are published.

To install an APK:

1. Download it on your Android phone.
2. Open it from the Downloads notification or Downloads folder.
3. Tap **Install**, then **Open**.
4. If Android asks, allow your browser or file manager to install apps from that source. You can turn that permission off again after installation.

**Updating from an earlier alpha:** the stable APK is signed with the project's private release key, which differs from the Android debug key used by the alpha builds. Android will not install it over a debug-signed alpha. Export any custom layouts you want to keep, then uninstall the old alpha app before installing the stable APK. Uninstalling also clears app-local preferences, remembered computers and trusted pairing credentials, so you will need to configure your preferences and pair your PC again. Later stable updates can be installed in place as long as the same release key is used.

### Desktop server

Choose the download for your PC:

| PC | Download | Start it |
|---|---|---|
| Windows x64 (Intel/AMD) | [TouchBridge for Windows (.zip)](https://github.com/Spont-427/touchbridge-controller/releases/download/v0.6.0/TouchBridge-Windows-x64.zip) | Extract the ZIP and double-click `Start-TouchBridge.bat`. Run `Setup-TouchBridge-Windows.bat` once if you want virtual gamepad support. |
| Linux x86_64 | [TouchBridge for Linux x86_64 (.tar.gz)](https://github.com/Spont-427/touchbridge-controller/releases/download/v0.6.0/TouchBridge-Linux-x86_64.tar.gz) | Extract the archive and run `./Install-TouchBridge-Linux.sh` in a terminal. |
| Linux ARM64 | [TouchBridge for Linux ARM64 (.tar.gz)](https://github.com/Spont-427/touchbridge-controller/releases/download/v0.6.0/TouchBridge-Linux-ARM64.tar.gz) | Extract the archive and run `./Install-TouchBridge-Linux.sh` in a terminal. |

The `Release downloads` GitHub Actions workflow builds and attaches the Windows and Linux packages plus the Android debug APK when a GitHub Release is published. Build the signed stable APK on your own release machine with `scripts/build-android-release.sh`, then upload `TouchBridge-0.6.0.apk` to the same release.

### Windows virtual gamepad

Windows mouse and keyboard input work without extra software. Xbox 360-compatible virtual gamepad output requires the optional ViGEmBus 1.22.0 driver:

1. Extract the Windows ZIP.
2. Double-click `Setup-TouchBridge-Windows.bat` and approve the Windows administrator prompt. The script downloads the pinned official driver, verifies its SHA-256 checksum and configures private-network firewall rules.
3. Start TouchBridge with `Start-TouchBridge.bat`.
4. Keep the server window open. It displays the six-digit pairing PIN.

The TouchBridge server itself runs without administrator rights. See [Windows gamepad setup](docs/windows-gamepad.md).

### Linux virtual gamepad

Linux exposes a generic virtual gamepad through `/dev/uinput`. The installer configures the required device permissions. See [Linux setup](docs/linux-setup.md).

## Connect your phone to your PC

1. Start the desktop server on your PC.
2. Connect the phone and PC to the same trusted Wi-Fi network.
3. Open TouchBridge on the phone and select **Computers**.
4. Tap **Find computers** and select your PC. You can enter its address manually if it is not found.
5. Enter the six-digit PIN shown by the desktop server.

Keep the server running while using the phone as a controller.

## Features

- Customizable controller layouts with two analog sticks, a D-pad, buttons and triggers.
- Analog accelerator and brake pedals mapped to RT and LT.
- A draggable steering wheel mapped to the left stick's horizontal axis.
- Optional gyro steering for layouts that contain a steering wheel, with neutral calibration, filtering, deadzone and inversion.
- Digital and analog trigger modes. Digital LT/RT send both the button and full trigger pressure.
- Live network round-trip time with p95 and jitter details.
- Layout creation, editing, duplication, import and export.
- Size-aware placement for large controls to reduce overlap and keep controls on-screen.
- Four themes: System, Dark, AMOLED black and Light.
- Settings pages for Appearance, Controls & haptics, Analog sticks and About.
- PIN pairing, trusted reconnection and automatic PC discovery.
- Linux mouse, keyboard and virtual gamepad support; Windows mouse and keyboard plus optional XInput gamepad support.

## Platform support

- **Android:** Android 6.0 or newer.
- **Linux:** mouse, keyboard and generic virtual gamepad through `/dev/uinput`, subject to kernel support and permissions.
- **Windows:** mouse and keyboard out of the box; Xbox 360-compatible virtual gamepad through ViGEmBus 1.22.0 when installed.

## Build from source

Most users should use the downloads above. These steps are for developers.

### Android

Requirements: JDK 17, Android SDK Platform 35, and Android SDK Build Tools 36.0.0. The project uses Android Gradle Plugin 9.5.2 and the Gradle 9.6.0 wrapper.

Build a debug APK for testing:

```sh
./scripts/build-android-debug.sh
```

Build a signed stable APK after creating your private release keystore and setting the environment variables described in [Android APK build instructions](android-client/BUILD-APK.md):

```sh
./scripts/build-android-release.sh
```

The scripts write the APKs to `dist/`. Do not commit generated APKs or signing keys.

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
- [Validation record](docs/validation.md)
- [Security notes](SECURITY.md)
- [Changelog](CHANGELOG.md)
- [Release notes](RELEASE-NOTES.md)

## License

TouchBridge is licensed under the [MIT License](LICENSE).
