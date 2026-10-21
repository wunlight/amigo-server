package v1

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"amigo-server/internal/config"
)

func RegisterRoutes(r chi.Router, pool *pgxpool.Pool, cfg *config.Config) {
	r.Get("/health", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"success":true,"data":{"status":"ok","version":"v1"}}`))
	})

	RegisterAuthRoutes(r, pool, cfg)
	RegisterUserRoutes(r, pool, cfg)
	RegisterSparepartRoutes(r, pool, cfg)
	RegisterOperationRoutes(r, pool, cfg)
	RegisterSaleRoutes(r, pool, cfg)
	RegisterStockOpnameRoutes(r, pool, cfg)
	RegisterRestockRoutes(r, pool, cfg)
}
