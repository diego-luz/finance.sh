package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/finance-sh/finance-sh/internal/dto"
	"github.com/finance-sh/finance-sh/internal/services"
	"github.com/finance-sh/finance-sh/pkg/response"
	"github.com/finance-sh/finance-sh/pkg/validator"
)

// SetupHandler exposes the first-run wizard endpoints. Both routes are PUBLIC
// (no auth, no tenant header). The Status endpoint is hit at most a few times
// per browser session by the SPA's bootstrap gate, so it is uncached on
// purpose.
type SetupHandler struct {
	setup *services.SetupService
}

func NewSetupHandler(setup *services.SetupService) *SetupHandler {
	return &SetupHandler{setup: setup}
}

// Status GET /setup/status — reports whether the platform still needs first-run
// initialization. Always returns {"needs_setup": bool}; never 4xx for callers
// without auth (this is the public bootstrap probe).
func (h *SetupHandler) Status(w http.ResponseWriter, r *http.Request) {
	needs, err := h.setup.NeedsSetup(r.Context())
	if err != nil {
		slog.ErrorContext(r.Context(), "setup status: count failed", "error", err)
		response.Error(w, http.StatusInternalServerError, "internal_error", "Erro ao consultar status de inicialização")
		return
	}
	response.OK(w, dto.SetupStatusResponse{NeedsSetup: needs})
}

// Initialize POST /setup/initialize — bootstraps the platform with the first
// super-admin user and their first organization. Public, but guarded server-
// side by an in-transaction users-count == 0 check that races safely with any
// concurrent caller.
func (h *SetupHandler) Initialize(w http.ResponseWriter, r *http.Request) {
	var req dto.SetupInitializeRequest
	if fields, err := validator.BindJSON(r, &req); err != nil || len(fields) > 0 {
		// Setup wizard contract: 400 invalid_request with per-field details (the
		// SPA renders them next to the inputs). Distinct from the 422 envelope
		// used by /auth so the frontend can branch cleanly.
		writeInvalidRequest(w, fields)
		return
	}

	res, err := h.setup.Initialize(req, services.AuthMeta{UserAgent: r.UserAgent(), IP: r.RemoteAddr})
	if err != nil {
		switch {
		case errors.Is(err, services.ErrAlreadyInitialized):
			response.Error(w, http.StatusConflict, "already_initialized",
				"A plataforma já foi inicializada.")
		case errors.Is(err, services.ErrUnsupportedCurrency):
			response.Error(w, http.StatusBadRequest, "invalid_currency",
				"Moeda não suportada.")
		case errors.Is(err, services.ErrWeakPassword):
			response.Error(w, http.StatusBadRequest, "weak_password",
				"Senha muito fraca (mínimo 8 caracteres).")
		default:
			slog.ErrorContext(r.Context(), "setup initialize failed", "error", err)
			response.Error(w, http.StatusInternalServerError, "internal_error",
				"Erro ao inicializar a plataforma")
		}
		return
	}
	response.Created(w, res)
}

// writeInvalidRequest emits the setup-specific 400 invalid_request envelope.
// The error body carries the per-field validation details under "fields" so the
// SPA can place the messages next to the right inputs.
func writeInvalidRequest(w http.ResponseWriter, fields map[string]string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	_ = json.NewEncoder(w).Encode(response.Envelope{
		Success: false,
		Error: &response.ErrorBody{
			Code:    "invalid_request",
			Message: "Dados inválidos",
			Fields:  fields,
		},
	})
}
