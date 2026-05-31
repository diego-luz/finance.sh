package repositories

import (
	"errors"

	"github.com/finance-sh/finance-sh/internal/entities"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ContactRepository persists contacts (clientes/fornecedores), always scoped by
// organization_id.
type ContactRepository struct{ db *gorm.DB }

func NewContactRepository(db *gorm.DB) *ContactRepository { return &ContactRepository{db: db} }

func (r *ContactRepository) Create(c *entities.Contact) error {
	return r.db.Create(c).Error
}

func (r *ContactRepository) Update(c *entities.Contact) error {
	return r.db.Save(c).Error
}

func (r *ContactRepository) Delete(orgID, id uuid.UUID) error {
	res := r.db.Where("organization_id = ? AND id = ?", orgID, id).Delete(&entities.Contact{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *ContactRepository) FindByID(orgID, id uuid.UUID) (*entities.Contact, error) {
	var c entities.Contact
	err := r.db.Where("organization_id = ? AND id = ?", orgID, id).First(&c).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &c, err
}

func (r *ContactRepository) List(orgID uuid.UUID) ([]entities.Contact, error) {
	var contacts []entities.Contact
	err := r.db.Where("organization_id = ?", orgID).Order("name asc").Find(&contacts).Error
	return contacts, err
}

// Search returns the org's contacts whose name, document or email match the
// query (ILIKE), ordered by name, capped at limit. Used by global search.
func (r *ContactRepository) Search(orgID uuid.UUID, q string, limit int) ([]entities.Contact, error) {
	var contacts []entities.Contact
	like := "%" + q + "%"
	err := r.db.Where("organization_id = ? AND (name ILIKE ? OR document ILIKE ? OR email ILIKE ?)",
		orgID, like, like, like).
		Order("name asc").
		Limit(limit).
		Find(&contacts).Error
	return contacts, err
}
