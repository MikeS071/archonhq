package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

type Postgres struct {
	DB *sql.DB
}

func Open(ctx context.Context, dsn string) (*Postgres, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}

	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(20)
	db.SetConnMaxLifetime(30 * time.Minute)

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	return &Postgres{DB: db}, nil
}

func (p *Postgres) Health(ctx context.Context) error {
	return p.DB.PingContext(ctx)
}

func (p *Postgres) Close() error {
	return p.DB.Close()
}
