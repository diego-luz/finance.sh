package services

import (
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/finance-sh/finance-sh/internal/entities"
	"github.com/finance-sh/finance-sh/internal/repositories"
	"github.com/finance-sh/finance-sh/pkg/hash"
	"github.com/finance-sh/finance-sh/pkg/logger"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// LGPD-related typed errors mapped by the handler layer.
var (
	ErrWrongPassword = errors.New("senha incorreta")
	// ErrOwnedOrgHasMembers blocks account deletion when the user is the sole
	// owner of an organization that still has other active members. Deleting would
	// silently destroy data belonging to those members. Mapped to HTTP 409.
	ErrOwnedOrgHasMembers = errors.New("você é o único proprietário de uma organização com outros membros. Transfira a propriedade ou remova os demais membros antes de excluir sua conta")
)

// LGPDService implements the data-subject rights endpoints (export + deletion)
// required by the LGPD: the right to data portability and the right to erasure.
type LGPDService struct {
	db    *gorm.DB
	users *repositories.UserRepository
}

func NewLGPDService(db *gorm.DB, users *repositories.UserRepository) *LGPDService {
	return &LGPDService{db: db, users: users}
}

// ExportData assembles the full data set the user can access across every
// organization they belong to, plus their own profile and personal audit logs.
// The result is a plain map so it serialises to a self-describing JSON document.
func (s *LGPDService) ExportData(userID uuid.UUID) (map[string]interface{}, error) {
	user, err := s.users.FindByID(userID)
	if err != nil {
		return nil, ErrUserNotFound
	}
	memberships, err := s.users.Memberships(userID)
	if err != nil {
		return nil, err
	}

	orgIDs := make([]uuid.UUID, 0, len(memberships))
	orgsOut := make([]map[string]interface{}, 0, len(memberships))
	membershipsOut := make([]map[string]interface{}, 0, len(memberships))
	for _, m := range memberships {
		orgIDs = append(orgIDs, m.OrganizationID)
		membershipsOut = append(membershipsOut, map[string]interface{}{
			"organization_id": m.OrganizationID,
			"role":            m.Role,
			"joined_at":       m.CreatedAt,
		})
		if m.Organization != nil {
			orgsOut = append(orgsOut, map[string]interface{}{
				"id":       m.Organization.ID,
				"name":     m.Organization.Name,
				"slug":     m.Organization.Slug,
				"currency": m.Organization.Currency,
				"owner_id": m.Organization.OwnerID,
			})
		}
	}

	doc := map[string]interface{}{
		"export_version":    "1.0",
		"generated_at":      time.Now().UTC(),
		"data_subject_lgpd": "Documento de portabilidade de dados (LGPD)",
		"user": map[string]interface{}{
			"id":                user.ID,
			"name":              user.Name,
			"email":             user.Email,
			"email_verified":    user.EmailVerified,
			"avatar_url":        user.AvatarURL,
			"terms_accepted_at": user.TermsAcceptedAt,
			"terms_version":     user.TermsVersion,
			"two_factor":        user.TwoFactorEnabled,
			"created_at":        user.CreatedAt,
		},
		"organizations": orgsOut,
		"memberships":   membershipsOut,
	}

	// Financial data is scoped to the orgs the user belongs to.
	if len(orgIDs) > 0 {
		doc["accounts"] = s.fetch(&entities.Account{}, orgIDs)
		doc["categories"] = s.fetch(&entities.Category{}, orgIDs)
		doc["contacts"] = s.fetch(&entities.Contact{}, orgIDs)
		doc["transactions"] = s.fetchTransactions(orgIDs)
		doc["credit_cards"] = s.fetch(&entities.CreditCard{}, orgIDs)
		doc["goals"] = s.fetch(&entities.Goal{}, orgIDs)
		doc["budgets"] = s.fetch(&entities.Budget{}, orgIDs)
		doc["notifications"] = s.fetch(&entities.Notification{}, orgIDs)
	}

	// Audit logs attributable to this user.
	var audits []entities.AuditLog
	s.db.Where("user_id = ?", userID).Order("created_at desc").Find(&audits)
	doc["audit_logs"] = audits

	slog.Info("lgpd export generated", "user_id", userID, "email", logger.MaskEmail(user.Email), "orgs", len(orgIDs))
	return doc, nil
}

// fetch loads all rows of a model scoped to the given organizations.
func (s *LGPDService) fetch(model interface{}, orgIDs []uuid.UUID) interface{} {
	var rows []map[string]interface{}
	s.db.Model(model).Where("organization_id IN ?", orgIDs).Find(&rows)
	return rows
}

// fetchTransactions loads transactions via the entity so EncryptedString Notes
// are decrypted before export.
func (s *LGPDService) fetchTransactions(orgIDs []uuid.UUID) interface{} {
	var rows []entities.Transaction
	s.db.Where("organization_id IN ?", orgIDs).Find(&rows)
	out := make([]map[string]interface{}, 0, len(rows))
	for i := range rows {
		t := &rows[i]
		out = append(out, map[string]interface{}{
			"id":              t.ID,
			"organization_id": t.OrganizationID,
			"account_id":      t.AccountID,
			"category_id":     t.CategoryID,
			"contact_id":      t.ContactID,
			"type":            t.Type,
			"amount":          t.Amount,
			"description":     t.Description,
			"date":            t.Date,
			"due_date":        t.DueDate,
			"paid":            t.Paid,
			"paid_at":         t.PaidAt,
			"recurring":       t.Recurring,
			"notes":           t.Notes.String(),
			"created_at":      t.CreatedAt,
		})
	}
	return out
}

// DeleteAccount erases the user's data per the LGPD right to erasure. After
// verifying the password it: soft-deletes the financial data and the
// organization for every org the user is the SOLE owner of; removes all of the
// user's memberships; anonymises the user record; revokes all refresh tokens;
// and writes an audit log. Everything runs in one transaction.
func (s *LGPDService) DeleteAccount(userID uuid.UUID, password string) error {
	user, err := s.users.FindByID(userID)
	if err != nil {
		return ErrUserNotFound
	}
	if !hash.Check(user.PasswordHash, password) {
		return ErrWrongPassword
	}

	memberships, err := s.users.Memberships(userID)
	if err != nil {
		return err
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		for _, m := range memberships {
			orgID := m.OrganizationID

			// Only purge an org if this user is its sole owner. Otherwise we just
			// drop this user's membership and leave the org intact for others.
			var ownerCount int64
			if err := tx.Model(&entities.Membership{}).
				Where("organization_id = ? AND role = ?", orgID, entities.RoleOwner).
				Count(&ownerCount).Error; err != nil {
				return err
			}
			isSoleOwner := m.Role == entities.RoleOwner && ownerCount <= 1

			if isSoleOwner {
				// Refuse to purge an org that still has OTHER members: doing so would
				// silently destroy their data. The user must transfer ownership or
				// remove the other members first. Returning here rolls back the whole
				// transaction (no partial deletion / anonymisation).
				var memberCount int64
				if err := tx.Model(&entities.Membership{}).
					Where("organization_id = ?", orgID).
					Count(&memberCount).Error; err != nil {
					return err
				}
				if memberCount > 1 {
					return ErrOwnedOrgHasMembers
				}

				// Soft-delete all financial data scoped to the org.
				for _, model := range []interface{}{
					&entities.Transaction{},
					&entities.Budget{},
					&entities.Goal{},
					&entities.CreditCard{},
					&entities.Category{},
					&entities.Contact{},
					&entities.Account{},
					&entities.Notification{},
					&entities.Invitation{},
				} {
					if err := tx.Where("organization_id = ?", orgID).Delete(model).Error; err != nil {
						return err
					}
				}
				// Remove all memberships of this org, then the org itself.
				if err := tx.Where("organization_id = ?", orgID).Delete(&entities.Membership{}).Error; err != nil {
					return err
				}
				if err := tx.Where("id = ?", orgID).Delete(&entities.Organization{}).Error; err != nil {
					return err
				}
			} else {
				// Shared org: drop only this user's membership.
				if err := tx.Where("user_id = ? AND organization_id = ?", userID, orgID).
					Delete(&entities.Membership{}).Error; err != nil {
					return err
				}
			}
		}

		// Anonymise the user (irreversible). A scrambled password hash makes the
		// account unusable; the email is replaced with a non-routable placeholder.
		scrambled, err := hash.RandomToken(32)
		if err != nil {
			return err
		}
		pwHash, err := hash.Password(scrambled)
		if err != nil {
			return err
		}
		anonEmail := fmt.Sprintf("deleted-%s@anonimizado.local", uuid.New().String())
		if err := tx.Model(&entities.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
			"name":               "Usuário removido",
			"email":              anonEmail,
			"password_hash":      pwHash,
			"email_verified":     false,
			"two_factor_enabled": false,
			"two_factor_secret":  "",
			"avatar_url":         "",
		}).Error; err != nil {
			return err
		}

		// Revoke all refresh tokens and remove recovery codes / pending verifications.
		if err := tx.Model(&entities.RefreshToken{}).
			Where("user_id = ? AND revoked = false", userID).
			Update("revoked", true).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id = ?", userID).Delete(&entities.RecoveryCode{}).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id = ?", userID).Delete(&entities.EmailVerification{}).Error; err != nil {
			return err
		}

		// Audit trail of the erasure request (user_id retained for accountability).
		uid := userID
		return tx.Create(&entities.AuditLog{
			UserID:   &uid,
			Action:   "DELETE",
			Entity:   "account",
			EntityID: userID.String(),
			Metadata: `{"reason":"lgpd_erasure"}`,
		}).Error
	})
	if err != nil {
		return err
	}

	slog.Info("lgpd account deleted/anonymised", "user_id", userID)
	return nil
}
