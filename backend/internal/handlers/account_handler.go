package handlers

import (
	"net/http"

	"github.com/finance-sh/finance-sh/internal/dto"
	"github.com/finance-sh/finance-sh/internal/middlewares"
	"github.com/finance-sh/finance-sh/internal/services"
	"github.com/finance-sh/finance-sh/pkg/response"
	"github.com/finance-sh/finance-sh/pkg/validator"
)

type AccountHandler struct {
	accounts *services.AccountService
}

func NewAccountHandler(accounts *services.AccountService) *AccountHandler {
	return &AccountHandler{accounts: accounts}
}

// List GET /accounts
func (h *AccountHandler) List(w http.ResponseWriter, r *http.Request) {
	orgID := middlewares.OrgID(r.Context())
	items, err := h.accounts.List(orgID)
	if writeServiceError(w, err) {
		return
	}
	response.OK(w, items)
}

// Get GET /accounts/{id}
func (h *AccountHandler) Get(w http.ResponseWriter, r *http.Request) {
	orgID := middlewares.OrgID(r.Context())
	id, ok := urlUUID(w, r, "id")
	if !ok {
		return
	}
	item, err := h.accounts.Get(orgID, id)
	if writeServiceError(w, err) {
		return
	}
	response.OK(w, item)
}

// Create POST /accounts
func (h *AccountHandler) Create(w http.ResponseWriter, r *http.Request) {
	orgID := middlewares.OrgID(r.Context())
	var req dto.AccountRequest
	if fields, err := validator.BindJSON(r, &req); err != nil || len(fields) > 0 {
		response.ValidationError(w, fields)
		return
	}
	item, err := h.accounts.Create(orgID, req)
	if writeServiceError(w, err) {
		return
	}
	response.Created(w, item)
}

// Update PUT /accounts/{id}
func (h *AccountHandler) Update(w http.ResponseWriter, r *http.Request) {
	orgID := middlewares.OrgID(r.Context())
	id, ok := urlUUID(w, r, "id")
	if !ok {
		return
	}
	var req dto.AccountRequest
	if fields, err := validator.BindJSON(r, &req); err != nil || len(fields) > 0 {
		response.ValidationError(w, fields)
		return
	}
	item, err := h.accounts.Update(orgID, id, req)
	if writeServiceError(w, err) {
		return
	}
	response.OK(w, item)
}

// Delete DELETE /accounts/{id}
func (h *AccountHandler) Delete(w http.ResponseWriter, r *http.Request) {
	orgID := middlewares.OrgID(r.Context())
	id, ok := urlUUID(w, r, "id")
	if !ok {
		return
	}
	if writeServiceError(w, h.accounts.Delete(orgID, id)) {
		return
	}
	response.NoContent(w)
}
