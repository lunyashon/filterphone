package database

import (
	"context"
	"database/sql"
	"time"

	"github.com/lunyashon/filterphone/internal/lib/structure"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type RetargetProvider interface {
	GetArrayRetarget(ctx context.Context) ([]RetargetArray, error)
	GetRetargetsByUniqueIDArray(ctx context.Context, uniqueIDArray string) ([]structure.Retarget, error)
	AddRetarget(ctx context.Context, retarget *structure.Retarget) error
	CheckArrayRetarget(ctx context.Context, nameArray string) (bool, error)
	UpdateFollowLink(ctx context.Context, finalLink string) error
	GetRetargetByFinalLink(ctx context.Context, finalLink string) (*RetargetByFinalLink, error)
	DeleteRetargetByUniqueIDArray(ctx context.Context, uniqueIDArray string) error
}

type RetargetArray struct {
	NameArray     string    `db:"name_array" json:"name_array"`
	Count         int       `db:"count" json:"count"`
	Links         int       `db:"links" json:"links"`
	CreatedAt     time.Time `db:"created_at" json:"created_at"`
	UniqueIDArray string    `db:"unique_id_array" json:"unique_id_array"`
	UtmCompaign   string    `db:"utm_compaign" json:"utm_campaign"`
}

type RetargetByFinalLink struct {
	UtmCompaign string    `db:"utm_compaign" json:"utm_compaign"`
	UtmSource   string    `db:"utm_source" json:"utm_source"`
	UtmContent  string    `db:"utm_content" json:"utm_content"`
	UtmMedium   string    `db:"utm_medium" json:"utm_medium"`
	UtmTerm     string    `db:"utm_term" json:"utm_term"`
	TargetUrl   string    `db:"target_url" json:"target_url"`
	TTL         time.Time `db:"ttl" json:"ttl"`
}

func (s *SDatabase) GetRetargetsByUniqueIDArray(ctx context.Context, uniqueIDArray string) ([]structure.Retarget, error) {
	const q = `
		SELECT 
			final_link, operator, phone, territory, 
			unique_token, target_url, utm_compaign, 
			utm_source, utm_content, utm_medium, 
			utm_term, ttl, follow_link, last_follow_link,
			name_array
		FROM 
			retarget_array 
		WHERE 
			unique_id_array = $1
		ORDER BY
			created_at DESC
	`

	var retargets []structure.Retarget
	if err := s.db.SelectContext(ctx, &retargets, q, uniqueIDArray); err != nil {
		return nil, err
	}
	return retargets, nil
}

func (s *SDatabase) AddRetarget(ctx context.Context, retarget *structure.Retarget) error {
	const q = `
		INSERT INTO retarget_array
			(name_array, final_link, operator, 
			phone, territory, unique_token, target_url, 
			utm_compaign, utm_source, utm_content, 
			utm_medium, utm_term, ttl, unique_id_array)
		VALUES
			($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`

	if _, err := s.db.ExecContext(
		ctx,
		q,
		retarget.NameArray, retarget.FinalLink, retarget.Operator,
		retarget.Phone, retarget.Territory, retarget.UniqueToken,
		retarget.TargetUrl, retarget.UtmCompaign, retarget.UtmSource,
		retarget.UtmContent, retarget.UtmMedium, retarget.UtmTerm,
		retarget.TTL, retarget.UniqueIDArray,
	); err != nil {
		s.logger.ErrorContext(
			ctx,
			"ERROR database",
			"method", "AddRetarget",
			"message", err.Error(),
		)
		return status.Errorf(codes.Internal, "database error")
	}
	return nil
}

func (s *SDatabase) GetArrayRetarget(ctx context.Context) ([]RetargetArray, error) {
	const q = `
		SELECT 
			name_array, COUNT(*) as count, SUM(follow_link) as links, 
			MIN(created_at) as created_at, unique_id_array, utm_compaign
		FROM 
			retarget_array 
		GROUP BY 
			name_array, unique_id_array, utm_compaign
	`

	var retargets []RetargetArray
	if err := s.db.SelectContext(ctx, &retargets, q); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		s.logger.ErrorContext(
			ctx,
			"ERROR database",
			"method", "GetArrayRetarget",
			"message", err.Error(),
		)
		return nil, status.Errorf(codes.Internal, "database error")
	}
	return retargets, nil
}

func (s *SDatabase) CheckArrayRetarget(ctx context.Context, nameArray string) (bool, error) {
	const q = `
		SELECT EXISTS (SELECT 1 FROM retarget_array WHERE name_array = $1 LIMIT 1)
	`

	var exists bool
	if err := s.db.GetContext(ctx, &exists, q, nameArray); err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		s.logger.ErrorContext(
			ctx,
			"ERROR database",
			"method", "CheckArrayRetarget",
			"message", err.Error(),
		)
		return false, status.Errorf(codes.Internal, "database error")
	}
	return exists, nil
}

func (s *SDatabase) UpdateFollowLink(ctx context.Context, finalLink string) error {
	const q = `
		UPDATE 
			retarget_array 
		SET 
			last_follow_link = $1, follow_link = follow_link + 1 
		WHERE 
			final_link = $2 
	`

	if _, err := s.db.ExecContext(ctx, q, time.Now().UTC(), finalLink); err != nil {
		s.logger.ErrorContext(
			ctx,
			"ERROR database",
			"method", "UpdateFollowLink",
			"message", err.Error(),
		)
		return status.Errorf(codes.Internal, "database error")
	}
	return nil
}

func (s *SDatabase) GetRetargetByFinalLink(ctx context.Context, finalLink string) (*RetargetByFinalLink, error) {
	const q = `
		SELECT 
			utm_compaign, utm_source, utm_content, 
			utm_medium, utm_term, target_url, ttl
		FROM 
			retarget_array 
		WHERE 
			final_link = $1
		LIMIT 1
	`

	var retarget RetargetByFinalLink
	if err := s.db.GetContext(ctx, &retarget, q, finalLink); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		s.logger.ErrorContext(
			ctx,
			"ERROR database",
			"method", "GetRetargetByFinalLink",
			"message", err.Error(),
		)
		return nil, status.Errorf(codes.Internal, "database error")
	}
	return &retarget, nil
}

func (s *SDatabase) DeleteRetargetByUniqueIDArray(ctx context.Context, uniqueIDArray string) error {
	const q = `
		DELETE FROM retarget_array WHERE unique_id_array = $1
	`

	if _, err := s.db.ExecContext(ctx, q, uniqueIDArray); err != nil {
		s.logger.ErrorContext(
			ctx,
			"ERROR database",
			"method", "DeleteRetargetByUniqueIDArray",
			"message", err.Error(),
		)
		return status.Errorf(codes.Internal, "database error")
	}
	return nil
}
