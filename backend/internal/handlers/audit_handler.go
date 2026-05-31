package handlers

import (
	"net/http"
	"time"

	"github.com/finance-sh/finance-sh/internal/dto"
	"github.com/finance-sh/finance-sh/internal/middlewares"
	"github.com/finance-sh/finance-sh/internal/services"
	"github.com/finance-sh/finance-sh/pkg/response"
	"github.com/google/uuid"
)

type AuditHandler struct {
	audit *services.AuditService
}

func NewAuditHandler(audit *services.AuditService) *AuditHandler {
	return &AuditHandler{audit: audit}
}

// List GET /audit-logs — lists the org's audit trail with filters and
// pagination. Owner/admin only (gated by the router).
func (h *AuditHandler) List(w http.ResponseWriter, r *http.Request) {
	orgID := middlewares.OrgID(r.Context())
	f := parseAuditFilter(r)

	items, meta, err := h.audit.List(orgID, f)
	if writeServiceError(w, err) {
		return
	}
	response.Paginated(w, items, meta)
}

// parseAuditFilter reads the supported query params into an AuditFilter. Dates
// accept RFC3339 or YYYY-MM-DD; an invalid user_id is ignored (no filter).
func parseAuditFilter(r *http.Request) dto.AuditFilter {
	q := r.URL.Query()
	f := dto.AuditFilter{
		Action:  q.Get("action"),
		Entity:  q.Get("entity"),
		Page:    atoiDefault(q.Get("page"), 1),
		PerPage: atoiDefault(q.Get("per_page"), 20),
	}
	if v := q.Get("user_id"); v != "" {
		if id, err := uuid.Parse(v); err == nil {
			f.UserID = &id
		}
	}
	if v := q.Get("from"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			f.From = &t
		} else if t, err := time.Parse("2006-01-02", v); err == nil {
			f.From = &t
		}
	}
	if v := q.Get("to"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			f.To = &t
		} else if t, err := time.Parse("2006-01-02", v); err == nil {
			// inclusive end of day
			t = t.Add(24*time.Hour - time.Second)
			f.To = &t
		}
	}
	return f
}
