# HoaryDash

A lightweight dashboard for old Android tablets, designed to run comfortably on
Chromium 44 (Android 6+) via Fully Kiosk Browser.

## Installation

1. In Home Assistant, go to **Settings → Add-ons → Add-on Store**
2. Click the three-dot menu (⋮) in the top right and choose **Repositories**
3. Add the repository URL: `https://github.com/YOUR_USERNAME/HoaryDash`
4. Find **HoaryDash** in the store and click **Install**
5. Start the addon — this creates the config folder on first run

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
dashboard:
  nightlight: true
  sensors:
    - entity_id: sensor.living_room_temperature # Or whatever you have
      label: Temperature
      unit: °C

localization:
  locale: "en-US"
  timezone: "America/New_York"
  hour12: false # Or true I guess if your weird

fully_kiosk:
  screensaver_timeout: 60

home_assistant:
  HA_URL: "http://homeassistant.local:8123"
  HA_TOKEN: "your_long_lived_token_here"
```

Save the file. The dashboard will regenerate automatically — no restart needed. Restart the addon if it doesn't work

## Getting a Long-Lived Access Token

1. In Home Assistant, click your username in the bottom-left
2. Scroll to **Long-Lived Access Tokens**
3. Click **Create Token**, give it a name (e.g. `hoarydash`)
4. Copy the token and paste it into `HA_TOKEN` in your config

## Fully Kiosk Setup

In Fully Kiosk Browser settings, set:

- **Web Content Settings → Start URL:** `http://<your-ha-ip>:4567/dash.html`
- **Screensaver → Screensaver URL:** `http://<your-ha-ip>:4567/dash.html`

Enable the javascript interface:

- **Advanced Web Settings → Enable JavaScript Interface:** Enable



## Config Reference

| Key | Description | Default |
|-----|-------------|---------|
| `dashboard.nightlight` | Show nightlight button | `true` |
| `dashboard.sensors` | List of HA entities to display | `[]` |
| `localization.locale` | BCP 47 locale tag (e.g. `sv-SE`) | Device locale |
| `localization.timezone` | IANA timezone in `Area/Location` format | Device timezone |
| `localization.hour12` | 12-hour clock | `false` |
| `localization.capitalise_day` | Capitalise day name | `false` |
| `fully_kiosk.password` | Fully Kiosk REST API password | — |
| `fully_kiosk.screensaver_timeout` | Fallback screensaver timeout (seconds) | `60` |
| `home_assistant.HA_URL` | Your HA instance URL | — |
| `home_assistant.HA_TOKEN` | Long-lived access token | — |

> **Timezone note:** Chromium 44 requires strict `Area/Location` IANA format.
> `Asia/Tokyo` works. Bare aliases like `Japan` do not.