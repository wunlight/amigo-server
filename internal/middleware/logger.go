package middleware

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

func Logger() func(http.Handler) http.Handler {
	return middleware.Logger
}

func RequestID() func(http.Handler) http.Handler {
	return middleware.RequestID
}

func RealIP() func(http.Handler) http.Handler {
	return middleware.RealIP
}

func Recoverer() func(http.Handler) http.Handler {
	return middleware.Recoverer
}

func Timeout(d time.Duration) func(http.Handler) http.Handler {
	return middleware.Timeout(d)
}
