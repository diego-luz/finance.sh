package repositories

import (
	"errors"
	"time"

	"github.com/finance-sh/finance-sh/internal/entities"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// RecurrenceRuleRepository persists recurrence rules. The CRUD methods are
// org-scoped (every WHERE carries organization_id); DueRules is the single
// GLOBAL read the worker uses to walk every org's due rules in one pass.
type RecurrenceRuleRepository struct{ db *gorm.DB }

func NewRecurrenceRuleRepository(db *gorm.DB) *RecurrenceRuleRepository {
	return &RecurrenceRuleRepository{db: db}
}

func (r *RecurrenceRuleRepository) Create(rule *entities.RecurrenceRule) error {
	return r.db.Create(rule).Error
}

func (r *RecurrenceRuleRepository) Update(rule *entities.RecurrenceRule) error {
	return r.db.Save(rule).Error
}

func (r *RecurrenceRuleRepository) Delete(orgID, id uuid.UUID) error {
	res := r.db.Where("organization_id = ? AND id = ?", orgID, id).Delete(&entities.RecurrenceRule{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *RecurrenceRuleRepository) FindByID(orgID, id uuid.UUID) (*entities.RecurrenceRule, error) {
	var rule entities.RecurrenceRule
	err := r.db.Preload("Account").Preload("Category").
		Where("organization_id = ? AND id = ?", orgID, id).First(&rule).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &rule, err
}

// List returns every recurrence rule of the org, ordered by next_run_date so the
// soonest upcoming recurrence is first. Account and Category are preloaded for
// display in the DTO.
func (r *RecurrenceRuleRepository) List(orgID uuid.UUID) ([]entities.RecurrenceRule, error) {
	var rules []entities.RecurrenceRule
	err := r.db.Preload("Account").Preload("Category").
		Where("organization_id = ?", orgID).
		Order("next_run_date asc").
		Find(&rules).Error
	return rules, err
}

// DueRules returns the active rules whose next_run_date is at or before `now`,
// across ALL organizations (the worker fans out per-org generation from this).
// A rule is due only while it still has occurrences left to generate:
//   - active = true
//   - next_run_date <= now
//   - (max_occurrences = 0 OR occurrences_count < max_occurrences)
//   - (end_date IS NULL OR next_run_date <= end_date)
//
// Account and Category are preloaded so generation can clone the template without
// extra round-trips.
func (r *RecurrenceRuleRepository) DueRules(now time.Time) ([]entities.RecurrenceRule, error) {
	var rules []entities.RecurrenceRule
	// Only rules whose account and organization still exist: the worker kept
	// generating into deleted accounts and into organizations erased by LGPD.
	err := r.db.Preload("Account").Preload("Category").
		Joins("JOIN accounts acc ON acc.id = recurrence_rules.account_id AND acc.deleted_at IS NULL").
		Joins("JOIN organizations org ON org.id = recurrence_rules.organization_id AND org.deleted_at IS NULL").
		Where("recurrence_rules.active = ? AND recurrence_rules.next_run_date <= ?", true, now).
		Where("recurrence_rules.max_occurrences = 0 OR recurrence_rules.occurrences_count < recurrence_rules.max_occurrences").
		Where("recurrence_rules.end_date IS NULL OR recurrence_rules.next_run_date <= recurrence_rules.end_date").
		Order("recurrence_rules.next_run_date asc").
		Find(&rules).Error
	return rules, err
}
