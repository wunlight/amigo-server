package sale

import (
	"amigo-server/internal/middleware"
	"amigo-server/pkg/apperror"
	"amigo-server/pkg/response"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
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

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userUUID, ok := middleware.GetUserUUID(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	page := 1
	perPage := 10

	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if perPageStr := r.URL.Query().Get("per_page"); perPageStr != "" {
		if pp, err := strconv.Atoi(perPageStr); err == nil && pp > 0 && pp <= 100 {
			perPage = pp
		}
	}

	sales, total, err := h.svc.ListSales(r.Context(), userUUID, page, perPage)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to list sales")
		return
	}

	totalPages := (total + perPage - 1) / perPage
	meta := response.Meta{
		Page:      page,
		PerPage:   perPage,
		Total:     total,
		TotalPage: totalPages,
	}

	response.Paginated(w, http.StatusOK, sales, meta)
}

func (h *Handler) Refund(w http.ResponseWriter, r *http.Request) {
	userUUID, ok := middleware.GetUserUUID(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	saleID := chi.URLParam(r, "id")
	if saleID == "" {
		response.Error(w, http.StatusBadRequest, "sale id is required")
		return
	}

	err := h.svc.RefundSale(r.Context(), saleID, userUUID)
	if err != nil {
		statusCode := http.StatusInternalServerError
		errMsg := err.Error()

		if errMsg == "sale not found" {
			statusCode = http.StatusNotFound
		} else if errMsg == "unauthorized: sale does not belong to user" {
			statusCode = http.StatusForbidden
		} else if errMsg == "sale is already refunded" {
			statusCode = http.StatusConflict
		}

		response.Error(w, statusCode, errMsg)
		return
	}

	response.Success(w, http.StatusOK, map[string]string{
		"message": "sale refunded successfully",
		"sale_id": saleID,
	})
}