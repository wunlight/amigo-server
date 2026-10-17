package v1

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func New(pool *pgxpool.Pool) http.Handler {
	r := chi.NewRouter()

	r.Get("/health", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"success":true,"data":{"status":"ok","version":"v1"}}`))
	})

	RegisterUserRoutes(r, pool)

	return r
}
