package app

import (
	"log/slog"

	"github.com/lunyashon/filterphone/internal/database"
	"github.com/lunyashon/filterphone/internal/lib/structure"
	"github.com/lunyashon/filterphone/internal/transport"
)

func Run(cfg *structure.Config, log *slog.Logger, db *database.Database) {
	log.Info("Starting filterphone", "config", cfg)
	transport.Init(cfg, log, db)
}
