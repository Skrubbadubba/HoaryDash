# v0.13.0

### ✨ Features
- **Build-time entity enrichment** -- entity labels and icons are now pre-populated from Home Assistant at build time, so the rendered HTML is fully defined before any client connects; the options serve as overrides
    - sensor units are also pre-populated from HA state using the `unit_of_measurement` attribute
- **Stateful icons** -- Tile cards (buttons, toggleables, toggleables with sliders), now change their icon based on state according to HA conventions
- **Tile display options** -- new `show_icon` and `show_pill` options for entity tile cards, controllable per card, per group, per screen, and per dashboard, overrides cascadingly
- **HA websocket warming** -- When a client visits a dashboard, a websocket connection to HA is warmed up, subscribing to entities present in that dashboard, before being handed off the client when it has finished loading. This speeds up the load time to get live state
- **Supervisor API** -- When run as an app/addon, HoaryDash uses the supervisor APi, meaning no token or url to homeassistant has to be provided
- **Buttons hiding** -- Buttons now hide after a timeout when nightlight is active

### ⚠ Breaking changes
- `show_forecast` on weather cards, `unit` on sensors has moved under an `options` key alongside other domain-specific fields. See [domain-specific-fields](https://github.com/Skrubbadubba/HoaryDash/blob/main/docs/CONFIG.md#domain-specific-fields)

### 🏗 Fixes & Improvements
- Position of nightlight brightness slider now appears centered

### ⚙️ Internal
- Generic `nodeWalker` pattern added; It can walk a struct of arbitrary shape looking for nodes of a certain type, allowing mutation. `SliceElem` implemented on the walker so slice-resident structs are correctly visited
- `Dashboard` now exposes `Cards()`, `Sensors()`, and `EntityIDs()` using a `nodeWalker`
- `RenderContext` struct introduced as a typed context frame; context stack managed via `pushCtx`/`popCtx`/`getCtx` template funcs, fresh per render. This allows outer configuration to propagate to deeply nested partials without prop drilling
- `ComponentIconMap` filtered per dashboard to only domain+class combinations present in that dashboard; SVGs resolved at build time and inlined as `window.ICON_MAP`
- Deprecated `api.go`, as it was not used and probably wont be. The philosophy is that everything the dashboard needs will be bundled with the html.

## Devlog

Hi guys! This update is a bit more subtle. What I wanted to focus on here is connecting to HA and prepopulating information. I feel the entire project so far has required a very heavy setup for users, needing a token, needing to write labels for every entity, etc. So while there arent any big new features or cards, hopefully its significantly simpler to setup.

Don't really have much else to say regarding features. Along that note of no new big cards or such, alot of effort was spent on code architechture. I tried my best to make some new systems and use some new patterns that were more "enterprisey", in case anyone ever would feel like taking a look at it, maybe contributing 👀. Even though I tried, I'm not satisfied at all, it's still quite the mess 😅.

You might have noticed this update came a bit slow compared to previous development. That trend will continue, I really have other stuff I need to spend more time on. That being said I'd really appreciate if someone felt like contributing, or suggest new features, because at the same time not having all the time for this, I'm also quite out of ideas that are within scope currently. If anyone has a speicific entity domain or ui component they would really like to see, you are more then free to make an issue!

See ya 