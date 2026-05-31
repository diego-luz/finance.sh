package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/finance-sh/finance-sh/internal/dto"
	"github.com/finance-sh/finance-sh/internal/middlewares"
	"github.com/finance-sh/finance-sh/internal/services"
	"github.com/finance-sh/finance-sh/pkg/response"
	"github.com/finance-sh/finance-sh/pkg/validator"
)

type TransactionHandler struct {
	txs *services.TransactionService
}

func NewTransactionHandler(txs *services.TransactionService) *TransactionHandler {
	return &TransactionHandler{txs: txs}
}

// List GET /transactions — supports filters via query string.
func (h *TransactionHandler) List(w http.ResponseWriter, r *http.Request) {
	orgID := middlewares.OrgID(r.Context())
	f := parseFilter(r)

	items, meta, err := h.txs.List(orgID, f)
	if writeServiceError(w, err) {
		return
	}
	response.Paginated(w, items, meta)
}

// Get GET /transactions/{id}
func (h *TransactionHandler) Get(w http.ResponseWriter, r *http.Request) {
	orgID := middlewares.OrgID(r.Context())
	id, ok := urlUUID(w, r, "id")
	if !ok {
		return
	}
	item, err := h.txs.Get(orgID, id)
	if writeServiceError(w, err) {
		return
	}
	response.OK(w, item)
}

// Create POST /transactions
func (h *TransactionHandler) Create(w http.ResponseWriter, r *http.Request) {
	orgID := middlewares.OrgID(r.Context())
	var req dto.TransactionRequest
	if fields, err := validator.BindJSON(r, &req); err != nil || len(fields) > 0 {
		response.ValidationError(w, fields)
		return
	}
	item, err := h.txs.Create(orgID, req)
	if writeServiceError(w, err) {
		return
	}
	response.Created(w, item)
}

// Update PUT /transactions/{id}
func (h *TransactionHandler) Update(w http.ResponseWriter, r *http.Request) {
	orgID := middlewares.OrgID(r.Context())
	id, ok := urlUUID(w, r, "id")
	if !ok {
		return
	}
	var req dto.TransactionRequest
	if fields, err := validator.BindJSON(r, &req); err != nil || len(fields) > 0 {
		response.ValidationError(w, fields)
		return
	}
	item, err := h.txs.Update(orgID, id, req)
	if writeServiceError(w, err) {
		return
	}
	response.OK(w, item)
}

// Delete DELETE /transactions/{id}?scope=one|all
//
// scope=all removes the whole installment group when the transaction belongs to
// one; scope=one (default) removes only this transaction.
func (h *TransactionHandler) Delete(w http.ResponseWriter, r *http.Request) {
	orgID := middlewares.OrgID(r.Context())
	id, ok := urlUUID(w, r, "id")
	if !ok {
		return
	}
	scope := r.URL.Query().Get("scope")
	if scope != "all" {
		scope = "one"
	}
	if writeServiceError(w, h.txs.DeleteScoped(orgID, id, scope)) {
		return
	}
	response.NoContent(w)
}

// Settle POST /transactions/{id}/settle
func (h *TransactionHandler) Settle(w http.ResponseWriter, r *http.Request) {
	orgID := middlewares.OrgID(r.Context())
	id, ok := urlUUID(w, r, "id")
	if !ok {
		return
	}
	// Body is optional: an empty body settles with defaults (paid now, same account).
	var req dto.SettleRequest
	if r.ContentLength != 0 {
		if fields, err := validator.BindJSON(r, &req); err != nil || len(fields) > 0 {
			response.ValidationError(w, fields)
			return
		}
	}
	item, err := h.txs.Settle(orgID, id, req)
	if writeServiceError(w, err) {
		return
	}
	response.OK(w, item)
}

// Unsettle POST /transactions/{id}/unsettle
func (h *TransactionHandler) Unsettle(w http.ResponseWriter, r *http.Request) {
	orgID := middlewares.OrgID(r.Context())
	id, ok := urlUUID(w, r, "id")
	if !ok {
		return
	}
	item, err := h.txs.Unsettle(orgID, id)
	if writeServiceError(w, err) {
		return
	}
	response.OK(w, item)
}

// BulkSettle POST /transactions/bulk-settle
func (h *TransactionHandler) BulkSettle(w http.ResponseWriter, r *http.Request) {
	orgID := middlewares.OrgID(r.Context())
	var req dto.BulkSettleRequest
	if fields, err := validator.BindJSON(r, &req); err != nil || len(fields) > 0 {
		response.ValidationError(w, fields)
		return
	}
	n, err := h.txs.BulkSettle(orgID, req)
	if writeServiceError(w, err) {
		return
	}
	response.OK(w, dto.BulkResult{Updated: n})
}

// BulkUnsettle POST /transactions/bulk-unsettle
func (h *TransactionHandler) BulkUnsettle(w http.ResponseWriter, r *http.Request) {
	orgID := middlewares.OrgID(r.Context())
	var req dto.BulkIDsRequest
	if fields, err := validator.BindJSON(r, &req); err != nil || len(fields) > 0 {
		response.ValidationError(w, fields)
		return
	}
	n, err := h.txs.BulkUnsettle(orgID, req)
	if writeServiceError(w, err) {
		return
	}
	response.OK(w, dto.BulkResult{Updated: n})
}

// BulkCategorize POST /transactions/bulk-categorize
func (h *TransactionHandler) BulkCategorize(w http.ResponseWriter, r *http.Request) {
	orgID := middlewares.OrgID(r.Context())
	var req dto.BulkCategorizeRequest
	if fields, err := validator.BindJSON(r, &req); err != nil || len(fields) > 0 {
		response.ValidationError(w, fields)
		return
	}
	n, err := h.txs.BulkCategorize(orgID, req)
	if writeServiceError(w, err) {
		return
	}
	response.OK(w, dto.BulkResult{Updated: n})
}

// BulkDelete POST /transactions/bulk-delete — soft-deletes several transactions.
func (h *TransactionHandler) BulkDelete(w http.ResponseWriter, r *http.Request) {
	orgID := middlewares.OrgID(r.Context())
	var req dto.BulkIDsRequest
	if fields, err := validator.BindJSON(r, &req); err != nil || len(fields) > 0 {
		response.ValidationError(w, fields)
		return
	}
	n, err := h.txs.BulkDelete(orgID, req)
	if writeServiceError(w, err) {
		return
	}
	response.OK(w, dto.BulkResult{Updated: n})
}

// Payables GET /payables — expenses filtered by paid status and due window.
func (h *TransactionHandler) Payables(w http.ResponseWriter, r *http.Request) {
	orgID := middlewares.OrgID(r.Context())
	f := parsePayableFilter(r)
	items, meta, summary, err := h.txs.Payables(orgID, f)
	if writeServiceError(w, err) {
		return
	}
	response.Paginated(w, items, dto.PayablesMeta{
		Page: meta.Page, PerPage: meta.PerPage, Total: meta.Total, Pages: meta.Pages, Summary: summary,
	})
}

// Receivables GET /receivables — income filtered by paid status and due window.
func (h *TransactionHandler) Receivables(w http.ResponseWriter, r *http.Request) {
	orgID := middlewares.OrgID(r.Context())
	f := parsePayableFilter(r)
	items, meta, summary, err := h.txs.Receivables(orgID, f)
	if writeServiceError(w, err) {
		return
	}
	response.Paginated(w, items, dto.PayablesMeta{
		Page: meta.Page, PerPage: meta.PerPage, Total: meta.Total, Pages: meta.Pages, Summary: summary,
	})
}

// parsePayableFilter reads the AP/AR query params into a PayableFilter.
func parsePayableFilter(r *http.Request) dto.PayableFilter {
	q := r.URL.Query()
	f := dto.PayableFilter{
		Status:    q.Get("status"),
		ContactID: q.Get("contact_id"),
		Page:      atoiDefault(q.Get("page"), 1),
		PerPage:   atoiDefault(q.Get("per_page"), 20),
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
			t = t.Add(24*time.Hour - time.Second)
			f.To = &t
		}
	}
	return f
}

// parseFilter reads the supported query params into a TransactionFilter.
func parseFilter(r *http.Request) dto.TransactionFilter {
	q := r.URL.Query()
	f := dto.TransactionFilter{
		Type:       q.Get("type"),
		AccountID:  q.Get("account_id"),
		CategoryID: q.Get("category_id"),
		ContactID:  q.Get("contact_id"),
		TagID:      q.Get("tag_id"),
		Search:     q.Get("search"),
		Page:       atoiDefault(q.Get("page"), 1),
		PerPage:    atoiDefault(q.Get("per_page"), 20),
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

func atoiDefault(s string, def int) int {
	if n, err := strconv.Atoi(s); err == nil {
		return n
	}
	return def
}
