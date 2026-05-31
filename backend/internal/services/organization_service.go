package services

import (
	"errors"
	"log/slog"
	"strings"

	"github.com/finance-sh/finance-sh/internal/database"
	"github.com/finance-sh/finance-sh/internal/dto"
	"github.com/finance-sh/finance-sh/internal/entities"
	"github.com/finance-sh/finance-sh/internal/money"
	"github.com/finance-sh/finance-sh/internal/repositories"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ErrUnsupportedCurrency is returned when a currency code is not in the
// supported ISO-4217 table. Handlers map it to HTTP 422.
var ErrUnsupportedCurrency = errors.New("moeda não suportada")

// ErrInvalidOrgName is returned when a new organization name is blank.
var ErrInvalidOrgName = errors.New("nome da organização é obrigatório")

// OrganizationService reads/updates the active organization and creates new ones
// for the caller. Get/Update are org-scoped (tenant from X-Organization-ID,
// gated by RequireRole); Create is tenant-agnostic (it makes a new tenant).
type OrganizationService struct {
	orgs *repositories.OrganizationRepository
	db   *gorm.DB
}

func NewOrganizationService(orgs *repositories.OrganizationRepository, db *gorm.DB) *OrganizationService {
	return &OrganizationService{orgs: orgs, db: db}
}

// Create makes a new organization owned by the caller (e.g. a personal org and a
// separate microempresa org under the same account) plus the owner membership,
// then seeds default categories/accounts so it isn't empty. Tenant-agnostic.
func (s *OrganizationService) Create(userID uuid.UUID, req dto.OrgCreateRequest) (*dto.OrgDTO, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, ErrInvalidOrgName
	}
	currency := money.DefaultCode
	if req.Currency != "" {
		if !money.IsSupported(req.Currency) {
			return nil, ErrUnsupportedCurrency
		}
		currency = money.Normalize(req.Currency)
	}
	org := &entities.Organization{Name: name, Slug: slugify(name), Currency: currency}
	if err := s.orgs.CreateForUser(org, userID); err != nil {
		return nil, err
	}
	// Onboarding: default categories/accounts. Best-effort (never blocks).
	if s.db != nil {
		if err := database.SeedDefaults(s.db, org.ID); err != nil {
			slog.Error("create org: default seed failed", "org_id", org.ID, "error", err)
		}
	}
	d := orgDTO(org, entities.RoleOwner)
	return &d, nil
}

// Get returns the active organization as an OrgDTO, stamped with the caller's
// role in it.
func (s *OrganizationService) Get(orgID uuid.UUID, role entities.Role) (*dto.OrgDTO, error) {
	o, err := s.orgs.FindByID(orgID)
	if err != nil {
		return nil, err
	}
	d := orgDTO(o, role)
	return &d, nil
}

// Update changes the active organization's name and/or currency. Empty fields
// are left untouched. A provided currency must be supported (ErrUnsupportedCurrency
// → 422 otherwise); it is stored normalized (uppercase). The caller's role is
// echoed back on the returned DTO.
func (s *OrganizationService) Update(orgID uuid.UUID, role entities.Role, req dto.OrgUpdateRequest) (*dto.OrgDTO, error) {
	o, err := s.orgs.FindByID(orgID)
	if err != nil {
		return nil, err
	}

	fields := map[string]interface{}{}
	if req.Name != "" {
		o.Name = req.Name
		fields["name"] = req.Name
	}
	if req.Currency != "" {
		if !money.IsSupported(req.Currency) {
			return nil, ErrUnsupportedCurrency
		}
		code := money.Normalize(req.Currency)
		o.Currency = code
		fields["currency"] = code
	}

	if err := s.orgs.Update(orgID, fields); err != nil {
		return nil, err
	}
	d := orgDTO(o, role)
	return &d, nil
}

// Currencies returns the static list of supported currencies for the picker.
func (s *OrganizationService) Currencies() []dto.CurrencyDTO {
	supported := money.Supported()
	out := make([]dto.CurrencyDTO, 0, len(supported))
	for _, c := range supported {
		out = append(out, dto.CurrencyDTO{Code: c.Code, Name: c.Name, Symbol: c.Symbol})
	}
	return out
}
