package middleware

import (
	"context"
	"net/http"

	"amigo-server/internal/domain/auth"
	"amigo-server/pkg/apperror"
	"amigo-server/pkg/response"
)

type contextKey string

const UserIDKey contextKey = "user_id"

func AuthMiddleware(authSvc auth.Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, err := getAuthCookie(r)
			if err != nil {
				response.Error(w, http.StatusUnauthorized, "unauthorized")
				return
			}

			meResp, err := authSvc.Me(r.Context(), token)
			if err != nil {
				response.Error(w, apperror.Status(err), apperror.Message(err))
				return
			}

			ctx := r.Context()
			ctx = contextWithUserID(ctx, meResp.Name)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func contextWithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, UserIDKey, userID)
}

func GetUserID(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(UserIDKey).(string)
	return userID, ok
}

const authCookieName = "access_token"

func getAuthCookie(r *http.Request) (string, error) {
	cookie, err := r.Cookie(authCookieName)
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

func OptionalAuthMiddleware(authSvc auth.Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, err := getAuthCookie(r)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			meResp, err := authSvc.Me(r.Context(), token)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			ctx := r.Context()
			ctx = contextWithUserID(ctx, meResp.Name)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func CorsPreflightMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodOptions {
				w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-CSRF-Token")
				w.Header().Set("Access-Control-Allow-Credentials", "true")
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}