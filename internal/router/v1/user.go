package v1

import (
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"amigo-server/internal/domain/user"
)

func RegisterUserRoutes(r chi.Router, pool *pgxpool.Pool) {
	userRepo := user.NewRepository(pool)
	userSvc := user.NewService(userRepo)
	userHandler := user.NewHandler(userSvc)

	r.Route("/users", func(r chi.Router) {
		r.Get("/", userHandler.List)
		r.Post("/", userHandler.Create)
		r.Get("/{id}", userHandler.GetByID)
		r.Put("/{id}", userHandler.Update)
		r.Patch("/{id}", userHandler.Update)
		r.Delete("/{id}", userHandler.Delete)
	})
}
