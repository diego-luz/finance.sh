package handlers

import (
	"net/http"

	"github.com/finance-sh/finance-sh/internal/middlewares"
	"github.com/finance-sh/finance-sh/internal/services"
	"github.com/finance-sh/finance-sh/pkg/response"
)

type SearchHandler struct {
	search *services.SearchService
}

func NewSearchHandler(search *services.SearchService) *SearchHandler {
	return &SearchHandler{search: search}
}

// Search GET /search?q=&limit=5 — org-scoped global search across the main
// domains. An empty q returns empty groups; limit is clamped to 1..20 (default 5).
func (h *SearchHandler) Search(w http.ResponseWriter, r *http.Request) {
	orgID := middlewares.OrgID(r.Context())
	q := r.URL.Query().Get("q")
	limit := atoiDefault(r.URL.Query().Get("limit"), 5)

	result, err := h.search.Search(orgID, q, limit)
	if writeServiceError(w, err) {
		return
	}
	response.OK(w, result)
}
