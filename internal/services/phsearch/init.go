package phsearch

import (
	"database/sql"
	"errors"
	"log/slog"
	"regexp"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/lunyashon/filterphone/internal/database"
	"github.com/lunyashon/filterphone/internal/lib/auth"
	"github.com/lunyashon/filterphone/internal/lib/structure"
	"google.golang.org/grpc/codes"
)

type PhSearch struct {
	cfg *structure.Config
	log *slog.Logger
	db  database.NumbersProvider
}

func GetInstance(
	server *gin.Engine,
	cfg *structure.Config,
	log *slog.Logger,
	db database.NumbersProvider,
) {
	phsearch := &PhSearch{
		cfg: cfg,
		log: log,
		db:  db,
	}

	server.GET("/api/v1/phone.search", phsearch.GetPhSearch)
	server.POST("/api/v1/phone.search", phsearch.GetPhSearch)
}

type PhSearchRequest struct {
	Phone string `json:"phone"`
}

func (phsearch *PhSearch) GetPhSearch(c *gin.Context) {
	valid, err := auth.ValidateToken(c, phsearch.cfg)
	if err != nil {
		c.JSON(structure.Status[codes.Unauthenticated], gin.H{"error": err.Error()})
		return
	}
	if !valid {
		c.JSON(structure.Status[codes.Unauthenticated], gin.H{"error": "unauthorized"})
		return
	}

	var request PhSearchRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(structure.Status[codes.InvalidArgument], gin.H{"error": err.Error()})
		return
	}

	abc, tail, err := ParsingPhone(request.Phone)
	if err != nil {
		c.JSON(structure.Status[codes.InvalidArgument], gin.H{"error": err.Error()})
		return
	}

	number, err := phsearch.db.GetNumbers(c.Request.Context(), int16(abc), tail)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(structure.Status[codes.NotFound], gin.H{"error": "phone not found", "phone": request.Phone})
			return
		}
		c.JSON(structure.Status[codes.Internal], gin.H{"error": "internal server error"})
		return
	}

	c.JSON(structure.Status[codes.OK], gin.H{"number": number})

}

func ParsingPhone(phone string) (int, int, error) {
	nonDigits := regexp.MustCompile(`\D+`)
	digitsOnly := nonDigits.ReplaceAllString(phone, "")

	switch {
	case len(digitsOnly) == 11 && (digitsOnly[0] == '7' || digitsOnly[0] == '8'):
		digitsOnly = digitsOnly[1:]
	case len(digitsOnly) == 10:
	default:
		return 0, 0, errors.New("invalid phone: expected 10 digits (or 11 starting with 7/8)")
	}

	abcStr := digitsOnly[0:3]
	tailStr := digitsOnly[3:]

	abc, err := strconv.Atoi(abcStr)
	if err != nil {
		return 0, 0, err
	}
	tail, err := strconv.Atoi(tailStr)
	if err != nil {
		return 0, 0, err
	}
	return abc, tail, nil
}
