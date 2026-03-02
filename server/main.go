package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "4567"
	}

	fs := http.FileServer(http.Dir("/app/frontend/static"))
	http.Handle("/", fs)
	http.HandleFunc("/api", helloHandler)

	log.Printf("Serving on :%s", port)
	log.Fatal(http.ListenAndServe("0.0.0.0:"+port, nil))
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(map[string]string{"message": "Hello, World!"})
}
