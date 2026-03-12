package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/joho/godotenv"

	"github.com/MikeS071/archonhq/apps/api/internal/httpserver"
	"github.com/MikeS071/archonhq/pkg/config"
	"github.com/MikeS071/archonhq/pkg/db"
	"github.com/MikeS071/archonhq/pkg/events"
	natsclient "github.com/MikeS071/archonhq/pkg/nats"
	"github.com/MikeS071/archonhq/pkg/objectstore"
	redisclient "github.com/MikeS071/archonhq/pkg/redis"
	"github.com/MikeS071/archonhq/pkg/telemetry"
)

func main() {
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	logger := telemetry.NewLogger()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pg, err := db.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("postgres init error: %v", err)
	}
	defer pg.Close()

	nc, err := natsclient.Connect(cfg.NATSURL)
	if err != nil {
		log.Fatalf("nats init error: %v", err)
	}
	defer nc.Close()

	rc, err := redisclient.Connect(cfg.RedisURL)
	if err != nil {
		log.Fatalf("redis init error: %v", err)
	}

	oc, err := objectstore.New(objectstore.Config{
		Endpoint:  cfg.S3Endpoint,
		AccessKey: cfg.S3AccessKey,
		SecretKey: cfg.S3SecretKey,
		Bucket:    cfg.S3Bucket,
		Region:    cfg.S3Region,
	})
	if err != nil {
		log.Fatalf("object store init error: %v", err)
	}

	server := httpserver.New(
		logger,
		pg,
		nc,
		rc,
		oc,
		events.NewPostgresStore(pg.DB),
	)

	addr := ":" + cfg.Port
	logger.Info("starting api server", "addr", addr, "env", cfg.AppEnv)
	if err := http.ListenAndServe(addr, server.Handler()); err != nil {
		log.Fatalf("api server stopped: %v", err)
	}
}
