package stock_opname

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrEmptyItems         = errors.New("stock opname items cannot be empty")
	ErrInvalidItem        = errors.New("invalid stock opname item: quantity must be non-negative")
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

func (s *Service) CreateStockOpname(ctx context.Context, userID string, payload CreateStockOpnamePayload) (*CreateStockOpnameResult, error) {
	if payload.IdempotencyKey == "" {
		return nil, ErrMissingIdempotency
	}

	if len(payload.Items) == 0 {
		return nil, ErrEmptyItems
	}

	for _, item := range payload.Items {
		if item.Quantity < 0 || item.SparepartID == "" {
			return nil, ErrInvalidItem
		}
	}

	if s.db == nil {
		return nil, errors.New("database pool not configured")
	}

	// Check if stock opname with same idempotency key already exists
	existingOpname, err := s.repo.GetStockOpnameByIdempotencyKey(ctx, payload.IdempotencyKey)
	if err != nil {
		return nil, err
	}

	// If stock opname already exists, return it
	if existingOpname != nil {
		return &CreateStockOpnameResult{StockOpnameID: existingOpname.ID}, nil
	}

	now := time.Now()
	opnameID := uuid.New().String()

	var opnameItems []StockOpnameItemDetail

	for _, item := range payload.Items {
		opnameItems = append(opnameItems, StockOpnameItemDetail{
			ID:            uuid.New().String(),
			StockOpnameID: opnameID,
			SparepartID:   item.SparepartID,
			Quantity:      item.Quantity,
			CreatedAt:     now,
		})
	}

	opname := &StockOpname{
		ID:             opnameID,
		UserID:         userID,
		IdempotencyKey: payload.IdempotencyKey,
		Notes:          payload.Notes,
		Items:          opnameItems,
		CreatedAt:      now,
	}

	err = pgx.BeginFunc(ctx, s.db, func(tx pgx.Tx) error {
		if err := s.repo.CreateStockOpname(ctx, tx, opname); err != nil {
			return err
		}

		if len(opnameItems) > 0 {
			if err := s.repo.CreateStockOpnameItems(ctx, tx, opnameItems); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &CreateStockOpnameResult{StockOpnameID: opnameID}, nil
}

func (s *Service) GetStockOverview(ctx context.Context) (*StockOverview, error) {
	return s.repo.GetStockOverview(ctx)
}
