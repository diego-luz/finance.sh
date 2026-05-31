package handlers

import (
	"net/http"

	"github.com/finance-sh/finance-sh/internal/dto"
	"github.com/finance-sh/finance-sh/internal/entities"
	"github.com/finance-sh/finance-sh/internal/middlewares"
	"github.com/finance-sh/finance-sh/internal/services"
	"github.com/finance-sh/finance-sh/pkg/response"
	"github.com/finance-sh/finance-sh/pkg/validator"
)

type MemberHandler struct {
	members *services.MemberService
}

func NewMemberHandler(members *services.MemberService) *MemberHandler {
	return &MemberHandler{members: members}
}

// ----- Members -----

// ListMembers GET /members
func (h *MemberHandler) ListMembers(w http.ResponseWriter, r *http.Request) {
	orgID := middlewares.OrgID(r.Context())
	items, err := h.members.ListMembers(orgID)
	if writeServiceError(w, err) {
		return
	}
	response.OK(w, items)
}

// UpdateMember PUT /members/{id}
func (h *MemberHandler) UpdateMember(w http.ResponseWriter, r *http.Request) {
	orgID := middlewares.OrgID(r.Context())
	id, ok := urlUUID(w, r, "id")
	if !ok {
		return
	}
	var req dto.UpdateMemberRequest
	if fields, err := validator.BindJSON(r, &req); err != nil || len(fields) > 0 {
		response.ValidationError(w, fields)
		return
	}
	item, err := h.members.UpdateRole(orgID, id, entities.Role(req.Role), middlewares.UserRole(r.Context()))
	if writeServiceError(w, err) {
		return
	}
	response.OK(w, item)
}

// RemoveMember DELETE /members/{id}
func (h *MemberHandler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	orgID := middlewares.OrgID(r.Context())
	id, ok := urlUUID(w, r, "id")
	if !ok {
		return
	}
	if writeServiceError(w, h.members.RemoveMember(orgID, id, middlewares.UserRole(r.Context()))) {
		return
	}
	response.NoContent(w)
}

// ----- Invitations -----

// ListInvitations GET /invitations
func (h *MemberHandler) ListInvitations(w http.ResponseWriter, r *http.Request) {
	orgID := middlewares.OrgID(r.Context())
	items, err := h.members.ListInvitations(orgID)
	if writeServiceError(w, err) {
		return
	}
	response.OK(w, items)
}

// CreateInvitation POST /invitations
func (h *MemberHandler) CreateInvitation(w http.ResponseWriter, r *http.Request) {
	orgID := middlewares.OrgID(r.Context())
	var req dto.InvitationRequest
	if fields, err := validator.BindJSON(r, &req); err != nil || len(fields) > 0 {
		response.ValidationError(w, fields)
		return
	}
	item, err := h.members.CreateInvitation(orgID, req)
	if writeServiceError(w, err) {
		return
	}
	response.Created(w, item)
}

// RevokeInvitation DELETE /invitations/{id}
func (h *MemberHandler) RevokeInvitation(w http.ResponseWriter, r *http.Request) {
	orgID := middlewares.OrgID(r.Context())
	id, ok := urlUUID(w, r, "id")
	if !ok {
		return
	}
	if writeServiceError(w, h.members.RevokeInvitation(orgID, id)) {
		return
	}
	response.NoContent(w)
}

// AcceptInvitation POST /invitations/accept — the authenticated user joins the
// inviting org. Tenant-agnostic: the org is resolved from the invitation.
func (h *MemberHandler) AcceptInvitation(w http.ResponseWriter, r *http.Request) {
	userID := middlewares.UserID(r.Context())
	var req dto.AcceptInvitationRequest
	if fields, err := validator.BindJSON(r, &req); err != nil || len(fields) > 0 {
		response.ValidationError(w, fields)
		return
	}
	item, err := h.members.AcceptInvitation(userID, req.Token)
	if writeServiceError(w, err) {
		return
	}
	response.OK(w, item)
}
