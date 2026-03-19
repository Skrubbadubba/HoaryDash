# ⚠🛠 Under heavy development 🛠⚠

This is very early alpha, looks fine IMO but possibly hard to setup. Is alpha even the right word? Not like its a product lol im just messing around.

---

# HoaryDash

<img width="2540" height="1420" alt="image" src="https://github.com/user-attachments/assets/b35e3998-f761-43a1-8738-caa63ee76f68" />

<img width="2540" height="1420" alt="image" src="https://github.com/user-attachments/assets/fae82044-86a4-4aec-9b61-1eb8df014a49" />


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


## How

HoaryDash runs as a Home Assistant addon (or just a docker container anywhere you'd like). When it starts, a Go server reads your `hoarydash.yaml` and generates a completely static `dash.html` using templates — no framework, no client-side rendering.Its all plain HTML with a small amount of hand-written ES5 JavaScript that Chromium 44 can handle.

Live entity state comes through a single WebSocket connection from the tablet to the Go server, which proxies it to HA's WebSocket API. Your HA token never leaves the server. When you edit the YAML, the server detects the change, regenerates the HTML, and tells the tablet to reload through the existing websocket.

In the future, I have plans for CSS and JS to get run through Babel and PostCSS at startup to make sure nothing modern sneaks in. The output would be guaranteed ES5 and prefixed CSS, so it stays compatible even as I add things.

### Why this stack

The other option were to have it be a custom integration. That would allow for a pretty straight forward way of adding entities to control the config, such as toggling or even scheduling the nightlight. Even more ambititous would be to make an android app and have the integration just be a thin controller instead of an entire server.

The current stack was chosen mostly because adding PostCSS and such requires node. Docker allows that easily and then we just control it from go. I dont know if thats even possibly in HAs python environment. The other reason is I like go and docker. I think python is and should stay a scripting language. The friction of adding features in python when you also need to follow HAs requirments is much higher than in just go + html.

## Security

### There are no regards to security yet!

If the project gathers interest I _will_ add security before a v1 release. Right now, the go server allows anyone to connect to its websocket, after which it will automatically authenticate with HA and proxy any messages. **Anyone with access to the HoaryDash server has access to everything in HA! In practice this means anyone on your LAN can do anything in HA.** However HoaryDash is never exposed to the internet unless you explicitly port forward it on your router or something.
