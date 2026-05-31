package handlers

import (
	"net/http"

	"github.com/finance-sh/finance-sh/internal/dto"
	"github.com/finance-sh/finance-sh/internal/middlewares"
	"github.com/finance-sh/finance-sh/internal/services"
	"github.com/finance-sh/finance-sh/pkg/response"
	"github.com/finance-sh/finance-sh/pkg/validator"
)

// RecurrenceHandler exposes the recurring-transaction rules CRUD plus the manual
// "run now" action. Reads (List) are available to any member; mutations and the
// run action are gated by RequireWrite in the router.
type RecurrenceHandler struct {
	svc *services.RecurrenceService
}

func NewRecurrenceHandler(svc *services.RecurrenceService) *RecurrenceHandler {
	return &RecurrenceHandler{svc: svc}
}

// List GET /recurrences
func (h *RecurrenceHandler) List(w http.ResponseWriter, r *http.Request) {
	orgID := middlewares.OrgID(r.Context())
	items, err := h.svc.List(orgID)
	if writeServiceError(w, err) {
		return
	}
	response.OK(w, items)
}

// Create POST /recurrences
func (h *RecurrenceHandler) Create(w http.ResponseWriter, r *http.Request) {
	orgID := middlewares.OrgID(r.Context())
	var req dto.RecurrenceRuleRequest
	if fields, err := validator.BindJSON(r, &req); err != nil || len(fields) > 0 {
		response.ValidationError(w, fields)
		return
	}
	item, err := h.svc.Create(orgID, req)
	if writeServiceError(w, err) {
		return
	}
	response.Created(w, item)
}

// Update PUT /recurrences/{id}
func (h *RecurrenceHandler) Update(w http.ResponseWriter, r *http.Request) {
	orgID := middlewares.OrgID(r.Context())
	id, ok := urlUUID(w, r, "id")
	if !ok {
		return
	}
	var req dto.RecurrenceRuleRequest
	if fields, err := validator.BindJSON(r, &req); err != nil || len(fields) > 0 {
		response.ValidationError(w, fields)
		return
	}
	item, err := h.svc.Update(orgID, id, req)
	if writeServiceError(w, err) {
		return
	}
	response.OK(w, item)
}

// Delete DELETE /recurrences/{id}
func (h *RecurrenceHandler) Delete(w http.ResponseWriter, r *http.Request) {
	orgID := middlewares.OrgID(r.Context())
	id, ok := urlUUID(w, r, "id")
	if !ok {
		return
	}
	if writeServiceError(w, h.svc.Delete(orgID, id)) {
		return
	}
	response.NoContent(w)
}

// Run POST /recurrences/{id}/run — generates the due occurrences for one rule now.
func (h *RecurrenceHandler) Run(w http.ResponseWriter, r *http.Request) {
	orgID := middlewares.OrgID(r.Context())
	id, ok := urlUUID(w, r, "id")
	if !ok {
		return
	}
	created, err := h.svc.RunNow(orgID, id)
	if writeServiceError(w, err) {
		return
	}
	response.OK(w, dto.RecurrenceRunResult{Created: created})
}
