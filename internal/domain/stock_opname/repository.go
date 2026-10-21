package stock_opname

import (
	"context"
	"errors"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RepositoryInterface interface {
	CreateStockOpname(ctx context.Context, tx pgx.Tx, opname *StockOpname) error
	CreateStockOpnameItems(ctx context.Context, tx pgx.Tx, items []StockOpnameItemDetail) error
	GetStockOpnameByIdempotencyKey(ctx context.Context, idempotencyKey string) (*StockOpname, error)
	GetStockOverview(ctx context.Context) (*StockOverview, error)
}

type Repository struct {
	db     *pgxpool.Pool
	logger *slog.Logger
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{
		db:     db,
		logger: slog.Default(),
	}
}

func (r *Repository) WithLogger(logger *slog.Logger) *Repository {
	r.logger = logger
	return r
}

func (r *Repository) CreateStockOpname(ctx context.Context, tx pgx.Tx, opname *StockOpname) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO stock_opnames (id, user_id, idempotency_key, notes, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`, opname.ID, opname.UserID, opname.IdempotencyKey, opname.Notes, opname.CreatedAt)

	if err != nil {
		r.logger.Error("failed to create stock opname", "stock_opname_id", opname.ID, "user_id", opname.UserID, "error", err)
		return err
	}

	r.logger.Info("stock opname created", "stock_opname_id", opname.ID, "user_id", opname.UserID)
	return nil
}

func (r *Repository) CreateStockOpnameItems(ctx context.Context, tx pgx.Tx, items []StockOpnameItemDetail) error {
	for _, item := range items {
		_, err := tx.Exec(ctx, `
			INSERT INTO stock_opname_items (id, stock_opname_id, sparepart_id, quantity, created_at)
			VALUES ($1, $2, $3, $4, $5)
		`, item.ID, item.StockOpnameID, item.SparepartID, item.Quantity, item.CreatedAt)

		if err != nil {
			r.logger.Error("failed to create stock opname item", "stock_opname_id", item.StockOpnameID, "sparepart_id", item.SparepartID, "error", err)
			return err
		}
	}

	r.logger.Debug("stock opname items created", "stock_opname_id", items[0].StockOpnameID, "count", len(items))
	return nil
}

func (r *Repository) GetStockOpnameByIdempotencyKey(ctx context.Context, idempotencyKey string) (*StockOpname, error) {
	var opname StockOpname

	err := r.db.QueryRow(ctx, `
		SELECT id, user_id, idempotency_key, notes, created_at
		FROM stock_opnames
		WHERE idempotency_key = $1
	`, idempotencyKey).Scan(&opname.ID, &opname.UserID, &opname.IdempotencyKey, &opname.Notes, &opname.CreatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		r.logger.Error("failed to get stock opname by idempotency key", "idempotency_key", idempotencyKey, "error", err)
		return nil, err
	}

	r.logger.Debug("found stock opname by idempotency key", "idempotency_key", idempotencyKey, "stock_opname_id", opname.ID)
	return &opname, nil
}

func (r *Repository) GetStockOverview(ctx context.Context) (*StockOverview, error) {
	var totalSpareparts int
	var lowStockItems int
	var outOfStockItems int
	var totalInventoryValue int

	err := r.db.QueryRow(ctx, `
		SELECT 
			COUNT(*) as total_spareparts,
			COUNT(CASE WHEN stock > 0 AND stock <= $1 THEN 1 END) as low_stock_items,
			COUNT(CASE WHEN stock = 0 THEN 1 END) as out_of_stock_items,
			COALESCE(SUM(stock * selling_price), 0) as total_inventory_value
		FROM spareparts
	`, LowStockThreshold).Scan(&totalSpareparts, &lowStockItems, &outOfStockItems, &totalInventoryValue)

	if err != nil {
		r.logger.Error("failed to get stock overview", "error", err)
		return nil, err
	}

	return &StockOverview{
		TotalSpareparts:     totalSpareparts,
		LowStockItems:       lowStockItems,
		OutOfStockItems:     outOfStockItems,
		TotalInventoryValue: totalInventoryValue,
	}, nil
}
