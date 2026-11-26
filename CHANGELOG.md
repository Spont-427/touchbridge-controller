# Changelog

## 0.9.0 — 2026-11-26 — Android interface redesign

- Redesign the portrait layout library with clearer cards, controller previews, an `In use` marker for the active layout, and distinct Play/Edit actions.
- Rework computer discovery and manual pairing guidance, labels, and saved/nearby computer sections.
- Reorganize Settings into clearer groups with revised labels and descriptions.
- Replace interface text-glyph icons with consistent inline SVGs and align the settings icon containers.
- Start newly opened screens at the top and add subtle screen transitions that respect reduced-motion preferences.
- Recreate the centered phone mark and apply it to the Layouts header and Android launcher icon, including adaptive and density-specific resources.
- Add Material 3-inspired light, dark and AMOLED surfaces, larger touch targets, visible focus styles and safe-area spacing.
- Add the TouchBridge design-system reference and update Android UI documentation.
- Keep the landscape controller and layout editor visuals outside the redesign.
- Bump Android and desktop version to `0.9.0` (`versionCode` 15) and the service-worker shell cache to v14.

## 0.8.0 — 2026-11-25 — in-app updates and controller refinements

- Add an Android in-app update flow that checks the latest stable GitHub release and selects the matching stable or debug APK asset.
- Show update-check status and download progress in Settings → About. Open Android's package installer only after the download completes.
- Verify the downloaded APK's package name, release version, increasing version code and signing-certificate fingerprint before offering installation.
- Require the same stable signing key for future in-app updates; debug builds remain separate tester builds.
- Improve Android JSON layout import for document providers that do not classify `.json` files consistently. Use a wildcard picker with preferred MIME types and retain web-side validation.
- Add a `FileReader` fallback for browsers and WebViews without `Blob.text()`.
- Migrate older `Stick click L` and `Stick click R` layout labels to `L3 click` and `R3 click`, including stored and imported layouts.
- Change adaptive-stick behaviour so the complete stick face moves to the initial touch point and stays there until release; remove the old activation-range setting.
- Keep the active stick above nearby controls while the adaptive stick face moves.
- Shorten long button labels in the editor so newly added controls fit their buttons.
- Add a native Android launcher icon using the existing TouchBridge brand mark, including adaptive and legacy density resources.
- Bump Android and desktop version to `0.8.0` (`versionCode` 14) and the service-worker shell cache to v11.
- Restore executable permissions on the Gradle wrapper and project shell scripts so the documented build and packaging commands work from a clean checkout.

## 0.7.0 — 2026-11-24 — multiplayer controllers and screen-awake support

- Allow up to four authenticated phones to connect to one desktop server. Each receives the next free player slot (`P1`–`P4`); a fifth receives `controller_busy` with the four-player limit, and disconnected slots are reused.
- Add independent virtual gamepad slots for each player: Linux creates up to four uinput devices, Windows creates one ViGEm provider per player, and the mock backend models all four slots.
- Scope gamepad resets to the player's own slot. Add shared ownership tracking for keyboard keys and mouse buttons so one phone cannot release an input another connected phone is still holding.
- Show the player badge and “Connected as Player N of 4” status in the controller UI; include player index and maximum controller count in authentication and capability messages.
- Add a persistent **Keep screen on** setting under Controls & haptics, enabled by default while the controller screen is open.
- Use the Android bridge to set `FLAG_KEEP_SCREEN_ON` while needed and clear it on leaving the controller or destroying the activity.
- Document four-player protocol and lifecycle behaviour, plus the screen-awake setting.
- Bump Android and desktop version to `0.7.0` (`versionCode` 13) and the service-worker shell cache to v10.
- Add regression tests for player-slot assignment/limits/reuse, per-player disconnect cleanup, and shared keyboard/mouse input ownership.

## 0.6.0 — 2026-11-24 — wheel controls, pedals and settings refresh

- Fix digital LT/RT handling by sending both the trigger button and full axis pressure, including for Linux evdev/gamepad backends.
- Show rolling median round-trip time beside the controller connection; expose p95 latency and jitter in the indicator details.
- Add analog accelerator and brake pedals mapped to RT and LT. Pedal pressure increases as the finger moves toward the bottom of the pedal.
- Add a draggable steering-wheel control mapped to `left_x`; the wheel returns to centre when released.
- Add optional gyro steering for layouts containing a wheel, with neutral calibration, filtering, deadzone and direction inversion.
- Replace the theme choices with System, Dark, AMOLED black and Light. Dark now uses a neutral `#0f141b` palette.
- Reorganize settings into Appearance, Controls & haptics, Analog sticks and About, including app/server versions, connection status, repository link and gamepad backend status.
- Include the desktop server version in the successful pairing response so the Android About screen can report the connected build.
- Improve layout placement for large controls so new wheels and pedals avoid existing controls and remain within the controller canvas where space permits.
- Add regression coverage for pedal pressure, wheel mapping, gyro steering and large-control placement.
- Bump Android and desktop version to `0.6.0` (`versionCode` 12) and the service-worker shell cache to v9.

## 0.5.3-alpha — 2026-11-21 — Windows virtual gamepad and Android pairing fix

- Add an optional Windows virtual gamepad backend using ViGEmBus. When the driver is installed, TouchBridge presents an Xbox 360-compatible controller through XInput; mouse and keyboard control is unaffected when it is not.
- Create the virtual controller lazily on the first gamepad input, reset it on disconnect, and unplug it on server shutdown. The server stays unelevated; only the third-party driver installation requires administrator rights.
- Split the Windows gamepad transport behind a provider interface so a future HIDMaestro backend can be added without changing the protocol or mouse/keyboard path.
- Keep the capability response accurate: without the driver, `capabilities.gamepad` is `false` and the client stays in mouse/keyboard mode.
- Add `Setup-TouchBridge-Windows.bat`, a rerunnable one-time setup that downloads the pinned ViGEmBus 1.22.0 installer, verifies its SHA-256, installs it silently and adds private-network firewall rules for TCP 8765 and UDP 8766.
- Fix the Android PIN pairing row on narrow screens: place the Pair device and Disconnect buttons in separate grid columns so the disabled Disconnect button no longer overlaps or intercepts taps on Pair device.
- Print the detected gamepad status in the server startup banner so a missing driver is visible before pairing.
- Document the ViGEmBus 1.22.0 requirement, uninstall path and the retired driver's updater caveat.
- Bump the Android app and desktop server to 0.5.3-alpha (Android `versionCode` 11) and the service-worker shell cache to v8.

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
