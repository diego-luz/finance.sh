package services

import (
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/finance-sh/finance-sh/internal/entities"
	"github.com/finance-sh/finance-sh/internal/money"
	"github.com/finance-sh/finance-sh/pkg/crypto"
	"github.com/finance-sh/finance-sh/pkg/validator"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ErrImportInvalid is returned when the uploaded document can't be parsed as a
// finance.sh export.
var ErrImportInvalid = errors.New("arquivo de importação inválido")

// Caps per collection of an LGPD import. The file is already capped at 25 MB;
// these stop it from being one huge list inserted row by row in a single
// transaction.
const (
	maxImportCatalog      = 2000   // accounts, categories, cards, goals
	maxImportContacts     = 20000  // contacts, budgets
	maxImportTransactions = 200000 // transactions
)

// importValid applies the same struct-tag rules as the regular endpoints'
// DTOs. The import used to take any value: negative amounts, unknown types and
// currencies, unbounded names and colors.
func importValid(v interface{}) bool { return len(validator.Validate(v)) == 0 }

// ImportSummary reports how many rows were restored.
type ImportSummary struct {
	OrganizationID   string `json:"organization_id"`
	OrganizationName string `json:"organization_name"`
	Accounts         int    `json:"accounts"`
	Categories       int    `json:"categories"`
	Contacts         int    `json:"contacts"`
	CreditCards      int    `json:"credit_cards"`
	Budgets          int    `json:"budgets"`
	Goals            int    `json:"goals"`
	Transactions     int    `json:"transactions"`
	Skipped          int    `json:"skipped"`
}

// ---- parse structs (subset of the export we restore). JSON keys mirror
// LGPDService.ExportData. ----

type impDoc struct {
	ExportVersion string        `json:"export_version"`
	Organizations []impOrg      `json:"organizations"`
	Accounts      []impAccount  `json:"accounts"`
	Categories    []impCategory `json:"categories"`
	Contacts      []impContact  `json:"contacts"`
	CreditCards   []impCard     `json:"credit_cards"`
	Budgets       []impBudget   `json:"budgets"`
	Goals         []impGoal     `json:"goals"`
	Transactions  []impTx       `json:"transactions"`
}

type impOrg struct {
	Name     string `json:"name"`
	Currency string `json:"currency"`
}
type impAccount struct {
	ID             string `json:"id"`
	Name           string `json:"name" validate:"required,max=120"`
	Type           string `json:"type" validate:"omitempty,oneof=bank wallet investment credit_card"`
	InitialBalance int64  `json:"initial_balance"`
	Color          string `json:"color" validate:"max=9"`
	Icon           string `json:"icon" validate:"max=40"`
	Archived       bool   `json:"archived"`
}
type impCategory struct {
	ID    string `json:"id"`
	Name  string `json:"name" validate:"required,max=120"`
	Kind  string `json:"kind" validate:"omitempty,oneof=income expense"`
	Color string `json:"color" validate:"max=9"`
	Icon  string `json:"icon" validate:"max=40"`
}
type impContact struct {
	ID       string `json:"id"`
	Name     string `json:"name" validate:"required,max=120"`
	Type     string `json:"type" validate:"omitempty,oneof=customer supplier both"`
	Document string `json:"document" validate:"max=20"`
	Email    string `json:"email" validate:"omitempty,email"`
	Phone    string `json:"phone" validate:"max=30"`
	Notes    string `json:"notes" validate:"max=500"`
}
type impCard struct {
	ID         string `json:"id"`
	Name       string `json:"name" validate:"required,max=120"`
	Limit      int64  `json:"limit" validate:"gte=0"`
	ClosingDay int    `json:"closing_day" validate:"min=1,max=31"`
	DueDay     int    `json:"due_day" validate:"min=1,max=31"`
	Color      string `json:"color" validate:"max=9"`
}
type impBudget struct {
	CategoryID string `json:"category_id"`
	Amount     int64  `json:"amount" validate:"gt=0"`
	Month      int    `json:"month" validate:"min=1,max=12"`
	Year       int    `json:"year" validate:"min=2000,max=2100"`
}
type impGoal struct {
	Name          string     `json:"name" validate:"required,max=120"`
	TargetAmount  int64      `json:"target_amount" validate:"gt=0"`
	CurrentAmount int64      `json:"current_amount" validate:"gte=0"`
	Deadline      *time.Time `json:"deadline"`
	Color         string     `json:"color" validate:"max=9"`
}
type impTx struct {
	AccountID string `json:"account_id"`
	// TransferAccountID is the destination of a transfer (exports before this
	// field existed have none: such transfers are skipped)
	TransferAccountID string     `json:"transfer_account_id"`
	CategoryID        string     `json:"category_id"`
	ContactID         string     `json:"contact_id"`
	Type              string     `json:"type" validate:"required,oneof=income expense transfer"`
	Amount            int64      `json:"amount" validate:"gt=0"`
	Description       string     `json:"description" validate:"max=200"`
	Date              time.Time  `json:"date"`
	DueDate           *time.Time `json:"due_date"`
	Paid              bool       `json:"paid"`
	PaidAt            *time.Time `json:"paid_at"`
	Recurring         bool       `json:"recurring"`
	Notes             string     `json:"notes" validate:"max=500"`
}

// ImportData restores a finance.sh export into a BRAND-NEW organization owned by
// the importing user. Old UUIDs are remapped to fresh ones so foreign keys stay
// consistent. Everything runs in a single transaction; on any error nothing is
// written. Notifications/audit logs are intentionally not restored.
func (s *LGPDService) ImportData(userID uuid.UUID, raw []byte) (*ImportSummary, error) {
	var doc impDoc
	if err := json.Unmarshal(raw, &doc); err != nil {
		return nil, ErrImportInvalid
	}
	if len(doc.Accounts) > maxImportCatalog || len(doc.Categories) > maxImportCatalog ||
		len(doc.CreditCards) > maxImportCatalog || len(doc.Goals) > maxImportCatalog ||
		len(doc.Contacts) > maxImportContacts || len(doc.Budgets) > maxImportContacts ||
		len(doc.Transactions) > maxImportTransactions {
		return nil, ErrImportInvalid
	}

	orgName := "Dados importados"
	currency := "BRL"
	if len(doc.Organizations) > 0 {
		if n := strings.TrimSpace(doc.Organizations[0].Name); n != "" {
			if r := []rune(n); len(r) > 100 {
				n = string(r[:100])
			}
			orgName = n + " (importado)"
		}
		if c := strings.TrimSpace(doc.Organizations[0].Currency); money.IsSupported(c) {
			currency = money.Normalize(c)
		}
	}

	sum := &ImportSummary{OrganizationName: orgName}
	err := s.db.Transaction(func(tx *gorm.DB) error {
		org := &entities.Organization{Name: orgName, Slug: slugify(orgName), OwnerID: userID, Currency: currency}
		if err := tx.Create(org).Error; err != nil {
			return err
		}
		if err := tx.Create(&entities.Membership{UserID: userID, OrganizationID: org.ID, Role: entities.RoleOwner}).Error; err != nil {
			return err
		}
		sum.OrganizationID = org.ID.String()

		accMap := map[string]uuid.UUID{}
		for _, a := range doc.Accounts {
			if !importValid(a) {
				sum.Skipped++
				continue
			}
			n := &entities.Account{
				OrganizationID: org.ID, Name: a.Name,
				Type:           entities.AccountType(orDefault(a.Type, "bank")),
				InitialBalance: a.InitialBalance,
				Color:          orDefault(a.Color, "#10b981"), Icon: orDefault(a.Icon, "wallet"),
				Archived: a.Archived,
			}
			if err := tx.Create(n).Error; err != nil {
				return err
			}
			if a.ID != "" {
				accMap[a.ID] = n.ID
			}
			sum.Accounts++
		}

		catMap := map[string]uuid.UUID{}
		for _, c := range doc.Categories {
			if !importValid(c) {
				sum.Skipped++
				continue
			}
			n := &entities.Category{
				OrganizationID: org.ID, Name: c.Name,
				Kind:  entities.CategoryKind(orDefault(c.Kind, "expense")),
				Color: orDefault(c.Color, "#6366f1"), Icon: orDefault(c.Icon, "tag"),
			}
			if err := tx.Create(n).Error; err != nil {
				return err
			}
			if c.ID != "" {
				catMap[c.ID] = n.ID
			}
			sum.Categories++
		}

		contactMap := map[string]uuid.UUID{}
		for _, c := range doc.Contacts {
			if !importValid(c) {
				sum.Skipped++
				continue
			}
			n := &entities.Contact{
				OrganizationID: org.ID, Name: c.Name,
				Type:     entities.ContactType(orDefault(c.Type, "both")),
				Document: c.Document, Email: c.Email, Phone: c.Phone, Notes: c.Notes,
			}
			if err := tx.Create(n).Error; err != nil {
				return err
			}
			if c.ID != "" {
				contactMap[c.ID] = n.ID
			}
			sum.Contacts++
		}

		for _, c := range doc.CreditCards {
			if !importValid(c) {
				sum.Skipped++
				continue
			}
			n := &entities.CreditCard{
				OrganizationID: org.ID, Name: c.Name, Limit: c.Limit,
				ClosingDay: c.ClosingDay, DueDay: c.DueDay, Color: orDefault(c.Color, "#0f1115"),
			}
			if err := tx.Create(n).Error; err != nil {
				return err
			}
			sum.CreditCards++
		}

		for _, b := range doc.Budgets {
			if !importValid(b) {
				sum.Skipped++
				continue
			}
			cat, ok := catMap[b.CategoryID]
			if !ok {
				sum.Skipped++
				continue
			}
			n := &entities.Budget{OrganizationID: org.ID, CategoryID: cat, Amount: b.Amount, Month: b.Month, Year: b.Year}
			if err := tx.Create(n).Error; err != nil {
				return err
			}
			sum.Budgets++
		}

		for _, g := range doc.Goals {
			if !importValid(g) {
				sum.Skipped++
				continue
			}
			n := &entities.Goal{
				OrganizationID: org.ID, Name: g.Name, TargetAmount: g.TargetAmount,
				CurrentAmount: g.CurrentAmount, Deadline: g.Deadline, Color: orDefault(g.Color, "#10b981"),
			}
			if err := tx.Create(n).Error; err != nil {
				return err
			}
			sum.Goals++
		}

		novas := make([]*entities.Transaction, 0, len(doc.Transactions))
		for _, t := range doc.Transactions {
			acc, ok := accMap[t.AccountID]
			if !ok || t.Date.IsZero() || !importValid(t) {
				sum.Skipped++
				continue
			}
			// a transfer needs its destination among the imported accounts
			var destino *uuid.UUID
			if t.Type == string(entities.TxTransfer) {
				d, ok := accMap[t.TransferAccountID]
				if !ok || d == acc {
					sum.Skipped++
					continue
				}
				destino = &d
			}
			n := &entities.Transaction{
				OrganizationID: org.ID, AccountID: acc,
				Type: entities.TransactionType(t.Type), Amount: t.Amount,
				Description: orDefault(t.Description, "Importado"),
				Date:        t.Date, DueDate: t.DueDate, Paid: t.Paid, PaidAt: t.PaidAt,
				Recurring: t.Recurring, Notes: crypto.EncryptedString(t.Notes),
				TransferAccountID: destino,
			}
			if id, ok := catMap[t.CategoryID]; ok {
				n.CategoryID = &id
			}
			if id, ok := contactMap[t.ContactID]; ok {
				n.ContactID = &id
			}
			novas = append(novas, n)
		}
		// in batches: one INSERT per row made a large export a long transaction
		if len(novas) > 0 {
			if err := tx.CreateInBatches(novas, 500).Error; err != nil {
				return err
			}
		}
		sum.Transactions = len(novas)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return sum, nil
}

func orDefault(s, def string) string {
	if strings.TrimSpace(s) == "" {
		return def
	}
	return s
}
