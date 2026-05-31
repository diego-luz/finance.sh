package handlers

import (
	"net/http"

	"github.com/finance-sh/finance-sh/internal/dto"
	"github.com/finance-sh/finance-sh/internal/middlewares"
	"github.com/finance-sh/finance-sh/internal/services"
	"github.com/finance-sh/finance-sh/pkg/response"
	"github.com/finance-sh/finance-sh/pkg/validator"
)

// OrganizationHandler serves the active-organization config endpoints
// (read/update) and the supported-currency list.
type OrganizationHandler struct {
	orgs *services.OrganizationService
}

func NewOrganizationHandler(orgs *services.OrganizationService) *OrganizationHandler {
	return &OrganizationHandler{orgs: orgs}
}

// Get GET /organization — the active organization (any member).
func (h *OrganizationHandler) Get(w http.ResponseWriter, r *http.Request) {
	orgID := middlewares.OrgID(r.Context())
	role := middlewares.UserRole(r.Context())
	item, err := h.orgs.Get(orgID, role)
	if writeServiceError(w, err) {
		return
	}
	response.OK(w, item)
}

// Update PUT /organization — updates name/currency (owner/admin only, enforced
// via RequireRole on the route).
func (h *OrganizationHandler) Update(w http.ResponseWriter, r *http.Request) {
	var req dto.OrgUpdateRequest
	if fields, err := validator.BindJSON(r, &req); err != nil || len(fields) > 0 {
		response.ValidationError(w, fields)
		return
	}
	orgID := middlewares.OrgID(r.Context())
	role := middlewares.UserRole(r.Context())
	item, err := h.orgs.Update(orgID, role, req)
	if writeServiceError(w, err) {
		return
	}
	response.OK(w, item)
}

// Create POST /organizations — self-service: create an ADDITIONAL organization
// owned by the caller (tenant-agnostic; no active org required).
func (h *OrganizationHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.OrgCreateRequest
	if fields, err := validator.BindJSON(r, &req); err != nil || len(fields) > 0 {
		response.ValidationError(w, fields)
		return
	}
	userID := middlewares.UserID(r.Context())
	item, err := h.orgs.Create(userID, req)
	if writeServiceError(w, err) {
		return
	}
	response.Created(w, item)
}

// Currencies GET /currencies — the supported-currency table (any member).
func (h *OrganizationHandler) Currencies(w http.ResponseWriter, r *http.Request) {
	response.OK(w, h.orgs.Currencies())
}
