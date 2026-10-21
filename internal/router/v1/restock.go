package v1

import (
	"amigo-server/internal/config"
	"amigo-server/internal/domain/auth"
	"amigo-server/internal/domain/restock"
	"amigo-server/internal/middleware"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func RegisterRestockRoutes(r chi.Router, pool *pgxpool.Pool, cfg *config.Config) {
	restockSvc := restock.NewService(pool)
	restockHandler := restock.NewHandler(restockSvc)

	authRepo := auth.NewRepository(pool)
	authSvc := auth.NewService(authRepo, cfg.JWTSecret, parseDuration(cfg.JWTExpiry))

	r.Route("/stocks/restock", func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(authSvc))
		r.Post("/", restockHandler.Create)
	})
}
