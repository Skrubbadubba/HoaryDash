package main

import (
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

    log.Printf("Serving on :%s", port)
    log.Fatal(http.ListenAndServe(":"+port, nil))
}
