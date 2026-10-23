package sale

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrInvalidCartItem   = errors.New("invalid cart item")
	ErrEmptyCart         = errors.New("cart is empty")
	ErrIdempotencyExists = errors.New("idempotency key already exists")
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

func (s *Service) Checkout(ctx context.Context, userID string, payload CheckoutPayload) (*CheckoutResult, error) {
	if len(payload.Items) == 0 {
		return nil, ErrEmptyCart
	}

	for _, item := range payload.Items {
		if item.Type != CartItemTypeSparepart && item.Type != CartItemTypeOperation {
			return nil, ErrInvalidCartItem
		}
		if item.Quantity <= 0 {
			return nil, ErrInvalidCartItem
		}
		if item.ID == "" {
			return nil, ErrInvalidCartItem
		}
		if item.Name == "" {
			return nil, ErrInvalidCartItem
		}
	}

	existingSale, err := s.repo.GetSaleByIdempotencyKey(ctx, payload.IdempotencyKey)
	if err == nil {
		// Return cached result with items and details
		return s.buildCheckoutResult(existingSale)
	}
	if !errors.Is(err, ErrIdempotencyKeyNotFound) {
		return nil, err
	}

	if s.db == nil {
		return nil, errors.New("database pool not configured")
	}

	now := time.Now()
	saleID := uuid.New().String()

	var sparepartItems []SaleSparepartItem
	var operationItems []SaleOperationItem
	var checkoutItems []CheckoutItem
	totalAmount := 0

	for _, item := range payload.Items {
		switch item.Type {
		case CartItemTypeSparepart:
			subtotal := item.UnitPrice * item.Quantity
			sparepartItems = append(sparepartItems, SaleSparepartItem{
				ID:          uuid.New().String(),
				SaleID:      saleID,
				SparepartID: item.ID,
				Quantity:    item.Quantity,
				UnitPrice:   item.UnitPrice,
				Subtotal:    subtotal,
				CreatedAt:   now,
			})
			checkoutItems = append(checkoutItems, CheckoutItem{
				Type:      item.Type,
				ID:        item.ID,
				Name:      item.Name,
				Quantity:  item.Quantity,
				UnitPrice: item.UnitPrice,
				Subtotal:  subtotal,
			})
			totalAmount += subtotal

		case CartItemTypeOperation:
			subtotal := item.UnitPrice * item.Quantity
			operationItems = append(operationItems, SaleOperationItem{
				ID:          uuid.New().String(),
				SaleID:      saleID,
				OperationID: item.ID,
				Quantity:    item.Quantity,
				UnitPrice:   item.UnitPrice,
				Subtotal:    subtotal,
				CreatedAt:   now,
			})
			checkoutItems = append(checkoutItems, CheckoutItem{
				Type:      item.Type,
				ID:        item.ID,
				Name:      item.Name,
				Quantity:  item.Quantity,
				UnitPrice: item.UnitPrice,
				Subtotal:  subtotal,
			})
			totalAmount += subtotal
		}
	}

	sale := &Sale{
		ID:             saleID,
		UserID:         userID,
		IdempotencyKey: payload.IdempotencyKey,
		TotalAmount:    totalAmount,
		Notes:          payload.Notes,
		Status:         SaleStatusCompleted,
		SparepartItems: sparepartItems,
		OperationItems: operationItems,
		CreatedAt:      now,
	}

	err = pgx.BeginFunc(ctx, s.db, func(tx pgx.Tx) error {
		if err := s.repo.CreateSale(ctx, tx, sale); err != nil {
			return err
		}

		if len(sparepartItems) > 0 {
			if err := s.repo.CreateSparepartItems(ctx, tx, sparepartItems); err != nil {
				return err
			}
		}

		if len(operationItems) > 0 {
			if err := s.repo.CreateOperationItems(ctx, tx, operationItems); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		if errors.Is(err, ErrIdempotencyKeyExists) {
			// Race condition: another request created the sale with same key
			// Fetch and return the existing sale
			existingSale, fetchErr := s.repo.GetSaleByIdempotencyKey(ctx, payload.IdempotencyKey)
			if fetchErr == nil {
				return s.buildCheckoutResult(existingSale)
			}
			return nil, fetchErr
		}
		return nil, err
	}

	return &CheckoutResult{
		SaleID:      saleID,
		CreatedAt:   now,
		TotalAmount: totalAmount,
		Items:       checkoutItems,
		Notes:       payload.Notes,
	}, nil
}

func (s *Service) buildCheckoutResult(sale *Sale) (*CheckoutResult, error) {
	checkoutItems := make([]CheckoutItem, 0, len(sale.SparepartItems)+len(sale.OperationItems))

	for _, item := range sale.SparepartItems {
		checkoutItems = append(checkoutItems, CheckoutItem{
			Type:      CartItemTypeSparepart,
			ID:        item.SparepartID,
			Name:      "", // Name is not stored in DB, will be empty on cache hit
			Quantity:  item.Quantity,
			UnitPrice: item.UnitPrice,
			Subtotal:  item.Subtotal,
		})
	}

	for _, item := range sale.OperationItems {
		checkoutItems = append(checkoutItems, CheckoutItem{
			Type:      CartItemTypeOperation,
			ID:        item.OperationID,
			Name:      "", // Name is not stored in DB, will be empty on cache hit
			Quantity:  item.Quantity,
			UnitPrice: item.UnitPrice,
			Subtotal:  item.Subtotal,
		})
	}

	return &CheckoutResult{
		SaleID:      sale.ID,
		CreatedAt:   sale.CreatedAt,
		TotalAmount: sale.TotalAmount,
		Items:       checkoutItems,
		Notes:       sale.Notes,
	}, nil
}

func (s *Service) ListSales(ctx context.Context, userID string, page int, perPage int) ([]Sale, int, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 10
	}

	return s.repo.ListSales(ctx, userID, page, perPage)
}

func (s *Service) RefundSale(ctx context.Context, saleID string, userID string) error {
	// Verify sale exists and belongs to user
	sale, err := s.repo.GetSaleByID(ctx, saleID)
	if err != nil {
		return err
	}

	if sale.UserID != userID {
		return errors.New("unauthorized: sale does not belong to user")
	}

	// Check if sale can be refunded
	if sale.Status == SaleStatusRefunded {
		return errors.New("sale is already refunded")
	}

	// Update status to refunded
	return s.repo.UpdateSaleStatus(ctx, saleID, SaleStatusRefunded)
}