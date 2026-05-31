package services

import (
	"context"
	"fmt"
	"time"

	"github.com/finance-sh/finance-sh/internal/dto"
	"github.com/finance-sh/finance-sh/internal/repositories"
	"github.com/finance-sh/finance-sh/pkg/cache"
	"github.com/google/uuid"
)

const (
	cashFlowMonths     = 6
	topCategoriesLimit = 5
	upcomingBillsLimit = 5
	dashboardCacheTTL  = 60 * time.Second
)

// DashboardCacheKey returns the cache key holding the dashboard payload for an
// org. Shared with the services that must invalidate it on mutation.
func DashboardCacheKey(orgID uuid.UUID) string {
	return fmt.Sprintf("dashboard:%s", orgID.String())
}

type DashboardService struct {
	txs      *repositories.TransactionRepository
	accounts *repositories.AccountRepository
	cache    *cache.Cache
}

func NewDashboardService(
	txs *repositories.TransactionRepository,
	accounts *repositories.AccountRepository,
	c *cache.Cache,
) *DashboardService {
	return &DashboardService{txs: txs, accounts: accounts, cache: c}
}

// Overview composes the repo aggregations into a single dashboard payload. The
// result is cached per org for a short TTL in process-local memory.
func (s *DashboardService) Overview(orgID uuid.UUID) (*dto.DashboardResponse, error) {
	ctx := context.Background()
	key := DashboardCacheKey(orgID)

	var cached dto.DashboardResponse
	if s.cache.GetJSON(ctx, key, &cached) {
		return &cached, nil
	}

	now := time.Now().UTC()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	monthEnd := monthStart.AddDate(0, 1, 0)

	balances, err := s.accounts.Balances(orgID)
	if err != nil {
		return nil, err
	}
	accts, err := s.accounts.List(orgID)
	if err != nil {
		return nil, err
	}

	var totalBalance int64
	summary := make([]dto.AccountDTO, 0, len(accts))
	for i := range accts {
		bal := balances[accts[i].ID]
		totalBalance += bal
		summary = append(summary, accountDTO(&accts[i], bal))
	}

	income, expense, err := s.txs.MonthTotals(orgID, now)
	if err != nil {
		return nil, err
	}

	cashFlow, err := s.txs.CashFlow(orgID, cashFlowMonths)
	if err != nil {
		return nil, err
	}

	topCats, err := s.txs.TopCategories(orgID, monthStart, monthEnd, topCategoriesLimit)
	if err != nil {
		return nil, err
	}

	bills, err := s.txs.UpcomingBills(orgID, upcomingBillsLimit)
	if err != nil {
		return nil, err
	}
	upcoming := make([]dto.TransactionDTO, 0, len(bills))
	for i := range bills {
		upcoming = append(upcoming, transactionDTO(&bills[i]))
	}

	// Ensure all slices are non-nil so the JSON renders `[]` instead of `null`
	// (a brand-new org with no transactions used to crash the frontend, which
	// does .length on top_categories).
	if cashFlow == nil {
		cashFlow = []dto.CashFlowPoint{}
	}
	if topCats == nil {
		topCats = []dto.CategorySpend{}
	}
	if summary == nil {
		summary = []dto.AccountDTO{}
	}
	resp := &dto.DashboardResponse{
		Balance:         totalBalance,
		MonthIncome:     income,
		MonthExpense:    expense,
		MonthNet:        income - expense,
		UpcomingBills:   upcoming,
		CashFlow:        cashFlow,
		TopCategories:   topCats,
		AccountsSummary: summary,
	}
	s.cache.SetJSON(ctx, key, resp, dashboardCacheTTL)
	return resp, nil
}
