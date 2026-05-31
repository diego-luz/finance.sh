package entities

import (
	"time"

	"github.com/finance-sh/finance-sh/pkg/crypto"
	"github.com/google/uuid"
)

// Money is always stored as int64 minor units (cents) to avoid float rounding.

type AccountType string

const (
	AccountBank       AccountType = "bank"
	AccountWallet     AccountType = "wallet"
	AccountInvestment AccountType = "investment"
	AccountCreditCard AccountType = "credit_card"
)

// Account is a place where money lives: a bank account, wallet, investment or
// credit card.
type Account struct {
	Base
	OrganizationID uuid.UUID   `gorm:"type:uuid;not null;index" json:"organization_id"`
	Name           string      `gorm:"not null" json:"name"`
	Type           AccountType `gorm:"type:varchar(20);default:'bank'" json:"type"`
	InitialBalance int64       `gorm:"default:0" json:"initial_balance"`
	Color          string      `gorm:"type:varchar(9);default:'#10b981'" json:"color"`
	Icon           string      `gorm:"default:'wallet'" json:"icon"`
	Archived       bool        `gorm:"default:false" json:"archived"`
}

func (Account) TableName() string { return "accounts" }

type CategoryKind string

const (
	CategoryIncome  CategoryKind = "income"
	CategoryExpense CategoryKind = "expense"
)

// Category groups transactions for reporting and budgeting.
type Category struct {
	Base
	OrganizationID uuid.UUID    `gorm:"type:uuid;not null;index" json:"organization_id"`
	Name           string       `gorm:"not null" json:"name"`
	Kind           CategoryKind `gorm:"type:varchar(10);default:'expense'" json:"kind"`
	Color          string       `gorm:"type:varchar(9);default:'#6366f1'" json:"color"`
	Icon           string       `gorm:"default:'tag'" json:"icon"`
}

func (Category) TableName() string { return "categories" }

// Tag is a free-form rótulo attached to transactions (many-to-many). Tags are
// org-scoped and their name is unique per organization (composite unique index
// idx_org_tag_name on organization_id+name).
type Tag struct {
	Base
	OrganizationID uuid.UUID `gorm:"type:uuid;not null;index;uniqueIndex:idx_org_tag_name" json:"organization_id"`
	Name           string    `gorm:"not null;uniqueIndex:idx_org_tag_name" json:"name"`
	Color          string    `gorm:"type:varchar(9);default:'#6b7280'" json:"color"`
}

func (Tag) TableName() string { return "tags" }

// CategoryRule drives automatic categorization: when a transaction description
// matches Pattern (case-insensitively, by MatchType), CategoryID is assigned.
// Rules are org-scoped; on multiple matches the highest Priority wins. MatchType
// is one of contains | prefix | regex (regex uses Go's regexp, compiled at match
// time). Inactive rules are ignored.
type CategoryRule struct {
	Base
	OrganizationID uuid.UUID `gorm:"type:uuid;not null;index" json:"organization_id"`
	Pattern        string    `gorm:"not null" json:"pattern"`
	MatchType      string    `gorm:"type:varchar(10);default:'contains'" json:"match_type"`
	CategoryID     uuid.UUID `gorm:"type:uuid;not null;index" json:"category_id"`
	Priority       int       `gorm:"default:0" json:"priority"`
	Active         bool      `gorm:"default:true" json:"active"`

	Category *Category `json:"category,omitempty"`
}

func (CategoryRule) TableName() string { return "category_rules" }

// MatchType values for CategoryRule.
const (
	MatchContains = "contains"
	MatchPrefix   = "prefix"
	MatchRegex    = "regex"
)

type ContactType string

const (
	ContactCustomer ContactType = "customer"
	ContactSupplier ContactType = "supplier"
	ContactBoth     ContactType = "both"
)

// Contact is a person or company the org transacts with (cliente/fornecedor).
// Used to attribute payables and receivables to a counterparty.
type Contact struct {
	Base
	OrganizationID uuid.UUID   `gorm:"type:uuid;not null;index" json:"organization_id"`
	Name           string      `gorm:"not null" json:"name"`
	Type           ContactType `gorm:"type:varchar(10);default:'both'" json:"type"`
	Document       string      `gorm:"type:varchar(20)" json:"document,omitempty"` // CPF/CNPJ
	Email          string      `json:"email,omitempty"`
	Phone          string      `gorm:"type:varchar(30)" json:"phone,omitempty"`
	Notes          string      `gorm:"type:text" json:"notes,omitempty"`
}

func (Contact) TableName() string { return "contacts" }

type TransactionType string

const (
	TxIncome   TransactionType = "income"
	TxExpense  TransactionType = "expense"
	TxTransfer TransactionType = "transfer"
)

// Transaction is a single money movement. Transfers reference a second account
// via TransferAccountID. Amounts are stored as positive cents; the Type field
// determines the sign applied during aggregation.
type Transaction struct {
	Base
	OrganizationID    uuid.UUID  `gorm:"type:uuid;not null;index" json:"organization_id"`
	AccountID         uuid.UUID  `gorm:"type:uuid;not null;index" json:"account_id"`
	CategoryID        *uuid.UUID `gorm:"type:uuid;index" json:"category_id,omitempty"`
	CreditCardID      *uuid.UUID `gorm:"type:uuid;index" json:"credit_card_id,omitempty"`
	TransferAccountID *uuid.UUID `gorm:"type:uuid" json:"transfer_account_id,omitempty"`
	// ContactID links the movement to a counterparty (cliente/fornecedor).
	// Nullable so existing data and flows are unaffected.
	ContactID   *uuid.UUID      `gorm:"type:uuid;index" json:"contact_id,omitempty"`
	Type        TransactionType `gorm:"type:varchar(10);not null;index" json:"type"`
	Amount      int64           `gorm:"not null" json:"amount"`
	Description string          `gorm:"not null" json:"description"`
	Date        time.Time       `gorm:"not null;index" json:"date"`
	// DueDate is the vencimento (when the bill/receivable falls due), distinct
	// from Date which is the competência (accrual). Nullable: when nil the Date
	// is used as the due reference. Used by AP/AR and the cash-flow forecast.
	DueDate *time.Time `gorm:"index" json:"due_date,omitempty"`
	// PaidAt records when the movement was settled. Nullable; set on settle.
	PaidAt *time.Time `json:"paid_at,omitempty"`
	// No GORM default: a default:true tag makes GORM skip the zero value (false),
	// silently persisting paid=true and breaking unpaid-bill tracking. The service
	// layer decides the value; the API/form always sends it explicitly.
	Paid      bool `gorm:"not null" json:"paid"`
	Recurring bool `gorm:"default:false" json:"recurring"`
	// Installment fields (parcelamento). A parcelado purchase is materialised as a
	// GROUP of N sibling transactions sharing the same InstallmentGroupID, one per
	// month. All optional/zero by default so non-parcelado transactions and existing
	// data are unaffected.
	//   InstallmentGroupID — set on every parcela of a group; nil for a single tx.
	//   InstallmentNumber  — 1-based position within the group; 0 = not an installment.
	//   InstallmentTotal   — total parcelas in the group; 0 = not an installment.
	InstallmentGroupID *uuid.UUID `gorm:"type:uuid;index" json:"installment_group_id,omitempty"`
	InstallmentNumber  int        `gorm:"default:0" json:"installment_number"`
	InstallmentTotal   int        `gorm:"default:0" json:"installment_total"`
	// ExternalID is the dedup key for imported statements: the OFX FITID, or a
	// short sha256 hash of date+amount+description for CSV rows without an id. It
	// is indexed so import preview/commit can batch-check existing rows. Empty for
	// manually-created transactions; the omitempty tag keeps it out of their payload.
	ExternalID string `gorm:"index" json:"external_id,omitempty"`
	// Notes is free-text and may contain sensitive personal information, so it is
	// encrypted at rest (AES-256-GCM). It is never used in WHERE/ORDER clauses, so
	// encryption does not impact querying. See pkg/crypto for the tradeoff.
	Notes crypto.EncryptedString `gorm:"type:text" json:"notes,omitempty"`

	Category *Category `json:"category,omitempty"`
	Account  *Account  `json:"account,omitempty"`
	Contact  *Contact  `json:"contact,omitempty"`
	// Tags is the many-to-many association with Tag through the implicit join
	// table transaction_tags. AutoMigrate(&Transaction{}, &Tag{}) creates the join.
	Tags []Tag `gorm:"many2many:transaction_tags;" json:"tags,omitempty"`
}

func (Transaction) TableName() string { return "transactions" }

// Attachment is a file (receipt/comprovante) stored in Postgres BYTEA, linked
// to a transaction. TransactionID is nullable so the entity can later extend
// to attachments on other domains. ObjectKey is an unused reserved column kept
// nullable/empty (blobs live in Data). The Data column is the BYTEA blob and is
// never JSON-encoded — handlers stream it through pkg/storage instead.
type Attachment struct {
	Base
	OrganizationID uuid.UUID  `gorm:"type:uuid;not null;index" json:"organization_id"`
	TransactionID  *uuid.UUID `gorm:"type:uuid;index" json:"transaction_id,omitempty"`
	FileName       string     `gorm:"not null" json:"file_name"`
	ContentType    string     `json:"content_type"`
	Size           int64      `json:"size"`
	SizeBytes      int64      `gorm:"column:size_bytes" json:"-"`
	ObjectKey      string     `json:"-"`
	Data           []byte     `gorm:"type:bytea" json:"-"`
	UploadedBy     uuid.UUID  `gorm:"type:uuid" json:"uploaded_by"`
}

func (Attachment) TableName() string { return "attachments" }

// CreditCard tracks limit and billing cycle. Card purchases generate
// transactions that settle on the invoice due date.
type CreditCard struct {
	Base
	OrganizationID uuid.UUID `gorm:"type:uuid;not null;index" json:"organization_id"`
	Name           string    `gorm:"not null" json:"name"`
	Limit          int64     `gorm:"default:0" json:"limit"`
	ClosingDay     int       `gorm:"default:1" json:"closing_day"`
	DueDay         int       `gorm:"default:10" json:"due_day"`
	Color          string    `gorm:"type:varchar(9);default:'#0f1115'" json:"color"`
}

func (CreditCard) TableName() string { return "credit_cards" }
