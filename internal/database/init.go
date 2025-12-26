package database

import (
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lunyashon/filterphone/internal/lib/structure"
)

type Database struct {
	Base    Base
	Numbers NumbersProvider
}

func GetInstance(logger *slog.Logger, cfg *structure.Config) (*Database, error) {
	var (
		err error
	)

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.HostDb,
		cfg.PortDb,
		cfg.LoginDb,
		cfg.PassDb,
		cfg.NameDb,
	)

	sdb := NewWithDSN(logger, dsn, "postgres")
	if err = sdb.Connect(); err != nil {
		return nil, err
	}

	db := &SDatabase{
		db:     sdb.DB(),
		logger: logger,
	}

	return &Database{
		Base:    db,
		Numbers: sdb,
	}, nil
}

type Options struct {
	Driver string
	DSN    string

	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

var ErrMissingDSN = errors.New("database: missing DSN")

func New(logger *slog.Logger, opts Options) *SDatabase {
	return &SDatabase{
		logger: logger,
		opts:   opts,
	}
}

func NewWithDSN(logger *slog.Logger, dsn string, driver string) *SDatabase {
	switch driver {
	case "mysql":
		driver = "mysql"
	case "postgres":
		driver = "postgres"
	default:
		panic(errors.New("invalid driver"))
	}
	return New(logger, Options{
		Driver: driver,
		DSN:    dsn,
	})
}

func (d *SDatabase) DB() *sqlx.DB {
	return d.db
}
