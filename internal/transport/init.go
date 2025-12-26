package transport

import (
	"fmt"
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/lunyashon/filterphone/internal/database"
	"github.com/lunyashon/filterphone/internal/lib/structure"
	"github.com/lunyashon/filterphone/internal/services/parser"
	"github.com/lunyashon/filterphone/internal/services/phsearch"
)

func Init(cfg *structure.Config, log *slog.Logger, db *database.Database) {
	server := gin.Default()

	parser.GetInstance(server, cfg, log, db.Numbers)
	phsearch.GetInstance(server, cfg, log, db.Numbers)

	server.Run(fmt.Sprintf(":%s", cfg.TcpPort))
}
