# HoaryDash Configuration Reference

All configuration lives in a single file: `/addon_configs/hoarydash/hoarydash.yaml`. This is the heart of the entire operation

The file has two top-level sections: `dashboards` (a map of dashboard names to dashboard configs) and a set of global keys (`localization`, `fully_kiosk`, `home_assistant`).

---

## Top-level structure

```yaml
dashboards:
  my-dashboard:        # becomes available at /my-dashboard/
    screens: [...]
    theme: {...}
    navbar: {...}

  another-dashboard:
    screens: [...]

localization:
  locale: "en-US"
  timezone: "America/New_York"
  hour12: false

fully_kiosk:
  screensaver_timeout: 60

home_assistant:
  url: "http://homeassistant.local:8123" # Defaults to this, can probably ignore
  token: "your_long_lived_token_here"
```

Each key under `dashboards` becomes an endpoint. `dashboards.dash` is served at `/dash/`, `dashboards.living-room` at `/living-room/`, and so on.

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
| `screenonlock` | bool | — | Keep screen on when the dashboard is active |
| `swipe` | bool | `true` | Enable horizontal swipe to navigate between screens |
| `show_hints` | bool | `true` | Show chevron hints at screen edges pointing toward adjacent screens |
| `screens` | list | — | List of screens (see [Screens](#screens)) |
| `navbar` | object | — | Navigation bar config (see [Navbar](#navbar)) |
| `nightlight` | object | — | Nightlight overlay config (see [Nightlight](#nightlight)) |
| `theme` | object | — | Visual theme (see [Theme](#theme)) |

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

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `name` | string | — | Display name shown in the navbar |
| `icon` | string | — | Icon name shown in the navbar (MDI icon slug, e.g. `home`, `thermometer`) |
| `layout` | string | `centered` | Layout preset. `centered` or `tiled` |
| `dateclock` | object | — | Clock/date display (see [Dateclock](#dateclock)) |
| `entities` | list | — | Control buttons/toggles. Centered layout only |
| `sensors` | list | — | Sensor readouts. Centered layout only |
| `widgets` | list | — | Widget cards (weather, media, todo). Centered layout only |
| `order` | object | — | CSS flex order for the three zones. Centered layout only |
| `groups` | list | — | Groups of cards. Tiled layout only |

### Dateclock

```yaml
dateclock:
  enabled: true
  show_seconds: true
  capitalise_day: true
  hour12: false
```

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `enabled` | bool | `true` | Show the clock and date |
| `show_seconds` | bool | `false` | Show seconds in the time display |
| `capitalise_day` | bool | `false` | Capitalise the first letter of the day name |
| `hour12` | bool | inherits from `localization` | Override the global hour12 setting for this clock |

### Zone ordering (centered layout)

The three zones of the centered layout — entities, dateclock/widgets, and sensors — are flex children. Their order is set via CSS `order`:

```yaml
order:
  entities: 1
  widgets: 2
  sensors: 3
```

Lower numbers appear first (top of the screen). Values can be any integer. For example, to just set sensors top the top:

```yaml
order:
   sensors: -1
```

---

## Entities

Entities are control elements — buttons, toggles, and adjustable sliders. The `entity_id` domain determines what control is rendered.

| Domain | Control rendered |
|--------|-----------------|
| `switch`, `input_boolean` | Toggle |
| `light`, `fan` | Toggle + adjustable overlay (brightness/speed/CCT) |
| `button`, `input_button`, `scene`, `script` | Button |

```yaml
entities:
  - entity_id: light.ceiling
    label: Ceiling
    icon: lightbulb

  - entity_id: button.doorbell
    label: Ring doorbell
    icon: doorbell
```

| Key | Type | Description |
|-----|------|-------------|
| `entity_id` | string | Home Assistant entity ID |
| `label` | string | Display label. Falls back to entity ID if omitted |
| `icon` | string | Emoji or mdi icon to display |

---

## Sensors

Sensors display a state value with a label and unit.

```yaml
sensors:
  - entity_id: sensor.living_room_temperature
    label: Temperature
    unit: °C
```

| Key | Type | Description |
|-----|------|-------------|
| `entity_id` | string | Home Assistant entity ID |
| `label` | string | Display label |
| `unit` | string | Unit string shown below the value, e.g. `°C`, `%`, `µg/m³` |

---

## Widgets

Widgets are richer cards driven by entity state. The domain of `entity_id` determines the widget type.

| Domain | Widget |
|--------|--------|
| `weather.*` | Weather widget with current conditions and forecast |
| `media_player.*` | Media player with album art, controls, and volume |
| `todo.*` | To-do list with add/check/filter |

```yaml
widgets:
  - entity_id: weather.home
    forecast_interval: twice_daily
    forecast_times: 5

  - entity_id: media_player.spotify
    show_volume: true
    show_album: true
```

### Common widget fields

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `entity_id` | string | — | Home Assistant entity ID |
| `font_size` | string | — | Override font size for this widget, e.g. `18` |
| `internal_borders` | bool | `true` | Show internal dividers within the widget card |

### Weather-specific fields

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `forecast_interval` | string | `daily` | Forecast granularity. One of `daily`, `twice_daily`, `hourly` |
| `forecast_times` | int | `5` | Number of forecast periods to display |

### Media player-specific fields

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `show_volume` | bool | `true` | Show the volume slider |
| `show_album` | bool | `true` | Show album art |

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
            unit: °C
          - entity_id: light.ceiling
            label: Ceiling
      - name: Media
        cards:
          - entity_id: media_player.spotify
```

Card type is inferred from the `entity_id` domain, exactly as in the centered layout. All the same fields apply.

| Key | Type | Description |
|-----|------|-------------|
| `name` | string | Group header label |
| `icon` | string | Icon shown next to the group header |
| `cards` | list | List of cards (same fields as entities, sensors, and widgets) |

---

## Theme

The theme controls colours, borders, backgrounds, and font sizes. `cards` is the base default — `entities`, `sensors`, and `widgets` each inherit from `cards` and can override individual properties.

```yaml
theme:
  background_gradient: "linear-gradient(135deg, #1a1a2e 0%, #16213e 40%, #0f3460 100%)"
  font_color: "#e0f7fa"
  secondary_font_color: "#80cbc4"
  icon_color: "#4dd0e1"
  base_font_size: 20

  cards:
    borders: true
    border_color: "rgba(255, 255, 255, 0.12)"
    border_radius: "0.75em"
    background: "rgba(255, 255, 255, 0.08)"

  entities:
    background: "rgba(255, 255, 255, 0.06)"

  sensors:
    borders: false
```

A current limitation is that there are no per-screen theming, meaning if you set sensors with a transparent backgrounds, they will appear so on every screen, which mioght not look the best on the tiled layout.

### Dashboard-level theme fields

| Key | Type | Description |
|-----|------|-------------|
| `background_gradient` | CSS value | Background of the entire page. Any valid CSS `background` value |
| `body_background` | CSS value | Fallback solid background colour |
| `font_color` | CSS color | Primary text colour |
| `secondary_font_color` | CSS color | Secondary/muted text colour |
| `icon_color` | CSS color | Icon tint colour |
| `button_background` | CSS color | Background for entity control buttons |
| `base_font_size` | number | Base font size in px. Everything scales from this |

### Card theme fields (cards / entities / sensors / widgets)

| Key | Type | Description |
|-----|------|-------------|
| `borders` | bool | Show card borders |
| `border_color` | CSS color | Border colour |
| `border_radius` | CSS value | Corner rounding, e.g. `0.75em` |
| `background` | CSS value | Card background colour or gradient |
| `font_size` | CSS value | Font size for cards in this section, e.g. `18px` |

---

## Full example

See [example](https://github.com/Skrubbadubba/HoaryDash/blob/main/hoarydash/hoarydash.example.md)