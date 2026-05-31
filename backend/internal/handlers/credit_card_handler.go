package handlers

import (
	"net/http"
	"strconv"

	"github.com/finance-sh/finance-sh/internal/dto"
	"github.com/finance-sh/finance-sh/internal/middlewares"
	"github.com/finance-sh/finance-sh/internal/services"
	"github.com/finance-sh/finance-sh/pkg/response"
	"github.com/finance-sh/finance-sh/pkg/validator"
	"github.com/go-chi/chi/v5"
)

type CreditCardHandler struct {
	cards *services.CreditCardService
}

func NewCreditCardHandler(cards *services.CreditCardService) *CreditCardHandler {
	return &CreditCardHandler{cards: cards}
}

// List GET /credit-cards
func (h *CreditCardHandler) List(w http.ResponseWriter, r *http.Request) {
	orgID := middlewares.OrgID(r.Context())
	items, err := h.cards.List(orgID)
	if writeServiceError(w, err) {
		return
	}
	response.OK(w, items)
}

// Get GET /credit-cards/{id}
func (h *CreditCardHandler) Get(w http.ResponseWriter, r *http.Request) {
	orgID := middlewares.OrgID(r.Context())
	id, ok := urlUUID(w, r, "id")
	if !ok {
		return
	}
	item, err := h.cards.Get(orgID, id)
	if writeServiceError(w, err) {
		return
	}
	response.OK(w, item)
}

// Create POST /credit-cards
func (h *CreditCardHandler) Create(w http.ResponseWriter, r *http.Request) {
	orgID := middlewares.OrgID(r.Context())
	var req dto.CreditCardRequest
	if fields, err := validator.BindJSON(r, &req); err != nil || len(fields) > 0 {
		response.ValidationError(w, fields)
		return
	}
	item, err := h.cards.Create(orgID, req)
	if writeServiceError(w, err) {
		return
	}
	response.Created(w, item)
}

// Update PUT /credit-cards/{id}
func (h *CreditCardHandler) Update(w http.ResponseWriter, r *http.Request) {
	orgID := middlewares.OrgID(r.Context())
	id, ok := urlUUID(w, r, "id")
	if !ok {
		return
	}
	var req dto.CreditCardRequest
	if fields, err := validator.BindJSON(r, &req); err != nil || len(fields) > 0 {
		response.ValidationError(w, fields)
		return
	}
	item, err := h.cards.Update(orgID, id, req)
	if writeServiceError(w, err) {
		return
	}
	response.OK(w, item)
}

// Delete DELETE /credit-cards/{id}
func (h *CreditCardHandler) Delete(w http.ResponseWriter, r *http.Request) {
	orgID := middlewares.OrgID(r.Context())
	id, ok := urlUUID(w, r, "id")
	if !ok {
		return
	}
	if writeServiceError(w, h.cards.Delete(orgID, id)) {
		return
	}
	response.NoContent(w)
}

// Invoices GET /credit-cards/{id}/invoices?limit=12 — recent invoices (faturas).
func (h *CreditCardHandler) Invoices(w http.ResponseWriter, r *http.Request) {
	orgID := middlewares.OrgID(r.Context())
	id, ok := urlUUID(w, r, "id")
	if !ok {
		return
	}
	limit := 12
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limit = n
		}
	}
	items, err := h.cards.Invoices(orgID, id, limit)
	if writeServiceError(w, err) {
		return
	}
	response.OK(w, items)
}

// Invoice GET /credit-cards/{id}/invoices/{reference} — one invoice + its txns.
func (h *CreditCardHandler) Invoice(w http.ResponseWriter, r *http.Request) {
	orgID := middlewares.OrgID(r.Context())
	id, ok := urlUUID(w, r, "id")
	if !ok {
		return
	}
	reference := chi.URLParam(r, "reference")
	item, err := h.cards.Invoice(orgID, id, reference)
	if writeServiceError(w, err) {
		return
	}
	response.OK(w, item)
}

// PayInvoice POST /credit-cards/{id}/invoices/{reference}/pay — settle the whole
// invoice against the chosen paying account.
func (h *CreditCardHandler) PayInvoice(w http.ResponseWriter, r *http.Request) {
	orgID := middlewares.OrgID(r.Context())
	id, ok := urlUUID(w, r, "id")
	if !ok {
		return
	}
	reference := chi.URLParam(r, "reference")
	var req dto.InvoicePayRequest
	if fields, err := validator.BindJSON(r, &req); err != nil || len(fields) > 0 {
		response.ValidationError(w, fields)
		return
	}
	item, err := h.cards.PayInvoice(orgID, id, reference, req)
	if writeServiceError(w, err) {
		return
	}
	response.OK(w, item)
}

// UnpayInvoice POST /credit-cards/{id}/invoices/{reference}/unpay — reverse the
// settlement of every transaction in the invoice.
func (h *CreditCardHandler) UnpayInvoice(w http.ResponseWriter, r *http.Request) {
	orgID := middlewares.OrgID(r.Context())
	id, ok := urlUUID(w, r, "id")
	if !ok {
		return
	}
	reference := chi.URLParam(r, "reference")
	item, err := h.cards.UnpayInvoice(orgID, id, reference)
	if writeServiceError(w, err) {
		return
	}
	response.OK(w, item)
}
