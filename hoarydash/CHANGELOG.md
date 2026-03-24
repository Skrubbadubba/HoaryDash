# v0.6.2

## Hotfix:
- Fixed code typo in toggles

---
# Previous minor update

## v0.6.0

Big release 🎉! Lots more theming, and we now have some pretty cool widgets for weather and media players!

### Whats New

#### ✨ Features

##### Configurable widgets

These are a separate card type for entities that require a bit more space. They are configurable under `widgets` and are configured alongside entities and sensors in the yaml like so:
```yaml
...
    sensors:
        ...
    entities:
        ...
    widgets:
        - entity_id: weather.home
          forecast_times: 5
          forecast_interval: daily
        - entity_id: media_player.spotify
          show_album: true
          font_size: 30
```

These cards sit in the center, to the right of the dateclock

##### Theming

- `entities`, `sensors` and `widgets` are all now separate. They share the same schema
- `cards` shares the same schema as above, and serves as a default
- Per widget fontsize override
- Can configure border-radius of cards

##### Other

- Configurable nightlight color
- Nightlight keeps screen on automatically
- Animations can be toggled

#### ⚡️ Code changes & optimizations

- All scripts now use IIFEs
- Component style tags are now duplicated
- Addded a REST api to the go server
    - Used for translations and proxying media images for now