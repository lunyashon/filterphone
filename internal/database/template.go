package database

import (
	"context"
	"database/sql"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type TemplateProvider interface {
	GetTemplates(ctx context.Context) ([]TemplateExport, error)
}

type TemplateExport struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (s *SDatabase) GetTemplates(ctx context.Context) ([]TemplateExport, error) {
	const q = `
		SELECT id, name FROM template_export
	`
	var templates []TemplateExport
	if err := s.db.SelectContext(ctx, &templates, q); err != nil {
		if err == sql.ErrNoRows {
			return []TemplateExport{}, nil
		}
		s.logger.ErrorContext(
			ctx,
			"ERROR database",
			"method", "GetTemplates",
			"message", err.Error(),
		)
		return nil, status.Errorf(codes.Internal, "database error")
	}
	return templates, nil
}
