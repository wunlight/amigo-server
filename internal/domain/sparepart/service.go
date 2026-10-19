package sparepart

import (
	"amigo-server/pkg/apperror"
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type Service interface {
	Create(ctx context.Context, input CreateSparepartInput) (*Sparepart, error)
	GetByID(ctx context.Context, id string) (*Sparepart, error)
	List(ctx context.Context, p ListParams) ([]Sparepart, int, error)
	Update(ctx context.Context, id string, input UpdateSparepartInput) (*Sparepart, error)
	Delete(ctx context.Context, id string) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) Create(ctx context.Context, input CreateSparepartInput) (*Sparepart, error) {
	input.Name = strings.TrimSpace(input.Name)

	if input.Name == "" {
		return nil, apperror.BadRequest("name is required")
	}
	if input.SellingPrice < 0 {
		return nil, apperror.BadRequest("selling price must greater than or equal to 0")
	}
	if input.InitialStock < 0 {
		return nil, apperror.BadRequest("stock must greater than or equal to 0")
	}

	sp, err := s.repo.Create(ctx, input)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, apperror.Conflict("sparepart already exists")
		}
		return nil, apperror.Wrap(err, 500, "failed to create sparepart")
	}

	return sp, nil
}

func (s *service) GetByID(ctx context.Context, id string) (*Sparepart, error) {
	if err := validateUUID(id); err != nil {
		return nil, err
	}

	sp, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, apperror.Wrap(err, 500, "failed to get sparepart")
	}
	if sp == nil {
		return nil, apperror.NotFound("sparepart not found")
	}

	return sp, nil
}

func (s *service) List(ctx context.Context, p ListParams) ([]Sparepart, int, error) {
	spareparts, total, err := s.repo.List(ctx, p)
	if err != nil {
		return nil, 0, apperror.Wrap(err, 500, "failed to list spareparts")
	}
	if spareparts == nil {
		spareparts = []Sparepart{}
	}

	return spareparts, total, nil
}

func (s *service) Update(ctx context.Context, id string, input UpdateSparepartInput) (*Sparepart, error) {
	if err := validateUUID(id); err != nil {
		return nil, err
	}

	if input.Name != nil {
		v := strings.TrimSpace(*input.Name)
		if v == "" {
			return nil, apperror.BadRequest("name cannot be empty")
		}
		input.Name = &v
	}

	sp, err := s.repo.Update(ctx, id, input)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, apperror.Conflict("sparepart already exists")
		}
		return nil, apperror.Wrap(err, 500, "failed to update sparepart")
	}
	if sp == nil {
		return nil, apperror.NotFound("sparepart not found")
	}

	return sp, nil
}

func (s *service) Delete(ctx context.Context, id string) error {
	if err := validateUUID(id); err != nil {
		return err
	}

	err := s.repo.Delete(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return apperror.NotFound("sparepart not found")
	}
	if err != nil {
		return apperror.Wrap(err, 500, "failed to delete sparepart")
	}

	return nil
}

func validateUUID(id string) error {
	if err := uuid.Validate(id); err != nil {
		return apperror.BadRequest("invalid uuid")
	}
	return nil
}
