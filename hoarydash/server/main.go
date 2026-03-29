package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/radovskyb/watcher"
	"go.yaml.in/yaml/v4"
)

//go:embed mdi.json
var mdiData []byte

var mdiIcons map[string]string

type Dashboard struct {
	Animations   *bool
	Screenonlock *bool
	Nightlight   struct {
		Enabled        *bool
		Color          template.CSS
		OverrideColors bool `yaml:"override_colors"`
	}
	Theme struct {
		BodyBackground     template.CSS `yaml:"body_background"`
		BackgroundGradient template.CSS `yaml:"background_gradient"`
		Cards              CardTheme    // Default for widgets, entities and sensors
		Entities           CardTheme
		Sensors            CardTheme
		Widgets            CardTheme
		ButtonBackground   template.CSS `yaml:"button_background"`
		FontColor          template.CSS `yaml:"font_color"`
		SecondaryFontColor template.CSS `yaml:"secondary_font_color"`
		IconColor          template.CSS `yaml:"icon_color"`
		BaseFontSize       template.CSS `yaml:"base_font_size"`
	}
	ShowHints *bool `yaml:"show_hints"`
	Swipe     *bool
	Navbar    struct {
		Enabled  bool
		Position string
		Style    string
	}
	Screens []Screen
}

type Screen struct {
	Position int

	Navigation *string
	Name       string
	Icon       *string
	Dateclock  struct {
		Enabled       *bool
		Hour12        bool
		CapitaliseDay bool `yaml:"capitalise_day"`
		ShowSeconds   bool `yaml:"show_seconds"`
	}
	Widgets []struct {
		EntityID        string `yaml:"entity_id"`
		FontSize        string `yaml:"font_size"` // Per widget override
		InternalBorders *bool  `yaml:"internal_borders"`
		// Weather-specific
		ForecastInterval *ForecastInterval `yaml:"forecast_interval"`
		ForecastTimes    *int              `yaml:"forecast_times"`
		// Media-specific
		ShowVolume *bool
		ShowAlbum  *bool
	}
	Sensors []struct {
		EntityID string `yaml:"entity_id"`
		Label    string
		Unit     string
	}
	Entities []Entity
	Order    struct {
		Entities int
		Widgets  int
		Sensors  int
	}
}

type Entity struct {
	EntityID string `yaml:"entity_id"`
	Label    string
	Icon     string
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

type CardTheme struct {
	Borders      *bool
	BorderColor  template.CSS `yaml:"border_color"`
	BorderRadius template.CSS `yaml:"border_radius"`
	Background   template.CSS
	FontSize     template.CSS `yaml:"font_size"`
}
type Config struct {
	Localization struct {
		Locale   string
		Timezone string
	}
	FullyKiosk struct {
		ScreensaverTimeout int `yaml:"screensaver_timeout"`
	} `yaml:"fully_kiosk"`
	HomeAssistant struct {
		URL   string
		TOKEN string
	} `yaml:"home_assistant"`
}

type Yaml struct {
	Dashboards map[string]Dashboard
	Config     `yaml:",inline"`
	IsDev      bool
}

type TemplateData struct {
	Dashboard
	Config
	IsDev bool
}

func check(e error, message string, v ...any) {
	if e != nil {
		log.Print(e)
		return
	}
	log.Printf(message, v...)
}

func domain(entityID string) string {
	parts := strings.SplitN(entityID, ".", 2)
	if len(parts) < 2 {
		return ""
	}
	return parts[0]
}

func BuildDash() {
	cfg, err := parseYaml()
	if err != nil {
		log.Printf("Could not load config when building dashboard")
		return
	}

	var tmpl *template.Template

	funcMap := template.FuncMap{
		"default": func(def any, val any) any {
			if val == nil {
				return def
			}
			v := reflect.ValueOf(val)

			if v.Kind() == reflect.Ptr && v.IsNil() {
				return def
			}
			if v.Kind() == reflect.String && v.String() == "" {
				return def
			}
			if v.Kind() == reflect.Int && v.Int() == 0 {
				return def
			}

			return val
		},
		"css": func(val any) template.CSS {
			return template.CSS(fmt.Sprintf("%v", val))
		},
		"mergeTheme": func(specific CardTheme, base CardTheme) CardTheme {
			result := specific
			if result.BorderColor == "" {
				result.BorderColor = base.BorderColor
			}
			if result.Background == "" {
				result.Background = base.Background
			}
			if result.FontSize == "" {
				result.FontSize = base.FontSize
			}
			if result.Borders == nil {
				result.Borders = base.Borders
			}
			return result
		},
		"enabledByDefault": func(v *bool) bool {
			if v == nil {
				return true
			}
			return *v
		},
		"disabledByDefault": func(v *bool) bool {
			if v == nil {
				return false
			}
			return *v
		},
		"dict": func(values ...any) map[string]any {
			m := map[string]any{}
			for i := 0; i < len(values); i += 2 {
				key := values[i].(string)
				m[key] = values[i+1]
			}
			return m
		},
		"domainIn": func(entityID string, domains ...string) bool {
			domain := domain(entityID)
			for _, d := range domains {
				if domain == d {
					return true
				}
			}
			return false
		},
		"anyOfIn": func(anyOf []string, in ...string) bool { // O(n) is n^2 but the lists are tiny so its fine
			for _, is := range anyOf {
				for _, of := range in {
					if is == of {
						return true
					}
				}
			}
			return false
		},
		"domain": domain,
		"domains": func(entityIDs []string) []string {
			var out []string
			for _, id := range entityIDs {
				out = append(out, domain(id))
			}
			return out
		},
		"entityIDs": func(entities []Entity) []string {
			out := make([]string, len(entities))
			for i, e := range entities {
				out[i] = e.EntityID
			}
			return out
		},
		"icon": func(name string) template.HTML {
			path := mdiIcons[name]
			return template.HTML(fmt.Sprintf(
				`<svg class="icon" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg"><path d="%s"/></svg>`,
				path,
			))
		},
		"isEmoji": func(s string) bool {
			for _, r := range s {
				return r > 127
			}
			return false
		},
		"json": func(j interface{}) string { // For debugging
			var out []byte
			out, err = json.Marshal(j)
			if err != nil {
				return ""
			}
			return string(out)
		},
		"merge": func(maps ...any) (map[string]any, error) {
			result := map[string]any{}
			for _, m := range maps {
				switch v := m.(type) {
				case map[string]any:
					for k, val := range v {
						result[k] = val
					}
				default:
					rv := reflect.ValueOf(m)
					if rv.Kind() == reflect.Ptr {
						rv = rv.Elem()
					}
					if rv.Kind() != reflect.Struct {
						return nil, fmt.Errorf("merge: unsupported type %T", m)
					}
					rt := rv.Type()
					for i := 0; i < rv.NumField(); i++ {
						f := rt.Field(i)
						if f.IsExported() {
							result[f.Name] = rv.Field(i).Interface()
						}
					}
				}
			}
			return result, nil
		},
		"prevScreen": func(d Dashboard, i int) *Screen {
			if i > 0 {
				return &d.Screens[i-1]
			}
			return nil
		},
		"nextScreen": func(d Dashboard, i int) *Screen {
			if i+1 < len(d.Screens) {
				return &d.Screens[i+1]
			}
			return nil
		},
	}

	tmpl, err = template.New("").Funcs(funcMap).ParseGlob(frontendPath + "/templates/*.html.tmpl")
	if err != nil {
		log.Printf("Could not return root level templates %v", err)
		return
	}

	tmpl, err = tmpl.ParseGlob(frontendPath + "/templates/css/*.html.tmpl")
	tmpl, err = tmpl.ParseGlob(frontendPath + "/templates/css/*.css.tmpl")
	tmpl, err = tmpl.ParseGlob(frontendPath + "/templates/entities/*.html.tmpl")
	tmpl, err = tmpl.ParseGlob(frontendPath + "/templates/widgets/*.html.tmpl")
	tmpl, err = tmpl.ParseGlob(frontendPath + "/templates/navbar-styles/*.html.tmpl")
	check(err, "Created template object")

	for name, dash := range cfg.Dashboards {
		outputDir := frontendPath + "/static/" + name
		err = os.MkdirAll(outputDir, 0755)
		check(err, "Created %s", outputDir)
		out, err := os.Create(outputDir + "/index.html")
		check(err, "Created/opened output file")
		defer out.Close()
		// fmt.Printf("Parsed yaml:\n%+v\n", cfg)
		data := TemplateData{dash, cfg.Config, isDev}
		// fmt.Printf("Template data:\n%+v\n", data)
		err = tmpl.ExecuteTemplate(out, "main.html.tmpl", data)
		out.Sync()
		check(err, "Template executed")
	}
}

func parseYaml() (*Yaml, error) {
	yaml_file, err := os.ReadFile(yamlPath)
	parsed := Yaml{}
	if err != nil {
		return &parsed, err
	}
	err = yaml.Unmarshal(yaml_file, &parsed)
	return &parsed, err
}

var isDev bool

var yamlPath string
var frontendPath string

func init() {
	godotenv.Load()

	isDev = os.Getenv("IS_DEV") == "true"

	if isDev {
		yamlPath = "../hoarydash.dev.yaml"
		frontendPath = "../frontend"
	} else {
		yamlPath = "/config/hoarydash.yaml"
		frontendPath = "/app/frontend"
	}

	json.Unmarshal(mdiData, &mdiIcons)

	log.Printf("isDev=%v yamlPath=%s frontendPath=%s", isDev, yamlPath, frontendPath)
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "4567"
	}

	log.Printf("Serving on :%s", port)

	yamlWatcher := watcher.New()
	yamlWatcher.SetMaxEvents(1)
	yamlWatcher.FilterOps(watcher.Write)
	yamlWatcher.AddRecursive(yamlPath)
	defer yamlWatcher.Close()

	rebuildChan := make(chan struct{})

	go func() {
		for {
			select {
			case event := <-yamlWatcher.Event:
				fmt.Println(event)
				BuildDash()
				rebuildChan <- struct{}{}
			case err := <-yamlWatcher.Error:
				log.Fatalln(err)
			case <-yamlWatcher.Closed:
				return
			}
		}
	}()

	BuildDash()

	cfg, err := parseYaml()
	check(err, "Config loaded successfully")
	log.Printf("Config is: %v", cfg)
	go yamlWatcher.Start(1 * time.Second)

	fs := http.FileServer(http.Dir(frontendPath + "/static"))
	http.Handle("/", fs)
	http.HandleFunc("/api/ws", wsProxyHandler(cfg.HomeAssistant.URL, cfg.HomeAssistant.TOKEN, rebuildChan))
	http.HandleFunc("/api/translations/{widget}/{lang}", translationsHandler())
	http.HandleFunc("/api/media_cover", mediaCoverHandler(cfg.HomeAssistant.URL, cfg.HomeAssistant.TOKEN))
	log.Print("Starting server on http://localhost:" + port)
	log.Fatal(http.ListenAndServe("0.0.0.0:"+port, nil))
}

func getHaDefaults(baseUrl string, token string) (string, string) {
	if baseUrl == "" {
		log.Print("HA url not set, defaulting to 'http://homeassistant.local:8123'")
		baseUrl = "http://homeassistant.local:8123"
	}

	if token == "" {
		log.Print("Getting HA token fron environment")
		envToken := os.Getenv("HA_TOKEN")
		if envToken == "" {
			log.Printf("No HA token could be read")
			return baseUrl, ""
		}
		token = envToken
	}
	return baseUrl, token
}
