package sparepart

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	Create(ctx context.Context, input CreateSparepartInput) (*Sparepart, error)
	FindByID(ctx context.Context, id string) (*Sparepart, error)
	List(ctx context.Context, p ListParams) ([]Sparepart, int, error)
	Update(ctx context.Context, id string, input UpdateSparepartInput) (*Sparepart, error)
	Delete(ctx context.Context, id string) error
}

type repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) Repository {
	return &repository{pool: pool}
}

func (r *repository) Create(ctx context.Context, input CreateSparepartInput) (*Sparepart, error) {
	sp := &Sparepart{}

	query := `
		INSERT INTO spareparts (name, selling_price, stock)
		VALUES ($1, $2, $3)
		RETURNING id, name, selling_price, stock, created_at, updated_at
	`

	err := r.pool.QueryRow(ctx, query, input.Name, input.SellingPrice, input.InitialStock).Scan(
		&sp.ID,
		&sp.Name,
		&sp.SellingPrice,
		&sp.Stock,
		&sp.CreatedAt,
		&sp.UpdatedAt,
	)

	return sp, err
}

func (r *repository) FindByID(ctx context.Context, id string) (*Sparepart, error) {
	spid, err := parseUUID(id)
	if err != nil {
		return nil, err
	}

	sp := &Sparepart{}

	query := `
		SELECT id, name, selling_price, stock, created_at, updated_at
		FROM spareparts
		WHERE id = $1
	`

	err = r.pool.QueryRow(ctx, query, spid).Scan(
		&sp.ID,
		&sp.Name,
		&sp.SellingPrice,
		&sp.Stock,
		&sp.CreatedAt,
		&sp.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}

	return sp, err
}

func (r *repository) List(ctx context.Context, p ListParams) ([]Sparepart, int, error) {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.PerPage < 1 || p.PerPage > 100 {
		p.PerPage = 10
	}
	offset := (p.Page - 1) * p.PerPage

	var total int
	countQuery := `
		SELECT COUNT(*)
		FROM spareparts
		WHERE (
			$1='' OR
			name ILIKE '%'||$1||'%'
		)
	`
	if err := r.pool.QueryRow(ctx, countQuery, p.Search).Scan(&total); err != nil {
		return nil, 0, err
	}

	listQuery := `
	  SELECT id, name, selling_price, stock, created_at, updated_at
		FROM spareparts
		WHERE (
			$1 = '' OR
			name ILIKE '%'||$1||'%'
		)
		ORDER BY id DESC
		LIMIT $2
		OFFSET $3
	`
	rows, err := r.pool.Query(ctx, listQuery, p.Search, p.PerPage, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var spareparts []Sparepart

	for rows.Next() {
		var sp Sparepart
		if err := rows.Scan(&sp.ID, &sp.Name, &sp.SellingPrice, &sp.Stock, &sp.CreatedAt, &sp.UpdatedAt); err != nil {
			return nil, 0, err
		}
		spareparts = append(spareparts, sp)
	}

	return spareparts, total, rows.Err()
}

func (r *repository) Update(ctx context.Context, id string, input UpdateSparepartInput) (*Sparepart, error) {
	spid, err := parseUUID(id)
	if err != nil {
		return nil, err
	}

	sp := &Sparepart{}

	query := `
		UPDATE spareparts SET
			name = COALESCE($2, name),
			selling_price = COALESCE($3, selling_price),
			updated_at = NOW()
		WHERE id = $1
		RETURNING id, name, selling_price, stock, created_at, updated_at
	`

	err = r.pool.QueryRow(ctx, query, spid, input.Name, input.SellingPrice).Scan(
		&sp.ID,
		&sp.Name,
		&sp.SellingPrice,
		&sp.Stock,
		&sp.CreatedAt,
		&sp.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}

	return sp, err
}

func (r *repository) Delete(ctx context.Context, id string) error {
	spid, err := parseUUID(id)
	if err != nil {
		return err
	}

	query := `DELETE FROM spareparts WHERE id = $1`

	ct, err := r.pool.Exec(ctx, query, spid)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

func parseUUID(id string) (uuid.UUID, error) {
	return uuid.Parse(id)
}
