package database

import (
	"context"
	"database/sql"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type LinkProvider interface {
	GetLinks(ctx context.Context) ([]string, error)
	SetLink(ctx context.Context, link string) error
	DeleteLink(ctx context.Context, link string) error
}

func (s *SDatabase) GetLinks(ctx context.Context) ([]string, error) {
	const q = `
		SELECT link FROM retarged_links
	`
	var links []string
	if err := s.db.SelectContext(ctx, &links, q); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		s.logger.ErrorContext(
			ctx,
			"ERROR database",
			"method", "GetLinks",
			"message", err.Error(),
		)
	}
	return links, nil
}

func (s *SDatabase) SetLink(ctx context.Context, link string) error {
	const q = `
		INSERT INTO retarged_links (link) VALUES ($1)
	`
	if _, err := s.db.ExecContext(ctx, q, link); err != nil {
		s.logger.ErrorContext(
			ctx,
			"ERROR database",
			"method", "SetLink",
			"message", err.Error(),
		)
		return status.Errorf(codes.Internal, "database error")
	}
	return nil
}

func (s *SDatabase) DeleteLink(ctx context.Context, link string) error {
	const q = `
		DELETE FROM retarged_links WHERE link = $1
	`
	if _, err := s.db.ExecContext(ctx, q, link); err != nil {
		s.logger.ErrorContext(
			ctx,
			"ERROR database",
			"method", "DeleteLink",
			"message", err.Error(),
		)
		return status.Errorf(codes.Internal, "database error")
	}
	return nil
}
