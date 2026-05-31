package repositories

import (
	"errors"

	"github.com/finance-sh/finance-sh/internal/entities"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AccountRepository persists accounts. Every method is scoped by
// organization_id so one tenant can never read or mutate another tenant's data.
type AccountRepository struct{ db *gorm.DB }

func NewAccountRepository(db *gorm.DB) *AccountRepository { return &AccountRepository{db: db} }

func (r *AccountRepository) Create(a *entities.Account) error {
	return r.db.Create(a).Error
}

func (r *AccountRepository) Update(a *entities.Account) error {
	return r.db.Save(a).Error
}

func (r *AccountRepository) Delete(orgID, id uuid.UUID) error {
	res := r.db.Where("organization_id = ? AND id = ?", orgID, id).Delete(&entities.Account{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *AccountRepository) FindByID(orgID, id uuid.UUID) (*entities.Account, error) {
	var a entities.Account
	err := r.db.Where("organization_id = ? AND id = ?", orgID, id).First(&a).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &a, err
}

func (r *AccountRepository) List(orgID uuid.UUID) ([]entities.Account, error) {
	var accts []entities.Account
	err := r.db.Where("organization_id = ?", orgID).Order("created_at asc").Find(&accts).Error
	return accts, err
}

// SearchByName returns the org's accounts whose name matches the query (ILIKE),
// ordered by name, capped at limit. Used by global search.
func (r *AccountRepository) SearchByName(orgID uuid.UUID, q string, limit int) ([]entities.Account, error) {
	var accts []entities.Account
	err := r.db.Where("organization_id = ? AND name ILIKE ?", orgID, "%"+q+"%").
		Order("name asc").
		Limit(limit).
		Find(&accts).Error
	return accts, err
}

// Balances computes the current balance (in cents) for every account of the
// organization. The balance is the initial balance plus all settled (paid)
// movements: incoming income/transfers add, outgoing expense/transfers subtract.
// Returns a map account_id -> balance.
func (r *AccountRepository) Balances(orgID uuid.UUID) (map[uuid.UUID]int64, error) {
	accts, err := r.List(orgID)
	if err != nil {
		return nil, err
	}

	balances := make(map[uuid.UUID]int64, len(accts))
	for _, a := range accts {
		balances[a.ID] = a.InitialBalance
	}

	type movement struct {
		AccountID uuid.UUID
		Type      entities.TransactionType
		Total     int64
	}

	// Money entering or leaving the account it is booked against.
	var rows []movement
	err = r.db.Model(&entities.Transaction{}).
		Select("account_id, type, COALESCE(SUM(amount),0) as total").
		Where("organization_id = ? AND paid = ?", orgID, true).
		Group("account_id, type").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	for _, m := range rows {
		switch m.Type {
		case entities.TxIncome:
			balances[m.AccountID] += m.Total
		case entities.TxExpense, entities.TxTransfer:
			// A transfer leaves the source account (account_id).
			balances[m.AccountID] -= m.Total
		}
	}

	// Transfers also credit the destination account.
	var inbound []movement
	err = r.db.Model(&entities.Transaction{}).
		Select("transfer_account_id as account_id, COALESCE(SUM(amount),0) as total").
		Where("organization_id = ? AND paid = ? AND type = ? AND transfer_account_id IS NOT NULL",
			orgID, true, entities.TxTransfer).
		Group("transfer_account_id").
		Scan(&inbound).Error
	if err != nil {
		return nil, err
	}
	for _, m := range inbound {
		balances[m.AccountID] += m.Total
	}

	return balances, nil
}

// Balance returns the computed balance for a single account.
func (r *AccountRepository) Balance(orgID, id uuid.UUID) (int64, error) {
	balances, err := r.Balances(orgID)
	if err != nil {
		return 0, err
	}
	return balances[id], nil
}
