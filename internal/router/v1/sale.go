package v1

import (
	"amigo-server/internal/config"
	"amigo-server/internal/domain/auth"
	"amigo-server/internal/domain/sale"
	"amigo-server/internal/middleware"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func RegisterSaleRoutes(r chi.Router, pool *pgxpool.Pool, cfg *config.Config) {
	saleSvc := sale.NewService(pool)
	saleHandler := sale.NewHandler(saleSvc)

	authRepo := auth.NewRepository(pool)
	authSvc := auth.NewService(authRepo, cfg.JWTSecret, parseDuration(cfg.JWTExpiry))

	r.Route("/sales", func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(authSvc))
		r.Get("/", saleHandler.List)
		r.Post("/checkout", saleHandler.Checkout)
		r.Post("/{id}/refund", saleHandler.Refund)
	})
}