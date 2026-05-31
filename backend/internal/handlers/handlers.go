// Package handlers contains the thin HTTP layer: it decodes/validates requests,
// delegates to services and writes the standard JSON envelope. No business logic
// lives here.
package handlers

import (
	"errors"
	"net/http"

	"github.com/finance-sh/finance-sh/internal/repositories"
	"github.com/finance-sh/finance-sh/internal/services"
	"github.com/finance-sh/finance-sh/pkg/response"
	"github.com/finance-sh/finance-sh/pkg/storage"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// urlUUID parses a path parameter as UUID, writing a 400 and returning false on
// failure.
func urlUUID(w http.ResponseWriter, r *http.Request, name string) (uuid.UUID, bool) {
	id, err := uuid.Parse(chi.URLParam(r, name))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "bad_request", "Identificador inválido")
		return uuid.Nil, false
	}
	return id, true
}

// writeServiceError maps common service/repository errors to HTTP responses.
// Returns true when it handled the error.
func writeServiceError(w http.ResponseWriter, err error) bool {
	if err == nil {
		return false
	}
	switch {
	case errors.Is(err, repositories.ErrNotFound), errors.Is(err, services.ErrNotFound):
		response.Error(w, http.StatusNotFound, "not_found", "Registro não encontrado")
	case errors.Is(err, repositories.ErrTagExists):
		response.Error(w, http.StatusConflict, "tag_exists", err.Error())
	case errors.Is(err, services.ErrAccountNotInOrg):
		response.Error(w, http.StatusUnprocessableEntity, "invalid_account", err.Error())
	case errors.Is(err, services.ErrCreditCardNotInOrg):
		response.Error(w, http.StatusUnprocessableEntity, "invalid_credit_card", err.Error())
	case errors.Is(err, services.ErrContactNotInOrg):
		response.Error(w, http.StatusUnprocessableEntity, "invalid_contact", err.Error())
	case errors.Is(err, services.ErrCategoryNotExpense):
		response.Error(w, http.StatusUnprocessableEntity, "invalid_category", err.Error())
	case errors.Is(err, services.ErrUnsupportedCurrency):
		response.Error(w, http.StatusUnprocessableEntity, "unsupported_currency", err.Error())
	case errors.Is(err, services.ErrInvalidOrgName):
		response.Error(w, http.StatusUnprocessableEntity, "invalid_org_name", err.Error())
	case errors.Is(err, services.ErrInvalidRule):
		response.Error(w, http.StatusUnprocessableEntity, "invalid_rule", err.Error())
	case errors.Is(err, services.ErrInvalidRecurrence):
		response.Error(w, http.StatusUnprocessableEntity, "invalid_recurrence", err.Error())
	case errors.Is(err, services.ErrEmptyImport):
		response.Error(w, http.StatusUnprocessableEntity, "empty_import", err.Error())
	case errors.Is(err, services.ErrUnsupportedFormat):
		response.Error(w, http.StatusUnprocessableEntity, "unsupported_format", err.Error())
	case errors.Is(err, services.ErrUnsupportedType):
		response.Error(w, http.StatusUnprocessableEntity, "unsupported_type", err.Error())
	case errors.Is(err, services.ErrFileTooLarge):
		response.Error(w, http.StatusRequestEntityTooLarge, "file_too_large", err.Error())
	case errors.Is(err, storage.ErrStorageUnavailable):
		response.Error(w, http.StatusServiceUnavailable, "storage_unavailable", err.Error())
	case errors.Is(err, services.ErrLastOwner):
		response.Error(w, http.StatusConflict, "last_owner", err.Error())
	case errors.Is(err, services.ErrOwnerOnly):
		response.Error(w, http.StatusForbidden, "forbidden", err.Error())
	case errors.Is(err, services.ErrAlreadyMember):
		response.Error(w, http.StatusConflict, "already_member", err.Error())
	case errors.Is(err, services.ErrInvitationToken):
		response.Error(w, http.StatusNotFound, "invalid_invitation", err.Error())
	case errors.Is(err, services.ErrInvalidToken):
		response.Error(w, http.StatusBadRequest, "invalid_token", err.Error())
	case errors.Is(err, services.ErrAccountLocked):
		response.Error(w, http.StatusLocked, "account_locked", err.Error())
	case errors.Is(err, services.ErrInvalidCode):
		response.Error(w, http.StatusUnprocessableEntity, "invalid_code", err.Error())
	case errors.Is(err, services.ErrWrongPassword):
		response.Error(w, http.StatusUnauthorized, "wrong_password", err.Error())
	// ----- Platform back-office (super-admin) -----
	case errors.Is(err, services.ErrEmailTaken):
		response.Error(w, http.StatusConflict, "email_taken", err.Error())
	case errors.Is(err, services.ErrAccountDisabled):
		response.Error(w, http.StatusForbidden, "account_disabled", err.Error())
	case errors.Is(err, services.ErrCannotDisableSelf):
		response.Error(w, http.StatusConflict, "cannot_disable_self", err.Error())
	case errors.Is(err, services.ErrLastSuperAdmin):
		response.Error(w, http.StatusConflict, "last_super_admin", err.Error())
	default:
		response.Error(w, http.StatusInternalServerError, "internal_error", "Erro interno do servidor")
	}
	return true
}
