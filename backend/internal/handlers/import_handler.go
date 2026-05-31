package handlers

import (
	"net/http"
	"strconv"

	"github.com/finance-sh/finance-sh/internal/dto"
	"github.com/finance-sh/finance-sh/internal/middlewares"
	"github.com/finance-sh/finance-sh/internal/services"
	"github.com/finance-sh/finance-sh/pkg/imports"
	"github.com/finance-sh/finance-sh/pkg/response"
	"github.com/finance-sh/finance-sh/pkg/validator"
	"github.com/google/uuid"
)

// ImportHandler exposes the statement-import flow (OFX/CSV): a stateless
// preview (parse + flag duplicates) followed by a commit (persist reviewed
// rows). Both are write operations (RequireWrite).
type ImportHandler struct {
	imports  *services.ImportService
	maxBytes int64
}

func NewImportHandler(svc *services.ImportService, maxMB int) *ImportHandler {
	if maxMB <= 0 {
		maxMB = 10
	}
	return &ImportHandler{imports: svc, maxBytes: int64(maxMB) * 1024 * 1024}
}

// Preview POST /imports/preview — multipart/form-data.
//
// Fields:
//   - file        (required) the statement (OFX or CSV)
//   - account_id  (required) target account UUID
//   - format      (optional) "ofx" | "csv" | "auto" (default "auto")
//   - delimiter   (optional, CSV) single char separator; auto-detect when absent
//   - has_header  (optional, CSV) "true"/"false"; default false
//   - date_col    (optional, CSV) 0-based index; -1 / absent = auto-detect
//   - desc_col    (optional, CSV) 0-based index; -1 / absent = auto-detect
//   - amount_col  (optional, CSV) 0-based index; -1 / absent = auto-detect
//   - date_format (optional, CSV) Go layout for the date column
//   - decimal_sep (optional, CSV) "," or "."
func (h *ImportHandler) Preview(w http.ResponseWriter, r *http.Request) {
	orgID := middlewares.OrgID(r.Context())

	r.Body = http.MaxBytesReader(w, r.Body, h.maxBytes+(1<<20))
	if err := r.ParseMultipartForm(h.maxBytes); err != nil {
		response.Error(w, http.StatusRequestEntityTooLarge, "file_too_large",
			"Arquivo excede o tamanho máximo permitido")
		return
	}

	accountID, err := uuid.Parse(r.FormValue("account_id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "bad_request", "account_id inválido")
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		response.Error(w, http.StatusBadRequest, "bad_request",
			"Arquivo ausente (campo multipart \"file\")")
		return
	}
	defer file.Close()

	format := r.FormValue("format")
	opts := parseCSVOptions(r)

	preview, err := h.imports.Preview(orgID, accountID, format, file, opts)
	if writeServiceError(w, err) {
		return
	}
	response.OK(w, preview)
}

// Commit POST /imports/commit — JSON.
//
// Body: { account_id, category_id?, rows: [{date, description, amount, type, external_id?}] }
// Returns: { created, skipped }.
func (h *ImportHandler) Commit(w http.ResponseWriter, r *http.Request) {
	orgID := middlewares.OrgID(r.Context())
	userID := middlewares.UserID(r.Context())

	var req dto.ImportCommitRequest
	if fields, err := validator.BindJSON(r, &req); err != nil || len(fields) > 0 {
		response.ValidationError(w, fields)
		return
	}

	accountID, err := uuid.Parse(req.AccountID)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "bad_request", "account_id inválido")
		return
	}

	var categoryID *uuid.UUID
	if req.CategoryID != "" {
		cid, err := uuid.Parse(req.CategoryID)
		if err != nil {
			response.Error(w, http.StatusBadRequest, "bad_request", "category_id inválido")
			return
		}
		categoryID = &cid
	}

	created, skipped, err := h.imports.Commit(orgID, userID, accountID, categoryID, req.Rows)
	if writeServiceError(w, err) {
		return
	}
	response.OK(w, dto.ImportCommitResult{Created: created, Skipped: skipped})
}

// parseCSVOptions reads the optional CSV tuning fields from the multipart form
// into imports.CSVOptions. Absent column fields default to -1 (auto-detect).
func parseCSVOptions(r *http.Request) imports.CSVOptions {
	opts := imports.CSVOptions{
		DateCol:    formInt(r, "date_col", -1),
		DescCol:    formInt(r, "desc_col", -1),
		AmountCol:  formInt(r, "amount_col", -1),
		HasHeader:  r.FormValue("has_header") == "true",
		DateFormat: r.FormValue("date_format"),
	}
	if d := r.FormValue("delimiter"); d != "" {
		opts.Delimiter = []rune(d)[0]
	}
	if d := r.FormValue("decimal_sep"); d != "" {
		opts.DecimalSep = []rune(d)[0]
	}
	return opts
}

// formInt reads a multipart/form integer field, returning def when absent or
// unparseable.
func formInt(r *http.Request, name string, def int) int {
	v := r.FormValue(name)
	if v == "" {
		return def
	}
	if n, err := strconv.Atoi(v); err == nil {
		return n
	}
	return def
}
