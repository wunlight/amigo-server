package dashboard

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	repo RepositoryInterface
}

func NewService(db *pgxpool.Pool) *Service {
	return &Service{
		repo: NewRepository(db),
	}
}

func (s *Service) WithLogger(logger *slog.Logger) *Service {
	s.repo.(*Repository).WithLogger(logger)
	return s
}

// GetOverview mengambil data overview untuk dashboard
func (s *Service) GetOverview(ctx context.Context) (*OverviewResponse, error) {
	dailyRevenue, err := s.repo.GetDailyRevenue(ctx)
	if err != nil {
		return nil, err
	}

	todaysTransactions, err := s.repo.GetTodaysTransactionCount(ctx)
	if err != nil {
		return nil, err
	}

	lowStockItems, err := s.repo.GetLowStockItems(ctx, LowStockThreshold)
	if err != nil {
		return nil, err
	}

	outOfStockItems, err := s.repo.GetOutOfStockItems(ctx)
	if err != nil {
		return nil, err
	}

	return &OverviewResponse{
		DailyRevenue:       dailyRevenue,
		TodaysTransactions: todaysTransactions,
		LowStockItems:      lowStockItems,
		OutOfStockItems:    outOfStockItems,
	}, nil
}

// GetDashboard mengambil semua data untuk full dashboard
func (s *Service) GetDashboard(ctx context.Context) (*DashboardResponse, error) {
	revenueData, err := s.repo.GetRevenueTrendLastNDays(ctx, 7)
	if err != nil {
		return nil, err
	}

	topSoldSparepart, err := s.repo.GetTopSoldSpareparten(ctx, 5)
	if err != nil {
		return nil, err
	}

	recentSales, err := s.repo.GetRecentSales(ctx, 10)
	if err != nil {
		return nil, err
	}

	criticalItems, err := s.repo.GetCriticalStockItems(ctx)
	if err != nil {
		return nil, err
	}

	return &DashboardResponse{
		RevenueData:      revenueData,
		TopSoldSparepart: topSoldSparepart,
		RecentSales:      recentSales,
		CriticalItems:    criticalItems,
	}, nil
}
