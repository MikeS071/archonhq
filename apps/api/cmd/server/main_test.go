package main

import (
	"context"
	"database/sql"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"

	"github.com/MikeS071/archonhq/apps/api/internal/httpserver"
	"github.com/MikeS071/archonhq/pkg/config"
	"github.com/MikeS071/archonhq/pkg/db"
	"github.com/MikeS071/archonhq/pkg/events"
	natsclient "github.com/MikeS071/archonhq/pkg/nats"
	"github.com/MikeS071/archonhq/pkg/objectstore"
	redisclient "github.com/MikeS071/archonhq/pkg/redis"
)

func snapshotDeps() (func(...string) error, func() (config.Config, error), func() *slog.Logger, func(context.Context, string) (*db.Postgres, error), func(string) (*natsclient.Client, error), func(string) (*redisclient.Client, error), func(objectstore.Config) (*objectstore.Client, error), func(*sql.DB) *events.PostgresStore, func(*slog.Logger, *db.Postgres, *natsclient.Client, *redisclient.Client, *objectstore.Client, events.Store) *httpserver.Server, func(string, http.Handler) error) {
	return loadDotEnv, loadAppConfig, newAppLogger, openPostgres, connectNATS, connectRedis, newObjectClient, newEventStore, newHTTPServer, serveHTTP
}

func restoreDeps(loadEnv func(...string) error, loadCfg func() (config.Config, error), loggerFn func() *slog.Logger, openDB func(context.Context, string) (*db.Postgres, error), natsFn func(string) (*natsclient.Client, error), redisFn func(string) (*redisclient.Client, error), objFn func(objectstore.Config) (*objectstore.Client, error), evtFn func(*sql.DB) *events.PostgresStore, serverFn func(*slog.Logger, *db.Postgres, *natsclient.Client, *redisclient.Client, *objectstore.Client, events.Store) *httpserver.Server, serveFn func(string, http.Handler) error) {
	loadDotEnv = loadEnv
	loadAppConfig = loadCfg
	newAppLogger = loggerFn
	openPostgres = openDB
	connectNATS = natsFn
	connectRedis = redisFn
	newObjectClient = objFn
	newEventStore = evtFn
	newHTTPServer = serverFn
	serveHTTP = serveFn
}

func TestRunSuccessAndFailurePaths(t *testing.T) {
	origLoadEnv, origLoadCfg, origLogger, origOpenDB, origNATS, origRedis, origObj, origEvent, origServer, origServe := snapshotDeps()
	defer restoreDeps(origLoadEnv, origLoadCfg, origLogger, origOpenDB, origNATS, origRedis, origObj, origEvent, origServer, origServe)

	loadDotEnv = func(...string) error { return nil }
	newAppLogger = func() *slog.Logger { return slog.New(slog.NewTextHandler(io.Discard, nil)) }
	loadAppConfig = func() (config.Config, error) {
		return config.Config{
			AppEnv:      "test",
			Port:        "18080",
			DatabaseURL: "postgres://test",
			NATSURL:     "nats://localhost:4222",
			RedisURL:    "redis://localhost:6379",
			S3Endpoint:  "http://minio:9000",
			S3Bucket:    "archonhq",
		}, nil
	}

	sqlDB, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer sqlDB.Close()

	openPostgres = func(context.Context, string) (*db.Postgres, error) {
		return &db.Postgres{DB: sqlDB}, nil
	}
	connectNATS = func(string) (*natsclient.Client, error) { return &natsclient.Client{}, nil }
	connectRedis = func(string) (*redisclient.Client, error) {
		return &redisclient.Client{URL: "redis://localhost:6379"}, nil
	}
	newObjectClient = func(objectstore.Config) (*objectstore.Client, error) {
		return &objectstore.Client{Config: objectstore.Config{Endpoint: "http://minio", Bucket: "archonhq"}}, nil
	}
	serveHTTP = func(string, http.Handler) error { return nil }

	if err := run(); err != nil {
		t.Fatalf("run success path returned error: %v", err)
	}

	loadAppConfig = func() (config.Config, error) { return config.Config{}, errors.New("bad config") }
	if err := run(); err == nil {
		t.Fatalf("expected config error")
	}

	loadAppConfig = func() (config.Config, error) {
		return config.Config{AppEnv: "test", Port: "18080", DatabaseURL: "postgres://test", NATSURL: "nats://localhost:4222", RedisURL: "redis://localhost:6379", S3Endpoint: "http://minio", S3Bucket: "archonhq"}, nil
	}
	openPostgres = func(context.Context, string) (*db.Postgres, error) { return nil, errors.New("db fail") }
	if err := run(); err == nil {
		t.Fatalf("expected postgres error")
	}

	openPostgres = func(context.Context, string) (*db.Postgres, error) { return &db.Postgres{DB: sqlDB}, nil }
	connectNATS = func(string) (*natsclient.Client, error) { return nil, errors.New("nats fail") }
	if err := run(); err == nil {
		t.Fatalf("expected nats error")
	}

	connectNATS = func(string) (*natsclient.Client, error) { return &natsclient.Client{}, nil }
	connectRedis = func(string) (*redisclient.Client, error) { return nil, errors.New("redis fail") }
	if err := run(); err == nil {
		t.Fatalf("expected redis error")
	}

	connectRedis = func(string) (*redisclient.Client, error) {
		return &redisclient.Client{URL: "redis://localhost:6379"}, nil
	}
	newObjectClient = func(objectstore.Config) (*objectstore.Client, error) { return nil, errors.New("object fail") }
	if err := run(); err == nil {
		t.Fatalf("expected object store error")
	}

	newObjectClient = func(objectstore.Config) (*objectstore.Client, error) {
		return &objectstore.Client{Config: objectstore.Config{Endpoint: "http://minio", Bucket: "archonhq"}}, nil
	}
	serveHTTP = func(string, http.Handler) error { return errors.New("serve fail") }
	if err := run(); err == nil {
		t.Fatalf("expected serve error")
	}
}
