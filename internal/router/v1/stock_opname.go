package v1

import (
	"amigo-server/internal/config"
	"amigo-server/internal/domain/auth"
	"amigo-server/internal/domain/stock_opname"
	"amigo-server/internal/middleware"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func RegisterStockOpnameRoutes(r chi.Router, pool *pgxpool.Pool, cfg *config.Config) {
	opnameSvc := stock_opname.NewService(pool)
	opnameHandler := stock_opname.NewHandler(opnameSvc)

	authRepo := auth.NewRepository(pool)
	authSvc := auth.NewService(authRepo, cfg.JWTSecret, parseDuration(cfg.JWTExpiry))

	r.Route("/stocks", func(r chi.Router) {
		r.Get("/overview", opnameHandler.GetOverview)
		r.Route("/opnames", func(r chi.Router) {
			r.Use(middleware.AuthMiddleware(authSvc))
			r.Post("/", opnameHandler.Create)
		})
	})
}
