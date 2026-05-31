package handlers

import (
	"net/http"

	"github.com/finance-sh/finance-sh/internal/middlewares"
	"github.com/finance-sh/finance-sh/internal/services"
	"github.com/finance-sh/finance-sh/pkg/response"
)

type ForecastHandler struct {
	forecast *services.ForecastService
}

func NewForecastHandler(forecast *services.ForecastService) *ForecastHandler {
	return &ForecastHandler{forecast: forecast}
}

// Forecast GET /cashflow/forecast?months=3 — projects the org's cash flow.
func (h *ForecastHandler) Forecast(w http.ResponseWriter, r *http.Request) {
	orgID := middlewares.OrgID(r.Context())
	months := atoiDefault(r.URL.Query().Get("months"), 3)
	resp, err := h.forecast.Forecast(orgID, months)
	if writeServiceError(w, err) {
		return
	}
	response.OK(w, resp)
}
