# TouchBridge Android interface design system

The Android app uses a Material 3-inspired design for the portrait library, computer pairing, settings, layout management, and utility screens. The immersive landscape game controller and layout editor keep their established visual language.

## Principles

- Make the next action easy to find: Play on layout cards, Scan network on Computers, and clear categories in Settings.
- Keep touch targets at least 48 dp where practical. Use an 8 dp spacing rhythm and aligned 44 dp icon containers.
- Use a consistent set of hand-authored inline SVG icons with a 24 dp viewBox. Do not use text glyphs or emoji as interface icons.
- Use Roboto/system fonts, clear text hierarchy, rounded tonal surfaces, and visible keyboard focus.
- Preserve screen purpose and existing IDs, handlers, and navigation when revising presentation.
- Respect safe areas, narrow screens, system light/dark appearance, and reduced motion.

## Color roles

Light and dark values are in `tokens.json`; matching CSS variables are in `tokens.css` and `internal/web/web/app.css`.

| Role | Light | Dark |
|---|---|---|
| Primary | `#6750A4` | `#D0BCFF` |
| On primary | `#FFFFFF` | `#381E72` |
| Primary container | `#EADDFF` | `#4F378B` |
| On primary container | `#21005D` | `#EADDFF` |
| Secondary container | `#E8DEF8` | `#4A4458` |
| Background / surface | `#F7F5FA` | `#141218` |
| Surface container | `#EEEBF2` | `#211F26` |
| On surface | `#1D1B20` | `#E6E1E9` |
| On surface variant | `#49454F` | `#CAC4D0` |
| Outline | `#79747E` | `#938F99` |

AMOLED mode uses a black background and surface, with the dark tonal palette for raised containers. Keep text and outlines distinct from those surfaces.

## Layout and components

- Use 4, 8, 12, 16, 24, and 32 dp spacing steps. Separate major content groups by 24 dp.
- Use 12–16 dp corner radii for controls and 18–24 dp for cards and grouped surfaces.
- Use tonal surfaces and restrained outlines instead of layered shadows.
- Primary actions use the primary fill; secondary actions use a tonal container.
- Inputs have a persistent label, a 48 dp minimum height, and a visible focus outline.
- Keep bottom navigation clear of gesture and safe-area insets. Reserve content space beneath fixed navigation.
- Keep paired labels and icons aligned to a shared grid. Use SVG viewBoxes and consistent 2 dp strokes.

## Accessibility and motion

- Body text and controls need readable contrast against their actual theme surface.
- Provide visible focus states, accessible names for icon-only buttons, and non-color state cues such as the current-layout label.
- Preserve browser zoom and platform text scaling where the screen allows it.
- Honor `prefers-reduced-motion`; avoid transitions that shift surrounding content.

## Implementation boundary

The WebView UI is shared from `internal/web/web` and copied to `android-client/app/src/main/assets/controller`. Keep both copies in sync when changing the bundled Android interface. The controller and editor styling is outside the Material 3 browse/settings scope unless a change is explicitly needed for usability.
