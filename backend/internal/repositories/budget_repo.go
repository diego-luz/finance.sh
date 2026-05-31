package repositories

import (
	"errors"
	"time"

	"github.com/finance-sh/finance-sh/internal/entities"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BudgetRepository persists monthly category budgets, scoped by organization_id.
type BudgetRepository struct{ db *gorm.DB }

func NewBudgetRepository(db *gorm.DB) *BudgetRepository { return &BudgetRepository{db: db} }

func (r *BudgetRepository) Create(b *entities.Budget) error {
	return r.db.Create(b).Error
}

func (r *BudgetRepository) Update(b *entities.Budget) error {
	return r.db.Save(b).Error
}

func (r *BudgetRepository) Delete(orgID, id uuid.UUID) error {
	res := r.db.Where("organization_id = ? AND id = ?", orgID, id).Delete(&entities.Budget{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *BudgetRepository) FindByID(orgID, id uuid.UUID) (*entities.Budget, error) {
	var b entities.Budget
	err := r.db.Where("organization_id = ? AND id = ?", orgID, id).First(&b).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &b, err
}

// ListByPeriod returns the org's budgets for a given month/year.
func (r *BudgetRepository) ListByPeriod(orgID uuid.UUID, month, year int) ([]entities.Budget, error) {
	var budgets []entities.Budget
	err := r.db.Where("organization_id = ? AND month = ? AND year = ?", orgID, month, year).
		Order("created_at asc").Find(&budgets).Error
	return budgets, err
}

// ListAll returns every budget for the org (used by the worker for overflow
// notifications).
func (r *BudgetRepository) ListAll(orgID uuid.UUID) ([]entities.Budget, error) {
	var budgets []entities.Budget
	err := r.db.Where("organization_id = ?", orgID).Find(&budgets).Error
	return budgets, err
}

// SpentByCategory returns the total PAID expense for a category within the
// given month/year (in cents).
func (r *BudgetRepository) SpentByCategory(orgID, categoryID uuid.UUID, month, year int) (int64, error) {
	start := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 1, 0)
	var total int64
	err := r.db.Model(&entities.Transaction{}).
		Select("COALESCE(SUM(amount),0)").
		Where("organization_id = ? AND paid = ? AND type = ? AND category_id = ? AND date >= ? AND date < ?",
			orgID, true, entities.TxExpense, categoryID, start, end).
		Scan(&total).Error
	return total, err
}
