package database

import (
	"log/slog"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type SDatabase struct {
	db     *sqlx.DB
	logger *slog.Logger
	opts   Options
}

type Base interface {
	Connect() error
	Close()
	HealthCheck() error
}

func (d *SDatabase) Connect() error {
	var (
		err error
	)

	driver := d.opts.Driver
	if driver == "" {
		driver = "postgres"
	}

	if d.opts.DSN == "" {
		return ErrMissingDSN
	}

	d.db, err = sqlx.Open(driver, d.opts.DSN)
	if err != nil {
		return err
	}

	maxOpen := d.opts.MaxOpenConns
	if maxOpen <= 0 {
		maxOpen = 20
	}
	maxIdle := d.opts.MaxIdleConns
	if maxIdle <= 0 {
		maxIdle = 10
	}
	lifetime := d.opts.ConnMaxLifetime
	if lifetime <= 0 {
		lifetime = 30 * time.Minute
	}

	d.db.SetMaxOpenConns(maxOpen)
	d.db.SetMaxIdleConns(maxIdle)
	d.db.SetConnMaxLifetime(lifetime)
	return nil
}

func (d *SDatabase) Close() {
	d.db.Close()
}

func (d *SDatabase) HealthCheck() error {
	return d.db.Ping()
}
