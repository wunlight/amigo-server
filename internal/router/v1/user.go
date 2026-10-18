package v1

import (
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"amigo-server/internal/config"
	"amigo-server/internal/domain/auth"
	"amigo-server/internal/domain/user"
	"amigo-server/internal/middleware"
)

func RegisterUserRoutes(r chi.Router, pool *pgxpool.Pool, cfg *config.Config) {
	userRepo := user.NewRepository(pool)
	userSvc := user.NewService(userRepo)
	userHandler := user.NewHandler(userSvc)

	authRepo := auth.NewRepository(pool)
	authSvc := auth.NewService(authRepo, cfg.JWTSecret, parseDuration(cfg.JWTExpiry))

	r.Route("/users", func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(authSvc))
		r.Get("/", userHandler.List)
		r.Post("/", userHandler.Create)
		r.Get("/{id}", userHandler.GetByID)
		r.Put("/{id}", userHandler.Update)
		r.Patch("/{id}", userHandler.Update)
		r.Delete("/{id}", userHandler.Delete)
	})
}
