package main

import (
    "fmt"
    "log"
    "net/http"
    "os"

    "github.com/joho/godotenv"
)

func main() {
    _ = godotenv.Load() // optional for local .env

    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }

    http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        fmt.Fprintln(w, "OK - ArchonHQ API healthy")
    })

    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintf(w, "ArchonHQ API stub\n")
    })

    addr := ":" + port
    log.Printf("API listening on %s", addr)
    log.Fatal(http.ListenAndServe(addr, nil))
}
