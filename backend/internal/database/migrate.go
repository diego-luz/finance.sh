package database

import (
	"github.com/finance-sh/finance-sh/internal/entities"
	"gorm.io/gorm"
)

// AutoMigrate creates/updates the schema for every entity. UUID generation is
// handled in Go (Base.BeforeCreate), but pgcrypto is enabled defensively in case
// a future migration relies on gen_random_uuid().
func AutoMigrate(db *gorm.DB) error {
	if err := db.Exec(`CREATE EXTENSION IF NOT EXISTS "pgcrypto"`).Error; err != nil {
		return err
	}

	return db.AutoMigrate(
		&entities.User{},
		&entities.RefreshToken{},
		&entities.PasswordReset{},
		&entities.EmailVerification{},
		&entities.RecoveryCode{},
		&entities.Organization{},
		&entities.Membership{},
		&entities.Invitation{},
		&entities.Account{},
		&entities.Category{},
		&entities.CategoryRule{},
		&entities.Contact{},
		&entities.Tag{},
		// Transaction declares the many2many to Tag; AutoMigrate creates the
		// implicit join table transaction_tags (transaction_id + tag_id).
		&entities.Transaction{},
		&entities.Attachment{},
		&entities.CreditCard{},
		&entities.Budget{},
		&entities.Goal{},
		&entities.Notification{},
		&entities.AuditLog{},
		// RecurrenceRule: the proper recurring-transaction engine. Migration 000008
		// (written by the main thread) is the production source of truth for this
		// table; it is added here so AUTO_MIGRATE=true picks it up in dev.
		&entities.RecurrenceRule{},
	)
}
