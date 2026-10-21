package restock

import (
	"context"
	"errors"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RepositoryInterface interface {
	CreateRestock(ctx context.Context, tx pgx.Tx, restock *Restock) error
	CreateRestockItems(ctx context.Context, tx pgx.Tx, items []RestockItemDetail) error
	GetRestockByIdempotencyKey(ctx context.Context, idempotencyKey string) (*Restock, error)
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

func (r *Repository) CreateRestock(ctx context.Context, tx pgx.Tx, restock *Restock) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO restocks (id, user_id, idempotency_key, total_amount, notes, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, restock.ID, restock.UserID, restock.IdempotencyKey, restock.TotalAmount, restock.Notes, restock.CreatedAt)

	if err != nil {
		r.logger.Error("failed to create restock", "restock_id", restock.ID, "user_id", restock.UserID, "error", err)
		return err
	}

	r.logger.Info("restock created", "restock_id", restock.ID, "user_id", restock.UserID, "total_amount", restock.TotalAmount)
	return nil
}

func (r *Repository) CreateRestockItems(ctx context.Context, tx pgx.Tx, items []RestockItemDetail) error {
	for _, item := range items {
		_, err := tx.Exec(ctx, `
			INSERT INTO restock_items (id, restock_id, sparepart_id, quantity, unit_cost, subtotal, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`, item.ID, item.RestockID, item.SparepartID, item.Quantity, item.UnitCost, item.Subtotal, item.CreatedAt)

		if err != nil {
			r.logger.Error("failed to create restock item", "restock_id", item.RestockID, "sparepart_id", item.SparepartID, "error", err)
			return err
		}
	}

	r.logger.Debug("restock items created", "restock_id", items[0].RestockID, "count", len(items))
	return nil
}

func (r *Repository) GetRestockByIdempotencyKey(ctx context.Context, idempotencyKey string) (*Restock, error) {
	var restock Restock

	err := r.db.QueryRow(ctx, `
		SELECT id, user_id, idempotency_key, total_amount, notes, created_at
		FROM restocks
		WHERE idempotency_key = $1
	`, idempotencyKey).Scan(&restock.ID, &restock.UserID, &restock.IdempotencyKey, &restock.TotalAmount, &restock.Notes, &restock.CreatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		r.logger.Error("failed to get restock by idempotency key", "idempotency_key", idempotencyKey, "error", err)
		return nil, err
	}

	r.logger.Debug("found restock by idempotency key", "idempotency_key", idempotencyKey, "restock_id", restock.ID)
	return &restock, nil
}
