package main

import (
	"fmt"
	"html/template"
	"os"

	"go.yaml.in/yaml/v4"
)

type Dashboard struct {
	Animations   *bool
	Screenonlock *bool
	Nightlight   struct {
		Enabled        *bool
		Color          template.CSS
		OverrideColors bool `yaml:"override_colors"`
	}
	ThemeRef  ThemeRef `yaml:"theme"`
	Theme     Theme    `yaml:"_"`
	ShowHints *bool    `yaml:"show_hints"`
	Swipe     *bool
	Navbar    struct {
		Enabled  bool
		Position string
		Style    string
	}
	Screens []Screen
}

type Screen struct {
	Layout    string
	Name      string
	Icon      *string
	Dateclock struct {
		Enabled       *bool
		Hour12        bool
		CapitaliseDay bool `yaml:"capitalise_day"`
		ShowSeconds   bool `yaml:"show_seconds"`
	}
	// Centered-layout specific
	Widgets  *CardGroup
	Sensors  *CardGroup
	Entities *CardGroup
	Order    struct {
		Entities int
		Widgets  int
		Sensors  int
	}

	// Tiled-layout specific
	Groups []struct {
		Name      string
		Icon      string
		CardGroup `yaml:",inline"`
	}

	// Fullscreen-layout specific
	EntityID     string `yaml:"entity_id"`
	MediaOptions `yaml:",inline"`
	Badges       struct {
		Sensors []Sensor
		Badge   struct {
			Label string
			Icon  string
		}
	}
	ThemeRef ThemeRef `yaml:"theme"`
	Theme    *Theme   `yaml:"_"`
	Rotate   *bool
}

type ThemeRef struct {
	Name  string // set if user wrote `theme: aurora`
	Theme *Theme // set if user wrote `theme: background: ...`
}

func (t *ThemeRef) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind == yaml.ScalarNode {
		t.Name = value.Value
		return nil
	}
	return value.Decode(&t.Theme)
}

type Entity struct {
	EntityID string `yaml:"entity_id"`
	Label    string
	Icon     string
}

type Sensor struct {
	Entity `yaml:",inline"`
	Unit   string
}

type Card struct {
	Entity `yaml:",inline"`
	Unit   string
	Style  CardStyle
	// Widget specific
	InternalBorders *bool `yaml:"internal_borders"`
	// Weather-specific
	WeatherOptions `yaml:",inline"`
	// Media-specific
	MediaOptions `yaml:",inline"`
}

type WeatherOptions struct {
	ForecastInterval *ForecastInterval `yaml:"forecast_interval"`
	ForecastTimes    *int              `yaml:"forecast_times"`
	Hour12           *bool
}
type MediaOptions struct {
	ShowVolume  *bool `yaml:"show_volume"`
	ShowAlbum   *bool `yaml:"show_album"`
	ShowBrowser *bool `yaml:"show_browser"`
	Spotifyplus *bool
	Queue       *bool
}

type CardGroup struct {
	Style CardStyle
	Cards []Card
}

type ForecastInterval string

const (
	ForecastIntervalDaily      ForecastInterval = "daily"
	ForecastIntervalTwiceDaily ForecastInterval = "twice_daily"
	ForecastIntervalHourly     ForecastInterval = "hourly"
)

func (f ForecastInterval) Valid() bool {
	switch f {
	case ForecastIntervalDaily, ForecastIntervalTwiceDaily, ForecastIntervalHourly:
		return true
	}
	return false
}

func (f *ForecastInterval) UnmarshalYAML(value *yaml.Node) error {
	var s string
	if err := value.Decode(&s); err != nil {
		return err
	}
	*f = ForecastInterval(s)
	if !f.Valid() {
		return fmt.Errorf("invalid forecast_interval %q, must be daily, twice_daily or hourly", s)
	}
	return nil
}

type Navigation string

const (
	NavigationNavbar Navigation = "navbar"
	NavigationSwipe  Navigation = "swipe"
)

func (n Navigation) Valid() bool {
	switch n {
	case NavigationNavbar, NavigationSwipe:
		return true
	}
	return false
}

func (n *Navigation) UnmarshalYAML(value *yaml.Node) error {
	var s string
	if err := value.Decode(&s); err != nil {
		return err
	}
	*n = Navigation(s)
	if !n.Valid() {
		return fmt.Errorf("invalid navigation %q, must be 'swipe' or 'navbar'")
	}
	return nil
}

type Config struct {
	Localization struct {
		Locale   string
		Timezone string
	}
	FullyKiosk *struct {
		ScreensaverTimeout int `yaml:"screensaver_timeout"`
	} `yaml:"fully_kiosk"`
	WallPanel     *bool
	HomeAssistant struct {
		URL   string
		TOKEN string
	} `yaml:"home_assistant"`
}

type Yaml struct {
	Dashboards map[string]Dashboard
	Config     `yaml:",inline"`
	Themes     ThemesMap
	IsDev      bool
}

func parseConfig() (*Yaml, error) {
	var yaml_file []byte
	var err error
	if isDev {
		yaml_file, err = os.ReadFile(configPath + "/hoarydash.dev.yaml")
	} else {
		yaml_file, err = os.ReadFile(configPath + "/hoarydash.yaml")
	}
	parsed := Yaml{}
	if err != nil {
		return &parsed, err
	}
	err = yaml.Unmarshal(yaml_file, &parsed)
	return &parsed, err
}
