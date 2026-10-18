package user

import (
	"context"
	"errors"
	"strings"

	"amigo-server/pkg/apperror"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"
)

type Service interface {
	Create(ctx context.Context, input CreateUserInput) (*User, error)
	GetByID(ctx context.Context, id string) (*User, error)
	List(ctx context.Context, p ListParams) ([]User, int, error)
	Update(ctx context.Context, id string, input UpdateUserInput) (*User, error)
	Delete(ctx context.Context, id string) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) Create(ctx context.Context, input CreateUserInput) (*User, error) {
	input.Name = strings.TrimSpace(input.Name)
	input.Email = strings.TrimSpace(input.Email)
	input.Password = strings.TrimSpace(input.Password)

	if input.Name == "" {
		return nil, apperror.BadRequest("name is required")
	}
	if input.Email == "" || !strings.Contains(input.Email, "@") {
		return nil, apperror.BadRequest("valid email is required")
	}
	if input.Password == "" {
		return nil, apperror.BadRequest("password is required")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, apperror.Wrap(err, 500, "failed to hash password")
	}

	u, err := s.repo.Create(ctx, input, string(hashedPassword))
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, apperror.Conflict("email already exists")
		}
		return nil, apperror.Wrap(err, 500, "failed to create user")
	}

	return u, nil
}

func (s *service) GetByID(ctx context.Context, id string) (*User, error) {
	if err := validateUUID(id); err != nil {
		return nil, err
	}

	u, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, apperror.Wrap(err, 500, "failed to get user")
	}
	if u == nil {
		return nil, apperror.NotFound("user not found")
	}

	return u, nil
}

func (s *service) List(ctx context.Context, p ListParams) ([]User, int, error) {
	users, total, err := s.repo.List(ctx, p)
	if err != nil {
		return nil, 0, apperror.Wrap(err, 500, "failed to list users")
	}
	if users == nil {
		users = []User{}
	}

	return users, total, nil
}

func (s *service) Update(ctx context.Context, id string, input UpdateUserInput) (*User, error) {
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
	if input.Email != nil {
		v := strings.TrimSpace(*input.Email)
		if v == "" || !strings.Contains(v, "@") {
			return nil, apperror.BadRequest("valid email is required")
		}
		input.Email = &v
	}

	u, err := s.repo.Update(ctx, id, input)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, apperror.Conflict("email already exists")
		}
		return nil, apperror.Wrap(err, 500, "failed to update user")
	}
	if u == nil {
		return nil, apperror.NotFound("user not found")
	}

	return u, nil
}

func (s *service) Delete(ctx context.Context, id string) error {
	if err := validateUUID(id); err != nil {
		return err
	}

	err := s.repo.Delete(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return apperror.NotFound("user not found")
	}
	if err != nil {
		return apperror.Wrap(err, 500, "failed to delete user")
	}

	return nil
}

func validateUUID(id string) error {
	if err := uuid.Validate(id); err != nil {
		return apperror.BadRequest("invalid uuid")
	}
	return nil
}
