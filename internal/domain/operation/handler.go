package operation

import (
	"amigo-server/pkg/apperror"
	"amigo-server/pkg/response"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	svc Service
}

func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var input CreateOperationInput
	if err := response.DecodeJSON(r, &input); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	op, err := h.svc.Create(r.Context(), input)
	if err != nil {
		response.Error(w, apperror.Status(err), apperror.Message(err))
		return
	}

	response.Success(w, http.StatusCreated, op)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	page, _ := strconv.Atoi(q.Get("page"))
	perPage, _ := strconv.Atoi(q.Get("per_page"))
	params := ListParams{Page: page, PerPage: perPage, Search: q.Get("search")}
	operations, total, err := h.svc.List(r.Context(), params)

	if err != nil {
		response.Error(w, apperror.Status(err), apperror.Message(err))
		return
	}
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 10
	}
	totalPage := 0
	if total > 0 {
		totalPage = (total + perPage - 1) / perPage
	}

	response.Paginated(w, http.StatusOK, operations, response.Meta{
		Page: page, PerPage: perPage, Total: total, TotalPage: totalPage,
	})
}

func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	op, svcErr := h.svc.GetByID(r.Context(), id)
	if svcErr != nil {
		response.Error(w, apperror.Status(svcErr), apperror.Message(svcErr))
		return
	}

	response.Success(w, http.StatusOK, op)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var input UpdateOperationInput
	if err := response.DecodeJSON(r, &input); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	op, svcErr := h.svc.Update(r.Context(), id, input)
	if svcErr != nil {
		response.Error(w, apperror.Status(svcErr), apperror.Message(svcErr))
		return
	}

	response.Success(w, http.StatusOK, op)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if svcErr := h.svc.Delete(r.Context(), id); svcErr != nil {
		response.Error(w, apperror.Status(svcErr), apperror.Message(svcErr))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
