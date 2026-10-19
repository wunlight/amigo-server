package operation

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
	Create(ctx context.Context, input CreateOperationInput) (*Operation, error)
	GetByID(ctx context.Context, id string) (*Operation, error)
	List(ctx context.Context, p ListParams) ([]Operation, int, error)
	Update(ctx context.Context, id string, input UpdateOperationInput) (*Operation, error)
	Delete(ctx context.Context, id string) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) Create(ctx context.Context, input CreateOperationInput) (*Operation, error) {
	input.Name = strings.TrimSpace(input.Name)

	if input.Name == "" {
		return nil, apperror.BadRequest("name is required")
	}
	if input.SellingPrice < 0 {
		return nil, apperror.BadRequest("selling price must greater than or equal to 0")
	}

	op, err := s.repo.Create(ctx, input)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, apperror.Conflict("operation already exists")
		}
		return nil, apperror.Wrap(err, 500, "failed to create operation")
	}

	return op, nil
}

func (s *service) GetByID(ctx context.Context, id string) (*Operation, error) {
	if err := validateUUID(id); err != nil {
		return nil, err
	}

	op, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, apperror.Wrap(err, 500, "failed to get operation")
	}
	if op == nil {
		return nil, apperror.NotFound("operation not found")
	}

	return op, nil
}

func (s *service) List(ctx context.Context, p ListParams) ([]Operation, int, error) {
	operations, total, err := s.repo.List(ctx, p)
	if err != nil {
		return nil, 0, apperror.Wrap(err, 500, "failed to list operations")
	}
	if operations == nil {
		operations = []Operation{}
	}

	return operations, total, nil
}

func (s *service) Update(ctx context.Context, id string, input UpdateOperationInput) (*Operation, error) {
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

	op, err := s.repo.Update(ctx, id, input)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, apperror.Conflict("operation already exists")
		}
		return nil, apperror.Wrap(err, 500, "failed to update operation")
	}
	if op == nil {
		return nil, apperror.NotFound("operation not found")
	}

	return op, nil
}

func (s *service) Delete(ctx context.Context, id string) error {
	if err := validateUUID(id); err != nil {
		return err
	}

	err := s.repo.Delete(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return apperror.NotFound("operation not found")
	}
	if err != nil {
		return apperror.Wrap(err, 500, "failed to delete operation")
	}

	return nil
}

func validateUUID(id string) error {
	if err := uuid.Validate(id); err != nil {
		return apperror.BadRequest("invalid uuid")
	}
	return nil
}
