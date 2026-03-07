package main

import (
	// "bufio"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/radovskyb/watcher"
	"go.yaml.in/yaml/v4"
)

type Config struct {
	Dashboard struct {
		Nightlight bool
	}
}

var yamlPath string = "/app/config/test.yaml"
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
	cfg := loadConfig()

	out, err := os.Create(frontendPath + "/static/dash.html")
	check(err, "Created/opened output file")
	defer out.Close()

	tmpl, err := template.ParseGlob(frontendPath + "/templates/*.html.tmpl")
	check(err, "Created template object")

	err = tmpl.ExecuteTemplate(out, "dash.html.tmpl", cfg)
	out.Sync()
	check(err, "Template executed")
}

func loadConfig() Config {
	config_file, err := os.ReadFile(yamlPath)
	config := Config{}
	err = yaml.Unmarshal(config_file, &config)
	check(err, "Read yaml file: %v", config)
	return config
}

func init() {
	log.Printf("init ran")
	isDev := os.Getenv("IS_DEV") == "true"
	if isDev {
		log.Print("is dev")
		yamlPath = strings.ReplaceAll(yamlPath, "/app", "..")
		frontendPath = strings.ReplaceAll(frontendPath, "/app", "..")
		log.Printf("paths are now:\n%s\n%s", yamlPath, frontendPath)
	}
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
				fmt.Println(event) // Print the event's info.
				BuildDash()
			case err := <-yamlWatcher.Error:
				log.Fatalln(err)
			case <-yamlWatcher.Closed:
				return
			}
		}
	}()

	BuildDash()

	go yamlWatcher.Start(1 * time.Second)

	fs := http.FileServer(http.Dir(frontendPath + "/static"))
	http.Handle("/", fs)
	http.HandleFunc("/api", helloHandler)
	log.Print("Starting server on http://localhost:" + port)
	log.Fatal(http.ListenAndServe("0.0.0.0:"+port, nil))
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(map[string]string{"message": "Hello, World!"})
}
