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
      unit: °C

widgets:
  cards:
    - entity_id: media_player.spotify
    - entity_id: weather.home
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
  label: Ceiling light
  icon: lightbulb
  style:
    border_radius: "1em"
    background: "rgba(255,255,255,0.12)"
```

| Key | Type | Description |
|-----|------|-------------|
| `entity_id` | string | Home Assistant entity ID. Determines the card type — see tables below |
| `label` | string | Display label. Falls back to entity ID if omitted |
| `icon` | string | MDI icon name or emoji, e.g. `lightbulb`, `🔔` |
| `unit` | string | Unit string shown below the value (sensors). e.g. `°C`, `%`, `µg/m³` |
| `style` | [card style](#card-style-fields) | Style overrides for this individual card |

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

### Widget-specific fields

Some card types accept additional fields.

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
| `internal_borders` | bool | `true` | Show the divider between info and controls |

### Card style fields

A `style` block can be placed on an individual card or on a card group. Individual card style takes precedence. For theme-level styling that applies to all cards of a given type across the whole dashboard, see [Theming](#theming).

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

Group object:

| Key | Type | Description |
|-----|------|-------------|
| `name` | string | Group header label |
| `icon` | string | Icon shown next to the group header |
| `cards` | list | List of cards (same fields as entities, sensors, and widgets) |

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
          forecast_interval: hourly
          forecast_times: 5
        - entity_id: media_player.spotify
          show_album: true
    sensors:
      cards:
        - entity_id: sensor.living_room_temperature
          label: Temperature
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

## Theming

The theme controls colours, borders, backgrounds, and font sizes. `cards` is the base default — `entities`, `sensors`, and `widgets` each inherit from `cards` and can override individual properties.

```yaml
theme:
  background: "linear-gradient(135deg, #1a1a2e 0%, #16213e 40%, #0f3460 100%)"
  font_color: "#e0f7fa"
  secondary_font_color: "#80cbc4"
  icon_color: "#4dd0e1"
  base_font_size: 20

  cards:
    borders: true
    border_color: "rgba(255, 255, 255, 0.12)"
    border_radius: "0.75em"
    background: "rgba(255, 255, 255, 0.08)"
    font_size: 14

  entities:
    background: "rgba(255, 255, 255, 0.06)"

  sensors:
    borders: false
    font_size: 20
```


### Common theme fields

A `theme` field with these keys can be applied to either the entire dashboard or per screen.

| Key | Type | Description |
|-----|------|-------------|
| `background` | CSS value | Background of the entire page. Any valid CSS `background` value. Note that radial gradients are not supported on old browsers |
| `opaque_background` | CSS value | Background color to be used where transparancy is not wanted, such as modals. The recommendation is to style card backgrounds with hsl, and specify the same hue value with lower saturation and brightness here | 
| `font_color` | CSS color | Primary text colour |
| `secondary_font_color` | CSS color | Secondary/muted text colour |
| `icon_color` | CSS color | Icon tint colour |
| `button_background` | CSS color | Background for entity control buttons |
| `font_size` | number | Base font size in px. Everything scales from this |
| `entities` | `card-style` | Styles for all cards of this category |
| `sensors` | `card-style` | Styles for all cards of this category |
| `widgets` | `card-style` | Styles for all cards of this category |

### Card theme fields (cards / entities / sensors / widgets)

Entities, sensors and widgets are all types of `Cards`. They all share the same schema, as follows:

| Key | Type | Description |
|-----|------|-------------|
| `borders` | bool | Show card borders |
| `border_color` | CSS color | Border colour |
| `border_radius` | CSS value | Corner rounding, e.g. `0.75em` |
| `background` | CSS value | Card background colour or gradient |
| `font_size` | CSS value | Font size for cards in this section, e.g. `18px` |

### Card group theming

On the _tiled_ layout, a group can have its own theme, which follows the `card`theme schema.

Theming is currently a work in process. The fact that layouts (centered and tiled) have different schemas for specifying what cards to show makes it complicated. The hope is that in the future, any dashboard, screen, group of cards, indivudial card, and any other future component can be themed individually.

---

## Full example

See [example](https://github.com/Skrubbadubba/HoaryDash/blob/main/hoarydash/hoarydash.example.md)
