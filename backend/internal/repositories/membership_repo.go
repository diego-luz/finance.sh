package repositories

import (
	"errors"

	"github.com/finance-sh/finance-sh/internal/entities"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// MembershipRepository handles org membership and invitations. Everything here
// is scoped by organization_id.
type MembershipRepository struct{ db *gorm.DB }

func NewMembershipRepository(db *gorm.DB) *MembershipRepository {
	return &MembershipRepository{db: db}
}

// ----- Members -----

// List returns all memberships of an org with the User preloaded.
func (r *MembershipRepository) List(orgID uuid.UUID) ([]entities.Membership, error) {
	var members []entities.Membership
	err := r.db.Preload("User").
		Where("organization_id = ?", orgID).
		Order("created_at asc").Find(&members).Error
	return members, err
}

// FindByID returns a single membership of the org.
func (r *MembershipRepository) FindByID(orgID, id uuid.UUID) (*entities.Membership, error) {
	var m entities.Membership
	err := r.db.Preload("User").
		Where("organization_id = ? AND id = ?", orgID, id).First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &m, err
}

func (r *MembershipRepository) Create(m *entities.Membership) error {
	return r.db.Create(m).Error
}

func (r *MembershipRepository) UpdateRole(orgID, id uuid.UUID, role entities.Role) error {
	res := r.db.Model(&entities.Membership{}).
		Where("organization_id = ? AND id = ?", orgID, id).
		Update("role", role)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *MembershipRepository) Delete(orgID, id uuid.UUID) error {
	res := r.db.Where("organization_id = ? AND id = ?", orgID, id).Delete(&entities.Membership{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// CountMembers returns the number of members in the org.
func (r *MembershipRepository) CountMembers(orgID uuid.UUID) (int64, error) {
	var n int64
	err := r.db.Model(&entities.Membership{}).Where("organization_id = ?", orgID).Count(&n).Error
	return n, err
}

// CountOwners returns the number of owners in the org (used to protect the last
// owner from demotion/removal).
func (r *MembershipRepository) CountOwners(orgID uuid.UUID) (int64, error) {
	var n int64
	err := r.db.Model(&entities.Membership{}).
		Where("organization_id = ? AND role = ?", orgID, entities.RoleOwner).
		Count(&n).Error
	return n, err
}

// MembershipByUserOrg returns an existing membership for a user in an org, if any.
func (r *MembershipRepository) MembershipByUserOrg(userID, orgID uuid.UUID) (*entities.Membership, error) {
	var m entities.Membership
	err := r.db.Where("user_id = ? AND organization_id = ?", userID, orgID).First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &m, err
}

// ----- Invitations -----

func (r *MembershipRepository) CreateInvitation(i *entities.Invitation) error {
	return r.db.Create(i).Error
}

// PendingInvitations returns the org's non-accepted invitations.
func (r *MembershipRepository) PendingInvitations(orgID uuid.UUID) ([]entities.Invitation, error) {
	var invites []entities.Invitation
	err := r.db.Where("organization_id = ? AND accepted = ?", orgID, false).
		Order("created_at desc").Find(&invites).Error
	return invites, err
}

func (r *MembershipRepository) CountPendingInvitations(orgID uuid.UUID) (int64, error) {
	var n int64
	err := r.db.Model(&entities.Invitation{}).
		Where("organization_id = ? AND accepted = ?", orgID, false).
		Count(&n).Error
	return n, err
}

func (r *MembershipRepository) FindInvitationByID(orgID, id uuid.UUID) (*entities.Invitation, error) {
	var i entities.Invitation
	err := r.db.Where("organization_id = ? AND id = ?", orgID, id).First(&i).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &i, err
}

// FindInvitationByToken resolves an invitation by its opaque token regardless of
// the active tenant (used by the accept flow).
func (r *MembershipRepository) FindInvitationByToken(token string) (*entities.Invitation, error) {
	var i entities.Invitation
	err := r.db.Where("token = ? AND accepted = ?", token, false).First(&i).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &i, err
}

func (r *MembershipRepository) DeleteInvitation(orgID, id uuid.UUID) error {
	res := r.db.Where("organization_id = ? AND id = ?", orgID, id).Delete(&entities.Invitation{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// AcceptInvitation atomically marks the invitation accepted and creates the
// membership for the accepting user.
func (r *MembershipRepository) AcceptInvitation(inv *entities.Invitation, userID uuid.UUID) (*entities.Membership, error) {
	membership := &entities.Membership{
		UserID:         userID,
		OrganizationID: inv.OrganizationID,
		Role:           inv.Role,
	}
	err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&entities.Invitation{}).
			Where("id = ?", inv.ID).Update("accepted", true).Error; err != nil {
			return err
		}
		return tx.Create(membership).Error
	})
	if err != nil {
		return nil, err
	}
	return membership, nil
}
