package sale

import (
	"context"
	"errors"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrIdempotencyKeyNotFound = errors.New("idempotency key not found")
	ErrIdempotencyKeyExists   = errors.New("idempotency key already exists")
)

type RepositoryInterface interface {
	GetSaleByIdempotencyKey(ctx context.Context, key string) (*Sale, error)
	CreateSale(ctx context.Context, tx pgx.Tx, sale *Sale) error
	CreateSparepartItems(ctx context.Context, tx pgx.Tx, items []SaleSparepartItem) error
	CreateOperationItems(ctx context.Context, tx pgx.Tx, items []SaleOperationItem) error
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

func (r *Repository) GetSaleByIdempotencyKey(ctx context.Context, key string) (*Sale, error) {
	var sale Sale
	err := r.db.QueryRow(ctx, `
		SELECT id, user_id, idempotency_key, total_amount, notes, status, created_at
		FROM sales
		WHERE idempotency_key = $1
	`, key).Scan(&sale.ID, &sale.UserID, &sale.IdempotencyKey, &sale.TotalAmount, &sale.Notes, &sale.Status, &sale.CreatedAt)

	if err == pgx.ErrNoRows {
		r.logger.Debug("sale not found for idempotency key", "key", key)
		return nil, ErrIdempotencyKeyNotFound
	}

	if err != nil {
		r.logger.Error("failed to get sale by idempotency key", "key", key, "error", err)
		return nil, err
	}

	return &sale, nil
}

func (r *Repository) CreateSale(ctx context.Context, tx pgx.Tx, sale *Sale) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO sales (id, user_id, idempotency_key, total_amount, notes, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, sale.ID, sale.UserID, sale.IdempotencyKey, sale.TotalAmount, sale.Notes, sale.Status, sale.CreatedAt, sale.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			r.logger.Debug("idempotency key already exists", "key", sale.IdempotencyKey)
			return ErrIdempotencyKeyExists
		}
		r.logger.Error("failed to create sale", "sale_id", sale.ID, "user_id", sale.UserID, "error", err)
		return err
	}
	r.logger.Info("sale created", "sale_id", sale.ID, "user_id", sale.UserID, "total_amount", sale.TotalAmount)
	return nil
}

func (r *Repository) CreateSparepartItems(ctx context.Context, tx pgx.Tx, items []SaleSparepartItem) error {
	for _, item := range items {
		_, err := tx.Exec(ctx, `
			INSERT INTO sale_sparepart_items (id, sale_id, sparepart_id, quantity, unit_price, subtotal, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`, item.ID, item.SaleID, item.SparepartID, item.Quantity, item.UnitPrice, item.Subtotal, item.CreatedAt)
		if err != nil {
			r.logger.Error("failed to create sparepart item", "sale_id", item.SaleID, "sparepart_id", item.SparepartID, "error", err)
			return err
		}
	}
	r.logger.Debug("sparepart items created", "sale_id", items[0].SaleID, "count", len(items))
	return nil
}

func (r *Repository) CreateOperationItems(ctx context.Context, tx pgx.Tx, items []SaleOperationItem) error {
	for _, item := range items {
		_, err := tx.Exec(ctx, `
			INSERT INTO sale_operation_items (id, sale_id, operation_id, quantity, unit_price, subtotal, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`, item.ID, item.SaleID, item.OperationID, item.Quantity, item.UnitPrice, item.Subtotal, item.CreatedAt)
		if err != nil {
			r.logger.Error("failed to create operation item", "sale_id", item.SaleID, "operation_id", item.OperationID, "error", err)
			return err
		}
	}
	r.logger.Debug("operation items created", "sale_id", items[0].SaleID, "count", len(items))
	return nil
}
