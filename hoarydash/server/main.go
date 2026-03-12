package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/radovskyb/watcher"
	"go.yaml.in/yaml/v4"
)

type Dashboard struct {
	Nightlight struct {
		Enabled        bool
		OverrideColors bool `yaml:"override_colors"`
	}
	Dateclock struct {
		Enabled       *bool
		Hour12        bool
		CapitaliseDay bool `yaml:"capitalise_day"`
		ShowSeconds   bool `yaml:"show_seconds"`
	}
	Sensors []struct {
		EntityID string `yaml:"entity_id"`
		Label    string
		Unit     string
	}
	Entities []struct {
		EntityID string `yaml:"entity_id"`
		Label    string
		Icon     string
	}
	Theme struct {
		BodyBackground     template.CSS `yaml:"body_background"`
		BackgroundGradient template.CSS `yaml:"background_gradient"`
		ButtonBackground   template.CSS `yaml:"button_background"`
		CardBackground     template.CSS `yaml:"card_background"`
		FontColor          template.CSS `yaml:"font_color"`
		SecondaryFontColor template.CSS `yaml:"secondary_font_color"`
		IconColor          template.CSS `yaml:"icon_color"`
	}
}
type Config struct {
	Localization struct {
		Locale   string
		Timezone string
	}
	FullyKiosk struct {
		Password           string `yaml:"password"`
		ScreensaverTimeout int    `yaml:"screensaver_timeout"`
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

func BuildDash() {
	cfg, err := parseYaml()
	if err != nil {
		log.Printf("Could not load config when building dashboard")
		return
	}

	funcMap := template.FuncMap{
		"default": func(def template.CSS, val template.CSS) template.CSS {
			if val == "" {
				return def
			}
			return val
		},
		"enabledByDefault": func(v *bool) bool {
			if v == nil {
				return true
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
			parts := strings.SplitN(entityID, ".", 2)
			if len(parts) < 2 {
				return false
			}
			domain := parts[0]
			for _, d := range domains {
				if domain == d {
					return true
				}
			}
			return false
		},
	}

	tmpl, err := template.New("").Funcs(funcMap).ParseGlob(frontendPath + "/templates/*.html.tmpl")
	if err != nil {
		log.Printf("Could not return root level templates %v", err)
		return
	}

	tmpl, err = tmpl.ParseGlob(frontendPath + "/templates/css/*.html.tmpl")
	tmpl, err = tmpl.ParseGlob(frontendPath + "/templates/entities/*.html.tmpl")
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
	// log.Printf("Config is: %v", cfg)
	go yamlWatcher.Start(1 * time.Second)

	fs := http.FileServer(http.Dir(frontendPath + "/static"))
	http.Handle("/", fs)
	http.HandleFunc("/api/ws", wsProxyHandler(cfg.HomeAssistant.URL, cfg.HomeAssistant.TOKEN, rebuildChan))
	log.Print("Starting server on http://localhost:" + port)
	log.Fatal(http.ListenAndServe("0.0.0.0:"+port, nil))
}
