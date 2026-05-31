package handlers

import (
	"net/http"

	"github.com/finance-sh/finance-sh/internal/dto"
	"github.com/finance-sh/finance-sh/internal/middlewares"
	"github.com/finance-sh/finance-sh/internal/services"
	"github.com/finance-sh/finance-sh/pkg/response"
	"github.com/finance-sh/finance-sh/pkg/validator"
)

type BudgetHandler struct {
	budgets *services.BudgetService
}

func NewBudgetHandler(budgets *services.BudgetService) *BudgetHandler {
	return &BudgetHandler{budgets: budgets}
}

// List GET /budgets?month=&year=
func (h *BudgetHandler) List(w http.ResponseWriter, r *http.Request) {
	orgID := middlewares.OrgID(r.Context())
	q := r.URL.Query()
	month := atoiDefault(q.Get("month"), 0)
	year := atoiDefault(q.Get("year"), 0)
	items, err := h.budgets.List(orgID, month, year)
	if writeServiceError(w, err) {
		return
	}
	response.OK(w, items)
}

// Create POST /budgets
func (h *BudgetHandler) Create(w http.ResponseWriter, r *http.Request) {
	orgID := middlewares.OrgID(r.Context())
	var req dto.BudgetRequest
	if fields, err := validator.BindJSON(r, &req); err != nil || len(fields) > 0 {
		response.ValidationError(w, fields)
		return
	}
	item, err := h.budgets.Create(orgID, req)
	if writeServiceError(w, err) {
		return
	}
	response.Created(w, item)
}

// Update PUT /budgets/{id}
func (h *BudgetHandler) Update(w http.ResponseWriter, r *http.Request) {
	orgID := middlewares.OrgID(r.Context())
	id, ok := urlUUID(w, r, "id")
	if !ok {
		return
	}
	var req dto.BudgetRequest
	if fields, err := validator.BindJSON(r, &req); err != nil || len(fields) > 0 {
		response.ValidationError(w, fields)
		return
	}
	item, err := h.budgets.Update(orgID, id, req)
	if writeServiceError(w, err) {
		return
	}
	response.OK(w, item)
}

// Delete DELETE /budgets/{id}
func (h *BudgetHandler) Delete(w http.ResponseWriter, r *http.Request) {
	orgID := middlewares.OrgID(r.Context())
	id, ok := urlUUID(w, r, "id")
	if !ok {
		return
	}
	if writeServiceError(w, h.budgets.Delete(orgID, id)) {
		return
	}
	response.NoContent(w)
}
