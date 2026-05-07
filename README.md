> [!IMPORTANT]
> Under development - expect breaking changes. Check the changelog and configuration guide before updating.

---

<div align="center">
    <h1 align="centered">
        HoaryDash
    </h1>
    <p>
        <a href="https://github.com/Skrubbadubba/HoaryDash/releases/latest">
            <img src="https://img.shields.io/github/v/release/Skrubbadubba/HoaryDash?style=flat-square" alt="Release" />
        </a>
        <a href="https://go.dev">
            <img src="https://img.shields.io/badge/Go-1.24-00ADD8?style=flat-square&logo=go&logoColor=white" alt="Go" />
        </a>
        <a href="https://www.home-assistant.io">
            <img src="https://img.shields.io/badge/Home%20Assistant-app-41BDF5?style=flat-square&logo=homeassistant&logoColor=white" alt="Home Assistant" />
        </a>
    </p>
</div>

<p align="center">
  <img src="https://raw.githubusercontent.com/Skrubbadubba/HoaryDash/refs/heads/main/docs/centered.png" width="49%" />
  <img src="https://raw.githubusercontent.com/Skrubbadubba/HoaryDash/refs/heads/main/docs/media.png" width="49%" />
</p>
<div align="center">
  <sub>Centered layout &nbsp;·&nbsp; Media-full layout</sub>
</div>

## What

A lightweight Home Assistant frontend for old Android tablets. It runs as an addon (app), and is configurable with yaml. It runs a server in go exposed on port 4567, which you simply point your kiosk browser to.

### Features

- Runs comfortably on Chromium 44 / Android 6 tablets via Fully Kiosk Browser.
- Live entity state via HA WebSocket (no polling)
- Sensor widgets, clock, nightlight, keep-screen-on toggle, media 
- Multiple screens with swipe or navbar navigation
- Multiple dashboards on separate endpoints
- Fully Kiosk Browser integration (screensaver toggle)
- Config-driven - edit one YAML file, dashboard regenerates automatically
- Designed as an HA app, also runnable as a standalone Docker container

### Requirements

- Home Assistant with Supervisor (HAOS or supervised install)
  > (Or any machine with docker)
- A browser on your tablet


## Setup

See [DOCS.md](https://github.com/Skrubbadubba/HoaryDash/blob/main/hoarydash/DOCS.md)

## Plans/Goals

I have a bunch of features planned. The idea is to make a feature rich configurable dashboard that not only works with old browser, but is good enough to where someone might even want use as an alternative to lovelace. That being said, I hyper-focus and switch interests easily. I have no idea how long this will be motivating for me. I also don't know when I will have time for it. Anyways, goals:

- [x] More complex widgets
- [x] Multiple screens per dashboard
- [x] A browse media modal
- [x] Fullscreen-widget layouts (such as a full screen just for media)
- [x] System for preset themes, allowing easy switch between them
- [ ] Controllable HA entities via mqtt such as current screen and theme
- [ ] Some security
- [ ] Calendar widget
- [ ] Custom community cards system (altought not sure if thats a good idea)
- [ ] A logo
- [x] Prepopulating dashboard with information from HA
    - [x] Fill all entity icons based on domain, device type, or attribute
    - [x] Fill label with friendly name
    - [ ] Configure timezone and language based on HA configuration
- [ ] Create a default configuration based on available HA entities on first install
- [ ] A drag and drop editor (this one is really out of scope though, dont get your hopes up)

> [!NOTE]
> Last updated: v0.13.0

## Why

Home Assistant's own dashboard is too modern for decade-old hardware. Fully kiosk and similar apps uses androids built in webviews, which for old tablets is heavily outdated and won't run the javascript that comes with lovelace.

You can get firefox, which has modern version compatible with even android 4 I think, and it uses its own javascript engine which has modern feature parity. You could configure fully kiosk to open firefox in app mode, and put the tab in fullscreen or something. Even so, HAs interface is very heavy. Old tablets struggle to run it. In fact on my tablet it was barely usable. It even crashed when I opened the color wheel for a lamp, not firefox mind you, but the entire tablet, it restarted...

But I wanted to make use of my tablet. It was cheap and I dont want it to become e-waste. HoaryDash is a purpose-built alternative: a Go server that generates a static, minimal dashboard from a simple YAML config and proxies live entity state from HA's WebSocket API. It doesn't use any javascript framework or bundles anything.


## How

HoaryDash runs as a Home Assistant addon (or just a docker container anywhere you'd like). When it starts, a Go server reads your `hoarydash.yaml` and generates a completely static `dash.html` using templates — no framework, no client-side rendering.Its all plain HTML with a small amount of hand-written ES5 JavaScript that Chromium 44 can handle.

Live entity state comes through a single WebSocket connection from the tablet to the Go server, which proxies it to HA's WebSocket API. Your HA token never leaves the server. When you edit the YAML, the server detects the change, regenerates the HTML, and tells the tablet to reload through the existing websocket.

In the future, I have plans for CSS and JS to get run through Babel and PostCSS at startup to make sure nothing modern sneaks in. The output would be guaranteed ES5 and prefixed CSS, so it stays compatible even as I add things.

### Why this stack

The other option were to have it be a custom integration. That would allow for a pretty straight forward way of adding entities to control the config, such as toggling or even scheduling the nightlight. Even more ambititous would be to make an android app and have the integration just be a thin controller instead of an entire server, but I don't know anything about native android development.

The current stack was chosen mostly because adding PostCSS and such requires node. Docker allows that easily and then we just control it from go. I dont know if thats even possibly in HAs python environment. The other reason is I like go and docker. I think python is and should stay a scripting language. The friction of adding features in python when you also need to follow HAs requirments is much higher than in just go + html.

> [!WARNING]
> ## Security
>
> **There are no regards to security yet!**
>
> If the project gathers interest I _will_ add security before a v1 release. Right now, the go server allows anyone to connect to its websocket, after which it will automatically authenticate with HA and proxy any messages. **Anyone with access to the HoaryDash server has access to everything in HA!** However HoaryDash is never exposed to the internet unless you explicitly port forward it on your router or something. In practice, this simply only means someone in your home could use hoarydash for full access to home assistant.
