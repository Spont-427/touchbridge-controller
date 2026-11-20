# TouchBridge

[GitHub repository](https://github.com/Spont-427/touchbridge-controller) · [All releases](https://github.com/Spont-427/touchbridge-controller/releases)

TouchBridge lets you use an Android phone as a controller for a Windows or Linux PC over your local network.

**Current version:** `0.5.2-alpha` · **Status:** Alpha

## Download TouchBridge

You need two parts: the Android app on your phone and the desktop server on your PC. Download both from the links below. You do not need to clone this repository, install Go, or compile the project.

### 1. Android app

[**Download TouchBridge for Android (APK)**](https://github.com/Spont-427/touchbridge-controller/releases/download/v0.5.2-alpha/TouchBridge-0.5.2-alpha-debug.apk)

1. Download the APK on your Android phone.
2. Open it from your Downloads notification or Downloads folder.
3. Tap **Install**, then **Open**.
4. Android may ask you to allow your browser or file manager to install apps from that source. Allow it for the app you are using, finish the installation, and turn the permission off again afterwards if you prefer.

This is an alpha debug build for testing, not a Play Store release. It may contain bugs. Only install APKs downloaded from the official TouchBridge Releases page.

### 2. Desktop server (Windows or Linux)

Choose the download for your PC:

| Your PC | Download | How to start |
|---|---|---|
| Windows x64 (Intel/AMD) | [**TouchBridge for Windows x64 (.zip)**](https://github.com/Spont-427/touchbridge-controller/releases/download/v0.5.2-alpha/TouchBridge-Windows-x64.zip) | Extract the ZIP and double-click `Start-TouchBridge.bat`. |
| Linux, 64-bit Intel/AMD | [**TouchBridge for Linux x86_64 (.tar.gz)**](https://github.com/Spont-427/touchbridge-controller/releases/download/v0.5.2-alpha/TouchBridge-Linux-x86_64.tar.gz) | Extract the archive and run `./Install-TouchBridge-Linux.sh` in a terminal. |
| Linux, ARM64 | [**TouchBridge for Linux ARM64 (.tar.gz)**](https://github.com/Spont-427/touchbridge-controller/releases/download/v0.5.2-alpha/TouchBridge-Linux-ARM64.tar.gz) | Extract the archive and run `./Install-TouchBridge-Linux.sh` in a terminal. |

**Windows:** Extract the ZIP and double-click `Start-TouchBridge.bat`. Keep the server window open while using TouchBridge; the six-digit pairing PIN appears there. If Windows Firewall asks, allow TouchBridge on **Private networks only**. Because the executable is not code-signed, Windows may show a SmartScreen warning; only run it if you downloaded it from the official TouchBridge release. Windows currently supports mouse and keyboard input; virtual gamepad output is not available in this alpha release.

**Linux:** The installer asks for administrator permission to configure `/dev/uinput` and install the server. It then starts TouchBridge and displays the pairing PIN. See [Linux setup](docs/linux-setup.md) for details.

**Release downloads:** When a GitHub Release is published, the `Release downloads` workflow automatically builds and attaches the Windows and Linux packages. They appear under that release's **Assets** after the workflow finishes. The Android APK is uploaded separately as a release asset.

## Connect your phone to your PC

1. Start the desktop server on your PC using the instructions above.
2. Connect your phone and PC to the same trusted local Wi-Fi network.
3. Open TouchBridge on your phone and go to **Computers**.
4. Tap **Find computers** and select your PC. You can enter the PC address manually if it is not found.
5. Enter the six-digit pairing PIN shown by the desktop server.

The server must remain running while you use the phone as a controller.

## Features

- Customizable controller layouts with two analog sticks, a D-pad, buttons and triggers.
- Create, edit, duplicate, import and export layouts on Android.
- Offline layout editing and controller preview.
- Automatic PC discovery or manual server address entry.
- PIN pairing and trusted reconnection.
- Linux support for mouse, keyboard and a generic virtual gamepad.
- Windows support for mouse and keyboard input.

## Platform support

- **Android:** Android 6.0 or newer.
- **Linux:** mouse, keyboard and generic virtual gamepad through `/dev/uinput`, subject to kernel support and permissions.
- **Windows:** mouse and keyboard. Virtual gamepad output is not available in this alpha release.

## Build from source

These steps are for developers or users who want to build the project themselves. Most users should use the downloads above instead.

### Android app

Requirements: JDK 17 and Android SDK 35.

```sh
cd android-client
./gradlew assembleDebug
```

The debug APK is created at:

```text
android-client/app/build/outputs/apk/debug/app-debug.apk
```

### Desktop server

Requirement: Go 1.23 or newer.

Build and run for your current platform:

```sh
go build -o touchbridge ./cmd/touchbridge
./touchbridge
```

To build the Windows and Linux release packages locally:

```sh
./scripts/package-release.sh
```

The packages are created in `dist/`.

## Security

The default connection is not encrypted. Use TouchBridge only on a trusted private network unless you configure TLS. Do not expose its ports directly to the public internet.

## Documentation

- [Linux setup](docs/linux-setup.md)
- [Windows gamepad status](docs/windows-gamepad.md)
- [Controller UI](docs/controller-ui.md)
- [Network protocol](docs/protocol.md)
- [Validation record](docs/validation.md)
- [Security notes](SECURITY.md)
- [Changelog](CHANGELOG.md)

## License

TouchBridge is licensed under the [MIT License](LICENSE).
