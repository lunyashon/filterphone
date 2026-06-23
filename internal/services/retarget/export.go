package retarget

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/lunyashon/filterphone/internal/lib/auth"
	"github.com/lunyashon/filterphone/internal/lib/structure"
	"google.golang.org/grpc/codes"
)

type RetargetExportRequest struct {
	Name     string `json:"name_array"`
	Template string `json:"template"`
	Items    []struct {
		Phone     string `json:"phone"`
		Retarget  string `json:"retarget"`
		Operator  string `json:"operator"`
		Territory string `json:"territory"`
		TargetUrl string `json:"target_url"`
		TTL       string `json:"ttl"`
		CountLink int    `json:"count_link"`
	} `json:"items"`
}

func (r *Retarget) ExportArrayRetarget(c *gin.Context) {
	valid, err := auth.ValidateToken(c, r.cfg)
	if err != nil {
		c.JSON(structure.Status[codes.Unauthenticated], gin.H{"error": err.Error()})
		return
	}
	if !valid {
		c.JSON(structure.Status[codes.Unauthenticated], gin.H{"error": "unauthorized"})
		return
	}

	var req RetargetExportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(structure.Status[codes.InvalidArgument], gin.H{"error": err.Error()})
		return
	}

	if req.Template == "" {
		c.JSON(structure.Status[codes.InvalidArgument], gin.H{"error": "template is required"})
		return
	}

	var buffer bytes.Buffer
	// Prepend UTF-8 BOM so that Excel opens the file correctly.
	buffer.Write([]byte{0xEF, 0xBB, 0xBF})

	w := csv.NewWriter(&buffer)
	w.Comma = ';'

	switch req.Template {
	case "Мегафон/Теле2":

		if err := w.Write([]string{"msisdn", "Имя"}); err != nil {
			c.JSON(structure.Status[codes.Internal], gin.H{"error": err.Error()})
			return
		}

		for _, item := range req.Items {
			row := []string{
				item.Phone,
				item.Retarget,
			}
			if err := w.Write(row); err != nil {
				continue
			}
		}
		break
	case "Билайн/СМСРитеил":
		for _, item := range req.Items {
			row := []string{
				item.Phone,
				item.Retarget,
			}
			if err := w.Write(row); err != nil {
				continue
			}
		}
		break
	case "МТС":
		if err := w.Write([]string{"Номер телефона", "Имя"}); err != nil {
			c.JSON(structure.Status[codes.Internal], gin.H{"error": err.Error()})
			return
		}
		for _, item := range req.Items {
			row := []string{
				item.Phone,
				item.Retarget,
			}
			if err := w.Write(row); err != nil {
				continue
			}
		}
		break
	default:
		c.JSON(structure.Status[codes.InvalidArgument], gin.H{"error": "invalid template"})
		return
	}

	w.Flush()
	if err := w.Error(); err != nil {
		c.JSON(structure.Status[codes.Internal], gin.H{"error": err.Error()})
		return
	}

	filename := "export.csv"
	if req.Name != "" {
		filename = fmt.Sprintf("%s.csv", req.Name)
	}

	escapedFilename := url.PathEscape(filename)
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename*=UTF-8''%s", escapedFilename))
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Data(structure.Status[codes.OK], "text/csv; charset=utf-8", buffer.Bytes())
}

func (r *Retarget) GetTemplates(c *gin.Context) {
	valid, err := auth.ValidateToken(c, r.cfg)
	if err != nil {
		c.JSON(structure.Status[codes.Unauthenticated], gin.H{"error": err.Error()})
		return
	}
	if !valid {
		c.JSON(structure.Status[codes.Unauthenticated], gin.H{"error": "unauthorized"})
		return
	}

	templates, err := r.dbTemplate.GetTemplates(c.Request.Context())
	if err != nil {
		c.JSON(structure.Status[codes.Internal], gin.H{"error": err.Error()})
		return
	}

	c.JSON(structure.Status[codes.OK], gin.H{"message": "success", "data": templates})
}
