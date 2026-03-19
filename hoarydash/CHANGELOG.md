# v0.5.3

## Features

- Allow styling font sizes
    - base font size
    - font size for sensors and entities separately
The entire components are now based on the font size, so the font size scales their size.
Example:
```yaml
theme:
    entities:
        font_size: 18
        ...
    sensors:
        font_size: 25
        ...
    base_font_size: 20
```

- Removed orange focus outline on tap
- Removed blue background highlight on tap


---
### Previous minor release:

## v0.5.0 🎉

This release brings a bunch of polish, and more controls. Might be actually useful now :D

### Features

##### Support and controls for more entities:

There is now cards for a bunch of simple entities. Toggleable entities now get a pill automatically, and lights gets a expandable popup with sliders for brightness and cct (if available)

Supported entities:
- light
- switch
- input_boolean
- input_button
- button
- scene
- script

Fan speed controls will hopefully be patched in soon

##### Theming

- Margins and alignment is slightly polished

###### New options

For sensors and entities under
```yaml
theme:
    entities:
        ...
    sensors:
        ...
```
- borders: `true/false`
- border_color: `css`
- background: `css`