package entities

import (
	"time"

	"github.com/google/uuid"
)

// RecurrenceFrequency values accepted by a RecurrenceRule.
const (
	FreqDaily   = "daily"
	FreqWeekly  = "weekly"
	FreqMonthly = "monthly"
	FreqYearly  = "yearly"
)

// RecurrenceRule is the recurring-transaction engine: it holds a template for
// the transaction to materialise plus a schedule that the worker walks to
// generate occurrences. It replaced the legacy Transaction.Recurring boolean
// flag job, which has been retired. The Recurring column on Transaction is kept
// read-only for backward compatibility, but all recurrences are modelled here.
//
// Money is int64 minor units (cents), like Transaction.Amount. Every query is
// scoped by organization_id; the worker's DueRules read is the only global one.
type RecurrenceRule struct {
	Base
	OrganizationID uuid.UUID `gorm:"type:uuid;not null;index" json:"organization_id"`

	// ----- TEMPLATE: the transaction each occurrence is cloned from -----
	Type        TransactionType `gorm:"type:varchar(10);not null" json:"type"` // income | expense
	Amount      int64           `gorm:"not null" json:"amount"`
	Description string          `gorm:"type:text;not null" json:"description"`
	AccountID   uuid.UUID       `gorm:"type:uuid;not null;index" json:"account_id"`
	CategoryID  *uuid.UUID      `gorm:"type:uuid;index" json:"category_id,omitempty"`
	ContactID   *uuid.UUID      `gorm:"type:uuid;index" json:"contact_id,omitempty"`
	// Paid is the paid-state stamped on every generated transaction (so a rule can
	// produce already-settled income or unpaid bills). No GORM default for the same
	// reason Transaction.Paid carries none: a default would mask an explicit false.
	Paid bool `gorm:"not null;default:false" json:"paid"`

	// ----- SCHEDULE -----
	Frequency        string     `gorm:"type:varchar(10);not null" json:"frequency"` // daily | weekly | monthly | yearly
	Interval         int        `gorm:"not null;default:1" json:"interval"`         // every N periods
	StartDate        time.Time  `gorm:"not null" json:"start_date"`
	EndDate          *time.Time `json:"end_date,omitempty"`
	MaxOccurrences   int        `gorm:"not null;default:0" json:"max_occurrences"`   // 0 = unlimited
	OccurrencesCount int        `gorm:"not null;default:0" json:"occurrences_count"` // generated so far
	NextRunDate      time.Time  `gorm:"not null;index" json:"next_run_date"`
	LastGeneratedAt  *time.Time `json:"last_generated_at,omitempty"`
	Active           bool       `gorm:"not null;default:true" json:"active"`

	// Associations (preloaded for the DTO; never used in WHERE clauses).
	Account  *Account  `json:"account,omitempty"`
	Category *Category `json:"category,omitempty"`
}

func (RecurrenceRule) TableName() string { return "recurrence_rules" }
