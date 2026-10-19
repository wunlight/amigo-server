package v1

import (
	"amigo-server/internal/config"
	"amigo-server/internal/domain/auth"
	"amigo-server/internal/domain/sparepart"
	"amigo-server/internal/middleware"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func RegisterSparepartRoutes(r chi.Router, pool *pgxpool.Pool, cfg *config.Config) {
	userRepo := sparepart.NewRepository(pool)
	userSvc := sparepart.NewService(userRepo)
	userHandler := sparepart.NewHandler(userSvc)

	authRepo := auth.NewRepository(pool)
	authSvc := auth.NewService(authRepo, cfg.JWTSecret, parseDuration(cfg.JWTExpiry))

	r.Route("/spareparts", func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(authSvc))
		r.Get("/", userHandler.List)
		r.Post("/", userHandler.Create)
		r.Get("/{id}", userHandler.GetByID)
		r.Put("/{id}", userHandler.Update)
		r.Patch("/{id}", userHandler.Update)
		r.Delete("/{id}", userHandler.Delete)
	})
}
