# TouchBridge 0.9.0

TouchBridge 0.9.0 refreshes the Android app's portrait interface with a clearer layout library, improved computer pairing, reorganized settings, consistent SVG icons, and Material 3-inspired themes. The landscape controller and layout editor visuals remain unchanged.

## Highlights

- Layout cards now include controller previews, a current-layout indicator, and clear Play/Edit actions.
- Computer discovery and manual pairing have clearer instructions and saved/nearby sections.
- Settings are grouped and relabelled for easier navigation.
- Light, dark and AMOLED themes, larger touch targets, visible focus styles and safe-area spacing.
- Screen navigation starts at the top and uses subtle transitions that respect reduced-motion preferences.
- Refreshed centered-phone brand mark and Android launcher icon, including adaptive and density-specific assets.
- Added the TouchBridge design-system reference and updated Android UI documentation.

## Release assets

- **Android stable:** `TouchBridge-0.9.0.apk` — signed with the same private release key used for v0.8.0.
- **Android debug:** `TouchBridge-0.9.0-debug.apk` — debug-signed and attached by GitHub Actions for testing.
- **Desktop:** Windows x64, Linux x86_64 and Linux ARM64 packages are attached by the release workflow.

The stable APK must be signed with the same release certificate as v0.8.0 so existing installations can update in place. Do not publish a debug APK as the stable download.
