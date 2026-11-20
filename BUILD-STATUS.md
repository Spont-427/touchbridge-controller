# Build status — 0.5.2-alpha

Date: 2026-11-20. Android `versionCode` 10, `versionName` 0.5.2-alpha. Desktop `appVersion` 0.5.2-alpha.

## Implemented

- Offline-first Android app with the shared controller UI bundled at a fixed local origin.
- Dynamic home library built from four presets plus the custom-layout registry.
- Settings → Manage layouts: create from the Xbox-style base, duplicate any layout, rename custom layouts, and delete custom layouts after inline confirmation. Presets are protected in the UI and in the delete handler.
- Layout editor with drag positioning, control list, grid, 2% snap, smart placement, label/action/size/opacity editing, reset and JSON import/export.
- Add-menu groups for individual buttons and complete D-pad, face-button, bumper and trigger sets.
- Hostile-input validation for imported and stored layouts, including the 40-control limit and required left/right sticks.
- Landscape immersive controller and editor screens; system Back returns to the library or the previous screen. The on-screen back chip has been removed.
- Immediate button haptics through the platform predefined click waveform, gated by the persisted Button haptics setting. Sticks do not vibrate; controller long-press context menus are suppressed.
- System / Light / Dark theme selection with native system-bar icon updates.
- Adaptive or fixed stick centres, activation range, radial deadzone and three response curves.
- Analog or digital LT/RT modes.
- Android UDP discovery, manual address entry, PIN pairing and Keystore-encrypted trusted reconnect tokens.
- Optional mouse and keyboard tools.
- Linux x64/ARM64 and Windows x64 desktop builds; Windows gamepad output remains disabled.

## Validation performed

- `go test ./...` passed after the final changes.
- Go race tests and `go vet` passed during this source line.
- JavaScript syntax passed for the app, gamepad core, theme and service worker.
- DOM audit: 86 direct references, no missing IDs. CSS braces balanced at 769/769.
- Android assets are byte-identical to the shared web UI after sync.
- Android debug APK built successfully and was installed on an API 37 emulator.
- Emulator interaction tests covered layout create/duplicate/rename/delete, preset protection, active-layout fallback, persistence across force-stop, button sets, themes, system Back exit and haptics on/off.
- Haptics were verified through `dumpsys vibrator_manager`: an immediate `Prebaked=CLICK`, `usage: TOUCH` request with the toggle on, and no new request with it off.

See [docs/validation.md](docs/validation.md) for the full record.

## Not validated

- Physical Android hardware, including haptic motor feel, display cutouts and real touch sampling.
- Physical Windows/Linux input injection, discovery and gamepad enumeration.
- Real games, anti-cheat software, long sessions and unstable networks.
- `scripts/check.sh` in full: the environment's Node installation is broken (`libsimdjson.so.36`).
- `govulncheck`, CodeQL and GitHub Actions execution from this workspace.
- Release signing and store distribution. The built artifact is a debug APK.

## Important caveats

- Default HTTP/WebSocket transport is unencrypted. Tokens are protected at rest on Android and hashed on the server, but not encrypted in transit without TLS.
- Android targets API 35. Implement Android 17's `ACCESS_LOCAL_NETWORK` runtime permission before raising the target SDK to 37.
- Windows virtual gamepad output is not enabled; Linux generic uinput gamepad support depends on kernel and permission configuration.
- Desktop binaries are cross-compiled, not exercised on physical desktops in this environment.
