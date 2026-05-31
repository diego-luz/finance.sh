package repositories

import (
	"errors"
	"time"

	"github.com/finance-sh/finance-sh/internal/entities"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// PasswordResetRepository persists single-use password reset tokens.
type PasswordResetRepository struct{ db *gorm.DB }

func NewPasswordResetRepository(db *gorm.DB) *PasswordResetRepository {
	return &PasswordResetRepository{db: db}
}

func (r *PasswordResetRepository) Create(pr *entities.PasswordReset) error {
	return r.db.Create(pr).Error
}

// FindByTokenHash returns the (unused) reset record matching the token hash.
func (r *PasswordResetRepository) FindByTokenHash(tokenHash string) (*entities.PasswordReset, error) {
	var pr entities.PasswordReset
	err := r.db.Where("token_hash = ?", tokenHash).First(&pr).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &pr, err
}

func (r *PasswordResetRepository) MarkUsed(id uuid.UUID) error {
	return r.db.Model(&entities.PasswordReset{}).
		Where("id = ?", id).Update("used", true).Error
}

// ConsumeAndResetPassword atomically (in a single DB transaction): re-validates
// the token (unused + not expired) by hash, sets the user's password, marks the
// token used, and revokes all of the user's active refresh tokens. A partial
// failure rolls everything back so the token can never be left reusable.
//
// It returns ErrNotFound when the token is missing, already used or expired so
// the caller can map it to an "invalid token" response.
func (r *PasswordResetRepository) ConsumeAndResetPassword(tokenHash, passwordHash string, now time.Time) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var pr entities.PasswordReset
		// Lock the row to prevent concurrent reuse of the same token.
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("token_hash = ?", tokenHash).First(&pr).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFound
		}
		if err != nil {
			return err
		}
		if pr.Used || now.After(pr.ExpiresAt) {
			return ErrNotFound
		}

		if err := tx.Model(&entities.User{}).
			Where("id = ?", pr.UserID).
			Update("password_hash", passwordHash).Error; err != nil {
			return err
		}
		if err := tx.Model(&entities.PasswordReset{}).
			Where("id = ?", pr.ID).Update("used", true).Error; err != nil {
			return err
		}
		return tx.Model(&entities.RefreshToken{}).
			Where("user_id = ? AND revoked = false", pr.UserID).
			Update("revoked", true).Error
	})
}
