package handlers

import (
	"net/http"

	"github.com/finance-sh/finance-sh/internal/dto"
	"github.com/finance-sh/finance-sh/internal/middlewares"
	"github.com/finance-sh/finance-sh/internal/services"
	"github.com/finance-sh/finance-sh/pkg/response"
	"github.com/finance-sh/finance-sh/pkg/validator"
)

type GoalHandler struct {
	goals *services.GoalService
}

func NewGoalHandler(goals *services.GoalService) *GoalHandler {
	return &GoalHandler{goals: goals}
}

// List GET /goals
func (h *GoalHandler) List(w http.ResponseWriter, r *http.Request) {
	orgID := middlewares.OrgID(r.Context())
	items, err := h.goals.List(orgID)
	if writeServiceError(w, err) {
		return
	}
	response.OK(w, items)
}

// Get GET /goals/{id}
func (h *GoalHandler) Get(w http.ResponseWriter, r *http.Request) {
	orgID := middlewares.OrgID(r.Context())
	id, ok := urlUUID(w, r, "id")
	if !ok {
		return
	}
	item, err := h.goals.Get(orgID, id)
	if writeServiceError(w, err) {
		return
	}
	response.OK(w, item)
}

// Create POST /goals
func (h *GoalHandler) Create(w http.ResponseWriter, r *http.Request) {
	orgID := middlewares.OrgID(r.Context())
	var req dto.GoalRequest
	if fields, err := validator.BindJSON(r, &req); err != nil || len(fields) > 0 {
		response.ValidationError(w, fields)
		return
	}
	item, err := h.goals.Create(orgID, req)
	if writeServiceError(w, err) {
		return
	}
	response.Created(w, item)
}

// Update PUT /goals/{id}
func (h *GoalHandler) Update(w http.ResponseWriter, r *http.Request) {
	orgID := middlewares.OrgID(r.Context())
	id, ok := urlUUID(w, r, "id")
	if !ok {
		return
	}
	var req dto.GoalRequest
	if fields, err := validator.BindJSON(r, &req); err != nil || len(fields) > 0 {
		response.ValidationError(w, fields)
		return
	}
	item, err := h.goals.Update(orgID, id, req)
	if writeServiceError(w, err) {
		return
	}
	response.OK(w, item)
}

// Delete DELETE /goals/{id}
func (h *GoalHandler) Delete(w http.ResponseWriter, r *http.Request) {
	orgID := middlewares.OrgID(r.Context())
	id, ok := urlUUID(w, r, "id")
	if !ok {
		return
	}
	if writeServiceError(w, h.goals.Delete(orgID, id)) {
		return
	}
	response.NoContent(w)
}
