package retarget

import (
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lunyashon/filterphone/internal/lib/auth"
	"github.com/lunyashon/filterphone/internal/lib/structure"
	"google.golang.org/grpc/codes"
)

type AddArrayRetargetRequest struct {
	Items []ItemRetarget `json:"items"`
	Name  string         `json:"name"`
	TTL   string         `json:"ttl"`
}

type ItemRetarget struct {
	FinalLink string `json:"final_link"`
	Operator  string `json:"operator"`
	Phone     string `json:"phone"`
	Territory string `json:"territory"`
	Token     string `json:"token"`
	Params    struct {
		UtmCampaign string `json:"utm_campaign"`
		UtmSource   string `json:"utm_source"`
		UtmContent  string `json:"utm_content"`
		UtmMedium   string `json:"utm_medium"`
		UtmTerm     string `json:"utm_term"`
		TargetUrl   string `json:"target_url"`
	}
}

var (
	letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
)

func (r *Retarget) AddArrayRetarget(c *gin.Context) {
	valid, err := auth.ValidateToken(c, r.cfg)
	if err != nil {
		c.JSON(structure.Status[codes.Unauthenticated], gin.H{"error": err.Error()})
		return
	}
	if !valid {
		c.JSON(structure.Status[codes.Unauthenticated], gin.H{"error": "unauthorized"})
		return
	}

	var req AddArrayRetargetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if exists, err := r.db.CheckArrayRetarget(c.Request.Context(), req.Name); err != nil {
		c.JSON(structure.Status[codes.Internal], gin.H{"error": err.Error()})
		return
	} else if exists {
		c.JSON(structure.Status[codes.AlreadyExists], gin.H{"error": "array retarget already exists"})
		return
	}

	uniqueIDArray := generateUniqueIDArray()

	for _, item := range req.Items {

		ttl, err := time.Parse(time.RFC3339, req.TTL)
		if err != nil {
			c.JSON(structure.Status[codes.InvalidArgument], gin.H{"error": "Failed to parse TTL. Invalid format, need to be in 2026-01-01T00:00:00Z format"})
			return
		}
		if ttl.Before(time.Now()) {
			c.JSON(structure.Status[codes.InvalidArgument], gin.H{"error": "TTL is in the past"})
			return
		}

		retarget := &structure.Retarget{
			NameArray:     req.Name,
			FinalLink:     item.FinalLink,
			Operator:      item.Operator,
			Phone:         item.Phone,
			Territory:     item.Territory,
			UniqueToken:   item.Token,
			TargetUrl:     item.Params.TargetUrl,
			UtmCompaign:   item.Params.UtmCampaign,
			UtmSource:     item.Params.UtmSource,
			UtmContent:    item.Params.UtmContent,
			UtmMedium:     item.Params.UtmMedium,
			UtmTerm:       item.Params.UtmTerm,
			TTL:           ttl,
			UniqueIDArray: uniqueIDArray,
		}

		if err := r.db.AddRetarget(c.Request.Context(), retarget); err != nil {
			c.JSON(structure.Status[codes.Internal], gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(structure.Status[codes.OK], gin.H{"message": "retarget added", "data": uniqueIDArray})
}

func generateUniqueIDArray() string {
	b := make([]rune, 10)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b) + "_" + time.Now().Format("20060102150405")
}

func (r *Retarget) GetAllArrayRetarget(c *gin.Context) {
	valid, err := auth.ValidateToken(c, r.cfg)
	if err != nil {
		c.JSON(structure.Status[codes.Unauthenticated], gin.H{"error": err.Error()})
		return
	}
	if !valid {
		c.JSON(structure.Status[codes.Unauthenticated], gin.H{"error": "unauthorized"})
		return
	}

	arrayRetargets, err := r.db.GetArrayRetarget(c.Request.Context())
	if err != nil {
		c.JSON(structure.Status[codes.Internal], gin.H{"error": err.Error()})
		return
	}

	c.JSON(structure.Status[codes.OK], gin.H{"message": "success", "data": arrayRetargets})
}

type ItemsArrayRetarget struct {
	NameArray string                       `json:"name_array"`
	Items     []ItemsArrayRetargetResponse `json:"items"`
}

type ItemsArrayRetargetResponse struct {
	Phone          string    `json:"phone"`
	FinalLink      string    `json:"final_link"`
	Operator       string    `json:"operator"`
	Territory      string    `json:"territory"`
	TargetUrl      string    `json:"target_url"`
	TTL            time.Time `json:"ttl"`
	LastFollowLink time.Time `json:"last_follow_link"`
	FollowLink     int       `json:"follow_link"`
}

func (r *Retarget) GetItemsArrayRetarget(c *gin.Context) {
	valid, err := auth.ValidateToken(c, r.cfg)
	if err != nil {
		c.JSON(structure.Status[codes.Unauthenticated], gin.H{"error": err.Error()})
		return
	}
	if !valid {
		c.JSON(structure.Status[codes.Unauthenticated], gin.H{"error": "unauthorized"})
		return
	}

	uniqueIDArray := c.Param("id_array")
	if uniqueIDArray == "" {
		c.JSON(structure.Status[codes.InvalidArgument], gin.H{"error": "id_array is required"})
		return
	}

	retargets, err := r.db.GetRetargetsByUniqueIDArray(c.Request.Context(), uniqueIDArray)
	if err != nil {
		c.JSON(structure.Status[codes.Internal], gin.H{"error": err.Error()})
		return
	}

	var response ItemsArrayRetarget
	response.NameArray = retargets[0].NameArray

	for _, retarget := range retargets {
		var lastFollowLink time.Time
		if retarget.LastFollowLink.Valid {
			lastFollowLink = retarget.LastFollowLink.Time
		}

		r := ItemsArrayRetargetResponse{
			Phone:          retarget.Phone,
			FinalLink:      retarget.FinalLink,
			Operator:       retarget.Operator,
			Territory:      retarget.Territory,
			TargetUrl:      retarget.TargetUrl,
			TTL:            retarget.TTL,
			LastFollowLink: lastFollowLink,
			FollowLink:     retarget.FollowLink,
		}
		response.Items = append(response.Items, r)
	}

	c.JSON(structure.Status[codes.OK], gin.H{"message": "success", "data": response})
}

func (r *Retarget) DeleteArrayRetarget(c *gin.Context) {
	valid, err := auth.ValidateToken(c, r.cfg)
	if err != nil {
		c.JSON(structure.Status[codes.Unauthenticated], gin.H{"error": err.Error()})
		return
	}
	if !valid {
		c.JSON(structure.Status[codes.Unauthenticated], gin.H{"error": "unauthorized"})
		return
	}

	uniqueIDArray := c.Param("id_array")
	if uniqueIDArray == "" {
		c.JSON(structure.Status[codes.InvalidArgument], gin.H{"error": "id_array is required"})
		return
	}

	if err := r.db.DeleteRetargetByUniqueIDArray(c.Request.Context(), uniqueIDArray); err != nil {
		c.JSON(structure.Status[codes.Internal], gin.H{"error": err.Error()})
		return
	}

	c.JSON(structure.Status[codes.OK], gin.H{"message": "success"})
}

func (r *Retarget) CheckItemRetarget(c *gin.Context) {
	// valid, err := auth.ValidateToken(c, r.cfg)
	// if err != nil {
	// 	c.JSON(structure.Status[codes.Unauthenticated], gin.H{"error": err.Error()})
	// 	return
	// }
	// if !valid {
	// 	c.JSON(structure.Status[codes.Unauthenticated], gin.H{"error": "unauthorized"})
	// 	return
	// }

	finalLink := c.Query("final_link")
	if finalLink == "" {
		c.JSON(structure.Status[codes.InvalidArgument], gin.H{"error": "final_link is required"})
		return
	}

	r.log.Info("final_link", "final_link", finalLink)

	retarget, err := r.db.GetRetargetByFinalLink(c.Request.Context(), finalLink)
	if err != nil {
		c.JSON(structure.Status[codes.Internal], gin.H{"error": err.Error()})
		return
	}

	if retarget == nil {
		c.JSON(structure.Status[codes.NotFound], gin.H{"error": "retarget not found"})
		return
	}

	if retarget.TTL.Before(time.Now()) {
		c.JSON(structure.Status[codes.NotFound], gin.H{"error": "retarget expired"})
		return
	}

	if err := r.db.UpdateFollowLink(c.Request.Context(), finalLink); err != nil {
		c.JSON(structure.Status[codes.Internal], gin.H{"error": err.Error()})
		return
	}

	c.JSON(structure.Status[codes.OK], gin.H{"message": "success", "data": retarget.TargetUrl})
}
