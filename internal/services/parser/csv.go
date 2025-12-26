package parser

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/lunyashon/filterphone/internal/database"
	"github.com/lunyashon/filterphone/internal/lib/structure"
)

type CSVParser struct {
	cfg     *structure.Config
	log     *slog.Logger
	numbers database.NumbersProvider
}

func GetInstance(
	server *gin.Engine,
	cfg *structure.Config,
	log *slog.Logger,
	numbers database.NumbersProvider,
) {
	csp := &CSVParser{
		cfg:     cfg,
		log:     log,
		numbers: numbers,
	}

	server.POST("/api/v1/csv.filter", csp.parseCsv)
	server.POST("/api/v1/csv.export", csp.ExportCsv)

	server.POST("/api/v1/csv.restore", csp.Restore)
	server.GET("/api/v1/csv.restore", csp.Restore)
}
