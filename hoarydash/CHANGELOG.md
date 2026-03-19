# v0.5.4

## Hotfix

Fix toggle pills not showing on fans and lights


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