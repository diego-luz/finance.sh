package services

import (
	"errors"
	"strings"

	"github.com/finance-sh/finance-sh/internal/dto"
	"github.com/finance-sh/finance-sh/internal/entities"
	"github.com/finance-sh/finance-sh/internal/repositories"
	"github.com/finance-sh/finance-sh/pkg/hash"
	"github.com/google/uuid"
)

// Typed errors for member/invitation flows (mapped to HTTP by the handlers).
var (
	ErrLastOwner       = errors.New("não é possível remover ou rebaixar o último proprietário")
	ErrAlreadyMember   = errors.New("usuário já é membro da organização")
	ErrInvitationToken = errors.New("convite inválido ou já utilizado")
	// ErrOwnerOnly is returned when an admin attempts an action reserved to owners
	// (granting/altering/removing the owner role). Mapped to HTTP 403.
	ErrOwnerOnly = errors.New("apenas um proprietário pode conceder ou alterar o papel de proprietário")
)

type MemberService struct {
	members *repositories.MembershipRepository
	users   *repositories.UserRepository
}

func NewMemberService(
	members *repositories.MembershipRepository,
	users *repositories.UserRepository,
) *MemberService {
	return &MemberService{members: members, users: users}
}

// ----- Members -----

func (s *MemberService) ListMembers(orgID uuid.UUID) ([]dto.MemberDTO, error) {
	members, err := s.members.List(orgID)
	if err != nil {
		return nil, err
	}
	out := make([]dto.MemberDTO, 0, len(members))
	for i := range members {
		out = append(out, memberDTO(&members[i]))
	}
	return out, nil
}

// UpdateRole changes a member's role, refusing to demote the last owner.
//
// Privilege guard: only an owner may grant the owner role or change an existing
// owner's role. An admin may only manage admin/member/viewer members and may
// never escalate anyone (including themselves) to owner.
func (s *MemberService) UpdateRole(orgID, membershipID uuid.UUID, role entities.Role, actorRole entities.Role) (*dto.MemberDTO, error) {
	m, err := s.members.FindByID(orgID, membershipID)
	if err != nil {
		return nil, err
	}
	// Non-owners cannot grant the owner role nor touch an existing owner.
	if actorRole != entities.RoleOwner && (role == entities.RoleOwner || m.Role == entities.RoleOwner) {
		return nil, ErrOwnerOnly
	}
	if m.Role == entities.RoleOwner && role != entities.RoleOwner {
		owners, err := s.members.CountOwners(orgID)
		if err != nil {
			return nil, err
		}
		if owners <= 1 {
			return nil, ErrLastOwner
		}
	}
	if err := s.members.UpdateRole(orgID, membershipID, role); err != nil {
		return nil, err
	}
	updated, err := s.members.FindByID(orgID, membershipID)
	if err != nil {
		return nil, err
	}
	d := memberDTO(updated)
	return &d, nil
}

// RemoveMember deletes a membership, refusing to remove the last owner.
//
// Privilege guard: only an owner may remove an owner membership. An admin may
// only remove admin/member/viewer members.
func (s *MemberService) RemoveMember(orgID, membershipID uuid.UUID, actorRole entities.Role) error {
	m, err := s.members.FindByID(orgID, membershipID)
	if err != nil {
		return err
	}
	if m.Role == entities.RoleOwner && actorRole != entities.RoleOwner {
		return ErrOwnerOnly
	}
	if m.Role == entities.RoleOwner {
		owners, err := s.members.CountOwners(orgID)
		if err != nil {
			return err
		}
		if owners <= 1 {
			return ErrLastOwner
		}
	}
	return s.members.Delete(orgID, membershipID)
}

// ----- Invitations -----

func (s *MemberService) ListInvitations(orgID uuid.UUID) ([]dto.InvitationDTO, error) {
	invites, err := s.members.PendingInvitations(orgID)
	if err != nil {
		return nil, err
	}
	out := make([]dto.InvitationDTO, 0, len(invites))
	for i := range invites {
		out = append(out, invitationDTO(&invites[i], ""))
	}
	return out, nil
}

// CreateInvitation creates an invitation including its raw token (email delivery
// is future). Members are unlimited — open-source, no plan caps.
func (s *MemberService) CreateInvitation(orgID uuid.UUID, req dto.InvitationRequest) (*dto.InvitationDTO, error) {
	token, err := hash.RandomToken(24)
	if err != nil {
		return nil, err
	}
	inv := &entities.Invitation{
		OrganizationID: orgID,
		Email:          strings.ToLower(strings.TrimSpace(req.Email)),
		Role:           entities.Role(req.Role),
		Token:          token,
	}
	if err := s.members.CreateInvitation(inv); err != nil {
		return nil, err
	}
	d := invitationDTO(inv, token)
	return &d, nil
}

func (s *MemberService) RevokeInvitation(orgID, id uuid.UUID) error {
	return s.members.DeleteInvitation(orgID, id)
}

// AcceptInvitation lets the authenticated user join the inviting org. The org is
// resolved from the invitation, so the active tenant is irrelevant.
func (s *MemberService) AcceptInvitation(userID uuid.UUID, token string) (*dto.MemberDTO, error) {
	inv, err := s.members.FindInvitationByToken(strings.TrimSpace(token))
	if err != nil {
		return nil, ErrInvitationToken
	}
	if existing, err := s.members.MembershipByUserOrg(userID, inv.OrganizationID); err == nil && existing != nil {
		return nil, ErrAlreadyMember
	}
	m, err := s.members.AcceptInvitation(inv, userID)
	if err != nil {
		return nil, err
	}
	// Reload with the user preloaded for the response.
	full, err := s.members.FindByID(inv.OrganizationID, m.ID)
	if err != nil {
		return nil, err
	}
	d := memberDTO(full)
	return &d, nil
}

func memberDTO(m *entities.Membership) dto.MemberDTO {
	d := dto.MemberDTO{
		ID:   m.ID.String(),
		Role: string(m.Role),
	}
	if m.User != nil {
		d.User = userDTO(m.User)
	}
	return d
}

func invitationDTO(i *entities.Invitation, token string) dto.InvitationDTO {
	return dto.InvitationDTO{
		ID:        i.ID.String(),
		Email:     i.Email,
		Role:      string(i.Role),
		Accepted:  i.Accepted,
		Token:     token,
		CreatedAt: i.CreatedAt,
	}
}
