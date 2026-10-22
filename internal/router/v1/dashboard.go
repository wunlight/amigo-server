package v1

import (
	"amigo-server/internal/config"
	"amigo-server/internal/domain/auth"
	"amigo-server/internal/domain/dashboard"
	"amigo-server/internal/middleware"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func RegisterDashboardRoutes(r chi.Router, pool *pgxpool.Pool, cfg *config.Config) {
	dashboardSvc := dashboard.NewService(pool)
	dashboardHandler := dashboard.NewHandler(dashboardSvc)

	authRepo := auth.NewRepository(pool)
	authSvc := auth.NewService(authRepo, cfg.JWTSecret, parseDuration(cfg.JWTExpiry))

	r.Route("/dashboard", func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(authSvc))
		r.Get("/", dashboardHandler.GetDashboard)
		r.Get("/overview", dashboardHandler.GetOverview)
	})
}
