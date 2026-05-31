package handlers

import (
	"net/http"

	"github.com/finance-sh/finance-sh/internal/dto"
	"github.com/finance-sh/finance-sh/internal/middlewares"
	"github.com/finance-sh/finance-sh/internal/services"
	"github.com/finance-sh/finance-sh/pkg/response"
	"github.com/finance-sh/finance-sh/pkg/validator"
)

// AdminHandler serves the platform back-office (super-admin): read-only views of
// the instance plus a couple of user-safety operations. The routes are gated by
// Auth + RequireSuperAdmin in the router; they are NOT tenant-scoped.
type AdminHandler struct {
	admin *services.AdminService
}

func NewAdminHandler(admin *services.AdminService) *AdminHandler {
	return &AdminHandler{admin: admin}
}

// Stats GET /admin/stats — global platform counters.
func (h *AdminHandler) Stats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.admin.Stats()
	if writeServiceError(w, err) {
		return
	}
	response.OK(w, stats)
}

// ListOrganizations GET /admin/organizations?search=&page=&per_page=
func (h *AdminHandler) ListOrganizations(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	items, meta, err := h.admin.ListOrganizations(
		q.Get("search"),
		atoiDefault(q.Get("page"), 1),
		atoiDefault(q.Get("per_page"), 20),
	)
	if writeServiceError(w, err) {
		return
	}
	response.Paginated(w, items, meta)
}

// ListUsers GET /admin/users?search=&page=&per_page=
func (h *AdminHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	items, meta, err := h.admin.ListUsers(
		q.Get("search"),
		atoiDefault(q.Get("page"), 1),
		atoiDefault(q.Get("per_page"), 20),
	)
	if writeServiceError(w, err) {
		return
	}
	response.Paginated(w, items, meta)
}

// DisableUser POST /admin/users/{id}/disable
func (h *AdminHandler) DisableUser(w http.ResponseWriter, r *http.Request) {
	id, ok := urlUUID(w, r, "id")
	if !ok {
		return
	}
	callerID := middlewares.UserID(r.Context())
	if err := h.admin.SetUserDisabled(callerID, id, true); writeServiceError(w, err) {
		return
	}
	response.OK(w, map[string]string{"message": "Usuário desativado."})
}

// EnableUser POST /admin/users/{id}/enable
func (h *AdminHandler) EnableUser(w http.ResponseWriter, r *http.Request) {
	id, ok := urlUUID(w, r, "id")
	if !ok {
		return
	}
	callerID := middlewares.UserID(r.Context())
	if err := h.admin.SetUserDisabled(callerID, id, false); writeServiceError(w, err) {
		return
	}
	response.OK(w, map[string]string{"message": "Usuário reativado."})
}

// ResetUserPassword POST /admin/users/{id}/reset-password — super-admin sets a
// new password for a user; user is forced to change it on next login and all
// their refresh tokens are revoked.
func (h *AdminHandler) ResetUserPassword(w http.ResponseWriter, r *http.Request) {
	id, ok := urlUUID(w, r, "id")
	if !ok {
		return
	}
	var req dto.AdminResetPasswordRequest
	if fields, err := validator.BindJSON(r, &req); err != nil || len(fields) > 0 {
		response.ValidationError(w, fields)
		return
	}
	callerID := middlewares.UserID(r.Context())
	if err := h.admin.ResetUserPassword(callerID, id, req.NewPassword); writeServiceError(w, err) {
		return
	}
	response.OK(w, map[string]string{"message": "Senha redefinida. O usuário será obrigado a trocá-la no próximo acesso."})
}
