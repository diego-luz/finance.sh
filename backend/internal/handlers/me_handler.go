package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/finance-sh/finance-sh/internal/dto"
	"github.com/finance-sh/finance-sh/internal/middlewares"
	"github.com/finance-sh/finance-sh/internal/repositories"
	"github.com/finance-sh/finance-sh/internal/services"
	"github.com/finance-sh/finance-sh/pkg/response"
	"github.com/finance-sh/finance-sh/pkg/validator"
)

// MeHandler serves the authenticated user's self-service endpoints: 2FA
// management and the LGPD data export / account deletion.
type MeHandler struct {
	auth *services.AuthService
	lgpd *services.LGPDService
}

func NewMeHandler(auth *services.AuthService, lgpd *services.LGPDService) *MeHandler {
	return &MeHandler{auth: auth, lgpd: lgpd}
}

// SetupTwoFactor POST /me/2fa/setup
func (h *MeHandler) SetupTwoFactor(w http.ResponseWriter, r *http.Request) {
	userID := middlewares.UserID(r.Context())
	res, err := h.auth.SetupTwoFactor(userID)
	if err != nil {
		if errors.Is(err, services.ErrUserNotFound) {
			response.Error(w, http.StatusNotFound, "not_found", "Usuário não encontrado")
			return
		}
		response.Error(w, http.StatusInternalServerError, "internal_error", "Erro ao iniciar configuração do 2FA")
		return
	}
	response.OK(w, res)
}

// EnableTwoFactor POST /me/2fa/enable
func (h *MeHandler) EnableTwoFactor(w http.ResponseWriter, r *http.Request) {
	var req dto.TwoFactorCodeRequest
	if fields, err := validator.BindJSON(r, &req); err != nil || len(fields) > 0 {
		response.ValidationError(w, fields)
		return
	}
	userID := middlewares.UserID(r.Context())
	res, err := h.auth.EnableTwoFactor(userID, req.Code)
	if err != nil {
		switch {
		case errors.Is(err, services.Err2FANotPending):
			response.Error(w, http.StatusUnprocessableEntity, "2fa_not_pending", err.Error())
		case errors.Is(err, services.ErrInvalidCode):
			response.Error(w, http.StatusUnprocessableEntity, "invalid_code", "Código de verificação inválido")
		default:
			response.Error(w, http.StatusInternalServerError, "internal_error", "Erro ao ativar o 2FA")
		}
		return
	}
	response.OK(w, res)
}

// DisableTwoFactor POST /me/2fa/disable
func (h *MeHandler) DisableTwoFactor(w http.ResponseWriter, r *http.Request) {
	var req dto.TwoFactorCodeRequest
	if fields, err := validator.BindJSON(r, &req); err != nil || len(fields) > 0 {
		response.ValidationError(w, fields)
		return
	}
	userID := middlewares.UserID(r.Context())
	if err := h.auth.DisableTwoFactor(userID, req.Code); err != nil {
		if errors.Is(err, services.ErrInvalidCode) {
			response.Error(w, http.StatusUnprocessableEntity, "invalid_code", "Código de verificação inválido")
			return
		}
		response.Error(w, http.StatusInternalServerError, "internal_error", "Erro ao desativar o 2FA")
		return
	}
	response.OK(w, map[string]string{"message": "Autenticação de dois fatores desativada."})
}

// ExportData GET /me/export — returns a downloadable JSON document with all the
// user's data (LGPD portability). Bypasses the standard envelope to deliver a
// clean attachment.
func (h *MeHandler) ExportData(w http.ResponseWriter, r *http.Request) {
	userID := middlewares.UserID(r.Context())
	doc, err := h.lgpd.ExportData(userID)
	if err != nil {
		if errors.Is(err, services.ErrUserNotFound) {
			response.Error(w, http.StatusNotFound, "not_found", "Usuário não encontrado")
			return
		}
		response.Error(w, http.StatusInternalServerError, "internal_error", "Erro ao exportar os dados")
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="finance-sh-meus-dados.json"`)
	w.WriteHeader(http.StatusOK)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(doc)
}

// ImportData restores a previously-exported finance.sh JSON into a BRAND-NEW
// organization owned by the caller (data portability between instances). The
// body is the export JSON; capped at 25MB.
func (h *MeHandler) ImportData(w http.ResponseWriter, r *http.Request) {
	userID := middlewares.UserID(r.Context())
	raw, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 25<<20))
	if err != nil {
		response.Error(w, http.StatusRequestEntityTooLarge, "too_large", "Arquivo grande demais (máx 25MB)")
		return
	}
	sum, err := h.lgpd.ImportData(userID, raw)
	if err != nil {
		if errors.Is(err, services.ErrImportInvalid) {
			response.Error(w, http.StatusUnprocessableEntity, "import_invalid", "Arquivo de importação inválido")
			return
		}
		response.Error(w, http.StatusInternalServerError, "internal_error", "Erro ao importar os dados")
		return
	}
	response.Created(w, sum)
}

// Sessions GET /me/sessions — lists the caller's active sessions (non-revoked,
// non-expired refresh tokens). The token hash is never exposed.
func (h *MeHandler) Sessions(w http.ResponseWriter, r *http.Request) {
	userID := middlewares.UserID(r.Context())
	sessions, err := h.auth.Sessions(userID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "internal_error", "Erro ao listar sessões")
		return
	}
	response.OK(w, sessions)
}

// RevokeSession DELETE /me/sessions/{id} — revokes one of the caller's sessions.
func (h *MeHandler) RevokeSession(w http.ResponseWriter, r *http.Request) {
	userID := middlewares.UserID(r.Context())
	id, ok := urlUUID(w, r, "id")
	if !ok {
		return
	}
	if err := h.auth.RevokeSession(userID, id); err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			response.Error(w, http.StatusNotFound, "not_found", "Sessão não encontrada")
			return
		}
		response.Error(w, http.StatusInternalServerError, "internal_error", "Erro ao revogar a sessão")
		return
	}
	response.OK(w, map[string]string{"message": "Sessão revogada."})
}

// RevokeOtherSessions POST /me/sessions/revoke-others — revokes every active
// session of the caller. The body may carry the caller's current raw refresh
// token in `keep_refresh_token` so that session is preserved; when absent ALL
// sessions are revoked and the client must re-login.
func (h *MeHandler) RevokeOtherSessions(w http.ResponseWriter, r *http.Request) {
	var req dto.RevokeOtherSessionsRequest
	// Body is optional; ignore decode errors so an empty body revokes everything.
	if r.ContentLength != 0 {
		if fields, err := validator.BindJSON(r, &req); err != nil || len(fields) > 0 {
			response.ValidationError(w, fields)
			return
		}
	}
	userID := middlewares.UserID(r.Context())
	n, err := h.auth.RevokeOtherSessions(userID, req.KeepRefreshToken)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "internal_error", "Erro ao revogar as sessões")
		return
	}
	response.OK(w, dto.BulkResult{Updated: n})
}

// ChangePassword POST /me/change-password — rotates the authenticated user's
// password. Requires the current password (re-auth guard). On success clears
// must_change_password and revokes every other refresh token; pass the caller's
// raw refresh token in keep_refresh_token to keep the current session alive
// (when absent ALL sessions are revoked and the client must re-login).
func (h *MeHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	var req dto.ChangePasswordRequest
	if fields, err := validator.BindJSON(r, &req); err != nil || len(fields) > 0 {
		response.ValidationError(w, fields)
		return
	}
	userID := middlewares.UserID(r.Context())
	if err := h.auth.ChangePassword(userID, req.CurrentPassword, req.NewPassword, req.KeepRefreshToken); err != nil {
		switch {
		case errors.Is(err, services.ErrWrongPassword):
			response.Error(w, http.StatusUnauthorized, "wrong_password", "Senha atual incorreta")
		case errors.Is(err, services.ErrUserNotFound):
			response.Error(w, http.StatusNotFound, "not_found", "Usuário não encontrado")
		default:
			response.Error(w, http.StatusInternalServerError, "internal_error", "Erro ao alterar a senha")
		}
		return
	}
	response.OK(w, map[string]bool{"success": true})
}

// DeleteAccount DELETE /me/account — LGPD right to erasure.
func (h *MeHandler) DeleteAccount(w http.ResponseWriter, r *http.Request) {
	var req dto.DeleteAccountRequest
	if fields, err := validator.BindJSON(r, &req); err != nil || len(fields) > 0 {
		response.ValidationError(w, fields)
		return
	}
	userID := middlewares.UserID(r.Context())
	if err := h.lgpd.DeleteAccount(userID, req.Password); err != nil {
		switch {
		case errors.Is(err, services.ErrWrongPassword):
			response.Error(w, http.StatusUnauthorized, "wrong_password", "Senha incorreta")
		case errors.Is(err, services.ErrOwnedOrgHasMembers):
			response.Error(w, http.StatusConflict, "owned_org_has_members", err.Error())
		case errors.Is(err, services.ErrUserNotFound):
			response.Error(w, http.StatusNotFound, "not_found", "Usuário não encontrado")
		default:
			response.Error(w, http.StatusInternalServerError, "internal_error", "Erro ao remover a conta")
		}
		return
	}
	response.OK(w, map[string]string{
		"message": "Sua conta e seus dados foram removidos/anonimizados conforme a LGPD.",
	})
}
