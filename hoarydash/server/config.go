package main

import (
	"fmt"
	"html/template"
	"log"
	"os"
	"reflect"
	"strings"

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
	Theme     Theme    `yaml:"-"`
	ShowHints *bool    `yaml:"show_hints"`
	Swipe     *bool
	Navbar    struct {
		Enabled  bool
		Position string
		Style    string
	}
	TileOptions `yaml:"tile_options,omitempty"`
	Screens     []Screen

	// internals
	cardIndex   map[string][]*Card   `yaml:"-"`
	sensorIndex map[string][]*Sensor `yaml:"-"`
}

var walkCards = makeNodeWalker[Card]()
var walkSensors = makeNodeWalker[Sensor]()

func (d *Dashboard) buildIndex() {
	if d.cardIndex != nil || d.sensorIndex != nil {
		return
	}
	cardIndex := map[string][]*Card{}
	sensorIndex := map[string][]*Sensor{}
	walkCards(d, func(f reflect.StructField, c *Card) error {
		if c.EntityID != "" {
			cardIndex[c.EntityID] = append(cardIndex[c.EntityID], c)
		}
		return nil
	})
	walkSensors(d, func(f reflect.StructField, s *Sensor) error {
		if s.EntityID != "" {
			sensorIndex[s.EntityID] = append(sensorIndex[s.EntityID], s)
		}
		return nil
	})
	d.cardIndex = cardIndex
	d.sensorIndex = sensorIndex
}

func (d *Dashboard) EntityIDs() []string {
	d.buildIndex()
	seen := make(map[string]struct{}, len(d.cardIndex)+len(d.sensorIndex))
	for id := range d.cardIndex {
		seen[id] = struct{}{}
	}
	for id := range d.sensorIndex {
		seen[id] = struct{}{}
	}
	ids := make([]string, 0, len(seen))
	for id := range seen {
		ids = append(ids, id)
	}
	for _, screen := range d.Screens {
		if screen.EntityID != "" {
			ids = append(ids, screen.EntityID)
		}
	}
	return ids
}

func (d *Dashboard) Cards() []*Card {
	d.buildIndex()
	var cards []*Card
	for _, ptrs := range d.cardIndex {
		cards = append(cards, ptrs...)
	}
	return cards
}

func (d *Dashboard) Sensors() []*Sensor {
	d.buildIndex()
	var sensors []*Sensor
	for _, ptrs := range d.sensorIndex {
		sensors = append(sensors, ptrs...)
	}
	return sensors
}

type Screen struct {
	Layout string
	Name   string
	Icon   *string
	Dateclock
	// Centered-layout specific
	Widgets     *CardGroup
	Sensors     *CardGroup
	Entities    *CardGroup
	TileOptions `yaml:"tile_options,omitempty"`
	Order       struct {
		Entities int
		Widgets  int
		Sensors  int
	}

	// Tiled-layout specific
	Stretch bool
	Groups  []struct {
		Name            string
		Icon            string
		NormalizeHeight bool `yaml:"normalize_height"`
		Stretch         *bool
		CardGroup       `yaml:",inline"`
	}

	// Fullscreen-layout specific
	EntityID     string `yaml:"entity_id"`
	MediaOptions `yaml:"media_options"`
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

type Dateclock struct {
	Enabled       *bool
	Hour12        bool
	CapitaliseDay bool    `yaml:"capitalise_day"`
	ShowDate      *bool   `yaml:"show_date"`
	ShowTime      *bool   `yaml:"show_time"`
	ShowSeconds   bool    `yaml:"show_seconds"`
	DateSize      float64 `yaml:"date_size"`
	DateWeight    int     `yaml:"date_weight"`
	TimeSize      float64 `yaml:"time_size"`
	TimeWeight    int     `yaml:"time_weight"`
	Align         string
	FontSize      float64 `yaml:"font_size"`
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
	Class    string `yaml:"-"`
}

type TileOptions struct {
	ShowIcon *bool `yaml:"show_icon"`
	ShowPill *bool `yaml:"show_pill"`
}

type SensorOptions struct {
	Unit string
}
type Sensor struct {
	Entity        `yaml:",inline"`
	SensorOptions `yaml:",inline"`
}

type Card struct {
	Entity  `yaml:",inline"`
	Style   CardStyle
	Options struct {
		TileOptions   `yaml:",inline"`
		SensorOptions `yaml:",inline"`
		WidgetOptions `yaml:",inline"`
	} `yaml:"options"`
}

type WeatherOptions struct {
	ShowForecast     bool              `yaml:"show_forecast"`
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

type WidgetOptions struct {
	InternalBorders *bool `yaml:"internal_borders"`
	WeatherOptions  `yaml:",inline"`
	MediaOptions    `yaml:",inline"`
}

type CardGroup struct {
	Style       CardStyle
	TileOptions `yaml:"tile_options,omitempty"`
	Cards       []Card
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
		return fmt.Errorf("invalid navigation %q, must be 'swipe' or 'navbar'", s)
	}
	return nil
}

type Settings struct {
	Localization struct {
		Locale   string
		Timezone string
	}
	FullyKiosk *struct {
		ScreensaverTimeout int `yaml:"screensaver_timeout"`
	} `yaml:"fully_kiosk"`
	WallPanel *bool
	HA        HAConfig `yaml:"home_assistant"`
}

type HAConfig struct {
	HTTPURL string `yaml:"-"`
	WSURL   string `yaml:"-"`
	Token   string `yaml:"token"`
	baseURL string `yaml:"url"`
}

type UserConfig struct {
	Dashboards map[string]*Dashboard
	Settings   `yaml:",inline"`
	Themes     ThemesMap
	IsDev      bool
}

func loadConfig() (UserConfig, error) {
	var yaml_file []byte
	var err error
	if isDev {
		yaml_file, err = os.ReadFile(configPath + "/hoarydash.dev.yaml")
	} else {
		yaml_file, err = os.ReadFile(configPath + "/hoarydash.yaml")
	}
	parsed := struct {
		Dashboards map[string]Dashboard
		Settings   `yaml:",inline"`
		Themes     ThemesMap
	}{}
	if err = yaml.Unmarshal(yaml_file, &parsed); err != nil {
		return UserConfig{}, err
	}
	cfg := UserConfig{
		Settings:   parsed.Settings,
		Themes:     parsed.Themes,
		Dashboards: make(map[string]*Dashboard, len(parsed.Dashboards)),
	}
	for name, dash := range parsed.Dashboards {
		d := dash
		cfg.Dashboards[name] = &d
	}
	cfg.Settings.HA = resolveHA(cfg.Settings.HA)
	return cfg, nil
}

func resolveHA(ha HAConfig) HAConfig {
	resolved := ha
	if supervisorToken := os.Getenv("SUPERVISOR_TOKEN"); supervisorToken != "" {
		log.Printf("Supervisor token detected, using supervisor API")
		resolved.Token = supervisorToken
		resolved.HTTPURL = "http://supervisor/core"
		resolved.WSURL = "ws://supervisor/core/websocket"
		return resolved
	}
	if resolved.baseURL == "" {
		log.Print("HA url not set, defaulting to 'http://homeassistant.local:8123'")
		resolved.baseURL = "http://homeassistant.local:8123"
	}
	if resolved.Token == "" {
		log.Print("Getting HA token from environment")
		resolved.Token = os.Getenv("HA_TOKEN")
		if resolved.Token == "" {
			log.Print("No HA token could be read")
		}
	}
	resolved.HTTPURL = resolved.baseURL
	resolved.WSURL = strings.NewReplacer(
		"http://", "ws://",
		"https://", "wss://",
	).Replace(resolved.baseURL) + "/api/websocket"
	return resolved
}
