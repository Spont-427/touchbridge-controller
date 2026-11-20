# Validation record — 0.5.2-alpha

Date: 2026-11-20. Android build: debug APK, `versionCode` 10, `versionName` 0.5.2-alpha.

This record separates automated checks from interactive emulator testing. Passing checks do not prove physical-device stability.

## Automated checks

- `gofmt -l .` — no unformatted Go files in the 0.5.2 source tree.
- `go test ./...` — passed for all Go packages after the final source changes.
- `go test -race ./...` and `go vet ./...` — passed on this source line during the 0.5.1/0.5.2 work; the 0.5.2 changes are confined to web assets, Android host code and documentation.
- JavaScript syntax — `app.js`, `gamepad-core.js`, `theme.js` and `sw.js` parse successfully with Deno.
- HTML-to-JavaScript binding audit — 86 direct DOM ID references, no missing IDs.
- CSS — 769 opening and 769 closing braces.
- Shared Android assets — all eight bundled controller assets are byte-for-byte identical to `internal/web/web/` after `scripts/sync-android-assets.sh` (`index.html`, `app.css`, `app.js`, `theme.js`, `gamepad-core.js`, `manifest.webmanifest`, `icon.svg`, `sw.js`).
- Android build — `./gradlew assembleDebug --no-daemon` passed and produced `android-client/app/build/outputs/apk/debug/app-debug.apk` (about 911 KiB).
- Install — `adb install -r app-debug.apk` succeeded on the API 37 emulator.

### Not run

- `scripts/check.sh` could not run in full because the environment's Node installation is broken (`libsimdjson.so.36` missing). The individual JavaScript, DOM, asset-parity, CSS and Go checks above were run directly instead.
- `node --test internal/web/web/gamepad-core.test.cjs` was therefore not executed for this build. The last recorded result on this source line was 13 passing tests.
- `govulncheck`, CodeQL and the GitHub Actions workflows have not run in this environment. They run on push or schedule in CI.
- No release-signed APK was produced; the validated artifact is debug-signed.

## Interactive emulator tests

Device: Android emulator `Medium_Phone_API_37.0`, headless, API 37, 914 × 411 CSS viewport in landscape (2.625 device pixel ratio). Package `io.github.touchbridge.mobile`.

The following flows were exercised by a human-directed agent through the UI and the WebView DevTools socket:

### Layout library and manager

- Opened **Settings → Manage layouts**; presets showed Edit and Duplicate only.
- Created **Elden Ring**; it appeared on the home library with generic artwork and `4 PRESETS · 1 CUSTOM`.
- Duplicated it as **Elden Ring copy** with copied controls.
- Opened the editor from the manager, added a D-pad set, and confirmed the control count increased from 16 to 20.
- Pressed **Done**, returned to Manage layouts, renamed the layout to **Elden Ring 2**.
- Deleted both custom layouts through the inline confirmation. The registry returned to `[]` and the library showed `4 PRESETS`.
- Verified custom layouts and the active selection survived a force-stop and relaunch.
- Deleted the active custom layout and confirmed the app fell back to Xbox-style and updated `touchbridge.lastProfile`.
- Confirmed preset deletion is unavailable in the UI and rejected by `deleteLayout()`.
- Reviewed the manage screen in both dark and light themes.

### Editor

- Verified the add menu contains **Single buttons** and **Button sets** groups.
- Added a complete D-pad set; all four controls were placed in free space and selectable.
- Confirmed empty-canvas taps clear the selection and hide the property panel.
- Confirmed stick controls cannot be removed.

### Navigation and themes

- Verified there is no controller back chip in the DOM at runtime.
- Pressed Android Back from the controller; the app returned to the portrait library and restored portrait system bars.
- Switched System, Light and Dark themes; summary, settings, manager and editor surfaces updated and persisted across relaunch.
- Captured screenshots for the library, manager, create form, delete confirmation, editor with the D-pad set, controller without a back chip, manager in light theme and final settings.

### Haptics

- With **Button haptics** on, a tap on the A button produced a new vibration request immediately:
  - `usage: TOUCH`, `Prebaked=CLICK`, duration approximately 101–103 ms on the emulator HAL.
  - The request was attributed to `io.github.touchbridge.mobile (uid=10235)`.
- With haptics off, the same tap produced no new entry in `dumpsys vibrator_manager`.
- Re-enabling the toggle and tapping produced a new click entry.
- The previous build's 9 ms one-shot was also observable in the service log, but is below the perceptible threshold on many physical motors; 0.5.2 uses the predefined click waveform instead.
- Long-press context menus on the controller surface are suppressed via `contextmenu` prevention.

Haptics cannot be physically felt on a headless emulator. The check proves that an immediate request of the intended type reaches Android and that the toggle suppresses it; it does not prove the feel on a specific phone.

## Not verified

- Physical Android phone testing: motor feel, display cutouts, touch sampling, thermals, real Wi-Fi, Android 6–16 behaviour, biometric/Keystore edge cases.
- Physical Windows and Linux desktops: discovery on real networks, firewall prompts, mouse/keyboard injection, gamepad enumeration, focus handling.
- Real games and anti-cheat behavior.
- Long sessions, reconnect under packet loss, multiple controllers, suspended apps and rotation edge cases.
- End-to-end touch-to-PC latency and jitter.
- Release signing, Play Store packaging or update migration from earlier versions.
- CI results for this repository. The workflows are present, but they had not run from this workspace at the time of writing.

Keep this release labeled alpha until at least one physical Android device and one physical desktop have completed the interactive checklist.
