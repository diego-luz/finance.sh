package entities

import (
	"time"

	"github.com/finance-sh/finance-sh/pkg/crypto"
	"github.com/google/uuid"
)

// User is a global identity. A user may belong to many organizations through
// Membership rows, which is what enables multi-tenancy.
type User struct {
	Base
	Name          string `gorm:"not null" json:"name"`
	Email         string `gorm:"uniqueIndex;not null" json:"email"`
	PasswordHash  string `gorm:"not null" json:"-"`
	EmailVerified bool   `gorm:"default:false" json:"email_verified"`
	AvatarURL     string `json:"avatar_url,omitempty"`

	// SuperAdmin grants access to the platform back-office (/admin). This is a
	// PLATFORM-level role, distinct from per-organization RBAC (Membership.Role).
	// A super-admin is not tenant-scoped.
	SuperAdmin bool `gorm:"default:false" json:"super_admin"`
	// Disabled blocks login entirely (platform-level account lockout enforced by
	// a super-admin). Never serialised to JSON.
	Disabled bool `gorm:"default:false" json:"-"`

	// MustChangePassword forces the user to set a new password on their next
	// login. Set to true by AdminService.Provision (the operator types an initial
	// password) and cleared by AuthService.ChangePassword. The frontend reads the
	// flag from UserDTO.must_change_password and routes accordingly; the server
	// itself does NOT block authentication so the user can call /me/change-password.
	MustChangePassword bool `gorm:"default:false" json:"-"`

	// LGPD consent: timestamp + accepted terms version at registration.
	TermsAcceptedAt *time.Time `json:"terms_accepted_at,omitempty"`
	TermsVersion    string     `gorm:"type:varchar(20)" json:"terms_version,omitempty"`

	// Two-factor authentication (TOTP). The secret is encrypted at rest because
	// it is a credential equivalent; it is never serialised to JSON.
	TwoFactorEnabled bool                   `gorm:"default:false" json:"two_factor_enabled"`
	TwoFactorSecret  crypto.EncryptedString `gorm:"type:text" json:"-"`

	Memberships []Membership `json:"memberships,omitempty"`
}

func (User) TableName() string { return "users" }

// EmailVerification stores a single-use email-verification token. Only the
// SHA-256 hash of the raw token is persisted; tokens expire after 24h.
type EmailVerification struct {
	Base
	UserID    uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	TokenHash string    `gorm:"uniqueIndex;not null" json:"-"`
	ExpiresAt time.Time `gorm:"not null" json:"expires_at"`
	Used      bool      `gorm:"default:false" json:"used"`
}

func (EmailVerification) TableName() string { return "email_verifications" }

// RecoveryCode is a single-use 2FA backup code. Only the SHA-256 hash of the
// raw code is persisted, mirroring how refresh/reset tokens are handled.
type RecoveryCode struct {
	Base
	UserID   uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	CodeHash string    `gorm:"not null;index" json:"-"`
	Used     bool      `gorm:"default:false" json:"used"`
}

func (RecoveryCode) TableName() string { return "recovery_codes" }

// RefreshToken stores the hash of an issued refresh token so it can be rotated
// and revoked server-side. The raw token is never persisted.
type RefreshToken struct {
	Base
	UserID    uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	TokenHash string    `gorm:"uniqueIndex;not null" json:"-"`
	ExpiresAt time.Time `gorm:"not null" json:"expires_at"`
	Revoked   bool      `gorm:"default:false" json:"revoked"`
	UserAgent string    `json:"user_agent,omitempty"`
	IP        string    `json:"ip,omitempty"`
}

func (RefreshToken) TableName() string { return "refresh_tokens" }

// PasswordReset stores a single-use password-reset token. The raw token is never
// persisted; only its SHA-256 hash is kept. Tokens expire after one hour.
type PasswordReset struct {
	Base
	UserID    uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	TokenHash string    `gorm:"uniqueIndex;not null" json:"-"`
	ExpiresAt time.Time `gorm:"not null" json:"expires_at"`
	Used      bool      `gorm:"default:false" json:"used"`
}

func (PasswordReset) TableName() string { return "password_resets" }
