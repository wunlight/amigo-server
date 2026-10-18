package auth

import (
	"net/http"

	"amigo-server/pkg/apperror"
	"amigo-server/pkg/response"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	svc Service
}

func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var input LoginInput
	if err := response.DecodeJSON(r, &input); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	token, loginResp, err := h.svc.Login(r.Context(), input)
	if err != nil {
		response.Error(w, apperror.Status(err), apperror.Message(err))
		return
	}

	setAuthCookie(w, token)
	response.Success(w, http.StatusOK, loginResp)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	clearAuthCookie(w)
	response.Success(w, http.StatusOK, map[string]string{"message": "logged out"})
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	token, err := getAuthCookie(r)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	meResp, err := h.svc.Me(r.Context(), token)
	if err != nil {
		response.Error(w, apperror.Status(err), apperror.Message(err))
		return
	}

	response.Success(w, http.StatusOK, meResp)
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Post("/login", h.Login)
	r.Post("/logout", h.Logout)
	r.Get("/me", h.Me)
}

const authCookieName = "access_token"

func setAuthCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     authCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: http.SameSiteLaxMode,
		MaxAge:   60 * 60 * 24, // 24 hours
	})
}

func clearAuthCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     authCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
}

func getAuthCookie(r *http.Request) (string, error) {
	cookie, err := r.Cookie(authCookieName)
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}