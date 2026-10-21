package stock_opname

import "time"

type StockOpnameItem struct {
	SparepartID string `json:"sparepart_id"`
	Quantity    int    `json:"quantity"`
}

type CreateStockOpnamePayload struct {
	IdempotencyKey string               `json:"idempotency_key"`
	Items          []StockOpnameItem    `json:"items"`
	Notes          string               `json:"notes"`
}

type CreateStockOpnameResult struct {
	StockOpnameID string `json:"stock_opname_id"`
}

type StockOpname struct {
	ID             string                    `json:"id"`
	UserID         string                    `json:"user_id"`
	IdempotencyKey string                    `json:"idempotency_key"`
	Notes          string                    `json:"notes"`
	Items          []StockOpnameItemDetail   `json:"items"`
	CreatedAt      time.Time                 `json:"created_at"`
}

type StockOpnameItemDetail struct {
	ID            string    `json:"id"`
	StockOpnameID string    `json:"stock_opname_id"`
	SparepartID   string    `json:"sparepart_id"`
	Quantity      int       `json:"quantity"`
	CreatedAt     time.Time `json:"created_at"`
}

type StockOverview struct {
	TotalSpareparts     int `json:"total_spareparts"`
	LowStockItems       int `json:"low_stock_items"`
	OutOfStockItems     int `json:"out_of_stock_items"`
	TotalInventoryValue int `json:"total_inventory_value"`
}

const LowStockThreshold = 5

