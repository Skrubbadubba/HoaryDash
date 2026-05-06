# HoaryDash Configuration Reference

All configuration lives in a single file: `/addon_configs/hoarydash/hoarydash.yaml`. This is the heart of the entire operation.

The file has two top-level sections: `dashboards` (a map of dashboard names to dashboard configs) and a set of global keys (`localization`, `fully_kiosk`, `home_assistant`).

---

## Top-level structure

```yaml
dashboards:
  my-dashboard:        # becomes available at /my-dashboard/
    screens: [...]
    theme: {...}
    navbar: {...}

  screensaver:         # available at /screensaver/
    screens: [...]

localization:
  locale: "en-US"
  timezone: "America/New_York"
  hour12: false

fully_kiosk:
  screensaver_timeout: 60
```

Each key under `dashboards` becomes an endpoint. `dashboards.dash` is served at `/dash/`, `dashboards.living-room` at `/living-room/`, and so on.

### Home Assistant connection

HoaryDash installed as an addon/app uses the supervisor API, this means no further configuration is needed. If you are running it as a docker container, you must provide details to connect to HA.

```yaml
home_assistant:
  url: "http://homeassistant.local:8123" # Defaults to this
  token: "your_long_lived_token_here"
```

The token can also be provided as an environment variable `HA_TOKEN`.

---

## Global config

### `localization`

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `locale` | string | `"en-US"` | BCP 47 locale tag used for date/time formatting |
| `timezone` | string | `"UTC"` | IANA timezone. Must be `Area/Location` format — bare aliases like `Japan` do not work on Chromium 44 |
| `hour12` | bool | `false` | Use 12-hour clock format |

### `fully_kiosk`

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `screensaver_timeout` | int | `60` | Seconds of inactivity before Fully Kiosk activates the screensaver |

### `wallpanel`

`bool`

> Note `fully_kiosk` takes precedence. If you want to use wallpanel, make sure to omit `fully_kiosk` entierely.

### `home_assistant`

| Key | Type | Description |
|-----|------|-------------|
| `url` | string | Full URL to your Home Assistant instance, e.g. `http://192.168.1.100:8123` |
| `token` | string | Long-lived access token from your HA profile page |

---

## Dashboard config

Each dashboard supports the following fields.

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `animations` | bool | `true` | Enable CSS transition animations |
| `screenonlock` | bool | `true` | Show button to toggle Fully Kiosks screensaver |
| `swipe` | bool | `true` | Enable horizontal swipe to navigate between screens |
| `show_hints` | bool | `true` | Show chevron hints at screen edges pointing toward adjacent screens |
| `screens` | list | — | List of screens (see [Screens](#screens)) |
| `navbar` | object | — | Navigation bar config (see [Navbar](#navbar)) |
| `nightlight` | object | — | Nightlight overlay config (see [Nightlight](#nightlight)) |
| `theme` | object | — | Visual theme (see [Theming](#theming)) |

### Navbar

```yaml
navbar:
  enabled: true
  position: bottom   # top, bottom, left, right
  style: default     # default, rectangle
```

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `enabled` | bool | `false` | Show the navigation bar |
| `position` | string | `bottom` | Which edge the navbar sits on. One of `top`, `bottom`, `left`, `right` |
| `style` | string | `default` | Visual style. `default` is a floating pill; `rectangle` is flush with the screen edge |

### Nightlight

An orange-tinted overlay for use as a light. Brightness is adjustable by dragging on the overlay.

```yaml
nightlight:
  enabled: true
  override_colors: true
  color: "hsl(22, 100%, 55%)"
```

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `enabled` | bool | `false` | Enable the nightlight overlay |
| `override_colors` | bool | `false` | Override nightlights own fire style colors to match theme |
| `color` | CSS color | red | Color of the overlay. Any valid CSS color value |

---

### Screenonlock

A button to toggle the screensaver on/off. Either wallpanel or fully kiosk needs to be configured for this to work.

## Screens

Each dashboard has a `screens` list. Every screen has its own layout, entities, and widgets.

```yaml
screens:
  - name: Main
    icon: home
    layout: centered
    dateclock:
      enabled: true
    entities: [...]
    widgets: [...]
    sensors: [...]
    order:
      entities: 1
      widgets: 2
      sensors: 3

  - name: Tiles
    layout: tiled
    groups: [...]
```

### Common fields

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `name` | string | — | Display name shown in the navbar |
| `icon` | string | — | Icon name shown in the navbar (MDI icon slug, e.g. `home`, `thermometer`) |
| `layout` | string | `centered` | Layout preset. `centered` or `tiled` |
| `dateclock` | object | — | Clock/date display (see [Dateclock](#dateclock)) |

### Dateclock

```yaml
dateclock:
  enabled: true
  show_time: true
  show_date: true
  show_seconds: true
  capitalise_day: true
  hour12: false
  align: center
  font_size: 1.2
  time_size: 1.0
  time_weight: 600
  date_size: 0.5
  date_weight: 600
```

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `enabled` | bool | `true` | Show the dateclock |
| `show_time` | bool | `true` | Show the time |
| `show_date` | bool | `true` | Show the date and day name |
| `show_seconds` | bool | `false` | Show seconds in the time display |
| `capitalise_day` | bool | `false` | Capitalise the first letter of the day name |
| `hour12` | bool | `false` | Use 12-hour time format |
| `align` | string | `center` | Text alignment — `left`, `center`, or `right` |
| `font_size` | number | — | Base font size for the entire dateclock, in `em` |
| `time_size` | number | — | Font size of the time, in `em`, relative to `font_size` |
| `time_weight` | number | `600` | Font weight of the time |
| `date_size` | number | `0.5` | Font size of the date, in `em`, relative to `font_size` |
| `date_weight` | number | `600` | Font weight of the date |

---

## Cards

A **card** is the basic building block of a HoaryDash screen. Every HA-connected component (entity controls, sensor readouts, weather widgets, media players, to-do lists) is a card. They all share the same base styling (border, background, corner radius) and the same configuration schema.

Cards are always defined inside a *card group*. A card group is a list of cards with an optional shared style. In the centered layout the three zones (`entities`, `sensors`, `widgets`) are card groups. In the tiled layout each named `group` is a card group (with some additional fields).

### Card group schema

```yaml
# Centered layout — each zone is a card group
entities:
  style:
    background: "rgba(255,255,255,0.06)"
  cards:
    - entity_id: light.ceiling
      label: Ceiling
      icon: lightbulb
    - entity_id: button.doorbell
      label: Ring doorbell
      icon: doorbell

sensors:
  cards:
    - entity_id: sensor.living_room_temperature
      label: Temperature
      options:
        unit: °C

widgets:
  cards:
    - entity_id: media_player.spotify
    - entity_id: weather.home
      options:
        forecast_interval: twice_daily
```

```yaml
# Tiled layout — each group is a card group
groups:
  - name: Controls
    style:
      background: "rgba(255,255,255,0.06)"
    cards:
      - entity_id: light.ceiling
        label: Ceiling
        icon: lightbulb
  - name: Climate
    cards:
      - entity_id: sensor.living_room_temperature
        label: Temperature
        options:
          unit: °C
```

| Key | Type | Description |
|-----|------|-------------|
| `cards` | list\[[`card`](#card-fields)\] | The cards in this group |
| `style` | [`card style`](#card-style-fields) | Style overrides applied to every card in this group |

### Card fields

The `entity_id` domain determines what control is rendered. All other fields are optional.

```yaml
- entity_id: light.ceiling
  label: Ceiling light # Optional
  icon: lightbulb # Optional
  style:
    border_radius: "1em"
    background: "rgba(255,255,255,0.12)"
  options:
    show_pill: true
```

| Key | Type | Description |
|-----|------|-------------|
| `entity_id` | string | Home Assistant entity ID. Determines the card type — see tables below |
| `label` | string | Display label |
| `icon` | string | MDI icon name or emoji, e.g. `lightbulb`, `🔔` |
| `options` | object | Domain specific domain options, for example unit values for sensors e.g. `°C`, `%`, `µg/m³`. See [Domain-specific fields](#domain-specific-fields) |
| `style` | [card style](#card-style-fields) | Style overrides for this individual card |

Note that both label and icon is prepopulated. Label is either stylized from the entity id, or from friendly_name from HA if available. Icon is populated based on HAs conventions based on domain + device class.

### Card types by domain

The `entity_id` domain determines what is rendered. Domains not listed below render a plain button.

| Domain | Card type | Category |
|--------|-----------|----------|
| `switch`, `input_boolean` | Toggle | Entity |
| `light`, `fan` | Toggle with adjustable overlay (brightness / speed / CCT) | Entity |
| `button`, `input_button`, `scene`, `script` | Button | Entity |
| `sensor`, `binary_sensor` | Sensor readout | Sensor |
| `weather` | Weather widget — current conditions and forecast | Widget |
| `media_player` | Media player — album art, controls, volume, browser | Widget |
| `todo` | To-do widget — add, check, filter | Widget |

As you can see, each card falls into a category, which is just another arbitrary way of grouping them I decided upon. Each category can be styled on a per dashboard/per screen basis, see [theming](#theming)

### Domain-specific fields

Some card types accept additional fields under an `options` key.

**Sensor**

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `unit` | string | fetched from HA | Unit to be displayed alongside the value |

**Weather**

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `forecast_interval` | string | `daily` | Forecast granularity: `daily`, `twice_daily`, or `hourly` |
| `forecast_times` | int | `5` | Number of forecast periods to display |
| `hour12` | bool | `false` | Show forecast times in 12-hour format |

**Media player**

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `show_volume` | bool | `true` | Show the volume slider |
| `show_album` | bool | `true` | Show album name below artist |
| `show_browser` | bool | `true` | Show the media browser button |
| `spotifyplus` | bool | `false` | Assume media player is from the the [spotifyplus integration](https://github.com/thlucas1/homeassistantcomponent_spotifyplus/tree/master). Provides extra functionality |
| `queue` | bool | `true`| Whether or not to dhow the queue button for items in the browser. Only applicable if showing browser |

#### Common for widgets

Widgets are a category of richer cards for some comains. Currently including __media_player__, __weather__, and __todo__.

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `internal_borders` | bool | `true` | Show the divider between info and controls |

#### Common for tiles

Tiles is a common UI element used for many simple domains that are clickable, toggleable, and may have adjustable attributes. Currently including __switch__, __input_boolean__, __light__, and __fan__. _All_ other domains are rendered as a button tile, altough only domains __script__, __button__, __scene__, and __input_button__ are confirmed to work. 

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `show_icon` | bool | `true` | Show the entity icon, determined either from yaml or HA attributes |
| `show_pill` | bool | | `false` | Show the toggle pill. Only applicable to toggleable entities |´

Additionally, these tile options can be set as global defaults at either the dashboard, screen or card-group level. They override eachother cascadingly, with a lower level taking precedence. Example:

```yaml
dashboards:
  dash:
    tile_options:
      show_icon: true
    screens:
    - name: Home
      tile_options:
        show_icon: false
              entities:
        cards:
          - entity_id: switch.zibgee_square_plug_2 # will not show pill
          - entity_id: fan.starkvind_luftrenare # will show pill
            options:
              show_pill: true
    - name: Controls
      layout: tiled
      tile_options:
        show_pill: true
      groups:
        - name: Lampor
          tile_options:
            show_pill: false
          icon: lightbulb
          cards:
            - entity_id: switch.zibgee_square_plug_2 # will show pill
              label: Hall-lampa
              options:
                show_pill: true
              # rest of entities in group will not show pill
            # all entities in other groups will show pill
```

### Card style fields

A `style` block can be placed on an individual card or on a card group. Individual card style takes precedence. See [Theming](#theming) for more information.

| Key | Type | Description |
|-----|------|-------------|
| `borders` | bool | Show card border |
| `border_color` | CSS color | Border colour |
| `border_radius` | CSS value | Corner rounding, e.g. `0.75em` |
| `background` | CSS value | Card background |
| `font_size` | CSS value | Font size for this card or group |

---

## Tiled layout

The tiled layout organises cards into named groups. Groups stack vertically in a scrollable container. Within each group, cards sit in a wrapping flex row.

```yaml
screens:
  - name: Tiles
    layout: tiled
    dateclock:
      enabled: true
    groups:
      - name: Climate
        icon: thermometer
        cards:
          - entity_id: sensor.living_room_temperature
            label: Temperature
            options:
              unit: °C
          - entity_id: light.ceiling
            label: Ceiling
        style:
          border_color: '#eb5c14'
      - name: Media
        cards:
          - entity_id: media_player.spotify
```

### Additional fields

In addition to the common screen fields, the tiled layout has the followwing:

| Key | Type | Description |
|-----|------|-------------|
| `groups` | list | Groups of cards |
| `stretch` | bool | Whether to align the items by stretching their widths to fill the screen (default _false_) |

Group object:

| Key | Type | Description |
|-----|------|-------------|
| `name` | string | Group header label |
| `icon` | string | Icon shown next to the group header |
| `cards` | list | List of cards (same fields as entities, sensors, and widgets) |
| `stretch` | bool | Whether to align the items by stretching their widths to fill the screen |

The `stretch` option is prioritised per group -- screen level option serves as fallback.

---

## Centered layout

The centered layout is the first developed and default one. It displays everything in a centered column. At the top we have basic entity controls for lights, toggles and buttons. In the middle we have the dateclock and widgets. At the bottom we have sensors. As of v0.10.0 each zone can technically render any type of card, so these are just recomendations at this point. Note though that I have no idea how well aligned everything will look if you put another card type in a zone it was not intended for.

```yaml
screens:
  - name: Main
    icon: home
    layout: centered
    # Default order
    order:
      entities: 1
      widgets: 2
      sensors: 3
    dateclock:
      enabled: true
      show_seconds: true
      capitalise_day: true
    entities:
      cards:
        - entity_id: light.ceiling
          label: Ceiling
          icon: lightbulb
        - entity_id: switch.fan
          label: Fan
    widgets:
      cards:
        - entity_id: weather.home
          options:
            forecast_interval: hourly
            forecast_times: 5
        - entity_id: media_player.spotify
          options:
            show_album: true
    sensors:
      cards:
        - entity_id: sensor.living_room_temperature
          label: Temperature
          options:
            unit: °C
    theme:
      widgets:
        borders: false
        background: "rgba(0, 0, 0, 0)"
      sensors:
        borders: false
        background: "rgba(0, 0, 0, 0)"
```

### Additional fields

In addition to the common screen fields, the tiled layout has the followwing:

| Key | Type | Description |
|-----|------|-------------|
| `entities` | list\[[`CardGroup`](#card-group-schema)\] | Control buttons/toggles |
| `sensors` | list\[[`CardGroup`](#card-group-schema)\] | Sensor readouts |
| `widgets` | list\[[`CardGroup`](#card-group-schema)\] | Widget cards (weather, media, todo) |
| `order` | object | CSS flex order for the three zones |

#### Order object:

The three zones of the layout — entities, dateclock/widgets, and sensors — are flex children. Their order is set via CSS `order`:

| Key | Type | Description |
|-----|------|-------------|
| `entities` | int | Flexbox order |
| `sensors` | int | Flexbox order |
| `widgets` | int | Flexbox order |

Lower numbers appear first (top of the screen). Values can be any integer. For example, to just put sensors at the top:

```yaml
order:
   sensors: -1
```

---

## Fullscreen media layout

The fullscreen media layout turns the entire screen into a media player view. Album art is blurred and stretched behind a dark scrim to form the backdrop. Transport controls, song info, and cover art sit in the foreground. A row of badge pills in the top-right corner can show sensor readings alongside the source selector and music browser.

```yaml
screens:
  - name: Now Playing
    layout: fullscreen-media
    entity_id: media_player.spotify
    media_options:
      show_browser: true
    dateclock:
      enabled: true
    badges:
      badge:
        label: Living room
        icon: sofa
      sensors:
        - entity_id: sensor.living_room_temperature
          label: Temp
          unit: °C
          icon: thermometer
        - entity_id: sensor.living_room_humidity
          label: Hum
          unit: "%"
```

It is recommended to apply a dark theme to this layout, as the text might not read well against the backdrop otherwise.

### Additional fields

In addition to the common screen fields, the fullscreen media layout has the following:

| Key | Type | Description |
|-----|------|-------------|
| `entity_id` | string | The `media_player.*` entity to control |
| `media_options` | object | Media related options, same as the media widget. See [widget options](#domain-specific-fields) |
| `rotate` | bool | Whether or not to slowly rotate the backdrop (default: _true_) |

#### Badges object

| Key | Type | Description |
|-----|------|-------------|
| `badge` | object | A static label pill. Useful for a room or media player name |
| `badge.label` | string | Text shown in the pill |
| `badge.icon` | string | Icon shown in the pill |
| `sensors` | list | Live sensor readings shown as pills |

Sensor fields follow the same schema as sensor cards elsewhere — `entity_id`, `label`, `options.unit`, `icon`.

---

## Theming

HoaryDash has a two-level theme system. The **dashboard theme** is fully merged with built-in defaults, so every color and shape value is always defined. **Screen themes** are partial — only the fields you set are applied, and they layer on top of the dashboard theme via CSS cascade.

Themes can be defined inline or by referencing a named preset.

### Named themes

Named themes can come from two sources:
- **Bundled presets** — defined in `themes.yaml`. Available presets: `light`, `gruvbox-dark`, `gruvbox-light`, `sepia`, `nord`, `aurora`, `sunset`, `forest`, `synthwave`
- **Custom themes** — defined under a `themes:` block in your config. These take precedence over bundled presets if names collide.

```yaml
themes:
  my-theme:
    background:
      color: "#1a1a2e"
    colors:
      accent: "#ff6b6b"

dashboards:
  main:
    theme: my-theme
    screens:
      - name: Home
        theme: aurora        # reference a bundled preset
```

### Inline themes

Instead of a name, you can define the theme directly under the `theme:` key:

```yaml
dashboards:
  main:
    theme:
      background:
        color: "#0f0f0f"
      colors:
        accent: "hsl(210, 90%, 65%)"
    screens:
      - name: Home
        theme:
          colors:
            accent: "#ff7043"   # only overrides accent; everything else cascades from dashboard
```

### Theme inheritance

Named themes can extend other named themes via a `base` field. The base is fully resolved first, then your overrides are merged on top. Inheritance is recursive and cycles are not allowed.

```yaml
themes:
  my-theme:
    base: aurora
    colors:
      accent: "#ff6b6b"
```

### Variable references

Themes support `$variable` references within their own definition, resolved before the theme is applied. Variables can be used anywhere a CSS value appears — including inside gradients and raw `custom` CSS.

Two sources of variables are available:

- **Semantic vars** — every field in `colors`, `shapes`, and `typography` is automatically available by its YAML key name: `$accent`, `$medium_border_radius`, `$font_family`, `$font_xl`, etc.
- **Explicit vars** — defined under `vars:` in the theme. These take precedence over semantic vars on collision.

An optional `:mutlipier` suffix applies opacity to the resolved color, and scales units:

```yaml
themes:
  my-theme:
    vars:
      brand: "#d79921"
      dark: "#1a1a2e"
    colors:
      accent: "$brand"
      interactive: "$brand"
      surface: "$brand:0.1"
    background:
      color: "linear-gradient(180deg, $dark:0.8, #0f0f0f)"
```

Variable references also work in card-type override fields (including `custom`) anywhere in the structural config, resolved against the effective theme for that screen.

All themes, including partially defined ones, always inherit the variables from the above layered theme (screen -> dashboard -> default). That is to say, vars such as `$text_xl` will always be resolved.

### Dashboard vs. screen themes

| Level | Merged with defaults? | Effect |
|---|---|---|
| Dashboard | Yes — all fields guaranteed populated | Sets the full visual baseline for the page |
| Screen | No — partial, only set fields emitted | Overrides specific values; CSS cascade fills the rest from dashboard |

This means a screen theme only needs to contain what it actually changes. Setting just `colors.accent` on a screen is valid and safe.

### Theme fields

#### Top-level

| Key | Type | Description |
|-----|------|-------------|
| `background` | background layer | Page background layer — see below |
| `background_overlay` | background layer | Optional overlay layer composited on top of `background` |
| `is_light` | bool | Hints that this is a light theme; adjusts some derived values |
| `size` | number | Global size multiplier applied to the entire dashboard |
| `base` | string | Name of a theme to inherit from |
| `vars` | map | Named variables for use with `$name` references |
| `custom` | CSS | Raw CSS injected into the dashboard or screen scope |

#### Background layers

Both `background` and `background_overlay` accept the same fields:

| Key | Type | Description |
|-----|------|-------------|
| `color` | CSS value | Background color or gradient |
| `image` | CSS value | Background image, e.g. `url(...)` |
| `size` | CSS value | `background-size` |
| `position` | CSS value | `background-position` |
| `repeat` | CSS value | `background-repeat` |
| `blend_mode` | CSS value | `background-blend-mode` |
| `filter` | CSS value | CSS filter applied to the layer |
| `opacity` | number | Layer opacity (0–1) |

#### Colors

All color fields live under the `colors:` key.

| Key | Type | Description |
|-----|------|-------------|
| `surface` | CSS color | Default card/panel background (semi-transparent) |
| `surface_opaque` | CSS color | Opaque surface, used for modals and overlays |
| `surface_prominent` | CSS color | Elevated surface, used for buttons |
| `surface_subtle` | CSS color | Faint surface, used for badges, should be semi-transparent |
| `surface_alt` | CSS color | Alternative surface color I couldn't find a name for. Used as the navbar background currently. |
| `highlight` | CSS color | Hover/focus highlight layer |
| `border` | CSS color | Theme-wide border color |
| `on_surface` | CSS color | Primary text and icon color |
| `on_surface_muted` | CSS color | Secondary text and icons |
| `on_surface_subtle` | CSS color | Tertiary text, timestamps |
| `on_background` | CSS color | Text/icons directly on the page background |
| `accent` | CSS color | Active indicators, highlights |
| `accent_muted` | CSS color | Dimmed accent (derived automatically if omitted) |
| `interactive` | CSS color | Button icons, slider thumbs, controls |
| `interactive_muted` | CSS color | Dimmed interactive (derived automatically if omitted) |
| `interactive_disabled` | CSS color | Disabled controls (derived automatically if omitted) |
| `state_on` | CSS color | Icon/text color when an entity is active |
| `state_off` | CSS color | Icon/text color when an entity is inactive |
| `state_disabled` | CSS color | Icon/text color for unavailable entities |
| `positive` | CSS color | Success/good state indicators |
| `negative` | CSS color | Error/bad state indicators |

Several colors are derived automatically when omitted: `accent_muted`, `interactive_muted`, `interactive_disabled`, `on_surface_muted`, `on_surface_subtle`, `on_background`, `surface_prominent`, and `surface_subtle` are all computed from their base color. Muted variants of state and semantic colors used for toggle track backgrounds are always computed and cannot be overridden.

#### Shapes

All shape fields live under the `shapes:` key.

| Key | Type | Description |
|-----|------|-------------|
| `borders` | bool | Show borders on all card-like elements |
| `tight_border_radius` | CSS value | Corner radius for compact elements, e.g. badge buttons |
| `medium_border_radius` | CSS value | Corner radius for cards and buttons |
| `wide_border_radius` | CSS value | Corner radius for modals and badges |
| `border_thick` | CSS value | Border width for cards, modals, tooltips |
| `border_thin` | CSS value | Border width for badges and badge buttons (derived as half of `border_thick` if omitted) |
| `gap_inner` | CSS value | Inner gap between elements (derived as half of `gap_outer` if omitted) |
| `gap_outer` | CSS value | Outer gap between elements |
| `padding_inner` | CSS value | Padding for certain elements |

#### Typography

All typography fields live under the `typography:` key.

| Key | Type | Description |
|-----|------|-------------|
| `font_family` | CSS value | Font family stack |
| `font_weight` | CSS value | Default font weight |
| `font_style` | CSS value | Default font style |
| `text_transform` | CSS value | Default text transform |
| `letter_spacing` | CSS value | Default letter spacing |
| `font_xxs` | CSS value | Extra-extra-small font size |
| `font_xs` | CSS value | Extra-small font size |
| `font_sm` | CSS value | Small font size |
| `font_md` | CSS value | Base/medium font size |
| `font_lg` | CSS value | Large font size |
| `font_xl` | CSS value | Extra-large font size |
| `font_xxl` | CSS value | Extra-extra-large font size |
| `font_hero` | CSS value | Hero/display font size |

#### Card-type overrides

Each of these keys accepts a **card style** object and overrides the defaults for that specific type of card. All fields are optional.

| Key | Applies to |
|-----|------------|
| `cards` | All cards (base default) |
| `entities` | Entity control cards only |
| `sensors` | Sensor readout cards only |
| `widgets` | Widget cards only |
| `modals` | Modal overlays |
| `badges` | Badge labels |
| `badge_buttons` | Badge buttons |
| `buttons` | Control buttons |
| `tooltips` | Tooltip popups |

**Card style fields:**

| Key | Type | Description |
|-----|------|-------------|
| `borders` | bool | Show or hide borders for this card type |
| `border_radius` | CSS value | Corner rounding |
| `border_width` | CSS value | Border thickness |
| `border_color` | CSS color | Border color (overrides the theme-wide border color) |
| `background` | CSS color | Background color |
| `padding` | CSS value | Inner padding |
| `size` | number | Size multiplier for this card type |
| `custom` | CSS | Raw CSS injected into this card type's rule block; `$variable` references are resolved |

### Custom CSS

As evident from the fields, there is the option to inject CSS. The `custom` field under **card style** and under the theme itself behaves differently. On **card style** objects, the CSS is inlined into a selector for those cards, be that a type of card, group of entities, or a specific entity.

Directly under theme however, the CSS is injected raw into the stylesheet. This means selectors need to be written. However this CSS has access to a special variable `$this`, which will resolve to the selector that the theme applies to. For dashboard level themes, that would be `body`.

These fields are ofcourse for the utmost power-user. It is expected the user inspects the DOM themselves to understand how to style. 

### Example

```yaml
dashboards:
  main:
    theme:
      background:
        color: "linear-gradient(135deg, #1a1a2e 0%, #0f3460 100%)"
      colors:
        accent: "hsl(200, 80%, 60%)"
        surface: "rgba(255,255,255,0.07)"
      shapes:
        borders: true
        medium_border_radius: "0.75em"
      typography:
        font_family: "'Inter', sans-serif"
      sensors:
        borders: true
      custom: |
        $this .center-container .centered-dateclock-wrapper {
          width: 50% !important;
        }
        $this .center-container .weather-main .weather-temp {
            font-size: $font_xxl !important;
        }
    screens:
      - name: Home
        sensors:
          style:
            background: $accent
            borders: false
            custom: |
              color: $interactive:0.5 !important;
          cards:
            # ...
        theme:
          colors:
            accent: "#ff7043"   # warm accent on this screen only
      - name: Media
        theme: synthwave        # use a named preset for this screen
```

---

## Full example

See [example](https://github.com/Skrubbadubba/HoaryDash/blob/main/hoarydash/hoarydash.example.md)
