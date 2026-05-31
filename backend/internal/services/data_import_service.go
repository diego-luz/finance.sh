package services

import (
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/finance-sh/finance-sh/internal/entities"
	"github.com/finance-sh/finance-sh/pkg/crypto"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ErrImportInvalid is returned when the uploaded document can't be parsed as a
// finance.sh export.
var ErrImportInvalid = errors.New("arquivo de importação inválido")

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
	Name           string `json:"name"`
	Type           string `json:"type"`
	InitialBalance int64  `json:"initial_balance"`
	Color          string `json:"color"`
	Icon           string `json:"icon"`
	Archived       bool   `json:"archived"`
}
type impCategory struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Kind  string `json:"kind"`
	Color string `json:"color"`
	Icon  string `json:"icon"`
}
type impContact struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	Document string `json:"document"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Notes    string `json:"notes"`
}
type impCard struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Limit      int64  `json:"limit"`
	ClosingDay int    `json:"closing_day"`
	DueDay     int    `json:"due_day"`
	Color      string `json:"color"`
}
type impBudget struct {
	CategoryID string `json:"category_id"`
	Amount     int64  `json:"amount"`
	Month      int    `json:"month"`
	Year       int    `json:"year"`
}
type impGoal struct {
	Name          string     `json:"name"`
	TargetAmount  int64      `json:"target_amount"`
	CurrentAmount int64      `json:"current_amount"`
	Deadline      *time.Time `json:"deadline"`
	Color         string     `json:"color"`
}
type impTx struct {
	AccountID   string     `json:"account_id"`
	CategoryID  string     `json:"category_id"`
	ContactID   string     `json:"contact_id"`
	Type        string     `json:"type"`
	Amount      int64      `json:"amount"`
	Description string     `json:"description"`
	Date        time.Time  `json:"date"`
	DueDate     *time.Time `json:"due_date"`
	Paid        bool       `json:"paid"`
	PaidAt      *time.Time `json:"paid_at"`
	Recurring   bool       `json:"recurring"`
	Notes       string     `json:"notes"`
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

	orgName := "Dados importados"
	currency := "BRL"
	if len(doc.Organizations) > 0 {
		if n := strings.TrimSpace(doc.Organizations[0].Name); n != "" {
			orgName = n + " (importado)"
		}
		if c := strings.TrimSpace(doc.Organizations[0].Currency); c != "" {
			currency = c
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
			n := &entities.Goal{
				OrganizationID: org.ID, Name: g.Name, TargetAmount: g.TargetAmount,
				CurrentAmount: g.CurrentAmount, Deadline: g.Deadline, Color: orDefault(g.Color, "#10b981"),
			}
			if err := tx.Create(n).Error; err != nil {
				return err
			}
			sum.Goals++
		}

		for _, t := range doc.Transactions {
			acc, ok := accMap[t.AccountID]
			if !ok || t.Type == "" || t.Date.IsZero() {
				sum.Skipped++
				continue
			}
			n := &entities.Transaction{
				OrganizationID: org.ID, AccountID: acc,
				Type: entities.TransactionType(t.Type), Amount: t.Amount,
				Description: orDefault(t.Description, "Importado"),
				Date:        t.Date, DueDate: t.DueDate, Paid: t.Paid, PaidAt: t.PaidAt,
				Recurring: t.Recurring, Notes: crypto.EncryptedString(t.Notes),
			}
			if id, ok := catMap[t.CategoryID]; ok {
				n.CategoryID = &id
			}
			if id, ok := contactMap[t.ContactID]; ok {
				n.ContactID = &id
			}
			if err := tx.Create(n).Error; err != nil {
				return err
			}
			sum.Transactions++
		}
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
