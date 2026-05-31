package services

import (
	"context"
	"errors"
	"time"

	"github.com/finance-sh/finance-sh/internal/dto"
	"github.com/finance-sh/finance-sh/internal/entities"
	"github.com/finance-sh/finance-sh/internal/recurrence"
	"github.com/finance-sh/finance-sh/internal/repositories"
	"github.com/finance-sh/finance-sh/pkg/cache"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ErrInvalidRecurrence is returned when a recurrence rule is malformed (bad
// frequency, interval < 1, non-positive amount, …). Handlers map it to HTTP 422.
var ErrInvalidRecurrence = errors.New("regra de recorrência inválida")

// maxCatchUpPerRun caps how many occurrences a single rule may generate in one
// GenerateDue pass. It bounds the catch-up loop so a rule whose StartDate is far
// in the past cannot generate an unbounded backlog (or spin) in one cycle; the
// remaining occurrences are picked up on the next worker tick.
const maxCatchUpPerRun = 60

// RecurrenceService is the proper recurring-transaction engine. It owns the
// RecurrenceRule CRUD and the worker-driven generation of transactions from due
// rules. It depends only on repositories + the pure recurrence package, so it
// introduces no import cycle (the worker constructs it directly from repos).
type RecurrenceService struct {
	rules      *repositories.RecurrenceRuleRepository
	txs        *repositories.TransactionRepository
	accounts   *repositories.AccountRepository
	categories *repositories.CategoryRepository
	contacts   *repositories.ContactRepository
	cache      *cache.Cache
	db         *gorm.DB
}

func NewRecurrenceService(
	rules *repositories.RecurrenceRuleRepository,
	txs *repositories.TransactionRepository,
	accounts *repositories.AccountRepository,
	categories *repositories.CategoryRepository,
	contacts *repositories.ContactRepository,
	c *cache.Cache,
	db *gorm.DB,
) *RecurrenceService {
	return &RecurrenceService{
		rules:      rules,
		txs:        txs,
		accounts:   accounts,
		categories: categories,
		contacts:   contacts,
		cache:      c,
		db:         db,
	}
}

func (s *RecurrenceService) invalidateDashboard(orgID uuid.UUID) {
	if s.cache != nil {
		s.cache.Delete(context.Background(), DashboardCacheKey(orgID))
	}
}

// validFrequency reports whether the frequency is one of the accepted values.
func validFrequency(f string) bool {
	switch f {
	case entities.FreqDaily, entities.FreqWeekly, entities.FreqMonthly, entities.FreqYearly:
		return true
	default:
		return false
	}
}

// ----- CRUD -----

func (s *RecurrenceService) List(orgID uuid.UUID) ([]dto.RecurrenceRuleDTO, error) {
	rules, err := s.rules.List(orgID)
	if err != nil {
		return nil, err
	}
	out := make([]dto.RecurrenceRuleDTO, 0, len(rules))
	for i := range rules {
		out = append(out, recurrenceRuleDTO(&rules[i]))
	}
	return out, nil
}

func (s *RecurrenceService) Create(orgID uuid.UUID, req dto.RecurrenceRuleRequest) (*dto.RecurrenceRuleDTO, error) {
	rule := &entities.RecurrenceRule{OrganizationID: orgID}
	if err := s.apply(orgID, rule, req, true); err != nil {
		return nil, err
	}
	// Fresh rule: nothing generated yet and the first run is the start date.
	rule.OccurrencesCount = 0
	rule.NextRunDate = rule.StartDate
	rule.LastGeneratedAt = nil

	if err := s.rules.Create(rule); err != nil {
		return nil, err
	}
	return s.reload(orgID, rule.ID)
}

func (s *RecurrenceService) Update(orgID, id uuid.UUID, req dto.RecurrenceRuleRequest) (*dto.RecurrenceRuleDTO, error) {
	rule, err := s.rules.FindByID(orgID, id)
	if err != nil {
		return nil, err
	}
	if err := s.apply(orgID, rule, req, false); err != nil {
		return nil, err
	}
	if err := s.rules.Update(rule); err != nil {
		return nil, err
	}
	return s.reload(orgID, id)
}

func (s *RecurrenceService) Delete(orgID, id uuid.UUID) error {
	return s.rules.Delete(orgID, id)
}

// apply validates the request and maps it onto the rule entity. On create the
// template fields and the schedule anchor are all set; on update the same fields
// are re-applied (NextRunDate/OccurrencesCount are NOT recomputed here so an edit
// does not silently replay the schedule). isCreate drives the Paid/Active and
// Interval defaults.
func (s *RecurrenceService) apply(orgID uuid.UUID, rule *entities.RecurrenceRule, req dto.RecurrenceRuleRequest, isCreate bool) error {
	if req.Amount <= 0 {
		return ErrInvalidRecurrence
	}
	if !validFrequency(req.Frequency) {
		return ErrInvalidRecurrence
	}
	interval := req.Interval
	if interval == 0 {
		interval = 1
	}
	if interval < 1 {
		return ErrInvalidRecurrence
	}
	if req.StartDate.IsZero() {
		return ErrInvalidRecurrence
	}
	if req.MaxOccurrences < 0 {
		return ErrInvalidRecurrence
	}
	if req.EndDate != nil && req.EndDate.Before(req.StartDate) {
		return ErrInvalidRecurrence
	}

	// Account ownership (required).
	accountID, err := parseUUID(req.AccountID)
	if err != nil {
		return err
	}
	if _, err := s.accounts.FindByID(orgID, accountID); err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			return ErrAccountNotInOrg
		}
		return err
	}

	// Optional category ownership.
	var categoryID *uuid.UUID
	if req.CategoryID != "" {
		cid, err := parseUUID(req.CategoryID)
		if err != nil {
			return err
		}
		if _, err := s.categories.FindByID(orgID, cid); err != nil {
			if errors.Is(err, repositories.ErrNotFound) {
				return ErrNotFound
			}
			return err
		}
		categoryID = &cid
	}

	// Optional contact ownership.
	var contactID *uuid.UUID
	if req.ContactID != "" {
		cid, err := parseUUID(req.ContactID)
		if err != nil {
			return err
		}
		if _, err := s.contacts.FindByID(orgID, cid); err != nil {
			if errors.Is(err, repositories.ErrNotFound) {
				return ErrContactNotInOrg
			}
			return err
		}
		contactID = &cid
	}

	rule.OrganizationID = orgID
	rule.Type = entities.TransactionType(req.Type)
	rule.Amount = req.Amount
	rule.Description = req.Description
	rule.AccountID = accountID
	rule.CategoryID = categoryID
	rule.ContactID = contactID
	rule.Frequency = req.Frequency
	rule.Interval = interval
	rule.StartDate = req.StartDate
	rule.EndDate = req.EndDate
	rule.MaxOccurrences = req.MaxOccurrences

	// Paid: default false on create; on update keep current unless sent.
	if req.Paid != nil {
		rule.Paid = *req.Paid
	} else if isCreate {
		rule.Paid = false
	}

	// Active: default true on create; on update keep current unless sent.
	if req.Active != nil {
		rule.Active = *req.Active
	} else if isCreate {
		rule.Active = true
	}
	return nil
}

// ----- Generation -----

// GenerateDue is the worker entry point: it walks every due rule across all orgs
// and generates the transactions they owe up to `now`. Returns the total number
// of transactions created. Affected orgs have their dashboard cache invalidated.
func (s *RecurrenceService) GenerateDue(now time.Time) (int, error) {
	dueRules, err := s.rules.DueRules(now)
	if err != nil {
		return 0, err
	}
	total := 0
	affected := make(map[uuid.UUID]bool)
	for i := range dueRules {
		rule := dueRules[i] // copy: we mutate and persist per rule
		created, err := s.generateForRule(&rule, now)
		if err != nil {
			// One bad rule must not abort the whole pass; the worker logs the count.
			continue
		}
		if created > 0 {
			total += created
			affected[rule.OrganizationID] = true
		}
	}
	for orgID := range affected {
		s.invalidateDashboard(orgID)
	}
	return total, nil
}

// RunNow generates the due occurrences for ONE rule immediately (org-scoped),
// reusing the same generation logic as the worker. Returns the count created.
func (s *RecurrenceService) RunNow(orgID, id uuid.UUID) (int, error) {
	rule, err := s.rules.FindByID(orgID, id)
	if err != nil {
		return 0, err
	}
	now := time.Now().UTC()
	created, err := s.generateForRule(rule, now)
	if err != nil {
		return 0, err
	}
	if created > 0 {
		s.invalidateDashboard(orgID)
	}
	return created, nil
}

// generateForRule materialises every occurrence a rule owes up to `now`, in a
// single DB transaction. The loop runs while the rule is active, its NextRunDate
// is due (<= now) and it is still within its end/max bounds, capped at
// maxCatchUpPerRun. Each occurrence advances NextRunDate via recurrence.Next; the
// idempotency guard is that advance — once NextRunDate moves past `now` the rule
// is no longer due, so a re-run (or a later worker tick) will not re-generate it.
// When the schedule is exhausted (EndDate passed or MaxOccurrences reached) the
// rule is deactivated. The mutated rule is persisted inside the same transaction.
func (s *RecurrenceService) generateForRule(rule *entities.RecurrenceRule, now time.Time) (int, error) {
	created := 0
	err := s.db.Transaction(func(tx *gorm.DB) error {
		for i := 0; i < maxCatchUpPerRun; i++ {
			if !rule.Active {
				break
			}
			if rule.NextRunDate.After(now) {
				break
			}
			// Bound by end date: an occurrence past EndDate exhausts the rule.
			if rule.EndDate != nil && rule.NextRunDate.After(*rule.EndDate) {
				rule.Active = false
				break
			}
			// Bound by max occurrences (0 = unlimited).
			if rule.MaxOccurrences > 0 && rule.OccurrencesCount >= rule.MaxOccurrences {
				rule.Active = false
				break
			}

			occDate := rule.NextRunDate
			due := occDate
			t := &entities.Transaction{
				OrganizationID: rule.OrganizationID,
				AccountID:      rule.AccountID,
				CategoryID:     rule.CategoryID,
				ContactID:      rule.ContactID,
				Type:           rule.Type,
				Amount:         rule.Amount,
				Description:    rule.Description,
				Date:           occDate,
				DueDate:        &due,
				Paid:           rule.Paid,
			}
			if err := tx.Create(t).Error; err != nil {
				return err
			}

			created++
			rule.OccurrencesCount++
			gen := now
			rule.LastGeneratedAt = &gen
			rule.NextRunDate = recurrence.Next(occDate, rule.Frequency, rule.Interval)

			// Re-evaluate the bounds after advancing so the rule is deactivated as
			// soon as it is exhausted (no extra occurrence is generated).
			if rule.MaxOccurrences > 0 && rule.OccurrencesCount >= rule.MaxOccurrences {
				rule.Active = false
				break
			}
			if rule.EndDate != nil && rule.NextRunDate.After(*rule.EndDate) {
				rule.Active = false
				break
			}
		}
		// Persist the rule's advanced schedule/state inside the same transaction so
		// the occurrences and the cursor move atomically.
		return tx.Save(rule).Error
	})
	if err != nil {
		return 0, err
	}
	return created, nil
}

func (s *RecurrenceService) reload(orgID, id uuid.UUID) (*dto.RecurrenceRuleDTO, error) {
	rule, err := s.rules.FindByID(orgID, id)
	if err != nil {
		return nil, err
	}
	d := recurrenceRuleDTO(rule)
	return &d, nil
}

func recurrenceRuleDTO(r *entities.RecurrenceRule) dto.RecurrenceRuleDTO {
	d := dto.RecurrenceRuleDTO{
		ID:               r.ID.String(),
		Type:             string(r.Type),
		Amount:           r.Amount,
		Description:      r.Description,
		AccountID:        r.AccountID.String(),
		Paid:             r.Paid,
		Frequency:        r.Frequency,
		Interval:         r.Interval,
		StartDate:        r.StartDate,
		EndDate:          r.EndDate,
		MaxOccurrences:   r.MaxOccurrences,
		OccurrencesCount: r.OccurrencesCount,
		NextRunDate:      r.NextRunDate,
		Active:           r.Active,
	}
	if r.CategoryID != nil {
		d.CategoryID = r.CategoryID.String()
	}
	if r.ContactID != nil {
		d.ContactID = r.ContactID.String()
	}
	if r.Account != nil {
		d.Account = &dto.RecurrenceRuleAccountDTO{ID: r.Account.ID.String(), Name: r.Account.Name}
	}
	if r.Category != nil {
		c := categoryDTO(r.Category)
		d.Category = &c
	}
	return d
}
