package services

import (
	"context"
	"errors"

	"github.com/finance-sh/finance-sh/internal/dto"
	"github.com/finance-sh/finance-sh/internal/entities"
	"github.com/finance-sh/finance-sh/internal/repositories"
	"github.com/finance-sh/finance-sh/pkg/cache"
	"github.com/google/uuid"
)

// ErrNotFound is the service-level not-found signal (re-exported for handlers).
var ErrNotFound = repositories.ErrNotFound

type AccountService struct {
	accounts *repositories.AccountRepository
	cache    *cache.Cache
}

func NewAccountService(accounts *repositories.AccountRepository, c *cache.Cache) *AccountService {
	return &AccountService{accounts: accounts, cache: c}
}

// invalidateDashboard drops the org's cached dashboard after a mutation that
// changes balances. Best-effort.
func (s *AccountService) invalidateDashboard(orgID uuid.UUID) {
	s.cache.Delete(context.Background(), DashboardCacheKey(orgID))
}

// List returns all accounts of the org with computed balances.
func (s *AccountService) List(orgID uuid.UUID) ([]dto.AccountDTO, error) {
	accts, err := s.accounts.List(orgID)
	if err != nil {
		return nil, err
	}
	balances, err := s.accounts.Balances(orgID)
	if err != nil {
		return nil, err
	}
	out := make([]dto.AccountDTO, 0, len(accts))
	for i := range accts {
		out = append(out, accountDTO(&accts[i], balances[accts[i].ID]))
	}
	return out, nil
}

func (s *AccountService) Get(orgID, id uuid.UUID) (*dto.AccountDTO, error) {
	a, err := s.accounts.FindByID(orgID, id)
	if err != nil {
		return nil, err
	}
	bal, err := s.accounts.Balance(orgID, id)
	if err != nil {
		return nil, err
	}
	d := accountDTO(a, bal)
	return &d, nil
}

func (s *AccountService) Create(orgID uuid.UUID, req dto.AccountRequest) (*dto.AccountDTO, error) {
	a := &entities.Account{
		OrganizationID: orgID,
		Name:           req.Name,
		Type:           entities.AccountType(req.Type),
		InitialBalance: req.InitialBalance,
		Color:          defaultStr(req.Color, "#10b981"),
		Icon:           defaultStr(req.Icon, "wallet"),
	}
	if err := s.accounts.Create(a); err != nil {
		return nil, err
	}
	s.invalidateDashboard(orgID)
	d := accountDTO(a, a.InitialBalance)
	return &d, nil
}

func (s *AccountService) Update(orgID, id uuid.UUID, req dto.AccountRequest) (*dto.AccountDTO, error) {
	a, err := s.accounts.FindByID(orgID, id)
	if err != nil {
		return nil, err
	}
	a.Name = req.Name
	a.Type = entities.AccountType(req.Type)
	a.InitialBalance = req.InitialBalance
	if req.Color != "" {
		a.Color = req.Color
	}
	if req.Icon != "" {
		a.Icon = req.Icon
	}
	if err := s.accounts.Update(a); err != nil {
		return nil, err
	}
	s.invalidateDashboard(orgID)
	bal, err := s.accounts.Balance(orgID, id)
	if err != nil {
		return nil, err
	}
	d := accountDTO(a, bal)
	return &d, nil
}

func (s *AccountService) Delete(orgID, id uuid.UUID) error {
	if err := s.accounts.Delete(orgID, id); err != nil {
		return err
	}
	s.invalidateDashboard(orgID)
	return nil
}

func accountDTO(a *entities.Account, balance int64) dto.AccountDTO {
	return dto.AccountDTO{
		ID:             a.ID.String(),
		Name:           a.Name,
		Type:           string(a.Type),
		InitialBalance: a.InitialBalance,
		Balance:        balance,
		Color:          a.Color,
		Icon:           a.Icon,
		Archived:       a.Archived,
	}
}

func defaultStr(v, fallback string) string {
	if v == "" {
		return fallback
	}
	return v
}

// parseUUID is a small shared helper for services that accept string IDs.
func parseUUID(s string) (uuid.UUID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil, errors.New("identificador inválido")
	}
	return id, nil
}

// parseUUIDs parses a slice of string IDs, returning an error on the first
// invalid value. Used by the bulk actions.
func parseUUIDs(in []string) ([]uuid.UUID, error) {
	out := make([]uuid.UUID, 0, len(in))
	for _, s := range in {
		id, err := parseUUID(s)
		if err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, nil
}
