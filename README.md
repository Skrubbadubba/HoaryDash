# ⚠🛠 Under heavy development 🛠⚠

This is very early alpha, looks fine IMO but possibly hard to setup. Is alpha even the right word? Not like its a product lol im just messing around.

---

# HoaryDash

## What

A lightweight Home Assistant dashboard for old Android tablets. It runs as an addon, and is configurable with yaml. It runs a server in go exposed on port 4567, which you simply point your kiosk browser to.

### Features

- Runs comfortably on Chromium 44 / Android 6 tablets via Fully Kiosk Browser
- Live entity state via HA WebSocket (no polling)
- Sensor widgets, clock, nightlight mode, keep-screen-on toggle
- Fully Kiosk Browser integration (screensaver control, brightness)
- Config-driven — edit one YAML file, dashboard regenerates automatically
- Designed as an HA addon, also runnable as a standalone Docker container

### Requirements

- Home Assistant with Supervisor (HAOS or supervised install)
- Fully Kiosk Browser on your tablet
- A long-lived access token from Home Assistant

## Why

Home Assistant's own dashboard is too modern for decade-old hardware. Fully kiosk and similar apps uses androids built in webviews, which for old tablets is heavily outdated and won't run the javascript that comes with lovelace.

You can get firefox, which has modern version compatible with even android 4 I think, and it uses its own javascript engine which has modern feature parity. You could configure fully kiosk to open firefox in app mode, and put the tab in fullscreen or something. Even so, HAs interface is very heavy. Old tablets struggle to run it. In fact on my tablet it was barely usable. It even crashed when I opened the color wheel for a lamp, not firefox mind you, but the entire tablet, it restarted...

 But I wanted to make use of my tablet. It was cheap and I dont want it to become e-waste. HoaryDash is a purpose-built alternative: a Go server that generates a static, minimal dashboard from a simple YAML config and proxies live entity state from HA's WebSocket API. It doesn't use any javascript framework or bundles anything.


