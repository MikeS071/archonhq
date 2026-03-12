package main

import (
	"context"
	"fmt"
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

var (
	loadDotEnv      = godotenv.Load
	loadAppConfig   = config.Load
	newAppLogger    = telemetry.NewLogger
	openPostgres    = db.Open
	connectNATS     = natsclient.Connect
	connectRedis    = redisclient.Connect
	newObjectClient = objectstore.New
	newEventStore   = events.NewPostgresStore
	newHTTPServer   = httpserver.New
	serveHTTP       = http.ListenAndServe
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	_ = loadDotEnv()

	cfg, err := loadAppConfig()
	if err != nil {
		return fmt.Errorf("config error: %w", err)
	}

	logger := newAppLogger()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pg, err := openPostgres(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("postgres init error: %w", err)
	}
	defer func() {
		_ = pg.Close()
	}()

	nc, err := connectNATS(cfg.NATSURL)
	if err != nil {
		return fmt.Errorf("nats init error: %w", err)
	}
	defer nc.Close()

	rc, err := connectRedis(cfg.RedisURL)
	if err != nil {
		return fmt.Errorf("redis init error: %w", err)
	}

	oc, err := newObjectClient(objectstore.Config{
		Endpoint:  cfg.S3Endpoint,
		AccessKey: cfg.S3AccessKey,
		SecretKey: cfg.S3SecretKey,
		Bucket:    cfg.S3Bucket,
		Region:    cfg.S3Region,
	})
	if err != nil {
		return fmt.Errorf("object store init error: %w", err)
	}

	server := newHTTPServer(
		logger,
		pg,
		nc,
		rc,
		oc,
		newEventStore(pg.DB),
	)

	addr := ":" + cfg.Port
	logger.Info("starting api server", "addr", addr, "env", cfg.AppEnv)
	if err := serveHTTP(addr, server.Handler()); err != nil {
		return fmt.Errorf("api server stopped: %w", err)
	}
	return nil
}
