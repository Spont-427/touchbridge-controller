# Controller UI and profiles (0.5.2-alpha)

The Android application bundles the shared controller UI locally under `app/src/main/assets/controller/`. The library, profiles, layout editor, stick settings and controller preview work without a desktop server. A server is needed only to transmit controller, keyboard or mouse state to a PC.

The same UI is served by the desktop server for browser preview, with Android-specific bridges for orientation, haptics, native file dialogs and secure token storage.

## Screens and navigation

- **Home / Layouts:** portrait library of installed layouts with **Play** and **Edit** actions, plus bottom navigation to **Computers** and **Settings**.
- **Controller:** dedicated landscape, immersive-fullscreen play surface.
- **Layout Studio:** separate landscape, immersive-fullscreen editor. **Done** returns to the screen that opened it (Home or Manage layouts).
- **Computers:** discovery, manual address entry, pairing and saved servers.
- **Settings:** appearance, haptics, trigger mode, sticks and layout management.
- **Layouts manager:** portrait list of every installed layout with management actions.
- **Tools:** optional mouse and keyboard screen, secondary to the gamepad flow.

Android system Back is handled inside the UI:

- Editor → previous screen (Home or Manage layouts).
- Controller → Home.
- Manage layouts → Settings.
- Computers / Tools / Settings → Home.
- Home → exits the app.

The controller no longer has an on-screen back chip. The system Back gesture is the supported way to leave the play surface, which keeps the controller free of chrome that can overlap controls.

## Library

The home library is rendered from the installed layout registry instead of a fixed card list:

- Presets: **Xbox-style**, **GTA V**, **Red Dead Redemption 2** and **Forza Horizon**.
- Custom layouts, appended in creation order.
- The header count reads `4 PRESETS` when no custom layouts exist, or `4 PRESETS · 2 CUSTOM` otherwise.
- Presets keep their themed artwork; custom layouts use the generic controller art.
- Every card can be played or edited. Renaming or deleting a layout updates the library immediately.

## Layout management

Open **Settings → Manage layouts**. This screen is available offline and stores everything in local WebView storage.

### Create

- **New layout** reveals a name field. The name is trimmed, repeated whitespace is collapsed, and the result is limited to 30 characters.
- Names must be non-empty and unique across presets and custom layouts, ignoring case.
- A new layout is saved as a fresh copy of the Xbox-style base layout, not as an empty canvas.
- The layout appears on the home screen and in the controller's profile selector immediately.

### Duplicate

- Available for presets and custom layouts.
- Copies the source layout's current controls rather than resetting to the preset base.
- The generated name starts at `Name copy`, then `Name copy 2`, `Name copy 3`, and so on, truncated to 30 characters.

### Rename

- Custom layouts only. The row switches to an inline text field; **Save** validates the same rules as creation, **Cancel** leaves the name unchanged.
- Renaming the active layout updates the controller and editor titles and the profile selector.

### Delete

- Custom layouts only, and only after a second inline confirmation: **Delete “Name”?** with **Cancel** and **Delete** actions.
- Deleting removes both the registry entry and the stored controls for that layout. There is no undo.
- Deleting the active layout selects Xbox-style, updates the last-used layout and re-renders the controller.
- Preset rows render only **Edit** and **Duplicate**. The delete handler independently rejects preset IDs, so protection does not depend on the UI alone.

Rows display the layout name, `Preset` or `Custom`, the current number of controls, and `Active` for the selected layout. The header reports the total installed layout count.

## Layout Studio

The editor runs independently of a server and saves each layout separately.

- **Canvas:** controls are positioned as percentages, so a layout scales to the controller surface. Drag a control to reposition it.
- **Selection:** tap a control or an entry in the selectable control list. Tapping empty canvas clears the selection and hides the compact property panel.
- **Properties:** label, gamepad action, size in pixels and opacity. Sticks cannot be deleted; all other controls can.
- **Grid and snap:** optional grid overlay and 2% snap-to-grid, persisted locally.
- **Smart placement:** new controls are placed in the clearest available region rather than stacked at the centre. Button sets reserve space for every member.
- **Reset:** restores the selected starter profile.
- **Import/export:** JSON layout files through the browser picker/download flow, or Android's system document picker and save dialog.
- **Limit:** the editor supports up to 40 controls per layout.

### Add menu

The add-control dropdown is grouped into two optgroups:

**Single buttons** — every supported gamepad action as an individual button.

**Button sets** — one menu choice adds multiple positioned controls:

| Set | Controls |
| --- | --- |
| D-pad (4 buttons) | `DPAD_UP`, `DPAD_DOWN`, `DPAD_LEFT`, `DPAD_RIGHT` arranged as a cross |
| Face buttons (4) | `A`, `B`, `X`, `Y` arranged as a diamond |
| Bumpers (LB / RB) | `LB` and `RB` |
| Triggers (LT / RT) | `LT` and `RT` |

Sets are placed as a unit around the next free position and can be moved afterward; there is no group-drag behavior after placement. Any set that would exceed the 40-control cap is rejected as a whole.

### Import validation

The importer accepts either a TouchBridge layout v1 document or a raw control array. It rejects files over 64 KiB and validates:

- Control count and unique IDs.
- Supported actions and control kinds.
- Finite, in-range positions, sizes and opacity.
- Exactly one left stick and one right stick.

Invalid imports leave the existing layout untouched.

## Play screen

The controller fills the viewport in landscape. A compact connection pill and latency indicator sit at the top centre so they do not cover the corner triggers. Browser chrome, the app's bottom navigation and the editor tools are hidden.

- Controls remain tactile in offline preview mode: sticks move and buttons highlight locally.
- Network state is sent only when the connection is authenticated and the selected server advertises a gamepad backend.
- The system Back gesture returns to the library and restores portrait mode.
- Pointer capture keeps a button held while the finger moves slightly outside it; release, cancellation and focus loss clear the state.
- Sticks return to neutral on release, pointer cancellation, focus loss and connection loss.

## Haptics

- Feedback fires on **button pointer-down**, not on release and not when dragging sticks.
- The Android bridge uses `VibrationEffect.createPredefined(EFFECT_CLICK)` on Android 10+ with touch usage on Android 13+, a 22 ms one-shot on Android 8–9, and a legacy 22 ms vibration below that.
- The short 9 ms vibration used in earlier builds was below the perceptible threshold on many phone motors, which made it appear that haptics only arrived after WebView's long-press feedback.
- **Settings → Button haptics** is checked at press time. When it is off, no vibration request reaches Android.
- Long-press context menus on the controller surface are suppressed so a held control does not add delayed native feedback.
- Haptics are optional and failures in the bridge never block input.

## Sticks

Both sticks support:

- **Adaptive/floating centre:** a touch inside the activation region becomes the temporary origin. The visible centre dot and activation ring move to that origin and render above the moving knob; axis calculations use the same origin. Touches outside the region are clamped to its edge.
- **Fixed centre:** the origin remains the visual centre of the stick.
- **Radial deadzone:** 0–25%.
- **Response curves:** linear, precision and responsive.
- **Activation range:** default 48% of the stick radius, applying to both sticks.

The origin and knob reset on release, pointer cancellation, window focus loss and connection loss.

## Settings

- **Theme:** `System`, `Light` or `Dark`, persisted on the device. System follows the OS colour scheme; the native status and navigation bar icons are updated to match. The dark theme is an AMOLED-oriented near-black palette; the light theme uses a paper surface.
- **Button haptics:** all-or-nothing gate for controller button feedback.
- **LT/RT mode:** analog pressure or digital button. Digital triggers are sent in the button set; analog triggers report `[0, 1]` pressure.
- **Stick settings:** centre mode, activation range, deadzone and curve.
- **Manage layouts:** opens the layout manager described above.

## Storage

Layout data is local to the app's WebView origin. The keys used by 0.5.2-alpha are:

| Key | Purpose |
| --- | --- |
| `touchbridge.profiles` | Custom layout registry: JSON array of `{ id, name }`, IDs matching `custom-[a-z0-9-]{1,56}`. |
| `touchbridge.layout.<id>` | Stored controls for a preset or custom layout. |
| `touchbridge.lastProfile` | Last selected layout ID. |
| `touchbridge.theme` | `system`, `light` or `dark`. |
| `touchbridge.haptics` | Button haptics on/off. |
| `touchbridge.triggerMode` | `analog` or `digital`. |
| `touchbridge.stick.centerMode` | `adaptive` or `fixed`. |
| `touchbridge.stick.centerRange` | Activation range percentage. |
| `touchbridge.stick.deadzone` | Radial deadzone percentage. |
| `touchbridge.stick.curve` | `linear`, `precision` or `responsive`. |
| `touchbridge.editor.grid` | Grid visibility in the editor. |
| `touchbridge.editor.snap` | 2% snap-to-grid setting. |
| `touchbridge.mouse.sensitivity` | Mouse tool sensitivity. |
| `touchbridge.savedServers` | Paired-server metadata. Device tokens are stored through the native secure store, not in this key. |

Clearing Android app storage removes layouts, preferences, remembered servers and trusted tokens. Presets return the next time the app starts.

## Starter profiles

Generic/Xbox-style, GTA V, Red Dead Redemption 2 and Forza Horizon are starter layouts. They adjust placement and size to broadly suit each game's interaction style. They are **not** guaranteed copies of official in-game button maps: controls vary by platform, game version and user configuration.

## Device pairing

The Android app remembers paired servers. The first connection uses a six-digit PIN and requests a device token. The desktop stores only a hash; Android stores the token encrypted with Android Keystore AES-GCM. Subsequent connections send the token without asking for the PIN. The user can forget a saved server from the device list. Resetting the server's trust database revokes all previously issued tokens.

## Known UI limitations

- The starter layouts are editable starting points, not game macros or game-detection integrations.
- Layouts are stored per install; there is no cloud sync or cross-device transfer beyond JSON export/import.
- Linux gamepad creation requires `/dev/uinput` and permission. The exposed device is a generic evdev gamepad.
- Windows virtual gamepad emulation is not enabled; mouse and keyboard are available.
- RTT measures network round trip, not one-way input-to-photon latency.
- The app targets API 35. Android 17's `ACCESS_LOCAL_NETWORK` runtime permission is not implemented yet.

## Android local-network permission note

The app targets API 35. Android documents `ACCESS_LOCAL_NETWORK` as a runtime permission for apps targeting API 37 or later; this release does not target API 37, so it retains the current `INTERNET`-permission LAN behaviour. Before raising the target SDK, implement and test the permission request and denial/recovery flow. See the [Android local network permission guide](https://developer.android.com/privacy-and-security/local-network-permission).
