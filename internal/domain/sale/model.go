package sale

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"time"
)

type CartItemType string

const (
	CartItemTypeSparepart CartItemType = "sparepart"
	CartItemTypeOperation CartItemType = "operation"
)

type CartItem struct {
	Type      CartItemType `json:"type"`
	ID        string       `json:"id"`
	Name      string       `json:"name"`
	Quantity  int          `json:"quantity"`
	UnitPrice int          `json:"unit_price"`
}

type CheckoutPayload struct {
	Items          []CartItem `json:"items"`
	Notes          string     `json:"notes"`
	IdempotencyKey string     `json:"idempotency_key"`
}

type CheckoutResult struct {
	SaleID      string         `json:"sale_id"`
	CreatedAt   time.Time      `json:"created_at"`
	TotalAmount int            `json:"total_amount"`
	Items       []CheckoutItem `json:"items"`
	Notes       string         `json:"notes"`
}

type CheckoutItem struct {
	Type      CartItemType `json:"type"`
	ID        string       `json:"id"`
	Name      string       `json:"name"`
	Quantity  int          `json:"quantity"`
	UnitPrice int          `json:"unit_price"`
	Subtotal  int          `json:"subtotal"`
}

func (p *CheckoutPayload) Hash() string {
	data, _ := json.Marshal(p.Items)
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

type Sale struct {
	ID             string              `json:"id"`
	UserID         string              `json:"user_id"`
	IdempotencyKey string              `json:"idempotency_key"`
	TotalAmount    int                 `json:"total_amount"`
	Notes          string              `json:"notes"`
	Status         string              `json:"status"`
	SparepartItems []SaleSparepartItem `json:"sparepart_items"`
	OperationItems []SaleOperationItem `json:"operation_items"`
	CreatedAt      time.Time           `json:"created_at"`
}

type SaleSparepartItem struct {
	ID          string    `json:"id"`
	SaleID      string    `json:"sale_id"`
	SparepartID string    `json:"sparepart_id"`
	Quantity    int       `json:"quantity"`
	UnitPrice   int       `json:"unit_price"`
	Subtotal    int       `json:"subtotal"`
	CreatedAt   time.Time `json:"created_at"`
}

type SaleOperationItem struct {
	ID          string    `json:"id"`
	SaleID      string    `json:"sale_id"`
	OperationID string    `json:"operation_id"`
	Quantity    int       `json:"quantity"`
	UnitPrice   int       `json:"unit_price"`
	Subtotal    int       `json:"subtotal"`
	CreatedAt   time.Time `json:"created_at"`
}

type IdempotencyKey struct {
	Key            string    `json:"key"`
	UserID         string    `json:"user_id"`
	RequestHash    string    `json:"request_hash"`
	ResponseStatus *int      `json:"response_status,omitempty"`
	ResponseBody   []byte    `json:"response_body,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	ExpiresAt      time.Time `json:"expires_at"`
}

const (
	SaleStatusCompleted = "completed"
	SaleStatusRefunded  = "refunded"
	SaleStatusVoided    = "voided"
)

const IdempotencyTTL = 24 * time.Hour
