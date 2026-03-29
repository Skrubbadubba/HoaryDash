# HoaryDash

A lightweight dashboard for old Android tablets, designed to run comfortably on
Chromium 44 (Android 6+) via Fully Kiosk Browser.

## Installation

1. In Home Assistant, go to **Settings → Add-ons → Add-on Store**
2. Click the three-dot menu (⋮) in the top right and choose **Repositories**
3. Add the repository URL: `https://github.com/YOUR_USERNAME/HoaryDash`
4. Find **HoaryDash** in the store and click **Install**
5. Start the addon — this creates the config folder on first run

---

## Configuration

After first start, a folder called `hoarydash` will appear in your addon config
directory. You can reach it via:

- **Samba share** → `\\homeassistant\addon_configs\hoarydash\`
- **SSH** → `/addon_configs/hoarydash/`
- **Studio Code Server addon**
- **File Editor addon**

These are addons you may or may not have. If you do not now how to access your home assistants file system, youtube tutorials are your friend.


Create a file called `hoarydash.yaml` in that folder with the following content:
```yaml
dashboards:
  # animations: false
  dash:
    nightlight:
      enabled: true
      override_colors: true
    sensors:
      - entity_id: sensor.living_room_temperature # Or whatever you have
        label: Temperature
        unit: °C
    entities:
      - entity_id: button.doorbell
        label: Ring doorbell
        icon: 🔔
      - entity_id: light.ceiling
        label: Ceiling light
        icon: 💡
    widgets:
      - entity_id: weather.home
        internal_borders: false
        forecast_interval: hourly
        forecast_times: 5
    # Example of light theme
    # theme:
    #   background_gradient: "linear-gradient(135deg, #1a1a2e 0%, #16213e 40%, #0f3460 100%)"
    #   font_color: "#e0f7fa"
    #   secondary_font_color: "#80cbc4"
    #   icon_color: "#4dd0e1"
    #   cards:
    #     borders: true
    #     border_color: "rgba(255, 255, 255, 0.12)"
    #     background: "rgba(255, 255, 255, 0.08)"
    #   base_font_size: 20


localization:
  locale: "en-US"
  timezone: "America/New_York"
  hour12: false # Or true I guess if your weird

fully_kiosk:
  screensaver_timeout: 60

home_assistant:
  token: "your_long_lived_token_here"
```

Find more example [themes](./themes.example.yaml)

Save the file. The dashboard will regenerate automatically — no restart needed. Restart the addon if it doesn't work

### Separate dashboards

Each entry under `dashboards` will become its own dashboard with its own endpoint, reachable on \<HoaryDash-url\>/\<dashboard-name\>. So from the example above one dashboard would be created at <a href="">http://homeassistant.local:4567/dash</a>

Find more details in [CONFIG.md](https://github.com/Skrubbadubba/HoaryDash/blob/main/hoarydash/CONFIG.md)

---

## Getting a Long-Lived Access Token

1. In Home Assistant, click your username in the bottom-left
2. Scroll to **Long-Lived Access Tokens**
3. Click **Create Token**, give it a name (e.g. `hoarydash`)
4. Copy the token and paste it into `HA_TOKEN` in your config

## Security

### There are no regards to security yet!

If the project gathers interest I _will_ add security before a v1 release. Right now, the go server allows anyone to connect to its websocket, after which it will automatically authenticate with HA and proxy any messages. **Anyone with access to the HoaryDash server has access to everything in HA! In practice this means anyone on your LAN can do anything in HA.** However HoaryDash is never exposed to the internet unless you explicitly port forward it on your router or something.



## Fully Kiosk Setup

In Fully Kiosk Browser settings, set:

- **Web Content Settings → Start URL:** `http://<your-ha-ip>:4567/dash.html`
- **Screensaver → Screensaver URL:** `http://<your-ha-ip>:4567/dash.html`

Enable the javascript interface:

- **Advanced Web Settings → Enable JavaScript Interface:** Enable



## Config Reference

See [go struct](https://github.com/Skrubbadubba/HoaryDash/blob/main/hoarydash/server/main.go#L16). The yaml annotation to the right of fields dictate the field name in yaml. The ones that dont have an annotation are just lowercased in yaml.