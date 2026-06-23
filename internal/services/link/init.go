package link

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/lunyashon/filterphone/internal/database"
	"github.com/lunyashon/filterphone/internal/lib/auth"
	"github.com/lunyashon/filterphone/internal/lib/structure"
	"google.golang.org/grpc/codes"
)

type Link struct {
	cfg *structure.Config
	log *slog.Logger
	db  database.LinkProvider
}

func GetInstance(
	server *gin.Engine,
	cfg *structure.Config,
	log *slog.Logger,
	db database.LinkProvider,
) {
	link := &Link{
		cfg: cfg,
		log: log,
		db:  db,
	}

	server.GET("/api/v1/link.get", link.GetLinks)
	server.POST("/api/v1/link.set", link.SetLink)
	server.DELETE("/api/v1/link.delete", link.DeleteLink)
}

func (l *Link) GetLinks(c *gin.Context) {
	valid, err := auth.ValidateToken(c, l.cfg)
	if err != nil {
		c.JSON(structure.Status[codes.Unauthenticated], gin.H{"error": err.Error()})
		return
	}
	if !valid {
		c.JSON(structure.Status[codes.Unauthenticated], gin.H{"error": "unauthorized"})
		return
	}
	links, err := l.db.GetLinks(c.Request.Context())
	if err != nil {
		c.JSON(structure.Status[codes.Internal], gin.H{"error": err.Error()})
		return
	}
	c.JSON(structure.Status[codes.OK], gin.H{"links": links})
}

type SetLinkRequest struct {
	Link string `json:"link"`
}

func (l *Link) SetLink(c *gin.Context) {
	valid, err := auth.ValidateToken(c, l.cfg)
	if err != nil {
		c.JSON(structure.Status[codes.Unauthenticated], gin.H{"error": err.Error()})
		return
	}
	if !valid {
		c.JSON(structure.Status[codes.Unauthenticated], gin.H{"error": "unauthorized"})
		return
	}
	var request SetLinkRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(structure.Status[codes.InvalidArgument], gin.H{"error": err.Error()})
		return
	}
	if err := l.db.SetLink(c.Request.Context(), request.Link); err != nil {
		c.JSON(structure.Status[codes.Internal], gin.H{"error": err.Error()})
		return
	}
	c.JSON(structure.Status[codes.OK], gin.H{"message": "link set"})
}

type DeleteLinkRequest struct {
	Link string `json:"link"`
}

func (l *Link) DeleteLink(c *gin.Context) {
	valid, err := auth.ValidateToken(c, l.cfg)
	if err != nil {
		c.JSON(structure.Status[codes.Unauthenticated], gin.H{"error": err.Error()})
		return
	}
	if !valid {
		c.JSON(structure.Status[codes.Unauthenticated], gin.H{"error": "unauthorized"})
		return
	}
	var request DeleteLinkRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(structure.Status[codes.InvalidArgument], gin.H{"error": err.Error()})
		return
	}
	if err := l.db.DeleteLink(c.Request.Context(), request.Link); err != nil {
		c.JSON(structure.Status[codes.Internal], gin.H{"error": err.Error()})
		return
	}
	c.JSON(structure.Status[codes.OK], gin.H{"message": "link deleted"})
}
