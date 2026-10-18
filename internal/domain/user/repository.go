package user

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	Create(ctx context.Context, input CreateUserInput, passwordHash string) (*User, error)
	FindByID(ctx context.Context, id string) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindByEmailWithPassword(ctx context.Context, email string) (*User, string, error)
	List(ctx context.Context, p ListParams) ([]User, int, error)
	Update(ctx context.Context, id string, input UpdateUserInput) (*User, error)
	Delete(ctx context.Context, id string) error
}

type repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) Repository {
	return &repository{pool: pool}
}

func (r *repository) Create(ctx context.Context, input CreateUserInput, passwordHash string) (*User, error) {
	u := &User{}

	query := `
		INSERT INTO users(name, email, password_hash)
		VALUES ($1, $2, $3)
		RETURNING id, name, email, created_at, updated_at
	`

	err := r.pool.QueryRow(ctx, query, input.Name, input.Email, passwordHash).Scan(
		&u.ID,
		&u.Name,
		&u.Email,
		&u.CreatedAt,
		&u.UpdatedAt,
	)

	return u, err
}

func (r *repository) FindByID(ctx context.Context, id string) (*User, error) {
	uid, err := parseUUID(id)
	if err != nil {
		return nil, err
	}

	u := &User{}

	query := `
		SELECT id, name, email, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	err = r.pool.QueryRow(ctx, query, uid).Scan(
		&u.ID,
		&u.Name,
		&u.Email,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}

	return u, err
}

func (r *repository) FindByEmail(ctx context.Context, email string) (*User, error) {
	u := &User{}

	query := `
	  SELECT id, name, email, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	err := r.pool.QueryRow(ctx, query, email).Scan(
		&u.ID,
		&u.Name,
		&u.Email,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}

	return u, err
}

func (r *repository) FindByEmailWithPassword(ctx context.Context, email string) (*User, string, error) {
	u := &User{}
	var passwordHash string

	query := `
	  SELECT id, name, email, password_hash, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	err := r.pool.QueryRow(ctx, query, email).Scan(
		&u.ID,
		&u.Name,
		&u.Email,
		&passwordHash,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, "", nil
	}

	return u, passwordHash, err
}

func (r *repository) List(ctx context.Context, p ListParams) ([]User, int, error) {
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
		FROM users
		WHERE (
			$1='' OR
			name ILIKE '%'||$1||'%' OR
			email ILIKE '%'||$1||'%'
		)
	`
	if err := r.pool.QueryRow(ctx, countQuery, p.Search).Scan(&total); err != nil {
		return nil, 0, err
	}

	listQuery := `
	  SELECT id, name, email, created_at, updated_at
		FROM users
		WHERE (
			$1 = '' OR
			name ILIKE '%'||$1||'%' OR
			email ILIKE '%'||$1||'%'
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

	var users []User

	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, 0, err
		}
		users = append(users, u)
	}

	return users, total, rows.Err()
}

func (r *repository) Update(ctx context.Context, id string, input UpdateUserInput) (*User, error) {
	uid, err := parseUUID(id)
	if err != nil {
		return nil, err
	}

	u := &User{}

	query := `
		UPDATE users SET
			name = COALESCE($2, name),
			email = COALESCE($3, email),
			updated_at = NOW()
		WHERE id = $1
		RETURNING id, name, email, created_at, updated_at
	`

	err = r.pool.QueryRow(ctx, query, uid, input.Name, input.Email).Scan(
		&u.ID,
		&u.Name,
		&u.Email,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}

	return u, err
}

func (r *repository) Delete(ctx context.Context, id string) error {
	uid, err := parseUUID(id)
	if err != nil {
		return err
	}

	query := `DELETE FROM users WHERE id = $1`

	ct, err := r.pool.Exec(ctx, query, uid)
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
