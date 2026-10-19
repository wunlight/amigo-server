package sparepart

import "time"

type Sparepart struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	SellingPrice int       `json:"selling_price"`
	Stock        int       `json:"stock"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type CreateSparepartInput struct {
	Name         string `json:"name"`
	SellingPrice int    `json:"selling_price"`
	InitialStock int    `json:"initial_stock"`
}

type UpdateSparepartInput struct {
	Name         *string `json:"name"`
	SellingPrice *int    `json:"selling_price"`
}

type ListParams struct {
	Page    int
	PerPage int
	Search  string
}
