# v0.10.0: Media Browser 🎵

This release comes with a 2 main updates:

- A media browser
- Standardisation of cards

If you know your media player supports browsing media, you can enable a button which brings up a browser modal like so:

```yaml
- entity_id: media_player.spotify
  show_browser: true
```

## Breaking changes

Screen positions have been deprecated. The list of screens will appear as they are ordered in the yaml.

With how cards now have been standardised, changes to yaml for the centered layout has to be made. In short, change:

```yaml
entities:
  - entity_id: light.ceiling
    label: Ceiling
    icon: lightbulb
  - entity_id: switch.fan
    label: Fan
sensors:
  - entity_id: sensor.living_room_temperature
    label: Temperature
    unit: °C
```

to:

```yaml
entities:
  cards:
    - entity_id: light.ceiling
      label: Ceiling
      icon: lightbulb
    - entity_id: switch.fan
      label: Fan
sensors:
  cards:
    - entity_id: sensor.living_room_temperature
      label: Temperature
      unit: °C
```

See [CONFIG.md](https://github.com/Skrubbadubba/HoaryDash/blob/main/hoarydash/CONFIG.md#card-group-schema#cards) for a full explanation

## Features

- Media browser button
- Cards individually stylable
- All groups/zones of cards follow a standard format, allowing them to be styled together, see [card styling](https://github.com/Skrubbadubba/HoaryDash/blob/main/hoarydash/CONFIG.md#card-group-schema)
- Media widget now fully uses icons instead of emojis
