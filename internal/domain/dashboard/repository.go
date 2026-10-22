package dashboard

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
)

type RepositoryInterface interface {
	GetDailyRevenue(ctx context.Context) (int, error)
	GetTodaysTransactionCount(ctx context.Context) (int, error)
	GetLowStockItems(ctx context.Context, threshold int) ([]StockAlertItem, error)
	GetOutOfStockItems(ctx context.Context) ([]StockAlertItem, error)
	GetRevenueTrendLastNDays(ctx context.Context, days int) ([]RevenueTrend, error)
	GetTopSoldSpareparten(ctx context.Context, limit int) ([]TopSparepartItem, error)
	GetRecentSales(ctx context.Context, limit int) ([]RecentSaleItem, error)
	GetCriticalStockItems(ctx context.Context) ([]CriticalStockItem, error)
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

// GetDailyRevenue menghitung total revenue hari ini
func (r *Repository) GetDailyRevenue(ctx context.Context) (int, error) {
	var revenue int

	err := r.db.QueryRow(ctx, `
		SELECT COALESCE(SUM(total_amount), 0)
		FROM sales
		WHERE DATE(created_at) = CURRENT_DATE
		AND status = 'completed'
	`).Scan(&revenue)

	if err != nil {
		r.logger.Error("failed to get daily revenue", "error", err)
		return 0, err
	}

	return revenue, nil
}

// GetTodaysTransactionCount menghitung jumlah transaksi hari ini
func (r *Repository) GetTodaysTransactionCount(ctx context.Context) (int, error) {
	var count int

	err := r.db.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM sales
		WHERE DATE(created_at) = CURRENT_DATE
		AND status = 'completed'
	`).Scan(&count)

	if err != nil {
		r.logger.Error("failed to get todays transaction count", "error", err)
		return 0, err
	}

	return count, nil
}

// GetLowStockItems mengambil items dengan stock di bawah threshold (tapi tidak 0)
func (r *Repository) GetLowStockItems(ctx context.Context, threshold int) ([]StockAlertItem, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, name, stock, selling_price
		FROM spareparts
		WHERE stock > 0 AND stock <= $1
		ORDER BY stock ASC
	`, threshold)

	if err != nil {
		r.logger.Error("failed to get low stock items", "error", err)
		return nil, err
	}
	defer rows.Close()

	var items []StockAlertItem
	for rows.Next() {
		var item StockAlertItem
		if err := rows.Scan(&item.ID, &item.Name, &item.CurrentStock, &item.SellingPrice); err != nil {
			r.logger.Error("failed to scan low stock item", "error", err)
			continue
		}
		items = append(items, item)
	}

	return items, nil
}

// GetOutOfStockItems mengambil items dengan stock 0
func (r *Repository) GetOutOfStockItems(ctx context.Context) ([]StockAlertItem, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, name, stock, selling_price
		FROM spareparts
		WHERE stock = 0
		ORDER BY name ASC
	`)

	if err != nil {
		r.logger.Error("failed to get out of stock items", "error", err)
		return nil, err
	}
	defer rows.Close()

	var items []StockAlertItem
	for rows.Next() {
		var item StockAlertItem
		if err := rows.Scan(&item.ID, &item.Name, &item.CurrentStock, &item.SellingPrice); err != nil {
			r.logger.Error("failed to scan out of stock item", "error", err)
			continue
		}
		items = append(items, item)
	}

	return items, nil
}

// GetRevenueTrendLastNDays mengambil trend revenue N hari terakhir
func (r *Repository) GetRevenueTrendLastNDays(ctx context.Context, days int) ([]RevenueTrend, error) {
	rows, err := r.db.Query(ctx, `
		SELECT 
			DATE(created_at)::TEXT as date,
			COALESCE(SUM(total_amount), 0) as amount
		FROM sales
		WHERE created_at >= NOW() - INTERVAL '1 day' * $1
		AND status = 'completed'
		GROUP BY DATE(created_at)
		ORDER BY DATE(created_at) ASC
	`, days)

	if err != nil {
		r.logger.Error("failed to get revenue trend", "error", err)
		return nil, err
	}
	defer rows.Close()

	var trends []RevenueTrend
	for rows.Next() {
		var trend RevenueTrend
		if err := rows.Scan(&trend.Date, &trend.Amount); err != nil {
			r.logger.Error("failed to scan revenue trend", "error", err)
			continue
		}
		trends = append(trends, trend)
	}

	return trends, nil
}

// GetTopSoldSpareparten mengambil top N sparepart terlaris
func (r *Repository) GetTopSoldSpareparten(ctx context.Context, limit int) ([]TopSparepartItem, error) {
	rows, err := r.db.Query(ctx, `
		SELECT 
			sp.id,
			sp.name,
			SUM(ssi.quantity) as total_qty,
			SUM(ssi.subtotal) as revenue
		FROM sale_sparepart_items ssi
		JOIN spareparts sp ON ssi.sparepart_id = sp.id
		JOIN sales s ON ssi.sale_id = s.id
		WHERE s.status = 'completed'
		GROUP BY sp.id, sp.name
		ORDER BY total_qty DESC
		LIMIT $1
	`, limit)

	if err != nil {
		r.logger.Error("failed to get top sold spareparts", "error", err)
		return nil, err
	}
	defer rows.Close()

	var items []TopSparepartItem
	for rows.Next() {
		var item TopSparepartItem
		if err := rows.Scan(&item.ID, &item.Name, &item.TotalQty, &item.Revenue); err != nil {
			r.logger.Error("failed to scan top sparepart item", "error", err)
			continue
		}
		item.Thumbnail = "" // TODO: tambahkan field thumbnail ke database jika diperlukan
		items = append(items, item)
	}

	return items, nil
}

// GetRecentSales mengambil recent sales terbaru
func (r *Repository) GetRecentSales(ctx context.Context, limit int) ([]RecentSaleItem, error) {
	rows, err := r.db.Query(ctx, `
		SELECT 
			s.id,
			s.total_amount,
			COUNT(ssi.id) + COUNT(soi.id) as item_count,
			s.created_at
		FROM sales s
		LEFT JOIN sale_sparepart_items ssi ON s.id = ssi.sale_id
		LEFT JOIN sale_operation_items soi ON s.id = soi.sale_id
		WHERE s.status = 'completed'
		GROUP BY s.id, s.total_amount, s.created_at
		ORDER BY s.created_at DESC
		LIMIT $1
	`, limit)

	if err != nil {
		r.logger.Error("failed to get recent sales", "error", err)
		return nil, err
	}
	defer rows.Close()

	var sales []RecentSaleItem
	for rows.Next() {
		var sale RecentSaleItem
		if err := rows.Scan(&sale.SaleID, &sale.TotalAmount, &sale.ItemCount, &sale.CreatedAt); err != nil {
			r.logger.Error("failed to scan recent sale", "error", err)
			continue
		}
		sales = append(sales, sale)
	}

	return sales, nil
}

// GetCriticalStockItems mengambil items yang perlu di-restock segera (low stock & out of stock)
func (r *Repository) GetCriticalStockItems(ctx context.Context) ([]CriticalStockItem, error) {
	rows, err := r.db.Query(ctx, `
		SELECT 
			id,
			name,
			stock,
			selling_price,
			CASE 
				WHEN stock = 0 THEN 'out_of_stock'
				WHEN stock <= $1 THEN 'critical'
				ELSE 'low'
			END as status
		FROM spareparts
		WHERE stock <= $2
		ORDER BY stock ASC, name ASC
	`, CriticalStockThreshold, LowStockThreshold)

	if err != nil {
		r.logger.Error("failed to get critical stock items", "error", err)
		return nil, err
	}
	defer rows.Close()

	var items []CriticalStockItem
	for rows.Next() {
		var item CriticalStockItem
		if err := rows.Scan(&item.ID, &item.Name, &item.CurrentStock, &item.SellingPrice, &item.Status); err != nil {
			r.logger.Error("failed to scan critical stock item", "error", err)
			continue
		}
		items = append(items, item)
	}

	return items, nil
}
