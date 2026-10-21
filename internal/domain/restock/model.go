package restock

import "time"

type RestockItem struct {
	SparepartID string `json:"sparepart_id"`
	Quantity    int    `json:"quantity"`
	UnitCost    int    `json:"unit_cost"`
}

type CreateRestockPayload struct {
	IdempotencyKey string        `json:"idempotency_key"`
	Items          []RestockItem `json:"items"`
	Notes          string        `json:"notes"`
}

type CreateRestockResult struct {
	RestockID string `json:"restock_id"`
}

type Restock struct {
	ID             string                `json:"id"`
	UserID         string                `json:"user_id"`
	IdempotencyKey string                `json:"idempotency_key"`
	TotalAmount    int                   `json:"total_amount"`
	Notes          string                `json:"notes"`
	Items          []RestockItemDetail   `json:"items"`
	CreatedAt      time.Time             `json:"created_at"`
}

type RestockItemDetail struct {
	ID          string    `json:"id"`
	RestockID   string    `json:"restock_id"`
	SparepartID string    `json:"sparepart_id"`
	Quantity    int       `json:"quantity"`
	UnitCost    int       `json:"unit_cost"`
	Subtotal    int       `json:"subtotal"`
	CreatedAt   time.Time `json:"created_at"`
}
