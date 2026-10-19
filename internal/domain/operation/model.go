package operation

import "time"

type Operation struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	SellingPrice int       `json:"selling_price"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type CreateOperationInput struct {
	Name         string `json:"name"`
	SellingPrice int    `json:"selling_price"`
}

type UpdateOperationInput struct {
	Name         *string `json:"name"`
	SellingPrice *int    `json:"selling_price"`
}

type ListParams struct {
	Page    int
	PerPage int
	Search  string
}
