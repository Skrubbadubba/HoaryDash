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

func check(e error, message string, v ...any) {
	if e != nil {
		log.Print(e)
		return
	}
	log.Printf(message, v...)
}

var isDev bool

var yamlPath string
var frontendPath string

func init() {
	godotenv.Load()

	isDev = os.Getenv("IS_DEV") == "true"

	if isDev {
		yamlPath = "../config"
		frontendPath = "../frontend"
	} else {
		yamlPath = "/config"
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

	cfg, err := parseConfig()
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
