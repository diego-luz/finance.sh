package repositories

import (
	"errors"

	"github.com/finance-sh/finance-sh/internal/entities"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// GoalRepository persists savings goals, always scoped by organization_id.
type GoalRepository struct{ db *gorm.DB }

func NewGoalRepository(db *gorm.DB) *GoalRepository { return &GoalRepository{db: db} }

func (r *GoalRepository) Create(g *entities.Goal) error {
	return r.db.Create(g).Error
}

func (r *GoalRepository) Update(g *entities.Goal) error {
	return r.db.Save(g).Error
}

func (r *GoalRepository) Delete(orgID, id uuid.UUID) error {
	res := r.db.Where("organization_id = ? AND id = ?", orgID, id).Delete(&entities.Goal{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *GoalRepository) FindByID(orgID, id uuid.UUID) (*entities.Goal, error) {
	var g entities.Goal
	err := r.db.Where("organization_id = ? AND id = ?", orgID, id).First(&g).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &g, err
}

func (r *GoalRepository) List(orgID uuid.UUID) ([]entities.Goal, error) {
	var goals []entities.Goal
	err := r.db.Where("organization_id = ?", orgID).Order("created_at asc").Find(&goals).Error
	return goals, err
}

// SearchByName returns the org's goals whose name matches the query (ILIKE),
// ordered by name, capped at limit. Used by global search.
func (r *GoalRepository) SearchByName(orgID uuid.UUID, q string, limit int) ([]entities.Goal, error) {
	var goals []entities.Goal
	err := r.db.Where("organization_id = ? AND name ILIKE ?", orgID, "%"+q+"%").
		Order("name asc").
		Limit(limit).
		Find(&goals).Error
	return goals, err
}
