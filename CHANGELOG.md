# Changelog

## Unreleased — Windows virtual gamepad

- Add an optional Windows virtual gamepad backend using ViGEmBus. When the driver is installed, TouchBridge presents an Xbox 360-compatible controller through XInput; mouse and keyboard control is unaffected when it is not.
- Create the virtual controller lazily on the first gamepad input, reset it on disconnect, and unplug it on server shutdown. The server stays unelevated; only the third-party driver installation requires administrator rights.
- Split the Windows gamepad transport behind a provider interface so a future HIDMaestro backend can be added without changing the protocol or mouse/keyboard path.
- Keep the capability response accurate: without the driver, `capabilities.gamepad` is `false` and the client stays in mouse/keyboard mode.
- Add `Setup-TouchBridge-Windows.bat`, a rerunnable one-time setup that downloads the pinned ViGEmBus 1.22.0 installer, verifies its SHA-256, installs it silently and adds private-network firewall rules for TCP 8765 and UDP 8766.
- Fix the pairing card on narrow screens: the disabled Disconnect button used to overlap the Pair device button and absorb its taps, so pairing appeared unresponsive. The two buttons now occupy separate grid cells.
- Print the detected gamepad status in the server startup banner so a missing driver is visible before pairing.
- Document the ViGEmBus 1.22.0 requirement, uninstall path and the retired driver's updater caveat.

## 0.5.2-alpha — 2026-11-20 — theming, layout library and editor additions

- Add a System / Light / Dark theme setting, persisted on the device. The lightweight AMOLED theme remains the default; the new light paper theme covers the library, computers, settings, tools, controller and editor surfaces, and the Android system bar icons follow the chosen theme.
- Add Settings → Manage layouts: create named custom layouts, duplicate any installed layout, rename custom layouts, and delete custom layouts after an inline confirmation. Preset layouts can be duplicated but have no delete action.
- Render the library from the installed layout list so custom layouts appear alongside the presets and can be played or edited directly.
- Add D-pad, face-button, bumper and trigger sets to the layout editor's add menu, alongside the existing single-button actions.
- Remove the on-screen controller back chip; the system Back gesture already returns from the controller to the library.
- Make controller button haptics immediate and perceptible by using the platform click waveform instead of a 9 ms one-shot; the existing Button haptics toggle still gates all feedback.
- Replace the editor's text ellipsis with a sliders icon for the layout options panel.
- Enable WebView debugging in debug builds only, for automated UI inspection without affecting release APKs.
- Bump the Android version to 0.5.2-alpha (`versionCode` 10), the desktop server to 0.5.2-alpha, and the service-worker shell cache to v7.

## 0.5.1-alpha — minimal AMOLED controller UI

- Remove promotional hero copy and colourful profile-card treatment in favour of a restrained black AMOLED library screen.
- Keep the home header to the product name and use the bottom tab bar for Layouts, Computers and Settings.
- Keep the play and layout editor on the same full-viewport coordinate canvas; place the connection pill and ping indicator in the top centre so they do not cover corner triggers.
- Move editor tools to a compact floating dock centred above the canvas, with selected-control size and opacity controls in a small contextual panel.
- Clear the selected-control panel when the user taps empty canvas space, restoring an unobstructed editing view.
- Apply haptic feedback to button presses only; joystick movement does not vibrate.
- Retain the bottom navigation for Layouts, Computers and Settings; settings include button haptics and analog/digital LT/RT mode.
- Add regression coverage for digital trigger actions and their Linux evdev button mappings.
- Align the desktop server version, Android version and cache-busting asset version at 0.5.1-alpha.


## 0.5.1-alpha — Android controller UX refinement

- Fixed adaptive-stick origin visibility: the origin dot and ring now render above the knob so the user can see the tapped centre while moving.
- Improved Layout Studio with a selectable control list, grid toggle, 2% snap-to-grid and smart placement of newly added buttons.
- Added native Android document-picker import and save-dialog export for JSON layouts.
- Improved Android local-origin navigation restrictions and reapplies immersive landscape after window focus returns.
- Added a 10-second connection timeout and clearer handling for invalid WebSocket addresses.
- Remembered servers are shown newest-first; decrypted tokens and server IDs are validated before use.
- Updated LAN discovery documentation for stable server IDs and HTTP/HTTPS scheme fields.
- Expanded the automated suite with snap-to-grid and smart-placement tests.

## 0.5.1-alpha — offline-first Android controller

- Bundle the shared controller UI inside the Android APK and load it from a fixed local HTTPS origin. The PC server is no longer required for home, profile selection, editing, stick tuning or controller preview.
- Replace the native Android discovery-only shell with a WebView host that provides offline UI, native LAN scanning, persistent trusted-device storage and orientation control.
- Store trusted-server device tokens encrypted at rest using Android Keystore AES-GCM; keep server metadata in app-private preferences. The server continues to store only token hashes.
- Use trusted device credentials after first pairing so the user does not need to enter a PIN again. A reset/revocation of the server token store requires re-pairing.
- Add landscape immersive-fullscreen controller and Layout Studio screens; restore portrait mode on return to the home/devices/tools screens.
- Keep gamepad controls visually interactive in offline preview mode while preventing any input transmission until an authenticated server advertises a gamepad backend.
- Fix adaptive-stick visual feedback by moving the centre marker and activation ring to the exact touch-down origin, returning all stick visuals to neutral on release, cancellation and focus loss.
- Keep mouse and keyboard tools in a secondary screen; make gamepad profiles the primary Android flow.
- Add strict opacity validation for imported/saved layouts while retaining compatibility with older layouts that omit the opacity field.
- Automatically navigate from the pairing screen to the controller after successful authentication.
- Add a source-sync script and CI check ensuring Android's bundled HTML/CSS/JS is identical to the desktop web UI.
- Update server CSP to allow the dynamic inline style properties used by the controller/editor while continuing to disallow inline scripts.
- Document Android 17 local-network permission requirements before any future target-SDK 37 upgrade.

## 0.3.1-alpha — reliability and input UX patch

- Do not intercept hardware keyboard navigation or activation while a native input, select, textarea, button, link or editable control has focus.
- Always release an already-forwarded hardware key on key-up even if focus moved to a UI control, preventing stuck keys/modifiers.
- Enforce the protocol's 4 KiB limit in the WebSocket frame parser before allocating/reading a large payload; reject unsupported opcodes and non-canonical extended lengths.
- Validate the WebSocket upgrade key as a 16-byte Base64 nonce.
- Bound concurrent upgraded WebSocket connections to prevent unbounded unauthenticated socket retention.
- Prevent the layout editor from exceeding its 40-control import/export schema limit and use collision-resistant per-session custom control IDs.
- Add regression tests for WebSocket framing, connection limits, keyboard target handling and gamepad state validation.
- Add a Go vulnerability scan to CI using the Go team's `govulncheck` tool; this environment could not download the scanner because outbound DNS was unavailable.
- Add `scripts/check.sh` as a non-mutating local quality gate, and short JSON/WebSocket fuzz runs to CI.

## 0.3.0-alpha — controller and stability milestone

- Add a customizable dual-stick gamepad screen with Generic, GTA V, RDR2 and Forza Horizon starter layouts.
- Add adaptive/floating joystick centres for both sticks, fixed-centre mode, configurable activation range, radial deadzone and linear/precision/responsive response curves.
- Add analog pressure to LT/RT trigger controls.
- Add layout import/export JSON with strict schema validation and 64 KiB file-size limit.
- Coalesce gamepad state snapshots under WebSocket congestion so the newest state is kept rather than flooding the connection with stale states.
- Add mouse/key hold reference tracking so multiple pointers or hardware keys cannot release an input still held by another source.
- Release all local pressed state on page blur, backgrounding or disconnect.
- Add rolling RTT median, p95 and jitter display.
- Add gamepad message validation and protocol fuzz seeds.
- Close hijacked WebSocket clients before shutting down input devices, preventing a disconnect/driver-close race.
- Release held gamepad state before profile changes, layout reset/import and entering layout-edit mode.
- Add CodeQL scanning and Dependabot configuration; require review for security-sensitive fixes.
- Keep Windows virtual gamepad disabled until a maintained provider and elevation model are tested on physical Windows hardware. Linux uses the generic uinput gamepad backend when available.
- Add `--version`, update Android version metadata and document gamepad protocol and limitations.

## 0.2.1-alpha — maintenance release

- Reserve the TCP listener before starting input injection or LAN discovery.
- Detect and report an already-running TouchBridge instance via a loopback-only status endpoint. A second launch now reuses the running process and displays its current PIN.
- Report a clear error when an unrelated process occupies the configured TCP port.
- Print a newly generated six-digit PIN every ten minutes; local status probes rotate expired PINs safely.
- Do not announce readiness until the TCP port has been successfully bound.
- Chunk accumulated touchpad movement into protocol-safe ±500 deltas and drain queued movement after temporary WebSocket backpressure.
- Bound the movement backlog to avoid replaying a large amount of stale cursor movement after a long network stall.
- Align Android Java and Kotlin bytecode targets to JVM 17 in the included Android Studio project.
- Add regression tests for startup idempotency, unrelated listeners, local-only status, PIN renewal and expiry.
- Move the single-controller reservation until after authentication so an unauthenticated WebSocket cannot block a legitimate user.
- Add a client heartbeat and idle timeout; abandoned connections release held inputs and free the controller slot after 45 seconds.

## 0.2.0-alpha

- Add UDP LAN discovery and Android server selection.
- Add six-digit PIN pairing with per-IP and global rate limits.
- Add a shared touchpad/keyboard controller UI and Windows/Linux mouse-keyboard input backends.
- Add Linux one-command installation and the initial Android Studio wrapper source project.
