package parser

import (
	"bytes"
	"encoding/csv"

	"github.com/gin-gonic/gin"
	"github.com/lunyashon/filterphone/internal/lib/auth"
	"github.com/lunyashon/filterphone/internal/lib/structure"
	"google.golang.org/grpc/codes"
)

type CSVFilterRequest struct {
	Phones    []string `json:"phones"`
	Operators []string `json:"operators"`
	Regions   []string `json:"regions"`
}

func (csp *CSVParser) ExportCsv(c *gin.Context) {
	valid, err := auth.ValidateToken(c, csp.cfg)
	if err != nil {
		c.JSON(structure.Status[codes.Unauthenticated], gin.H{"error": err.Error()})
		return
	}
	if !valid {
		c.JSON(structure.Status[codes.Unauthenticated], gin.H{"error": "unauthorized"})
		return
	}

	var request CSVFilterRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(structure.Status[codes.InvalidArgument], gin.H{"error": err.Error()})
		return
	}

	var (
		tr            = []string{"Телефон"}
		haveOperators = false
		haveRegions   = false
	)

	if len(request.Phones) == 0 {
		c.JSON(structure.Status[codes.InvalidArgument], gin.H{"error": "phones is required"})
		return
	}
	if len(request.Operators) != 0 {
		tr = append(tr, "Оператор")
		haveOperators = true
	}
	if len(request.Regions) != 0 {
		tr = append(tr, "Регион")
		haveRegions = true
	}

	var buffer bytes.Buffer
	buffer.Write([]byte{0xEF, 0xBB, 0xBF})
	w := csv.NewWriter(&buffer)
	w.Comma = ';'

	if err := w.Write(tr); err != nil {
		c.JSON(structure.Status[codes.Internal], gin.H{"error": err.Error()})
		return
	}
	for key, phone := range request.Phones {
		row := []string{phone}
		if haveOperators {
			row = append(row, request.Operators[key])
		}
		if haveRegions {
			row = append(row, request.Regions[key])
		}
		if err := w.Write(row); err != nil {
			continue
		}
	}
	w.Flush()
	if err := w.Error(); err != nil {
		c.JSON(structure.Status[codes.Internal], gin.H{"error": err.Error()})
		return
	}
	c.Header("Content-Disposition", "attachment; filename=filtered_phones.csv")
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Data(structure.Status[codes.OK], "text/csv; charset=utf-8", buffer.Bytes())
}
