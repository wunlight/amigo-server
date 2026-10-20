package sale

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) GetSaleByIdempotencyKey(ctx context.Context, key string) (*Sale, error) {
	args := m.Called(ctx, key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Sale), args.Error(1)
}

func (m *MockRepository) CreateSale(ctx context.Context, tx pgx.Tx, sale *Sale) error {
	args := m.Called(ctx, tx, sale)
	return args.Error(0)
}

func (m *MockRepository) CreateSparepartItems(ctx context.Context, tx pgx.Tx, items []SaleSparepartItem) error {
	args := m.Called(ctx, tx, items)
	return args.Error(0)
}

func (m *MockRepository) CreateOperationItems(ctx context.Context, tx pgx.Tx, items []SaleOperationItem) error {
	args := m.Called(ctx, tx, items)
	return args.Error(0)
}

func TestCheckout_EmptyCart(t *testing.T) {
	mockRepo := new(MockRepository)
	svc := NewServiceWithDeps(mockRepo, nil)

	payload := CheckoutPayload{
		Items:          []CartItem{},
		Notes:          "Test",
		IdempotencyKey: "test-key",
	}

	result, err := svc.Checkout(context.Background(), "user-1", payload)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, ErrEmptyCart, err)
}

func TestCheckout_InvalidCartItem_Type(t *testing.T) {
	mockRepo := new(MockRepository)
	svc := NewServiceWithDeps(mockRepo, nil)

	payload := CheckoutPayload{
		Items: []CartItem{
			{Type: "invalid", ID: "1", Quantity: 1, UnitPrice: 10000},
		},
		Notes:          "Test",
		IdempotencyKey: "test-key",
	}

	result, err := svc.Checkout(context.Background(), "user-1", payload)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, ErrInvalidCartItem, err)
}

func TestCheckout_InvalidCartItem_ZeroQuantity(t *testing.T) {
	mockRepo := new(MockRepository)
	svc := NewServiceWithDeps(mockRepo, nil)

	payload := CheckoutPayload{
		Items: []CartItem{
			{Type: CartItemTypeSparepart, ID: "1", Quantity: 0, UnitPrice: 10000},
		},
		Notes:          "Test",
		IdempotencyKey: "test-key",
	}

	result, err := svc.Checkout(context.Background(), "user-1", payload)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, ErrInvalidCartItem, err)
}

func TestCheckout_InvalidCartItem_EmptyID(t *testing.T) {
	mockRepo := new(MockRepository)
	svc := NewServiceWithDeps(mockRepo, nil)

	payload := CheckoutPayload{
		Items: []CartItem{
			{Type: CartItemTypeSparepart, ID: "", Quantity: 1, UnitPrice: 10000},
		},
		Notes:          "Test",
		IdempotencyKey: "test-key",
	}

	result, err := svc.Checkout(context.Background(), "user-1", payload)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, ErrInvalidCartItem, err)
}

func TestCheckout_IdempotencyKeyExists(t *testing.T) {
	mockRepo := new(MockRepository)
	svc := NewServiceWithDeps(mockRepo, nil)

	existingSale := &Sale{
		ID:             "existing-sale-id",
		UserID:         "user-1",
		IdempotencyKey: "test-key",
		TotalAmount:    50000,
		Notes:          "Previous order",
		Status:         SaleStatusCompleted,
		CreatedAt:      time.Now(),
	}

	mockRepo.On("GetSaleByIdempotencyKey", mock.Anything, "test-key").Return(existingSale, nil)

	payload := CheckoutPayload{
		Items: []CartItem{
			{Type: CartItemTypeSparepart, ID: "1", Quantity: 1, UnitPrice: 10000},
		},
		Notes:          "Test",
		IdempotencyKey: "test-key",
	}

	result, err := svc.Checkout(context.Background(), "user-1", payload)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "existing-sale-id", result.SaleID)

	mockRepo.AssertExpectations(t)
}

func TestCheckout_NewSale_DbNil(t *testing.T) {
	mockRepo := new(MockRepository)
	svc := NewServiceWithDeps(mockRepo, nil)

	mockRepo.On("GetSaleByIdempotencyKey", mock.Anything, "test-key").Return(nil, ErrIdempotencyKeyNotFound)

	payload := CheckoutPayload{
		Items: []CartItem{
			{Type: CartItemTypeSparepart, ID: "sparepart-1", Quantity: 2, UnitPrice: 10000},
		},
		Notes:          "Test",
		IdempotencyKey: "test-key",
	}

	result, err := svc.Checkout(context.Background(), "user-1", payload)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "database pool not configured")

	mockRepo.AssertExpectations(t)
}

func TestCheckoutPayload_Hash(t *testing.T) {
	payload := CheckoutPayload{
		Items: []CartItem{
			{Type: CartItemTypeSparepart, ID: "1", Quantity: 2, UnitPrice: 10000},
			{Type: CartItemTypeOperation, ID: "2", Quantity: 1, UnitPrice: 50000},
		},
		Notes:          "Test notes",
		IdempotencyKey: "test-key-123",
	}

	hash1 := payload.Hash()
	hash2 := payload.Hash()

	assert.Equal(t, hash1, hash2)
	assert.Len(t, hash1, 64)
}

func TestCheckoutPayload_Hash_DifferentOrder(t *testing.T) {
	payload1 := CheckoutPayload{
		Items: []CartItem{
			{Type: CartItemTypeSparepart, ID: "1", Quantity: 2, UnitPrice: 10000},
			{Type: CartItemTypeOperation, ID: "2", Quantity: 1, UnitPrice: 50000},
		},
		Notes:          "Test notes",
		IdempotencyKey: "test-key-123",
	}

	payload2 := CheckoutPayload{
		Items: []CartItem{
			{Type: CartItemTypeOperation, ID: "2", Quantity: 1, UnitPrice: 50000},
			{Type: CartItemTypeSparepart, ID: "1", Quantity: 2, UnitPrice: 10000},
		},
		Notes:          "Test notes",
		IdempotencyKey: "test-key-123",
	}

	assert.NotEqual(t, payload1.Hash(), payload2.Hash())
}

func TestIdempotencyTTL(t *testing.T) {
	assert.Equal(t, 24*time.Hour, IdempotencyTTL)
}

func TestSaleStatusConstants(t *testing.T) {
	assert.Equal(t, "completed", SaleStatusCompleted)
	assert.Equal(t, "refunded", SaleStatusRefunded)
	assert.Equal(t, "voided", SaleStatusVoided)
}

func TestCartItemTypeConstants(t *testing.T) {
	assert.Equal(t, CartItemType("sparepart"), CartItemTypeSparepart)
	assert.Equal(t, CartItemType("operation"), CartItemTypeOperation)
}

func NewServiceWithDeps(repo RepositoryInterface, db *pgxpool.Pool) *Service {
	return &Service{
		repo: repo,
		db:   db,
	}
}