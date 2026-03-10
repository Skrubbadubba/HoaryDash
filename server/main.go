package main

import (
	// "bufio"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
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

func templHandler(data Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tmpl := template.Must(template.ParseFiles("../frontend/templates/index.html"))

		if err := tmpl.Execute(w, data); err != nil {
			panic(err)
		}
	}
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
	http.HandleFunc("/api", helloHandler)
	http.HandleFunc("/api/kiosk/screensaver/toggle", kioskToggleHandler(cfg))
	http.HandleFunc("/api/ws", wsProxyHandler(cfg.HomeAssistant.URL, cfg.HomeAssistant.TOKEN))
	log.Print("Starting server on http://localhost:" + port)
	log.Fatal(http.ListenAndServe("0.0.0.0:"+port, nil))
}

var kioskState = struct {
	sync.Mutex
	screensaverDisabled bool
	originalTimeout     string
}{}

func kioskToggleHandler(cfg *Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			http.Error(w, "could not determine client IP", http.StatusInternalServerError)
			return
		}
		if isDev {
			host = "192.168.1.141"
			log.Printf("host hardcoded to 192.168.1.141")
		}
		log.Printf("Got %s request from %s", r.Method, host)
		password := cfg.FullyKiosk.Password

		kioskState.Lock()
		defer kioskState.Unlock()

		if !kioskState.screensaverDisabled {
			// --- Toggle ON: read current value, then set to 0 ---
			original, err := getKioskSetting(host, password, "timeToScreensaverV2")
			if err != nil {
				log.Printf("getSettings failed, using fallback: %v", err)
				original = fmt.Sprintf("%d", cfg.FullyKiosk.ScreensaverTimeout)
			}
			kioskState.originalTimeout = original
			if err := setKioskSetting(host, password, "timeToScreensaverV2", "0"); err != nil {
				http.Error(w, "failed to set screensaver", http.StatusBadGateway)
				return
			}
			kioskState.screensaverDisabled = true
			log.Printf("screensaver disabled (original: %s)", original)
		} else {
			// --- Toggle OFF: restore ---
			if err := setKioskSetting(host, password, "timeToScreensaverV2", kioskState.originalTimeout); err != nil {
				http.Error(w, "failed to restore screensaver", http.StatusBadGateway)
				return
			}
			kioskState.screensaverDisabled = false
			log.Printf("screensaver restored to %s", kioskState.originalTimeout)
		}

		w.WriteHeader(http.StatusOK)
	}
}

func getKioskSetting(host, password, key string) (string, error) {
	url := fmt.Sprintf("http://%s:2323/?cmd=listSettings&type=json&password=%s", host, password)
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var settings map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&settings); err != nil {
		return "", err
	}
	val, ok := settings[key]
	if !ok {
		return "", fmt.Errorf("key %s not found in settings", key)
	}
	return fmt.Sprintf("%v", val), nil
}

func setKioskSetting(host, password, key, value string) error {
	url := fmt.Sprintf("http://%s:2323/?cmd=setStringSetting&key=%s&value=%s&password=%s",
		host, key, value, password)
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)
	return nil
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(map[string]string{"message": "Hello, World!"})
}
