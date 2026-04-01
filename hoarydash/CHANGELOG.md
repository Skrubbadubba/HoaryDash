# Patch v0.9.1

- Removed debug code that overwrote the eye icon on the screenonlock button

---

## Previous Minor Release

### v0.9.0

- Support for Wallpanel:
Use it by:
```yaml
wallpanel: true
```
at the top level. Make sure to not have a fully kiosk entry, it will take preceedence. This is used by the screenonlock button btw.

- Screen navigation now happens by changing the hash in the url to id of screen.

You could use this to for example control which screen is active from home assistant, by using fully kiosk commands to navigate to for example `http://homeassistant.local:4567/dash/#screen-tiles`.
