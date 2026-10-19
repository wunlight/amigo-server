package operation

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	Create(ctx context.Context, input CreateOperationInput) (*Operation, error)
	FindByID(ctx context.Context, id string) (*Operation, error)
	List(ctx context.Context, p ListParams) ([]Operation, int, error)
	Update(ctx context.Context, id string, input UpdateOperationInput) (*Operation, error)
	Delete(ctx context.Context, id string) error
}

type repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) Repository {
	return &repository{pool: pool}
}

func (r *repository) Create(ctx context.Context, input CreateOperationInput) (*Operation, error) {
	op := &Operation{}

	query := `
		INSERT INTO operations (name, selling_price)
		VALUES ($1, $2)
		RETURNING id, name, selling_price, created_at, updated_at
	`

	err := r.pool.QueryRow(ctx, query, input.Name, input.SellingPrice).Scan(
		&op.ID,
		&op.Name,
		&op.SellingPrice,
		&op.CreatedAt,
		&op.UpdatedAt,
	)

	return op, err
}

func (r *repository) FindByID(ctx context.Context, id string) (*Operation, error) {
	opID, err := parseUUID(id)
	if err != nil {
		return nil, err
	}

	op := &Operation{}

	query := `
		SELECT id, name, selling_price, created_at, updated_at
		FROM operations
		WHERE id = $1
	`

	err = r.pool.QueryRow(ctx, query, opID).Scan(
		&op.ID,
		&op.Name,
		&op.SellingPrice,
		&op.CreatedAt,
		&op.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}

	return op, err
}

func (r *repository) List(ctx context.Context, p ListParams) ([]Operation, int, error) {
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
		FROM operations
		WHERE (
			$1='' OR
			name ILIKE '%'||$1||'%'
		)
	`
	if err := r.pool.QueryRow(ctx, countQuery, p.Search).Scan(&total); err != nil {
		return nil, 0, err
	}

	listQuery := `
	  SELECT id, name, selling_price, created_at, updated_at
		FROM operations
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

	var operations []Operation

	for rows.Next() {
		var op Operation
		if err := rows.Scan(&op.ID, &op.Name, &op.SellingPrice, &op.CreatedAt, &op.UpdatedAt); err != nil {
			return nil, 0, err
		}
		operations = append(operations, op)
	}

	return operations, total, rows.Err()
}

func (r *repository) Update(ctx context.Context, id string, input UpdateOperationInput) (*Operation, error) {
	opID, err := parseUUID(id)
	if err != nil {
		return nil, err
	}

	op := &Operation{}

	query := `
		UPDATE operations SET
			name = COALESCE($2, name),
			selling_price = COALESCE($3, selling_price),
			updated_at = NOW()
		WHERE id = $1
		RETURNING id, name, selling_price, created_at, updated_at
	`

	err = r.pool.QueryRow(ctx, query, opID, input.Name, input.SellingPrice).Scan(
		&op.ID,
		&op.Name,
		&op.SellingPrice,
		&op.CreatedAt,
		&op.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}

	return op, err
}

func (r *repository) Delete(ctx context.Context, id string) error {
	opID, err := parseUUID(id)
	if err != nil {
		return err
	}

	query := `DELETE FROM operations WHERE id = $1`

	ct, err := r.pool.Exec(ctx, query, opID)
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
