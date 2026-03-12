# v0.4.0

## Multiple dashboards 🎉

You can now expose multiple endpoints with their own dashboards like so:
```yaml
dashboards:
    dash:
        nightlight:
            enabled: true
        theme:
            ...
    screensaver:
        nightlight:
            enabled: false
        sensors:
            ...
        theme:
            ...
    ...
home_assistant:
    ...
localization:
    ...
```

These would be available on \<url\>/dash and \<url\>/screensaver respectively.