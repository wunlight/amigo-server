package v1

import (
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"amigo-server/internal/config"
	"amigo-server/internal/domain/auth"
)

func RegisterAuthRoutes(r chi.Router, pool *pgxpool.Pool, cfg *config.Config) {
	authRepo := auth.NewRepository(pool)
	authSvc := auth.NewService(authRepo, cfg.JWTSecret, parseDuration(cfg.JWTExpiry))
	authHandler := auth.NewHandler(authSvc)

	authHandler.RegisterRoutes(r)
}

func parseDuration(s string) time.Duration {
	d, _ := time.ParseDuration(s)
	if d == 0 {
		return 24 * time.Hour
	}
	return d
}