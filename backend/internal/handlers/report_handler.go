package handlers

import (
	"net/http"
	"time"

	"github.com/finance-sh/finance-sh/internal/middlewares"
	"github.com/finance-sh/finance-sh/internal/services"
	"github.com/finance-sh/finance-sh/pkg/response"
)

type ReportHandler struct {
	reports *services.ReportService
}

func NewReportHandler(reports *services.ReportService) *ReportHandler {
	return &ReportHandler{reports: reports}
}

// TransactionsCSV GET /reports/transactions.csv?from=&to=&type=
func (h *ReportHandler) TransactionsCSV(w http.ResponseWriter, r *http.Request) {
	orgID := middlewares.OrgID(r.Context())
	f := parseFilter(r)
	// Export ignores pagination; clear the page/per-page so the repo returns all.
	f.Page, f.PerPage = 0, 0

	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="transacoes.csv"`)

	if err := h.reports.TransactionsCSV(orgID, f, w); err != nil {
		// Headers may already be flushed; still attempt a clean error envelope.
		response.Error(w, http.StatusInternalServerError, "internal_error", "Erro ao gerar relatório")
		return
	}
}

// ExcelExport GET /export/data.xlsx — multi-sheet workbook with the org's data.
func (h *ReportHandler) ExcelExport(w http.ResponseWriter, r *http.Request) {
	orgID := middlewares.OrgID(r.Context())
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", `attachment; filename="finance-sh-dados.xlsx"`)
	if err := h.reports.ExcelExport(orgID, w); err != nil {
		response.Error(w, http.StatusInternalServerError, "internal_error", "Erro ao gerar planilha")
		return
	}
}

// StatementPDF GET /reports/statement.pdf?from=&to= — bank-statement-style PDF.
func (h *ReportHandler) StatementPDF(w http.ResponseWriter, r *http.Request) {
	orgID := middlewares.OrgID(r.Context())
	from, to := parseDateRange(r)
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", `attachment; filename="extrato.pdf"`)
	if err := h.reports.StatementPDF(orgID, from, to, w); err != nil {
		response.Error(w, http.StatusInternalServerError, "internal_error", "Erro ao gerar extrato")
		return
	}
}

// MonthlyPDF GET /reports/monthly.pdf?month=&year= — monthly summary PDF.
func (h *ReportHandler) MonthlyPDF(w http.ResponseWriter, r *http.Request) {
	orgID := middlewares.OrgID(r.Context())
	q := r.URL.Query()
	month := atoiDefault(q.Get("month"), 0)
	year := atoiDefault(q.Get("year"), 0)
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", `attachment; filename="resumo-mensal.pdf"`)
	if err := h.reports.MonthlyPDF(orgID, month, year, w); err != nil {
		response.Error(w, http.StatusInternalServerError, "internal_error", "Erro ao gerar resumo mensal")
		return
	}
}

// Summary GET /reports/summary?from=&to= — aggregated data for the Relatórios page.
func (h *ReportHandler) Summary(w http.ResponseWriter, r *http.Request) {
	orgID := middlewares.OrgID(r.Context())
	from, to := parseDateRange(r)
	res, err := h.reports.Summary(orgID, from, to)
	if writeServiceError(w, err) {
		return
	}
	response.OK(w, res)
}

// parseDateRange reads the from/to query params (RFC3339 or YYYY-MM-DD). The
// `to` bound is pushed to end-of-day so a single-day inclusive value works.
// Either or both may be nil (the service applies the current-month default).
func parseDateRange(r *http.Request) (from, to *time.Time) {
	q := r.URL.Query()
	if v := q.Get("from"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			from = &t
		} else if t, err := time.Parse("2006-01-02", v); err == nil {
			from = &t
		}
	}
	if v := q.Get("to"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			to = &t
		} else if t, err := time.Parse("2006-01-02", v); err == nil {
			t = t.Add(24*time.Hour - time.Second)
			to = &t
		}
	}
	return from, to
}
