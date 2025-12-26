package database

import (
	"context"
	"database/sql"

	"github.com/lunyashon/filterphone/internal/lib/structure"
)

type NumbersProvider interface {
	CreateNumbers(ctx context.Context, numbers structure.Numbers) error
	GetNumbers(ctx context.Context, code int16, numbers int) (*structure.Numbers, error)
	DeleteNumbers(ctx context.Context) error
}

func (s *SDatabase) DeleteNumbers(ctx context.Context) error {
	var (
		err error
	)

	const q = `
		DELETE FROM numbers_diapason
	`
	_, err = s.db.ExecContext(ctx, q)
	if err != nil {
		return err
	}
	return nil
}

func (s *SDatabase) CreateNumbers(ctx context.Context, numbers structure.Numbers) error {
	var (
		err error
	)

	const q = `
		INSERT INTO numbers_diapason (code, from_n, to_n, capacity, operator, region, territory, inn)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err = s.db.ExecContext(ctx, q, numbers.Code, numbers.From, numbers.To, numbers.Capacity, numbers.Operator, numbers.Region, numbers.Territory, numbers.INN)
	if err != nil {
		return err
	}
	return nil
}

func (s *SDatabase) GetNumbers(ctx context.Context, code int16, numbers int) (*structure.Numbers, error) {
	var (
		err            error
		number         structure.Numbers
		mobileOperator sql.NullString
	)

	const q = `
		SELECT 
			code, from_n, to_n, capacity, operator, region, territory, inn, mobile_operator
		FROM numbers_diapason
		WHERE code = $1 AND $2::int <@ span
		LIMIT 1
	`
	err = s.db.QueryRowxContext(ctx, q, code, numbers).Scan(
		&number.Code,
		&number.From,
		&number.To,
		&number.Capacity,
		&number.Operator,
		&number.Region,
		&number.Territory,
		&number.INN,
		&mobileOperator,
	)
	if err != nil {
		return nil, err
	}
	if mobileOperator.Valid {
		number.MobileOperator = mobileOperator.String
	}
	return &number, nil
}
