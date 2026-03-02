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
var frontendTemplatesPath string = "/app/frontend/templates"

func check(e error, message string, v ...any) {
	if e != nil {
		panic(e)
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

func templBuilder() {
	data := readConfig()

	outputFile, err := os.Create(frontendTemplatesPath + "/generated.html")
	check(err, "Created/opened output file")
	defer outputFile.Close()

	tmpl, err := template.ParseFiles(frontendTemplatesPath + "/index.html")
	check(err, "Created template object")

	err = tmpl.Execute(outputFile, data)
	outputFile.Sync()
	check(err, "Template executed")
}

func readConfig() Config {
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
		frontendTemplatesPath = strings.ReplaceAll(frontendTemplatesPath, "/app", "..")
		log.Printf("paths are now:\n%s\n%s", yamlPath, frontendTemplatesPath)
	}
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "4567"
	}
	fs := http.FileServer(http.Dir(frontendTemplatesPath))
	// fs := http.FileServer(http.Dir("app/frontend/templates"))

	http.Handle("/", fs)
	http.HandleFunc("/api", helloHandler)

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
				templBuilder()
			case err := <-yamlWatcher.Error:
				log.Fatalln(err)
			case <-yamlWatcher.Closed:
				return
			}
		}
	}()

	templBuilder()

	yamlWatcher.Start(1 * time.Second)
	// http.HandleFunc("/template", templHandler(config))

	log.Fatal(http.ListenAndServe("0.0.0.0:"+port, nil))
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(map[string]string{"message": "Hello, World!"})
}
