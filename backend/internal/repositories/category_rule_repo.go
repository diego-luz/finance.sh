package repositories

import (
	"errors"

	"github.com/finance-sh/finance-sh/internal/entities"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CategoryRuleRepository persists the keyword->category rules used by automatic
// categorization. Every query is scoped by organization_id.
type CategoryRuleRepository struct{ db *gorm.DB }

func NewCategoryRuleRepository(db *gorm.DB) *CategoryRuleRepository {
	return &CategoryRuleRepository{db: db}
}

func (r *CategoryRuleRepository) Create(rule *entities.CategoryRule) error {
	return r.db.Create(rule).Error
}

func (r *CategoryRuleRepository) Update(rule *entities.CategoryRule) error {
	return r.db.Save(rule).Error
}

func (r *CategoryRuleRepository) Delete(orgID, id uuid.UUID) error {
	res := r.db.Where("organization_id = ? AND id = ?", orgID, id).Delete(&entities.CategoryRule{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *CategoryRuleRepository) FindByID(orgID, id uuid.UUID) (*entities.CategoryRule, error) {
	var rule entities.CategoryRule
	err := r.db.Preload("Category").
		Where("organization_id = ? AND id = ?", orgID, id).First(&rule).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &rule, err
}

// List returns every rule for the org with its Category preloaded, highest
// priority first (ties broken by newest first for a stable order).
func (r *CategoryRuleRepository) List(orgID uuid.UUID) ([]entities.CategoryRule, error) {
	var rules []entities.CategoryRule
	err := r.db.Preload("Category").
		Where("organization_id = ?", orgID).
		Order("priority desc, created_at desc").
		Find(&rules).Error
	return rules, err
}

// ListActive returns only the active rules for the org, Category preloaded,
// highest priority first. Used by the matching engine.
func (r *CategoryRuleRepository) ListActive(orgID uuid.UUID) ([]entities.CategoryRule, error) {
	var rules []entities.CategoryRule
	err := r.db.Preload("Category").
		Where("organization_id = ? AND active = ?", orgID, true).
		Order("priority desc, created_at desc").
		Find(&rules).Error
	return rules, err
}
