package restock

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrEmptyItems         = errors.New("restock items cannot be empty")
	ErrInvalidItem        = errors.New("invalid restock item: quantity must be positive and unit cost must be non-negative")
	ErrMissingIdempotency = errors.New("idempotency_key is required")
)

type Service struct {
	repo RepositoryInterface
	db   *pgxpool.Pool
}

func NewService(db *pgxpool.Pool) *Service {
	return &Service{
		repo: NewRepository(db),
		db:   db,
	}
}

func (s *Service) CreateRestock(ctx context.Context, userID string, payload CreateRestockPayload) (*CreateRestockResult, error) {
	if payload.IdempotencyKey == "" {
		return nil, ErrMissingIdempotency
	}

	if len(payload.Items) == 0 {
		return nil, ErrEmptyItems
	}

	for _, item := range payload.Items {
		if item.Quantity <= 0 || item.UnitCost < 0 || item.SparepartID == "" {
			return nil, ErrInvalidItem
		}
	}

	if s.db == nil {
		return nil, errors.New("database pool not configured")
	}

	// Check if restock with same idempotency key already exists
	existingRestock, err := s.repo.GetRestockByIdempotencyKey(ctx, payload.IdempotencyKey)
	if err != nil {
		return nil, err
	}

	// If restock already exists, return it
	if existingRestock != nil {
		return &CreateRestockResult{RestockID: existingRestock.ID}, nil
	}

	now := time.Now()
	restockID := uuid.New().String()

	var restockItems []RestockItemDetail
	totalAmount := 0

	for _, item := range payload.Items {
		subtotal := item.UnitCost * item.Quantity
		restockItems = append(restockItems, RestockItemDetail{
			ID:          uuid.New().String(),
			RestockID:   restockID,
			SparepartID: item.SparepartID,
			Quantity:    item.Quantity,
			UnitCost:    item.UnitCost,
			Subtotal:    subtotal,
			CreatedAt:   now,
		})
		totalAmount += subtotal
	}

	restock := &Restock{
		ID:             restockID,
		UserID:         userID,
		IdempotencyKey: payload.IdempotencyKey,
		TotalAmount:    totalAmount,
		Notes:          payload.Notes,
		Items:          restockItems,
		CreatedAt:      now,
	}

	err = pgx.BeginFunc(ctx, s.db, func(tx pgx.Tx) error {
		if err := s.repo.CreateRestock(ctx, tx, restock); err != nil {
			return err
		}

		if len(restockItems) > 0 {
			if err := s.repo.CreateRestockItems(ctx, tx, restockItems); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &CreateRestockResult{RestockID: restockID}, nil
}
