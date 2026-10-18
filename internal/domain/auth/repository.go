package auth

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	FindByEmailWithPassword(ctx context.Context, email string) (*User, string, error)
	FindByID(ctx context.Context, id string) (*User, error)
}

type repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) Repository {
	return &repository{pool: pool}
}

type User struct {
	ID        string
	Name      string
	Email     string
	CreatedAt string
	UpdatedAt string
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

func (r *repository) FindByID(ctx context.Context, id string) (*User, error) {
	u := &User{}

	query := `
		SELECT id, name, email, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	err := r.pool.QueryRow(ctx, query, id).Scan(
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

func parseUUID(id string) (uuid.UUID, error) {
	return uuid.Parse(id)
}