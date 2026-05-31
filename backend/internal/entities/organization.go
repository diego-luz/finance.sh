package entities

import "github.com/google/uuid"

// Role defines the permission level of a user within an organization (RBAC).
type Role string

const (
	RoleOwner  Role = "owner"
	RoleAdmin  Role = "admin"
	RoleMember Role = "member"
	RoleViewer Role = "viewer"
)

// Organization is the tenant boundary. Every financial record carries an
// OrganizationID and all queries are scoped by it for full data isolation.
type Organization struct {
	Base
	Name     string    `gorm:"not null" json:"name"`
	Slug     string    `gorm:"uniqueIndex;not null" json:"slug"`
	OwnerID  uuid.UUID `gorm:"type:uuid;not null;index" json:"owner_id"`
	Currency string    `gorm:"type:varchar(3);default:'BRL'" json:"currency"`
}

func (Organization) TableName() string { return "organizations" }

// Membership links a User to an Organization with a role. The composite unique
// index prevents a user from joining the same organization twice.
type Membership struct {
	Base
	UserID         uuid.UUID     `gorm:"type:uuid;not null;uniqueIndex:idx_user_org" json:"user_id"`
	OrganizationID uuid.UUID     `gorm:"type:uuid;not null;uniqueIndex:idx_user_org" json:"organization_id"`
	Role           Role          `gorm:"type:varchar(20);default:'member'" json:"role"`
	User           *User         `json:"user,omitempty"`
	Organization   *Organization `json:"organization,omitempty"`
}

func (Membership) TableName() string { return "memberships" }

// Invitation lets an organization admin invite a new member by email before
// the invitee has an account.
type Invitation struct {
	Base
	OrganizationID uuid.UUID `gorm:"type:uuid;not null;index" json:"organization_id"`
	Email          string    `gorm:"not null;index" json:"email"`
	Role           Role      `gorm:"type:varchar(20);default:'member'" json:"role"`
	Token          string    `gorm:"uniqueIndex;not null" json:"-"`
	Accepted       bool      `gorm:"default:false" json:"accepted"`
}

func (Invitation) TableName() string { return "invitations" }
