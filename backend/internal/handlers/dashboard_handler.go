package handlers

import (
	"net/http"

	"github.com/finance-sh/finance-sh/internal/middlewares"
	"github.com/finance-sh/finance-sh/internal/services"
	"github.com/finance-sh/finance-sh/pkg/response"
)

type DashboardHandler struct {
	dashboard *services.DashboardService
}

func NewDashboardHandler(dashboard *services.DashboardService) *DashboardHandler {
	return &DashboardHandler{dashboard: dashboard}
}

// Overview GET /dashboard
func (h *DashboardHandler) Overview(w http.ResponseWriter, r *http.Request) {
	orgID := middlewares.OrgID(r.Context())
	data, err := h.dashboard.Overview(orgID)
	if writeServiceError(w, err) {
		return
	}
	response.OK(w, data)
}
