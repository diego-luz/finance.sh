package repositories

import (
	"errors"

	"github.com/finance-sh/finance-sh/internal/entities"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// OrganizationRepository persists organization rows. All reads/writes are
// addressed by the organization id resolved from the active tenant.
type OrganizationRepository struct{ db *gorm.DB }

func NewOrganizationRepository(db *gorm.DB) *OrganizationRepository {
	return &OrganizationRepository{db: db}
}

// FindByID loads an organization by id.
func (r *OrganizationRepository) FindByID(id uuid.UUID) (*entities.Organization, error) {
	var o entities.Organization
	err := r.db.First(&o, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &o, err
}

// CreateForUser creates a new organization owned by userID plus the owner
// membership, in a single transaction. Used by the self-service "create another
// organization" flow (e.g. a personal org + a microempresa org).
func (r *OrganizationRepository) CreateForUser(org *entities.Organization, userID uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		org.OwnerID = userID
		if err := tx.Create(org).Error; err != nil {
			return err
		}
		return tx.Create(&entities.Membership{
			UserID: userID, OrganizationID: org.ID, Role: entities.RoleOwner,
		}).Error
	})
}

// Update persists the given mutable fields (name/currency) for the org. Only
// the supplied columns are written; the caller decides which ones changed.
func (r *OrganizationRepository) Update(id uuid.UUID, fields map[string]interface{}) error {
	if len(fields) == 0 {
		return nil
	}
	res := r.db.Model(&entities.Organization{}).Where("id = ?", id).Updates(fields)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}
