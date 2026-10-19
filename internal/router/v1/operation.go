package v1

import (
	"amigo-server/internal/config"
	"amigo-server/internal/domain/auth"
	"amigo-server/internal/domain/operation"
	"amigo-server/internal/middleware"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func RegisterOperationRoutes(r chi.Router, pool *pgxpool.Pool, cfg *config.Config) {
	opRepo := operation.NewRepository(pool)
	opSvc := operation.NewService(opRepo)
	opHandler := operation.NewHandler(opSvc)

	authRepo := auth.NewRepository(pool)
	authSvc := auth.NewService(authRepo, cfg.JWTSecret, parseDuration(cfg.JWTExpiry))

	r.Route("/operations", func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(authSvc))
		r.Get("/", opHandler.List)
		r.Post("/", opHandler.Create)
		r.Get("/{id}", opHandler.GetByID)
		r.Put("/{id}", opHandler.Update)
		r.Delete("/{id}", opHandler.Delete)
	})
}
