package main

import (
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

	cfg, err := loadConfig()
	check(err, "Could not load config")
	if isDev {
		// log.Printf("Config is: %s", jsonStr(cfg))
	}

	BuildDashFromConfig(cfg)

	go yamlWatcher.Start(1 * time.Second)

	fs := http.FileServer(http.Dir(frontendPath + "/static"))
	for name, dash := range cfg.Dashboards {
		fileMiddleware, wsHandler := handlers(dash.EntityIDs(), cfg.HA.WSURL, cfg.HA.Token, rebuildChan)
		http.Handle("/"+name+"/", fileMiddleware(fs))
		http.HandleFunc("/api/ws/"+name, wsHandler)
	}
	http.Handle("/", fs)
	// http.HandleFunc("/api/ws", wsProxyHandler(cfg.HA, nil, rebuildChan))
	http.HandleFunc("/api/proxy/", haProxyHandler(cfg.HA.HTTPURL, cfg.HA.Token))
	log.Print("Starting server on http://localhost:" + port)
	log.Fatal(http.ListenAndServe("0.0.0.0:"+port, nil))
}
