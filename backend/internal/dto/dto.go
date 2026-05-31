// Package dto holds request/response data-transfer objects. They decouple the
// transport layer from entities and carry validation tags.
package dto

import (
	"time"

	"github.com/google/uuid"
)

// ----- Auth -----

type RegisterRequest struct {
	Name             string `json:"name" validate:"required,min=2,max=120"`
	Email            string `json:"email" validate:"required,email"`
	Password         string `json:"password" validate:"required,min=8,max=72"`
	OrganizationName string `json:"organization_name" validate:"required,min=2,max=120"`
	// Currency is the optional ISO-4217 code for the new organization. When absent
	// it defaults to BRL; when present it must be a supported code (the service
	// returns ErrUnsupportedCurrency → 422 otherwise).
	Currency string `json:"currency" validate:"omitempty,len=3"`
	// LGPD: the user must explicitly accept the terms of use / privacy policy.
	AcceptedTerms bool `json:"accepted_terms" validate:"eq=true"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type AuthResponse struct {
	AccessToken  string  `json:"access_token"`
	RefreshToken string  `json:"refresh_token"`
	ExpiresIn    int64   `json:"expires_in"`
	User         UserDTO `json:"user"`
	Organization OrgDTO  `json:"organization"`
}

// LoginResult is what AuthService.Login returns. When the user has 2FA enabled,
// MFARequired is true and only MFAToken is populated (a short-lived token to be
// exchanged at /auth/2fa/verify). Otherwise Auth carries the normal response.
type LoginResult struct {
	MFARequired bool          `json:"mfa_required"`
	MFAToken    string        `json:"mfa_token,omitempty"`
	Auth        *AuthResponse `json:"-"`
}

// MFAChallengeResponse is returned to the client when 2FA is required.
type MFAChallengeResponse struct {
	MFARequired bool   `json:"mfa_required"`
	MFAToken    string `json:"mfa_token"`
}

// TwoFactorVerifyRequest exchanges an mfa_token + TOTP/recovery code for tokens.
type TwoFactorVerifyRequest struct {
	MFAToken string `json:"mfa_token" validate:"required"`
	Code     string `json:"code" validate:"required"`
}

type UserDTO struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	Email            string `json:"email"`
	EmailVerified    bool   `json:"email_verified"`
	TwoFactorEnabled bool   `json:"two_factor_enabled"`
	AvatarURL        string `json:"avatar_url,omitempty"`
	// SuperAdmin lets the frontend show the platform back-office area. Derived
	// from the User entity (platform-level role, not per-org RBAC).
	SuperAdmin bool `json:"super_admin"`
	// MustChangePassword is true when the user was admin-provisioned with an
	// initial password and has not changed it yet. The frontend uses this flag to
	// force the change-password UI right after login.
	MustChangePassword bool `json:"must_change_password"`
}

type OrgDTO struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Slug     string `json:"slug"`
	Currency string `json:"currency"`
	Role     string `json:"role"`
}

type MeResponse struct {
	User          UserDTO  `json:"user"`
	Organizations []OrgDTO `json:"organizations"`
}

// OrgUpdateRequest updates the active organization. Both fields are optional:
// an absent/empty value leaves the current one untouched. Currency, when set,
// must be a supported ISO-4217 code (the service returns ErrUnsupportedCurrency
// → 422 otherwise).
type OrgUpdateRequest struct {
	Name     string `json:"name" validate:"omitempty,min=2,max=120"`
	Currency string `json:"currency" validate:"omitempty,len=3"`
}

// OrgCreateRequest is the body of POST /organizations (self-service: create an
// additional organization owned by the caller).
type OrgCreateRequest struct {
	Name     string `json:"name" validate:"required,min=2,max=120"`
	Currency string `json:"currency" validate:"omitempty,len=3"`
}

// CurrencyDTO is one supported currency exposed by GET /currencies. Name is in
// pt-BR.
type CurrencyDTO struct {
	Code   string `json:"code"`
	Name   string `json:"name"`
	Symbol string `json:"symbol"`
}

// ----- Accounts -----

type AccountRequest struct {
	Name           string `json:"name" validate:"required,max=120"`
	Type           string `json:"type" validate:"required,oneof=bank wallet investment credit_card"`
	InitialBalance int64  `json:"initial_balance"`
	Color          string `json:"color" validate:"max=9"`
	Icon           string `json:"icon" validate:"max=40"`
}

type AccountDTO struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Type           string `json:"type"`
	InitialBalance int64  `json:"initial_balance"`
	Balance        int64  `json:"balance"`
	Color          string `json:"color"`
	Icon           string `json:"icon"`
	Archived       bool   `json:"archived"`
}

// ----- Categories -----

type CategoryRequest struct {
	Name  string `json:"name" validate:"required,max=120"`
	Kind  string `json:"kind" validate:"required,oneof=income expense"`
	Color string `json:"color" validate:"max=9"`
	Icon  string `json:"icon" validate:"max=40"`
}

type CategoryDTO struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Kind  string `json:"kind"`
	Color string `json:"color"`
	Icon  string `json:"icon"`
}

// ----- Categorization rules -----

// CategoryRuleRequest creates/updates a rule. MatchType defaults to "contains"
// when empty; Priority defaults to 0; Active is honoured as sent (the handler
// passes a pointer so an absent flag defaults to true on create).
type CategoryRuleRequest struct {
	Pattern    string `json:"pattern" validate:"required,max=200"`
	CategoryID string `json:"category_id" validate:"required,uuid"`
	MatchType  string `json:"match_type" validate:"omitempty,oneof=contains prefix regex"`
	Priority   int    `json:"priority"`
	// Active is a pointer so "absent" (create) differs from explicit false; nil
	// means default true.
	Active *bool `json:"active"`
}

type CategoryRuleDTO struct {
	ID         string       `json:"id"`
	Pattern    string       `json:"pattern"`
	MatchType  string       `json:"match_type"`
	CategoryID string       `json:"category_id"`
	Category   *CategoryDTO `json:"category,omitempty"`
	Priority   int          `json:"priority"`
	Active     bool         `json:"active"`
}

// SuggestResponse is returned by GET /categorization/suggest. Source is one of
// "rule" | "history" | "none". CategoryID/Category are present only when a match
// was found.
type SuggestResponse struct {
	CategoryID string       `json:"category_id,omitempty"`
	Source     string       `json:"source"`
	Category   *CategoryDTO `json:"category,omitempty"`
}

// ApplyResult reports how many uncategorized transactions were auto-assigned.
type ApplyResult struct {
	Updated int `json:"updated"`
}

// ----- Contacts -----

type ContactRequest struct {
	Name     string `json:"name" validate:"required,max=120"`
	Type     string `json:"type" validate:"omitempty,oneof=customer supplier both"`
	Document string `json:"document" validate:"max=20"`
	Email    string `json:"email" validate:"omitempty,email"`
	Phone    string `json:"phone" validate:"max=30"`
	Notes    string `json:"notes" validate:"max=500"`
}

type ContactDTO struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	Document string `json:"document,omitempty"`
	Email    string `json:"email,omitempty"`
	Phone    string `json:"phone,omitempty"`
	Notes    string `json:"notes,omitempty"`
}

// ----- Tags (rótulos) -----

type TagRequest struct {
	Name  string `json:"name" validate:"required,max=60"`
	Color string `json:"color" validate:"max=9"`
}

type TagDTO struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
}

// ----- Transactions -----

type TransactionRequest struct {
	AccountID         string     `json:"account_id" validate:"required,uuid"`
	CategoryID        string     `json:"category_id" validate:"omitempty,uuid"`
	CreditCardID      string     `json:"credit_card_id" validate:"omitempty,uuid"`
	TransferAccountID string     `json:"transfer_account_id" validate:"omitempty,uuid"`
	ContactID         string     `json:"contact_id" validate:"omitempty,uuid"`
	Type              string     `json:"type" validate:"required,oneof=income expense transfer"`
	Amount            int64      `json:"amount" validate:"required,gt=0"`
	Description       string     `json:"description" validate:"required,max=200"`
	Date              time.Time  `json:"date" validate:"required"`
	DueDate           *time.Time `json:"due_date"`
	Paid              bool       `json:"paid"`
	Notes             string     `json:"notes" validate:"max=500"`
	// Installments parcels the purchase: absent/0/1 = single transaction; >1 =
	// generate that many sibling transactions, one per month (parcelamento).
	Installments int `json:"installments" validate:"omitempty,min=1,max=420"`
	// TagIDs are the org's tag ids to attach to the transaction. Foreign ids
	// (not belonging to the org) are silently skipped by the service. An empty
	// or absent list clears the tags on update.
	TagIDs []string `json:"tag_ids" validate:"omitempty,dive,uuid"`
}

// TransactionContactDTO is the minimal contact shape embedded in a transaction.
type TransactionContactDTO struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type TransactionDTO struct {
	ID           string     `json:"id"`
	AccountID    string     `json:"account_id"`
	CreditCardID string     `json:"credit_card_id,omitempty"`
	ContactID    string     `json:"contact_id,omitempty"`
	Type         string     `json:"type"`
	Amount       int64      `json:"amount"`
	Description  string     `json:"description"`
	Date         time.Time  `json:"date"`
	DueDate      *time.Time `json:"due_date,omitempty"`
	PaidAt       *time.Time `json:"paid_at,omitempty"`
	Paid         bool       `json:"paid"`
	Recurring    bool       `json:"recurring"`
	// Status is derived (not persisted): "paid" | "overdue" | "open".
	Status string `json:"status"`
	// Installment fields (parcelamento). InstallmentGroupID is omitted when the
	// transaction is not part of a group; InstallmentTotal (and Number) are omitted
	// when total==0 (not an installment).
	InstallmentGroupID string                 `json:"installment_group_id,omitempty"`
	InstallmentNumber  int                    `json:"installment_number,omitempty"`
	InstallmentTotal   int                    `json:"installment_total,omitempty"`
	Notes              string                 `json:"notes,omitempty"`
	Category           *CategoryDTO           `json:"category,omitempty"`
	Contact            *TransactionContactDTO `json:"contact,omitempty"`
	// Tags are the rótulos attached to the transaction. Empty/omitted when none.
	Tags []TagDTO `json:"tags,omitempty"`
	// AttachmentCount is the number of receipts/comprovantes attached to the
	// transaction. Populated efficiently (one grouped query per page) in List and
	// on the single GET; 0 when none.
	AttachmentCount int `json:"attachment_count"`
}

// ----- Attachments (comprovantes) -----

// AttachmentDTO is the public shape of a transaction attachment. The raw object
// storage key is never exposed.
type AttachmentDTO struct {
	ID            string    `json:"id"`
	TransactionID string    `json:"transaction_id,omitempty"`
	FileName      string    `json:"file_name"`
	ContentType   string    `json:"content_type"`
	Size          int64     `json:"size"`
	CreatedAt     time.Time `json:"created_at"`
}

// ----- Recurrence rules (recurring-transaction engine) -----

// RecurrenceRuleRequest creates/updates a recurrence rule. Interval defaults to
// 1, MaxOccurrences to 0 (unlimited). Paid/Active are pointers so an absent flag
// keeps the documented default (Paid=false on create; Active=true on create).
type RecurrenceRuleRequest struct {
	Type        string `json:"type" validate:"required,oneof=income expense"`
	Amount      int64  `json:"amount" validate:"required,gt=0"`
	Description string `json:"description" validate:"required,max=200"`
	AccountID   string `json:"account_id" validate:"required,uuid"`
	CategoryID  string `json:"category_id" validate:"omitempty,uuid"`
	ContactID   string `json:"contact_id" validate:"omitempty,uuid"`
	// Paid is the paid-state stamped on each generated transaction. Pointer so
	// "absent" (create) differs from explicit false; nil means default false.
	Paid           *bool      `json:"paid"`
	Frequency      string     `json:"frequency" validate:"required,oneof=daily weekly monthly yearly"`
	Interval       int        `json:"interval" validate:"omitempty,min=1,max=365"`
	StartDate      time.Time  `json:"start_date" validate:"required"`
	EndDate        *time.Time `json:"end_date"`
	MaxOccurrences int        `json:"max_occurrences" validate:"omitempty,min=0"`
	// Active toggles generation. Pointer so absent keeps the default (true on
	// create); on update, nil leaves the current value untouched.
	Active *bool `json:"active"`
}

// RecurrenceRuleAccountDTO is the minimal account shape embedded in a rule.
type RecurrenceRuleAccountDTO struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type RecurrenceRuleDTO struct {
	ID               string                    `json:"id"`
	Type             string                    `json:"type"`
	Amount           int64                     `json:"amount"`
	Description      string                    `json:"description"`
	AccountID        string                    `json:"account_id"`
	CategoryID       string                    `json:"category_id,omitempty"`
	ContactID        string                    `json:"contact_id,omitempty"`
	Paid             bool                      `json:"paid"`
	Frequency        string                    `json:"frequency"`
	Interval         int                       `json:"interval"`
	StartDate        time.Time                 `json:"start_date"`
	EndDate          *time.Time                `json:"end_date,omitempty"`
	MaxOccurrences   int                       `json:"max_occurrences"`
	OccurrencesCount int                       `json:"occurrences_count"`
	NextRunDate      time.Time                 `json:"next_run_date"`
	Active           bool                      `json:"active"`
	Account          *RecurrenceRuleAccountDTO `json:"account,omitempty"`
	Category         *CategoryDTO              `json:"category,omitempty"`
}

// RecurrenceRunResult reports how many transactions a manual run generated.
type RecurrenceRunResult struct {
	Created int `json:"created"`
}

// ----- Statement import (OFX/CSV) -----

// ImportPreview is the result of parsing a statement without persisting: every
// parsed row plus a summary. The frontend renders it, lets the user
// edit/deselect rows, then posts the kept rows to /imports/commit.
type ImportPreview struct {
	Format  string         `json:"format"` // "ofx" | "csv"
	Rows    []ImportRowDTO `json:"rows"`
	Summary ImportSummary  `json:"summary"`
}

// ImportRowDTO is one parsed statement line. Amount is positive cents; Type is
// "income"|"expense". Duplicate flags a row that already exists (by ExternalID,
// or by date+amount+description when no external id); Reason explains why.
type ImportRowDTO struct {
	Index       int       `json:"index"`
	Date        time.Time `json:"date"`
	Description string    `json:"description"`
	Amount      int64     `json:"amount"`
	Type        string    `json:"type"`
	ExternalID  string    `json:"external_id,omitempty"`
	Duplicate   bool      `json:"duplicate"`
	Reason      string    `json:"reason,omitempty"`
	// SuggestedCategoryID is the category the auto-categorizer would assign to
	// this row (rule or history match); empty when nothing matched. The frontend
	// can pre-fill/accept it before committing.
	SuggestedCategoryID string `json:"suggested_category_id,omitempty"`
}

// ImportSummary counts the preview rows by disposition.
type ImportSummary struct {
	Total      int `json:"total"`
	New        int `json:"new"`
	Duplicates int `json:"duplicates"`
}

// ImportCommitRequest persists a set of reviewed rows into an account. The
// optional CategoryID (validated against the org) is applied to every row.
type ImportCommitRequest struct {
	AccountID  string            `json:"account_id" validate:"required,uuid"`
	CategoryID string            `json:"category_id" validate:"omitempty,uuid"`
	Rows       []ImportCommitRow `json:"rows" validate:"required,min=1,dive"`
}

// ImportCommitRow is one transaction to create. Amount is positive cents.
type ImportCommitRow struct {
	Date        time.Time `json:"date" validate:"required"`
	Description string    `json:"description" validate:"required,max=200"`
	Amount      int64     `json:"amount" validate:"required,gt=0"`
	Type        string    `json:"type" validate:"required,oneof=income expense"`
	ExternalID  string    `json:"external_id" validate:"omitempty,max=120"`
}

// ImportCommitResult reports how many rows were created vs skipped (duplicates).
type ImportCommitResult struct {
	Created int `json:"created"`
	Skipped int `json:"skipped"`
}

// SettleRequest marks a transaction as paid; both fields optional. When
// AccountID is set the transaction is moved to that account on settle.
type SettleRequest struct {
	AccountID string     `json:"account_id" validate:"omitempty,uuid"`
	PaidAt    *time.Time `json:"paid_at"`
}

// ----- Bulk transaction actions -----

// BulkSettleRequest marks several transactions paid in one call. AccountID and
// PaidAt are optional: when AccountID is set the rows are moved to that account;
// PaidAt defaults to now.
type BulkSettleRequest struct {
	IDs       []string   `json:"ids" validate:"required,min=1,dive,uuid"`
	AccountID string     `json:"account_id" validate:"omitempty,uuid"`
	PaidAt    *time.Time `json:"paid_at"`
}

// BulkIDsRequest is the shared shape for bulk operations that only need ids
// (unsettle, delete).
type BulkIDsRequest struct {
	IDs []string `json:"ids" validate:"required,min=1,dive,uuid"`
}

// BulkCategorizeRequest assigns one category to several transactions.
type BulkCategorizeRequest struct {
	IDs        []string `json:"ids" validate:"required,min=1,dive,uuid"`
	CategoryID string   `json:"category_id" validate:"required,uuid"`
}

// BulkResult reports how many rows a bulk operation affected.
type BulkResult struct {
	Updated int64 `json:"updated"`
}

// TransactionFilter captures query params for listing transactions.
type TransactionFilter struct {
	Type       string
	AccountID  string
	CategoryID string
	ContactID  string
	TagID      string
	From       *time.Time
	To         *time.Time
	Search     string
	Page       int
	PerPage    int
}

// PayableFilter captures query params for the AP/AR endpoints. Status is one of
// open | overdue | paid | all. From/To window the due date.
type PayableFilter struct {
	Status    string
	ContactID string
	From      *time.Time
	To        *time.Time
	Page      int
	PerPage   int
}

// PayablesSummary aggregates open/overdue totals (cents) for the AP/AR widgets.
type PayablesSummary struct {
	TotalOpen    int64 `json:"total_open"`
	TotalOverdue int64 `json:"total_overdue"`
	DueNext7d    int64 `json:"due_next_7d"`
	CountOpen    int64 `json:"count_open"`
}

// PayablesMeta carries pagination plus the AP/AR summary in the response meta.
type PayablesMeta struct {
	Page    int             `json:"page"`
	PerPage int             `json:"per_page"`
	Total   int64           `json:"total"`
	Pages   int             `json:"pages"`
	Summary PayablesSummary `json:"summary"`
}

// ----- Common -----

type PageMeta struct {
	Page    int   `json:"page"`
	PerPage int   `json:"per_page"`
	Total   int64 `json:"total"`
	Pages   int   `json:"pages"`
}

// ----- Dashboard -----

type DashboardResponse struct {
	Balance         int64            `json:"balance"`
	MonthIncome     int64            `json:"month_income"`
	MonthExpense    int64            `json:"month_expense"`
	MonthNet        int64            `json:"month_net"`
	UpcomingBills   []TransactionDTO `json:"upcoming_bills"`
	CashFlow        []CashFlowPoint  `json:"cash_flow"`
	TopCategories   []CategorySpend  `json:"top_categories"`
	AccountsSummary []AccountDTO     `json:"accounts_summary"`
}

type CashFlowPoint struct {
	Month   string `json:"month"`
	Income  int64  `json:"income"`
	Expense int64  `json:"expense"`
}

type CategorySpend struct {
	CategoryID string `json:"category_id"`
	Name       string `json:"name"`
	Color      string `json:"color"`
	Total      int64  `json:"total"`
}

// ----- Reports summary (Relatórios page) -----

// ReportSummary is the aggregated payload for the frontend Relatórios page: the
// period totals plus breakdowns by category and by month. All amounts in cents.
type ReportSummary struct {
	Income     int64                 `json:"income"`
	Expense    int64                 `json:"expense"`
	Net        int64                 `json:"net"`
	ByCategory []ReportCategoryTotal `json:"by_category"`
	ByMonth    []ReportMonthTotal    `json:"by_month"`
}

// ReportCategoryTotal is one category's settled total within the period. Kind is
// "income" or "expense"; uncategorised rows are omitted.
type ReportCategoryTotal struct {
	CategoryID string `json:"category_id"`
	Name       string `json:"name"`
	Color      string `json:"color"`
	Kind       string `json:"kind"`
	Total      int64  `json:"total"`
}

// ReportMonthTotal is the settled income/expense for one YYYY-MM in the period.
type ReportMonthTotal struct {
	Month   string `json:"month"`
	Income  int64  `json:"income"`
	Expense int64  `json:"expense"`
}

// ----- Cash-flow forecast -----

type ForecastResponse struct {
	CurrentBalance int64           `json:"current_balance"`
	Months         []ForecastMonth `json:"months"`
	EndBalance     int64           `json:"end_balance"`
	Lowest         ForecastLowest  `json:"lowest"`
	Alerts         []ForecastAlert `json:"alerts"`
}

type ForecastMonth struct {
	Month            string `json:"month"` // YYYY-MM
	Inflow           int64  `json:"inflow"`
	Outflow          int64  `json:"outflow"`
	Net              int64  `json:"net"`
	ProjectedBalance int64  `json:"projected_balance"`
}

type ForecastLowest struct {
	Month   string `json:"month"`
	Balance int64  `json:"balance"`
}

type ForecastAlert struct {
	Month   string `json:"month"`
	Message string `json:"message"`
}

// ----- Credit Cards -----

type CreditCardRequest struct {
	Name       string `json:"name" validate:"required,max=120"`
	Limit      int64  `json:"limit" validate:"gte=0"`
	ClosingDay int    `json:"closing_day" validate:"required,min=1,max=28"`
	DueDay     int    `json:"due_day" validate:"required,min=1,max=28"`
	Color      string `json:"color" validate:"max=9"`
}

type CreditCardDTO struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Limit      int64  `json:"limit"`
	ClosingDay int    `json:"closing_day"`
	DueDay     int    `json:"due_day"`
	Color      string `json:"color"`
	Used       int64  `json:"used"`
	Available  int64  `json:"available"`
}

// ----- Credit-card invoices (faturas) -----

// InvoiceDTO is one billing cycle of a credit card, computed by grouping the
// card's transactions by their invoice reference. Amounts are in cents.
//
//	Reference  — closing year-month "YYYY-MM" that identifies the invoice.
//	Total      — sum of all expenses in the cycle.
//	PaidTotal  — sum of the paid ones; OpenTotal = Total - PaidTotal.
//	Status     — open | closed | paid | overdue (see service for the rule).
type InvoiceDTO struct {
	Reference        string    `json:"reference"`
	PeriodStart      time.Time `json:"period_start"`
	PeriodEnd        time.Time `json:"period_end"`
	DueDate          time.Time `json:"due_date"`
	Total            int64     `json:"total"`
	PaidTotal        int64     `json:"paid_total"`
	OpenTotal        int64     `json:"open_total"`
	Status           string    `json:"status"`
	TransactionCount int       `json:"transaction_count"`
}

// InvoiceDetailDTO is a single invoice plus the card transactions in its cycle.
type InvoiceDetailDTO struct {
	Invoice      InvoiceDTO       `json:"invoice"`
	Transactions []TransactionDTO `json:"transactions"`
}

// InvoicePayRequest settles every transaction in an invoice: marks them paid,
// records paid_at (default now) and moves them to AccountID (the paying bank
// account, validated against the org) so the balance reflects the payment.
type InvoicePayRequest struct {
	AccountID string     `json:"account_id" validate:"required,uuid"`
	PaidAt    *time.Time `json:"paid_at"`
}

// ----- Goals -----

type GoalRequest struct {
	Name          string     `json:"name" validate:"required,max=120"`
	TargetAmount  int64      `json:"target_amount" validate:"required,gt=0"`
	CurrentAmount int64      `json:"current_amount" validate:"gte=0"`
	Deadline      *time.Time `json:"deadline"`
	Color         string     `json:"color" validate:"max=9"`
}

type GoalDTO struct {
	ID            string     `json:"id"`
	Name          string     `json:"name"`
	TargetAmount  int64      `json:"target_amount"`
	CurrentAmount int64      `json:"current_amount"`
	Deadline      *time.Time `json:"deadline,omitempty"`
	Color         string     `json:"color"`
	Progress      float64    `json:"progress"`
}

// ----- Budgets -----

type BudgetRequest struct {
	CategoryID string `json:"category_id" validate:"required,uuid"`
	Amount     int64  `json:"amount" validate:"required,gt=0"`
	Month      int    `json:"month" validate:"required,min=1,max=12"`
	Year       int    `json:"year" validate:"required,min=2000,max=2100"`
}

type BudgetDTO struct {
	ID         string       `json:"id"`
	CategoryID string       `json:"category_id"`
	Category   *CategoryDTO `json:"category,omitempty"`
	Amount     int64        `json:"amount"`
	Month      int          `json:"month"`
	Year       int          `json:"year"`
	Spent      int64        `json:"spent"`
	Percent    float64      `json:"percent"`
}

// ----- Members & Invitations -----

type MemberDTO struct {
	ID   string  `json:"id"` // membership id
	User UserDTO `json:"user"`
	Role string  `json:"role"`
}

type UpdateMemberRequest struct {
	Role string `json:"role" validate:"required,oneof=owner admin member viewer"`
}

type InvitationRequest struct {
	Email string `json:"email" validate:"required,email"`
	Role  string `json:"role" validate:"required,oneof=admin member viewer"`
}

type InvitationDTO struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	Accepted  bool      `json:"accepted"`
	Token     string    `json:"token,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type AcceptInvitationRequest struct {
	Token string `json:"token" validate:"required"`
}

// ----- Notifications -----

type NotificationDTO struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	Title     string    `json:"title"`
	Message   string    `json:"message"`
	Read      bool      `json:"read"`
	CreatedAt time.Time `json:"created_at"`
}

// ----- Password reset -----

type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type ResetPasswordRequest struct {
	Token    string `json:"token" validate:"required"`
	Password string `json:"password" validate:"required,min=8,max=72"`
}

// ChangePasswordRequest changes the authenticated user's password. Requires the
// current password (anti-CSRF / re-auth guard); on success clears
// must_change_password and revokes every other refresh token (the optional
// keep_refresh_token preserves the caller's current session).
type ChangePasswordRequest struct {
	CurrentPassword  string `json:"current_password" validate:"required"`
	NewPassword      string `json:"new_password" validate:"required,min=8,max=72"`
	KeepRefreshToken string `json:"keep_refresh_token"`
}

// ----- Email verification -----

type VerifyEmailRequest struct {
	Token string `json:"token" validate:"required"`
}

type ResendVerificationRequest struct {
	Email string `json:"email" validate:"required,email"`
}

// ----- Two-factor authentication (TOTP) -----

// TwoFactorSetupResponse returns the freshly generated secret and otpauth URL so
// the client can render a QR code.
type TwoFactorSetupResponse struct {
	Secret     string `json:"secret"`
	OTPAuthURL string `json:"otpauth_url"`
}

// TwoFactorCodeRequest carries a 6-digit TOTP (used for enable/disable).
type TwoFactorCodeRequest struct {
	Code string `json:"code" validate:"required"`
}

// TwoFactorEnableResponse returns the one-time recovery codes (shown once).
type TwoFactorEnableResponse struct {
	RecoveryCodes []string `json:"recovery_codes"`
}

// ----- Sessions (refresh tokens) -----

// SessionDTO is a single active session derived from a refresh token. The token
// hash is never exposed.
type SessionDTO struct {
	ID        string    `json:"id"`
	UserAgent string    `json:"user_agent"`
	IP        string    `json:"ip"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// RevokeOtherSessionsRequest optionally carries the caller's current raw refresh
// token so it is kept alive while every other session is revoked. The access
// token cannot identify the current refresh token, so the client passes it
// explicitly. When omitted, ALL sessions are revoked and the client must
// re-login.
type RevokeOtherSessionsRequest struct {
	KeepRefreshToken string `json:"keep_refresh_token"`
}

// ----- LGPD: account deletion -----

type DeleteAccountRequest struct {
	Password string `json:"password" validate:"required"`
}

// ----- Audit logs (auditoria) -----

// UserMiniDTO is the minimal id+name+email shape used to identify the actor of
// an audit log entry. nil on the AuditLogDTO when the log carries no user.
type UserMiniDTO struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// AuditLogDTO is the public shape of one audit log entry. User is nil when the
// log has no associated user (e.g. an anonymous/system action).
type AuditLogDTO struct {
	ID        string       `json:"id"`
	CreatedAt time.Time    `json:"created_at"`
	Action    string       `json:"action"`
	Entity    string       `json:"entity,omitempty"`
	EntityID  string       `json:"entity_id,omitempty"`
	IP        string       `json:"ip,omitempty"`
	User      *UserMiniDTO `json:"user,omitempty"`
	Metadata  string       `json:"metadata,omitempty"`
}

// AuditFilter captures query params for listing audit logs. Action/Entity are
// exact matches; From/To window created_at.
type AuditFilter struct {
	Action  string
	Entity  string
	UserID  *uuid.UUID
	From    *time.Time
	To      *time.Time
	Page    int
	PerPage int
}

// ----- Platform back-office (super-admin) -----

// AdminStats is the global counters payload for GET /admin/stats.
type AdminStats struct {
	Organizations int64 `json:"organizations"`
	Users         int64 `json:"users"`
	Transactions  int64 `json:"transactions"`
}

// AdminOrgOwnerDTO is the minimal owner shape embedded in an AdminOrgDTO.
type AdminOrgOwnerDTO struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// AdminOrgDTO is one organization row in the back-office org listing.
type AdminOrgDTO struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Slug         string            `json:"slug"`
	Currency     string            `json:"currency"`
	Owner        *AdminOrgOwnerDTO `json:"owner,omitempty"`
	Members      int64             `json:"members"`
	Transactions int64             `json:"transactions"`
	CreatedAt    time.Time         `json:"created_at"`
}

// AdminResetPasswordRequest — super-admin resets a user's password. The user
// will be forced to change it on their next login (must_change_password=true).
type AdminResetPasswordRequest struct {
	NewPassword string `json:"new_password" validate:"required,min=8,max=72"`
}

// AdminUserDTO is one user row in the back-office user listing.
type AdminUserDTO struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Email         string    `json:"email"`
	SuperAdmin    bool      `json:"super_admin"`
	Disabled      bool      `json:"disabled"`
	EmailVerified bool      `json:"email_verified"`
	CreatedAt     time.Time `json:"created_at"`
	Organizations []string  `json:"organizations"`
}

// ----- First-run setup wizard -----

// SetupStatusResponse is the public payload of GET /setup/status. needs_setup
// is true when the platform has zero users — the SPA then routes to the wizard.
type SetupStatusResponse struct {
	NeedsSetup bool `json:"needs_setup"`
}

// SetupUser is the user block of the initialize request.
type SetupUser struct {
	Name     string `json:"name" validate:"required,min=1,max=120"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=72"`
}

// SetupOrganization is the organization block of the initialize request. The
// currency must be a supported ISO-4217 code.
type SetupOrganization struct {
	Name     string `json:"name" validate:"required,min=1,max=120"`
	Currency string `json:"currency" validate:"required,len=3"`
}

// SetupInitializeRequest is the body of POST /setup/initialize. It bootstraps
// the platform with the first super-admin user and their first organization.
type SetupInitializeRequest struct {
	User         SetupUser         `json:"user" validate:"required"`
	Organization SetupOrganization `json:"organization" validate:"required"`
}

// ----- Global search -----

// SearchNamedDTO is the minimal id+name shape used by global-search groups that
// only need to identify and label a hit (credit_cards, goals).
type SearchNamedDTO struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// SearchGroups holds the per-domain hits of a global search. Each group is
// org-scoped and capped at the requested limit. Lists are never nil so the JSON
// always carries [] (empty) instead of null.
type SearchGroups struct {
	Transactions []TransactionDTO `json:"transactions"`
	Contacts     []ContactDTO     `json:"contacts"`
	Categories   []CategoryDTO    `json:"categories"`
	Accounts     []AccountDTO     `json:"accounts"`
	CreditCards  []SearchNamedDTO `json:"credit_cards"`
	Goals        []SearchNamedDTO `json:"goals"`
}

// SearchResult is the response of GET /search: the echoed query, the grouped
// results and the total number of hits (sum of group lengths).
type SearchResult struct {
	Query   string       `json:"query"`
	Results SearchGroups `json:"results"`
	Total   int          `json:"total"`
}
