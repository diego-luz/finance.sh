package repositories

import (
	"errors"

	"github.com/finance-sh/finance-sh/internal/entities"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CreditCardRepository persists credit cards, always scoped by organization_id.
type CreditCardRepository struct{ db *gorm.DB }

func NewCreditCardRepository(db *gorm.DB) *CreditCardRepository {
	return &CreditCardRepository{db: db}
}

func (r *CreditCardRepository) Create(c *entities.CreditCard) error {
	return r.db.Create(c).Error
}

func (r *CreditCardRepository) Update(c *entities.CreditCard) error {
	return r.db.Save(c).Error
}

func (r *CreditCardRepository) Delete(orgID, id uuid.UUID) error {
	res := r.db.Where("organization_id = ? AND id = ?", orgID, id).Delete(&entities.CreditCard{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *CreditCardRepository) FindByID(orgID, id uuid.UUID) (*entities.CreditCard, error) {
	var c entities.CreditCard
	err := r.db.Where("organization_id = ? AND id = ?", orgID, id).First(&c).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &c, err
}

func (r *CreditCardRepository) List(orgID uuid.UUID) ([]entities.CreditCard, error) {
	var cards []entities.CreditCard
	err := r.db.Where("organization_id = ?", orgID).Order("created_at asc").Find(&cards).Error
	return cards, err
}

// SearchByName returns the org's credit cards whose name matches the query
// (ILIKE), ordered by name, capped at limit. Used by global search.
func (r *CreditCardRepository) SearchByName(orgID uuid.UUID, q string, limit int) ([]entities.CreditCard, error) {
	var cards []entities.CreditCard
	err := r.db.Where("organization_id = ? AND name ILIKE ?", orgID, "%"+q+"%").
		Order("name asc").
		Limit(limit).
		Find(&cards).Error
	return cards, err
}

// UsedAmounts returns, per credit card, the sum of UNPAID expense transactions
// booked against it (the open invoice). Returns a map credit_card_id -> used.
func (r *CreditCardRepository) UsedAmounts(orgID uuid.UUID) (map[uuid.UUID]int64, error) {
	type row struct {
		CreditCardID uuid.UUID
		Total        int64
	}
	var rows []row
	err := r.db.Model(&entities.Transaction{}).
		Select("credit_card_id, COALESCE(SUM(amount),0) as total").
		Where("organization_id = ? AND paid = ? AND type = ? AND credit_card_id IS NOT NULL",
			orgID, false, entities.TxExpense).
		Group("credit_card_id").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	used := make(map[uuid.UUID]int64, len(rows))
	for _, rw := range rows {
		used[rw.CreditCardID] = rw.Total
	}
	return used, nil
}
