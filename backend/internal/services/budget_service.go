package services

import (
	"errors"
	"time"

	"github.com/finance-sh/finance-sh/internal/dto"
	"github.com/finance-sh/finance-sh/internal/entities"
	"github.com/finance-sh/finance-sh/internal/repositories"
	"github.com/google/uuid"
)

// ErrCategoryNotExpense is returned when a budget references a category that is
// not an expense category (budgets only cap spending).
var ErrCategoryNotExpense = errors.New("categoria de orçamento deve ser de despesa")

type BudgetService struct {
	budgets    *repositories.BudgetRepository
	categories *repositories.CategoryRepository
}

func NewBudgetService(
	budgets *repositories.BudgetRepository,
	categories *repositories.CategoryRepository,
) *BudgetService {
	return &BudgetService{budgets: budgets, categories: categories}
}

// List returns the org budgets for the given month/year (defaulting to the
// current month when either is zero), each with computed spent/percent.
func (s *BudgetService) List(orgID uuid.UUID, month, year int) ([]dto.BudgetDTO, error) {
	if month < 1 || month > 12 || year == 0 {
		now := time.Now().UTC()
		month, year = int(now.Month()), now.Year()
	}
	budgets, err := s.budgets.ListByPeriod(orgID, month, year)
	if err != nil {
		return nil, err
	}
	out := make([]dto.BudgetDTO, 0, len(budgets))
	for i := range budgets {
		d, err := s.toDTO(orgID, &budgets[i])
		if err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, nil
}

func (s *BudgetService) Create(orgID uuid.UUID, req dto.BudgetRequest) (*dto.BudgetDTO, error) {
	catID, err := s.validateCategory(orgID, req.CategoryID)
	if err != nil {
		return nil, err
	}
	b := &entities.Budget{
		OrganizationID: orgID,
		CategoryID:     catID,
		Amount:         req.Amount,
		Month:          req.Month,
		Year:           req.Year,
	}
	if err := s.budgets.Create(b); err != nil {
		return nil, err
	}
	d, err := s.toDTO(orgID, b)
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func (s *BudgetService) Update(orgID, id uuid.UUID, req dto.BudgetRequest) (*dto.BudgetDTO, error) {
	b, err := s.budgets.FindByID(orgID, id)
	if err != nil {
		return nil, err
	}
	catID, err := s.validateCategory(orgID, req.CategoryID)
	if err != nil {
		return nil, err
	}
	b.CategoryID = catID
	b.Amount = req.Amount
	b.Month = req.Month
	b.Year = req.Year
	if err := s.budgets.Update(b); err != nil {
		return nil, err
	}
	d, err := s.toDTO(orgID, b)
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func (s *BudgetService) Delete(orgID, id uuid.UUID) error {
	return s.budgets.Delete(orgID, id)
}

// validateCategory ensures the category belongs to the org and is an expense.
func (s *BudgetService) validateCategory(orgID uuid.UUID, raw string) (uuid.UUID, error) {
	catID, err := parseUUID(raw)
	if err != nil {
		return uuid.Nil, err
	}
	cat, err := s.categories.FindByID(orgID, catID)
	if err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			return uuid.Nil, ErrNotFound
		}
		return uuid.Nil, err
	}
	if cat.Kind != entities.CategoryExpense {
		return uuid.Nil, ErrCategoryNotExpense
	}
	return catID, nil
}

func (s *BudgetService) toDTO(orgID uuid.UUID, b *entities.Budget) (dto.BudgetDTO, error) {
	spent, err := s.budgets.SpentByCategory(orgID, b.CategoryID, b.Month, b.Year)
	if err != nil {
		return dto.BudgetDTO{}, err
	}
	var percent float64
	if b.Amount > 0 {
		percent = float64(spent) / float64(b.Amount)
	}
	d := dto.BudgetDTO{
		ID:         b.ID.String(),
		CategoryID: b.CategoryID.String(),
		Amount:     b.Amount,
		Month:      b.Month,
		Year:       b.Year,
		Spent:      spent,
		Percent:    percent,
	}
	// Embed the category for display (best-effort; ignore lookup failures).
	if cat, err := s.categories.FindByID(orgID, b.CategoryID); err == nil {
		c := categoryDTO(cat)
		d.Category = &c
	}
	return d, nil
}
