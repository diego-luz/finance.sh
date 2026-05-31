package handlers

import (
	"net/http"

	"github.com/finance-sh/finance-sh/internal/dto"
	"github.com/finance-sh/finance-sh/internal/middlewares"
	"github.com/finance-sh/finance-sh/internal/services"
	"github.com/finance-sh/finance-sh/pkg/response"
	"github.com/finance-sh/finance-sh/pkg/validator"
)

type ContactHandler struct {
	contacts *services.ContactService
}

func NewContactHandler(contacts *services.ContactService) *ContactHandler {
	return &ContactHandler{contacts: contacts}
}

// List GET /contacts
func (h *ContactHandler) List(w http.ResponseWriter, r *http.Request) {
	orgID := middlewares.OrgID(r.Context())
	items, err := h.contacts.List(orgID)
	if writeServiceError(w, err) {
		return
	}
	response.OK(w, items)
}

// Get GET /contacts/{id}
func (h *ContactHandler) Get(w http.ResponseWriter, r *http.Request) {
	orgID := middlewares.OrgID(r.Context())
	id, ok := urlUUID(w, r, "id")
	if !ok {
		return
	}
	item, err := h.contacts.Get(orgID, id)
	if writeServiceError(w, err) {
		return
	}
	response.OK(w, item)
}

// Create POST /contacts
func (h *ContactHandler) Create(w http.ResponseWriter, r *http.Request) {
	orgID := middlewares.OrgID(r.Context())
	var req dto.ContactRequest
	if fields, err := validator.BindJSON(r, &req); err != nil || len(fields) > 0 {
		response.ValidationError(w, fields)
		return
	}
	item, err := h.contacts.Create(orgID, req)
	if writeServiceError(w, err) {
		return
	}
	response.Created(w, item)
}

// Update PUT /contacts/{id}
func (h *ContactHandler) Update(w http.ResponseWriter, r *http.Request) {
	orgID := middlewares.OrgID(r.Context())
	id, ok := urlUUID(w, r, "id")
	if !ok {
		return
	}
	var req dto.ContactRequest
	if fields, err := validator.BindJSON(r, &req); err != nil || len(fields) > 0 {
		response.ValidationError(w, fields)
		return
	}
	item, err := h.contacts.Update(orgID, id, req)
	if writeServiceError(w, err) {
		return
	}
	response.OK(w, item)
}

// Delete DELETE /contacts/{id}
func (h *ContactHandler) Delete(w http.ResponseWriter, r *http.Request) {
	orgID := middlewares.OrgID(r.Context())
	id, ok := urlUUID(w, r, "id")
	if !ok {
		return
	}
	if writeServiceError(w, h.contacts.Delete(orgID, id)) {
		return
	}
	response.NoContent(w)
}
