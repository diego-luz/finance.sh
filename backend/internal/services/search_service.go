package services

import (
	"strings"

	"github.com/finance-sh/finance-sh/internal/dto"
	"github.com/finance-sh/finance-sh/internal/repositories"
	"github.com/google/uuid"
)

// SearchService performs an org-scoped global search across the main domains.
// Each group is matched case-insensitively (ILIKE %q%) and capped at a limit.
type SearchService struct {
	txs         *repositories.TransactionRepository
	contacts    *repositories.ContactRepository
	categories  *repositories.CategoryRepository
	accounts    *repositories.AccountRepository
	creditCards *repositories.CreditCardRepository
	goals       *repositories.GoalRepository
}

func NewSearchService(
	txs *repositories.TransactionRepository,
	contacts *repositories.ContactRepository,
	categories *repositories.CategoryRepository,
	accounts *repositories.AccountRepository,
	creditCards *repositories.CreditCardRepository,
	goals *repositories.GoalRepository,
) *SearchService {
	return &SearchService{
		txs:         txs,
		contacts:    contacts,
		categories:  categories,
		accounts:    accounts,
		creditCards: creditCards,
		goals:       goals,
	}
}

// Search runs the org-scoped query across every domain. The limit is clamped to
// 1..20 (default 5 when <= 0). An empty/blank query returns empty groups.
func (s *SearchService) Search(orgID uuid.UUID, q string, limit int) (dto.SearchResult, error) {
	q = strings.TrimSpace(q)
	res := dto.SearchResult{
		Query: q,
		Results: dto.SearchGroups{
			Transactions: []dto.TransactionDTO{},
			Contacts:     []dto.ContactDTO{},
			Categories:   []dto.CategoryDTO{},
			Accounts:     []dto.AccountDTO{},
			CreditCards:  []dto.SearchNamedDTO{},
			Goals:        []dto.SearchNamedDTO{},
		},
	}
	if q == "" {
		return res, nil
	}

	switch {
	case limit <= 0:
		limit = 5
	case limit > 20:
		limit = 20
	}

	// Transactions (description).
	txRows, err := s.txs.SearchByDescription(orgID, q, limit)
	if err != nil {
		return dto.SearchResult{}, err
	}
	for i := range txRows {
		res.Results.Transactions = append(res.Results.Transactions, transactionDTO(&txRows[i]))
	}

	// Contacts (name/document/email).
	contactRows, err := s.contacts.Search(orgID, q, limit)
	if err != nil {
		return dto.SearchResult{}, err
	}
	for i := range contactRows {
		res.Results.Contacts = append(res.Results.Contacts, contactDTO(&contactRows[i]))
	}

	// Categories (name).
	catRows, err := s.categories.SearchByName(orgID, q, limit)
	if err != nil {
		return dto.SearchResult{}, err
	}
	for i := range catRows {
		res.Results.Categories = append(res.Results.Categories, categoryDTO(&catRows[i]))
	}

	// Accounts (name). Minimal shape: the computed balance is intentionally
	// omitted (0) to avoid an extra aggregation per search.
	acctRows, err := s.accounts.SearchByName(orgID, q, limit)
	if err != nil {
		return dto.SearchResult{}, err
	}
	for i := range acctRows {
		res.Results.Accounts = append(res.Results.Accounts, accountDTO(&acctRows[i], 0))
	}

	// Credit cards (name → minimal id/name).
	cardRows, err := s.creditCards.SearchByName(orgID, q, limit)
	if err != nil {
		return dto.SearchResult{}, err
	}
	for i := range cardRows {
		res.Results.CreditCards = append(res.Results.CreditCards, dto.SearchNamedDTO{
			ID:   cardRows[i].ID.String(),
			Name: cardRows[i].Name,
		})
	}

	// Goals (name → minimal id/name).
	goalRows, err := s.goals.SearchByName(orgID, q, limit)
	if err != nil {
		return dto.SearchResult{}, err
	}
	for i := range goalRows {
		res.Results.Goals = append(res.Results.Goals, dto.SearchNamedDTO{
			ID:   goalRows[i].ID.String(),
			Name: goalRows[i].Name,
		})
	}

	res.Total = len(res.Results.Transactions) +
		len(res.Results.Contacts) +
		len(res.Results.Categories) +
		len(res.Results.Accounts) +
		len(res.Results.CreditCards) +
		len(res.Results.Goals)
	return res, nil
}
