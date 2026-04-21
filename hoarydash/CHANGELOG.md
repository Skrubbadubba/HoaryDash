# v0.12.0

### ✨ Features

- **Theme inheritance** - themes can now extend other named themes via a `base` field
- **Typography theme settings** - new `typography` block for font family, weight, style, transform, letter spacing, and a full font size scale
- **Semantic theme variables** - color, shape, and typography fields are now automatically available as `$variable` references without needing an explicit `vars:` declaration
- **Variable resolution across the full config** - `$variable` references are now resolved in card style fields and `custom` CSS throughout the entire YAML structure, not just inside theme definitions
- **Variable multiplier suffix** - append `:multiplier` to any variable reference to scale it; works on both colors (opacity) and unit values (via `calc`)
- **More dateclock customization** - independent control over time and date visibility, font sizes, font weights, and text alignment

### ⚠ Breaking changes

- Theme colors are now under a `colors` key instead of under the theme directly
- `font_size` everywhere is now just `size`. Use the `typography` or `custom` fields of a theme to use your own css units
- Much of the theming system is updated. See the config [guide](https://github.com/Skrubbadubba/HoaryDash/blob/main/docs/CONFIG.md#theming) for a full explanation
### 🏗 Fixes & Improvements

- Font size YAML fields now consistently use `em` units
- Media controller can now play non-Spotifyplus items
- UI more consistently respects `theme.shapes` values
- Dateclock respects `show_time` and `show_date` independently
- Entity overlays (lights, etc) now respect the theme

### ⚙️ Internal

- Introduced `mitchellh/reflectwalk` for recursive CSS field resolution across arbitrary structs
- Theme system now emits a significantly expanded set of utility classes; components bind semantic classes rather than receiving per-component CSS
- Shared flexbox and layout primitives moved to base utility classes, eliminating redundant per-component CSS rules - substantially reduces the rendered CSS bundle size