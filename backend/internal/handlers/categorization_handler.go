package handlers

import (
	"net/http"

	"github.com/finance-sh/finance-sh/internal/dto"
	"github.com/finance-sh/finance-sh/internal/middlewares"
	"github.com/finance-sh/finance-sh/internal/services"
	"github.com/finance-sh/finance-sh/pkg/response"
	"github.com/finance-sh/finance-sh/pkg/validator"
)

// CategorizationHandler exposes the automatic-categorization rules CRUD plus the
// apply/suggest helpers. Mutations are gated by RequireWrite in the router; the
// list and suggest reads are available to any member.
type CategorizationHandler struct {
	svc *services.CategorizationService
}

func NewCategorizationHandler(svc *services.CategorizationService) *CategorizationHandler {
	return &CategorizationHandler{svc: svc}
}

// ListRules GET /categorization/rules
func (h *CategorizationHandler) ListRules(w http.ResponseWriter, r *http.Request) {
	orgID := middlewares.OrgID(r.Context())
	items, err := h.svc.ListRules(orgID)
	if writeServiceError(w, err) {
		return
	}
	response.OK(w, items)
}

// CreateRule POST /categorization/rules
func (h *CategorizationHandler) CreateRule(w http.ResponseWriter, r *http.Request) {
	orgID := middlewares.OrgID(r.Context())
	var req dto.CategoryRuleRequest
	if fields, err := validator.BindJSON(r, &req); err != nil || len(fields) > 0 {
		response.ValidationError(w, fields)
		return
	}
	item, err := h.svc.CreateRule(orgID, req)
	if writeServiceError(w, err) {
		return
	}
	response.Created(w, item)
}

// UpdateRule PUT /categorization/rules/{id}
func (h *CategorizationHandler) UpdateRule(w http.ResponseWriter, r *http.Request) {
	orgID := middlewares.OrgID(r.Context())
	id, ok := urlUUID(w, r, "id")
	if !ok {
		return
	}
	var req dto.CategoryRuleRequest
	if fields, err := validator.BindJSON(r, &req); err != nil || len(fields) > 0 {
		response.ValidationError(w, fields)
		return
	}
	item, err := h.svc.UpdateRule(orgID, id, req)
	if writeServiceError(w, err) {
		return
	}
	response.OK(w, item)
}

// DeleteRule DELETE /categorization/rules/{id}
func (h *CategorizationHandler) DeleteRule(w http.ResponseWriter, r *http.Request) {
	orgID := middlewares.OrgID(r.Context())
	id, ok := urlUUID(w, r, "id")
	if !ok {
		return
	}
	if writeServiceError(w, h.svc.DeleteRule(orgID, id)) {
		return
	}
	response.NoContent(w)
}

// Apply POST /categorization/apply
func (h *CategorizationHandler) Apply(w http.ResponseWriter, r *http.Request) {
	orgID := middlewares.OrgID(r.Context())
	updated, err := h.svc.ApplyToUncategorized(orgID)
	if writeServiceError(w, err) {
		return
	}
	response.OK(w, dto.ApplyResult{Updated: updated})
}

// Suggest GET /categorization/suggest?description=&type=expense|income
func (h *CategorizationHandler) Suggest(w http.ResponseWriter, r *http.Request) {
	orgID := middlewares.OrgID(r.Context())
	description := r.URL.Query().Get("description")
	txType := r.URL.Query().Get("type")
	if txType != "income" && txType != "expense" {
		response.Error(w, http.StatusBadRequest, "bad_request", "type deve ser income ou expense")
		return
	}
	if description == "" {
		response.Error(w, http.StatusBadRequest, "bad_request", "description é obrigatório")
		return
	}
	resp, err := h.svc.Suggest(orgID, description, txType)
	if writeServiceError(w, err) {
		return
	}
	response.OK(w, resp)
}
