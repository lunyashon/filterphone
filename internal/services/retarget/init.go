package retarget

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/lunyashon/filterphone/internal/database"
	"github.com/lunyashon/filterphone/internal/lib/structure"
)

type Retarget struct {
	cfg        *structure.Config
	log        *slog.Logger
	db         database.RetargetProvider
	dbTemplate database.TemplateProvider
}

func GetInstance(
	server *gin.Engine,
	cfg *structure.Config,
	log *slog.Logger,
	db database.RetargetProvider,
	dbTemplate database.TemplateProvider,
) {
	retarget := &Retarget{
		cfg:        cfg,
		log:        log,
		db:         db,
		dbTemplate: dbTemplate,
	}

	server.POST("/api/v1/retarget.add", retarget.AddArrayRetarget)
	server.GET("/api/v1/retarget.arrays", retarget.GetAllArrayRetarget)
	server.GET("/api/v1/retarget.array/:id_array", retarget.GetItemsArrayRetarget)
	server.DELETE("/api/v1/retarget.array/:id_array", retarget.DeleteArrayRetarget)
	server.GET("/api/v1/retarget.check", retarget.CheckItemRetarget)

	server.POST("/api/v1/retarget.export", retarget.ExportArrayRetarget)
	server.GET("/api/v1/template.export", retarget.GetTemplates)
}
