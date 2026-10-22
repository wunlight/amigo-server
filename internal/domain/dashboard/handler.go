package dashboard

import (
	"amigo-server/pkg/response"
	"net/http"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// GetOverview menghandle GET /dashboard/overview
func (h *Handler) GetOverview(w http.ResponseWriter, r *http.Request) {
	overview, err := h.svc.GetOverview(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to get overview data")
		return
	}

	response.Success(w, http.StatusOK, overview)
}

// GetDashboard menghandle GET /dashboard
func (h *Handler) GetDashboard(w http.ResponseWriter, r *http.Request) {
	dashboard, err := h.svc.GetDashboard(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to get dashboard data")
		return
	}

	response.Success(w, http.StatusOK, dashboard)
}
