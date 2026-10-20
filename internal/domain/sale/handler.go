package sale

import (
	"amigo-server/internal/middleware"
	"amigo-server/pkg/apperror"
	"amigo-server/pkg/response"
	"net/http"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Checkout(w http.ResponseWriter, r *http.Request) {
	var payload CheckoutPayload
	if err := response.DecodeJSON(r, &payload); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	userUUID, ok := middleware.GetUserUUID(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	result, err := h.svc.Checkout(r.Context(), userUUID, payload)
	if err != nil {
		response.Error(w, apperror.Status(err), apperror.Message(err))
		return
	}

	response.Success(w, http.StatusCreated, result)
}