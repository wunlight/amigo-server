package dashboard

import "time"

// OverviewResponse untuk GET /dashboard/overview
type OverviewResponse struct {
	DailyRevenue       int                `json:"daily_revenue"`
	TodaysTransactions int                `json:"todays_transactions"`
	LowStockItems      []StockAlertItem   `json:"low_stock_items"`
	OutOfStockItems    []StockAlertItem   `json:"out_of_stock_items"`
}

type StockAlertItem struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	CurrentStock int    `json:"current_stock"`
	SellingPrice int    `json:"selling_price"`
}

// DashboardResponse untuk GET /dashboard
type DashboardResponse struct {
	RevenueData      []RevenueTrend      `json:"revenue_data"`
	TopSoldSparepart []TopSparepartItem  `json:"top_sold_sparepart"`
	RecentSales      []RecentSaleItem    `json:"recent_sales"`
	CriticalItems    []CriticalStockItem `json:"critical_items"`
}

type RevenueTrend struct {
	Date   string `json:"date"`
	Amount int    `json:"amount"`
}

type TopSparepartItem struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	TotalQty  int    `json:"total_qty"`
	Revenue   int    `json:"revenue"`
	Thumbnail string `json:"thumbnail"`
}

type RecentSaleItem struct {
	SaleID      string    `json:"sale_id"`
	TotalAmount int       `json:"total_amount"`
	ItemCount   int       `json:"item_count"`
	CreatedAt   time.Time `json:"created_at"`
}

type CriticalStockItem struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	CurrentStock int    `json:"current_stock"`
	SellingPrice int    `json:"selling_price"`
	Status       string `json:"status"` // "critical" atau "low"
}

// Constant untuk threshold
const (
	LowStockThreshold      = 5
	OutOfStockThreshold    = 0
	CriticalStockThreshold = 3
)
