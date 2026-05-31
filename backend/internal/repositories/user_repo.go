package repositories

import (
	"errors"
	"time"

	"github.com/finance-sh/finance-sh/internal/entities"
	"github.com/finance-sh/finance-sh/pkg/crypto"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var ErrNotFound = errors.New("registro não encontrado")

type UserRepository struct{ db *gorm.DB }

func NewUserRepository(db *gorm.DB) *UserRepository { return &UserRepository{db: db} }

// CreateWithOrganization atomically creates the user, their first organization
// and an owner membership. This is the bootstrap of a new tenant on self-signup.
func (r *UserRepository) CreateWithOrganization(user *entities.User, org *entities.Organization) (*entities.Membership, error) {
	var membership entities.Membership
	err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(user).Error; err != nil {
			return err
		}
		org.OwnerID = user.ID
		if err := tx.Create(org).Error; err != nil {
			return err
		}
		membership = entities.Membership{
			UserID:         user.ID,
			OrganizationID: org.ID,
			Role:           entities.RoleOwner,
		}
		if err := tx.Create(&membership).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &membership, nil
}

func (r *UserRepository) FindByEmail(email string) (*entities.User, error) {
	var u entities.User
	err := r.db.Where("email = ?", email).First(&u).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &u, err
}

func (r *UserRepository) FindByID(id uuid.UUID) (*entities.User, error) {
	var u entities.User
	err := r.db.First(&u, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &u, err
}

// FindByIDs batch-loads the users with the given ids in a single query. Empty
// ids yields an empty slice (no query). Used to resolve user names for listings
// (e.g. the audit log) without an N+1 round-trip.
func (r *UserRepository) FindByIDs(ids []uuid.UUID) ([]entities.User, error) {
	if len(ids) == 0 {
		return []entities.User{}, nil
	}
	var users []entities.User
	err := r.db.Where("id IN ?", ids).Find(&users).Error
	return users, err
}

func (r *UserRepository) EmailExists(email string) bool {
	var count int64
	r.db.Model(&entities.User{}).Where("email = ?", email).Count(&count)
	return count > 0
}

// CountAll returns the total number of users in the platform regardless of
// organization. Used by the first-run setup wizard to decide whether the
// instance still needs an initial super-admin (count == 0).
func (r *UserRepository) CountAll() (int64, error) {
	var n int64
	err := r.db.Model(&entities.User{}).Count(&n).Error
	return n, err
}

// Memberships returns all org memberships for a user with org preloaded, ordered
// by created_at ascending. The stable order makes the Tenant middleware's
// no-header fallback (memberships[0]) deterministic for multi-org users.
func (r *UserRepository) Memberships(userID uuid.UUID) ([]entities.Membership, error) {
	var m []entities.Membership
	err := r.db.Preload("Organization").Where("user_id = ?", userID).
		Order("created_at asc").Find(&m).Error
	return m, err
}

// FindOrganization loads an organization by id. Used by reports that need the
// org name/currency for headers (there is no dedicated org repository).
func (r *UserRepository) FindOrganization(id uuid.UUID) (*entities.Organization, error) {
	var o entities.Organization
	err := r.db.First(&o, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &o, err
}

// Membership returns the membership of a user in a specific org (tenant guard).
func (r *UserRepository) Membership(userID, orgID uuid.UUID) (*entities.Membership, error) {
	var m entities.Membership
	err := r.db.Where("user_id = ? AND organization_id = ?", userID, orgID).First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &m, err
}

// ----- Refresh tokens -----

func (r *UserRepository) SaveRefreshToken(t *entities.RefreshToken) error {
	return r.db.Create(t).Error
}

func (r *UserRepository) FindRefreshToken(hash string) (*entities.RefreshToken, error) {
	var t entities.RefreshToken
	err := r.db.Where("token_hash = ? AND revoked = false", hash).First(&t).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &t, err
}

func (r *UserRepository) RevokeRefreshToken(id uuid.UUID) error {
	return r.db.Model(&entities.RefreshToken{}).Where("id = ?", id).Update("revoked", true).Error
}

// RevokeAllRefreshTokens revokes every active refresh token of a user (used
// after a password reset to invalidate existing sessions).
func (r *UserRepository) RevokeAllRefreshTokens(userID uuid.UUID) error {
	return r.db.Model(&entities.RefreshToken{}).
		Where("user_id = ? AND revoked = false", userID).
		Update("revoked", true).Error
}

// ActiveRefreshTokens returns the user's non-revoked, non-expired refresh tokens
// (their active sessions), newest first. The token hash is never exposed by the
// caller — only metadata.
func (r *UserRepository) ActiveRefreshTokens(userID uuid.UUID) ([]entities.RefreshToken, error) {
	var tokens []entities.RefreshToken
	err := r.db.
		Where("user_id = ? AND revoked = false AND expires_at > ?", userID, time.Now().UTC()).
		Order("created_at desc").
		Find(&tokens).Error
	return tokens, err
}

// RevokeRefreshTokenForUser revokes a single refresh token that must belong to
// the user (defends against revoking another user's session by id). Returns
// ErrNotFound when no matching active token exists.
func (r *UserRepository) RevokeRefreshTokenForUser(userID, id uuid.UUID) error {
	res := r.db.Model(&entities.RefreshToken{}).
		Where("id = ? AND user_id = ?", id, userID).
		Update("revoked", true)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// RevokeOtherRefreshTokens revokes every active refresh token of the user EXCEPT
// the one whose hash equals keepHash. When keepHash is empty all are revoked.
// Returns the number revoked.
func (r *UserRepository) RevokeOtherRefreshTokens(userID uuid.UUID, keepHash string) (int64, error) {
	q := r.db.Model(&entities.RefreshToken{}).
		Where("user_id = ? AND revoked = false", userID)
	if keepHash != "" {
		q = q.Where("token_hash <> ?", keepHash)
	}
	res := q.Update("revoked", true)
	return res.RowsAffected, res.Error
}

// SetPassword updates a user's bcrypt password hash.
func (r *UserRepository) SetPassword(userID uuid.UUID, passwordHash string) error {
	return r.db.Model(&entities.User{}).
		Where("id = ?", userID).
		Update("password_hash", passwordHash).Error
}

// AdminResetPassword sets a new password hash AND marks must_change_password
// (super-admin reset: the user is forced to change it on their next login).
func (r *UserRepository) AdminResetPassword(userID uuid.UUID, passwordHash string) error {
	return r.db.Model(&entities.User{}).
		Where("id = ?", userID).
		Updates(map[string]interface{}{
			"password_hash":        passwordHash,
			"must_change_password": true,
		}).Error
}

// Save persists changes to an existing user row.
func (r *UserRepository) Save(u *entities.User) error {
	return r.db.Save(u).Error
}

// ----- Platform back-office (super-admin) -----

// SetDisabled toggles a user's platform-level disabled flag (back-office).
func (r *UserRepository) SetDisabled(id uuid.UUID, disabled bool) error {
	res := r.db.Model(&entities.User{}).Where("id = ?", id).Update("disabled", disabled)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// CountSuperAdmins returns the number of (enabled) platform super-admins. Used to
// guard against disabling the last super-admin.
func (r *UserRepository) CountSuperAdmins() (int64, error) {
	var n int64
	err := r.db.Model(&entities.User{}).
		Where("super_admin = ? AND disabled = ?", true, false).
		Count(&n).Error
	return n, err
}

// ----- Email verification -----

func (r *UserRepository) CreateEmailVerification(ev *entities.EmailVerification) error {
	return r.db.Create(ev).Error
}

func (r *UserRepository) FindEmailVerification(tokenHash string) (*entities.EmailVerification, error) {
	var ev entities.EmailVerification
	err := r.db.Where("token_hash = ?", tokenHash).First(&ev).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &ev, err
}

func (r *UserRepository) MarkEmailVerificationUsed(id uuid.UUID) error {
	return r.db.Model(&entities.EmailVerification{}).
		Where("id = ?", id).Update("used", true).Error
}

// MarkEmailVerified flips the user's verified flag.
func (r *UserRepository) MarkEmailVerified(userID uuid.UUID) error {
	return r.db.Model(&entities.User{}).
		Where("id = ?", userID).Update("email_verified", true).Error
}

// ----- Two-factor / recovery codes -----

// SetTwoFactorSecret stores a (pending) TOTP secret without enabling 2FA yet.
// The EncryptedString type encrypts the value at the DB boundary.
func (r *UserRepository) SetTwoFactorSecret(userID uuid.UUID, secret string) error {
	return r.db.Model(&entities.User{}).
		Where("id = ?", userID).
		Update("two_factor_secret", crypto.EncryptedString(secret)).Error
}

// SetTwoFactorEnabled toggles the enabled flag.
func (r *UserRepository) SetTwoFactorEnabled(userID uuid.UUID, enabled bool) error {
	return r.db.Model(&entities.User{}).
		Where("id = ?", userID).Update("two_factor_enabled", enabled).Error
}

// DisableTwoFactor clears the secret and disables 2FA, removing recovery codes.
func (r *UserRepository) DisableTwoFactor(userID uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&entities.User{}).Where("id = ?", userID).
			Updates(map[string]interface{}{"two_factor_enabled": false, "two_factor_secret": ""}).Error; err != nil {
			return err
		}
		return tx.Where("user_id = ?", userID).Delete(&entities.RecoveryCode{}).Error
	})
}

// ReplaceRecoveryCodes deletes existing codes and stores the new hashes.
func (r *UserRepository) ReplaceRecoveryCodes(userID uuid.UUID, hashes []string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ?", userID).Delete(&entities.RecoveryCode{}).Error; err != nil {
			return err
		}
		for _, h := range hashes {
			if err := tx.Create(&entities.RecoveryCode{UserID: userID, CodeHash: h}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// FindRecoveryCode returns an unused recovery code matching the hash.
func (r *UserRepository) FindRecoveryCode(userID uuid.UUID, codeHash string) (*entities.RecoveryCode, error) {
	var rc entities.RecoveryCode
	err := r.db.Where("user_id = ? AND code_hash = ? AND used = false", userID, codeHash).First(&rc).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &rc, err
}

// UseRecoveryCode marks a recovery code consumed.
func (r *UserRepository) UseRecoveryCode(id uuid.UUID) error {
	return r.db.Model(&entities.RecoveryCode{}).
		Where("id = ?", id).Update("used", true).Error
}
