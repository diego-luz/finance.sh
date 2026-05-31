package handlers

import (
	"net/http"

	"github.com/finance-sh/finance-sh/pkg/response"
)

type HealthHandler struct{}

func NewHealthHandler() *HealthHandler { return &HealthHandler{} }

// Health GET /health — liveness probe.
func (h *HealthHandler) Health(w http.ResponseWriter, _ *http.Request) {
	response.OK(w, map[string]string{"status": "ok"})
}
