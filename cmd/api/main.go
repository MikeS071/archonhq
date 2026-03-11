package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq" // Postgres driver
	"github.com/nats-io/nats.go"
)

func main() {
	_ = godotenv.Load() // safe if no .env

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/health", healthHandler)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "ArchonHQ API - stub running in Docker")
	})

	addr := ":" + port
	log.Printf("Starting server on %s", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")

	// Check Postgres
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		http.Error(w, "DATABASE_URL not set", http.StatusServiceUnavailable)
		return
	}
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		http.Error(w, fmt.Sprintf("DB connection failed: %v", err), http.StatusServiceUnavailable)
		return
	}
	defer db.Close()
	if err = db.Ping(); err != nil {
		http.Error(w, fmt.Sprintf("DB ping failed: %v", err), http.StatusServiceUnavailable)
		return
	}

	// Check NATS
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = "nats://nats:4222" // fallback to compose service
	}
	nc, err := nats.Connect(natsURL, nats.Timeout(2*time.Second))
	if err != nil {
		http.Error(w, fmt.Sprintf("NATS connection failed: %v", err), http.StatusServiceUnavailable)
		return
	}
	defer nc.Close()

	fmt.Fprintln(w, "OK - ArchonHQ API healthy")
	fmt.Fprintf(w, "DB connected: %s\n", dbURL)
	fmt.Fprintf(w, "NATS connected: %s\n", natsURL)
}