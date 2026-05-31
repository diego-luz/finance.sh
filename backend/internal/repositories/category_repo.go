package repositories

import (
	"errors"

	"github.com/finance-sh/finance-sh/internal/entities"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CategoryRepository persists categories, always scoped by organization_id.
type CategoryRepository struct{ db *gorm.DB }

func NewCategoryRepository(db *gorm.DB) *CategoryRepository { return &CategoryRepository{db: db} }

func (r *CategoryRepository) Create(c *entities.Category) error {
	return r.db.Create(c).Error
}

func (r *CategoryRepository) Update(c *entities.Category) error {
	return r.db.Save(c).Error
}

func (r *CategoryRepository) Delete(orgID, id uuid.UUID) error {
	res := r.db.Where("organization_id = ? AND id = ?", orgID, id).Delete(&entities.Category{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *CategoryRepository) FindByID(orgID, id uuid.UUID) (*entities.Category, error) {
	var c entities.Category
	err := r.db.Where("organization_id = ? AND id = ?", orgID, id).First(&c).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &c, err
}

func (r *CategoryRepository) List(orgID uuid.UUID) ([]entities.Category, error) {
	var cats []entities.Category
	err := r.db.Where("organization_id = ?", orgID).Order("kind asc, name asc").Find(&cats).Error
	return cats, err
}

// SearchByName returns the org's categories whose name matches the query
// (ILIKE), ordered by name, capped at limit. Used by global search.
func (r *CategoryRepository) SearchByName(orgID uuid.UUID, q string, limit int) ([]entities.Category, error) {
	var cats []entities.Category
	err := r.db.Where("organization_id = ? AND name ILIKE ?", orgID, "%"+q+"%").
		Order("name asc").
		Limit(limit).
		Find(&cats).Error
	return cats, err
}
