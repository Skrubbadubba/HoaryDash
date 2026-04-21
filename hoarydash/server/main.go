package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/radovskyb/watcher"
)

var isDev bool

var configPath string
var frontendPath string
var appPath string

// type AddonOptions struct {
// 	Port int `json:"port"`
// }

// func loadAddonOptions() AddonOptions {
// 	defaults := AddonOptions{Port: 4567}
// 	data, err := os.ReadFile("/data/options.json")
// 	if err != nil {
// 		return defaults // not running as addon, use defaults
// 	}
// 	var opts AddonOptions
// 	if err := json.Unmarshal(data, &opts); err != nil {
// 		return defaults
// 	}
// 	return opts
// }

func init() {
	godotenv.Load()

	isDev = os.Getenv("IS_DEV") == "true"

	if isDev {
		configPath = "../config"
		frontendPath = "../frontend"
		appPath = ".."
	} else {
		configPath = "/config"
		frontendPath = "/app/frontend"
		appPath = "/app"
	}

	json.Unmarshal(mdiData, &mdiIcons)

	log.Printf("isDev=%v configPath=%s frontendPath=%s appPath=%s", isDev, configPath, frontendPath, appPath)
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
	yamlWatcher.AddRecursive(configPath)
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

	cfg, err := parseConfig()
	check(err, "Config loaded successfully")
	if isDev {
		// log.Printf("Config is: %s", jsonStr(cfg))
	}
	go yamlWatcher.Start(1 * time.Second)

	fs := http.FileServer(http.Dir(frontendPath + "/static"))
	http.Handle("/", fs)
	http.HandleFunc("/api/ws", wsProxyHandler(cfg.HomeAssistant.URL, cfg.HomeAssistant.TOKEN, rebuildChan))
	http.HandleFunc("/api/translations", translationsHandler())
	http.HandleFunc("/api/media_cover", mediaCoverHandler(cfg.HomeAssistant.URL, cfg.HomeAssistant.TOKEN))
	log.Print("Starting server on http://localhost:" + port)
	log.Fatal(http.ListenAndServe("0.0.0.0:"+port, nil))
}

func defaultHaUrl(baseUrl string) string {
	if baseUrl == "" {
		log.Print("HA url not set, defaulting to 'http://homeassistant.local:8123'")
		return "http://homeassistant.local:8123"
	}
	return baseUrl
}

func defaultHaToken(token string) string {
	if token == "" {
		log.Print("Getting HA token fron environment")
		envToken := os.Getenv("HA_TOKEN")
		if envToken == "" {
			log.Printf("No HA token could be read")
			return ""
		}
		return envToken
	}
	return token
}
func getHaDefaults(baseUrl string, token string) (string, string) {
	return defaultHaUrl(baseUrl), defaultHaToken(token)
}
