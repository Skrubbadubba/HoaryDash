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

type Config struct {
	Dashboard struct {
		Nightlight bool
		Sensors    []struct {
			EntityID string `yaml:"entity_id"`
			Label    string
			Unit     string
		}
		Theme struct {
			BodyBackground   string `yaml:"body_background"`
			ButtonBackground string `yaml:"button_background"`
			FontColor        string `yaml:"font_color"`
		}
	}
	Localization struct {
		Locale        string
		Timezone      string
		Hour12        bool
		CapitaliseDay bool `yaml:"capitalise_day"`
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

type TemplateData struct {
	Config
	IsDev bool
}

var yamlPath string = "/app/config/"
var frontendPath string = "/app/frontend"

func check(e error, message string, v ...any) {
	if e != nil {
		log.Print(e)
	}
	log.Printf(message, v...)
}

func BuildDash() {
	cfg, err := loadConfig()
	if err != nil {
		log.Printf("Could not load config when building dashboard")
		return
	}

	data := TemplateData{*cfg, isDev}

	out, err := os.Create(frontendPath + "/static/dash.html")
	check(err, "Created/opened output file")
	defer out.Close()

	funcMap := template.FuncMap{
		"default": func(def string, val interface{}) template.CSS {
			s, ok := val.(string)
			if !ok || s == "" {
				return template.CSS(def)
			}
			return template.CSS(s)
		},
	}
	tmpl, err := template.New("").Funcs(funcMap).ParseGlob(frontendPath + "/templates/*.html.tmpl")
	if err != nil {
		log.Printf("Could not return root level templates %v", err)
		return
	}
	tmpl, err = tmpl.ParseGlob(frontendPath + "/templates/css/*.html.tmpl")
	check(err, "Created template object")

	err = tmpl.ExecuteTemplate(out, "main.html.tmpl", data)
	out.Sync()
	check(err, "Template executed")
}

func loadConfig() (*Config, error) {
	config_file, err := os.ReadFile(yamlPath)
	config := Config{}
	if err != nil {
		return &config, err
	}
	err = yaml.Unmarshal(config_file, &config)
	return &config, err
}

var isDev bool

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Print("Error loading .env file")
	}

	isDev = os.Getenv("IS_DEV") == "true"
	yamlFilename := "config.yaml"
	if isDev {
		log.Print("is dev")
		yamlPath = strings.ReplaceAll(yamlPath, "/app", "..")
		yamlFilename = "dev.yaml"

		frontendPath = strings.ReplaceAll(frontendPath, "/app", "..")
		log.Printf("paths are now:\n%s\n%s", yamlPath, frontendPath)
	}

	yamlPath += yamlFilename
	pwd, _ := os.Getwd()
	log.Printf("Paths are now %s and %s\n process running at: %s", yamlPath, frontendPath, pwd)
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

	go func() {
		for {
			select {
			case event := <-yamlWatcher.Event:
				fmt.Println(event)
				BuildDash()
			case err := <-yamlWatcher.Error:
				log.Fatalln(err)
			case <-yamlWatcher.Closed:
				return
			}
		}
	}()

	BuildDash()

	cfg, err := loadConfig()
	check(err, "Config loaded successfully")
	// log.Printf("Config is: %v", cfg)
	go yamlWatcher.Start(1 * time.Second)

	fs := http.FileServer(http.Dir(frontendPath + "/static"))
	http.Handle("/", fs)
	http.HandleFunc("/api/ws", wsProxyHandler(cfg.HomeAssistant.URL, cfg.HomeAssistant.TOKEN))
	log.Print("Starting server on http://localhost:" + port)
	log.Fatal(http.ListenAndServe("0.0.0.0:"+port, nil))
}
