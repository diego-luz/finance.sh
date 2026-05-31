package database

import (
	"math/rand"
	"strconv"
	"time"

	"github.com/finance-sh/finance-sh/internal/cards"
	"github.com/finance-sh/finance-sh/internal/entities"
	"github.com/finance-sh/finance-sh/pkg/hash"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	demoEmail       = "demo@finance.sh"
	superAdminEmail = "super@finance.sh"
)

// Seed inserts a demo tenant with realistic data so the dashboard is populated
// on a fresh database, and a single platform super-admin. Both steps are
// idempotent: if the respective user already exists they are skipped.
func Seed(db *gorm.DB) error {
	// Platform super-admin (back-office). Seeded independently of the demo tenant
	// so it exists even on databases that already carry the demo data.
	if err := SeedSuperAdmin(db); err != nil {
		return err
	}

	var count int64
	db.Model(&entities.User{}).Where("email = ?", demoEmail).Count(&count)
	if count > 0 {
		return nil
	}

	pw, err := hash.Password("senha123")
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	return db.Transaction(func(tx *gorm.DB) error {
		user := &entities.User{
			Name:          "Demo finance.sh",
			Email:         demoEmail,
			PasswordHash:  pw,
			EmailVerified: true,
			// LGPD consent so the demo account looks like a properly-onboarded user.
			TermsAcceptedAt:  &now,
			TermsVersion:     "1.0",
			TwoFactorEnabled: false, // explicit: demo must NOT require 2FA
		}
		if err := tx.Create(user).Error; err != nil {
			return err
		}

		org := &entities.Organization{
			Name:     "finance.sh Demo",
			Slug:     "finance-sh-demo",
			OwnerID:  user.ID,
			Currency: "BRL",
		}
		if err := tx.Create(org).Error; err != nil {
			return err
		}

		if err := tx.Create(&entities.Membership{
			UserID:         user.ID,
			OrganizationID: org.ID,
			Role:           entities.RoleOwner,
		}).Error; err != nil {
			return err
		}

		// Accounts.
		accounts := []*entities.Account{
			{OrganizationID: org.ID, Name: "Conta Corrente", Type: entities.AccountBank, InitialBalance: 350000, Color: "#10b981", Icon: "bank"},
			{OrganizationID: org.ID, Name: "Carteira", Type: entities.AccountWallet, InitialBalance: 20000, Color: "#f59e0b", Icon: "wallet"},
			{OrganizationID: org.ID, Name: "Investimentos", Type: entities.AccountInvestment, InitialBalance: 1500000, Color: "#6366f1", Icon: "trending-up"},
		}
		for _, a := range accounts {
			if err := tx.Create(a).Error; err != nil {
				return err
			}
		}

		// Categories (income + expense) with colors/icons.
		incomeCats := []*entities.Category{
			{OrganizationID: org.ID, Name: "Salário", Kind: entities.CategoryIncome, Color: "#10b981", Icon: "briefcase"},
			{OrganizationID: org.ID, Name: "Freelance", Kind: entities.CategoryIncome, Color: "#22c55e", Icon: "laptop"},
			{OrganizationID: org.ID, Name: "Rendimentos", Kind: entities.CategoryIncome, Color: "#14b8a6", Icon: "trending-up"},
		}
		expenseCats := []*entities.Category{
			{OrganizationID: org.ID, Name: "Moradia", Kind: entities.CategoryExpense, Color: "#ef4444", Icon: "home"},
			{OrganizationID: org.ID, Name: "Alimentação", Kind: entities.CategoryExpense, Color: "#f97316", Icon: "utensils"},
			{OrganizationID: org.ID, Name: "Transporte", Kind: entities.CategoryExpense, Color: "#eab308", Icon: "car"},
			{OrganizationID: org.ID, Name: "Lazer", Kind: entities.CategoryExpense, Color: "#a855f7", Icon: "film"},
			{OrganizationID: org.ID, Name: "Saúde", Kind: entities.CategoryExpense, Color: "#ec4899", Icon: "heart-pulse"},
			{OrganizationID: org.ID, Name: "Assinaturas", Kind: entities.CategoryExpense, Color: "#3b82f6", Icon: "repeat"},
		}
		for _, c := range append(incomeCats, expenseCats...) {
			if err := tx.Create(c).Error; err != nil {
				return err
			}
		}

		// Categorization rules so the demo shows automatic categorization in
		// action. All "contains" (case-insensitive), mapped to the demo
		// categories above.
		rules := []*entities.CategoryRule{
			{OrganizationID: org.ID, Pattern: "salário", MatchType: entities.MatchContains, CategoryID: incomeCats[0].ID, Priority: 10, Active: true},
			{OrganizationID: org.ID, Pattern: "aluguel", MatchType: entities.MatchContains, CategoryID: expenseCats[0].ID, Priority: 10, Active: true},
			{OrganizationID: org.ID, Pattern: "supermercado", MatchType: entities.MatchContains, CategoryID: expenseCats[1].ID, Priority: 10, Active: true},
			{OrganizationID: org.ID, Pattern: "mercado", MatchType: entities.MatchContains, CategoryID: expenseCats[1].ID, Priority: 5, Active: true},
			{OrganizationID: org.ID, Pattern: "combustível", MatchType: entities.MatchContains, CategoryID: expenseCats[2].ID, Priority: 10, Active: true},
			{OrganizationID: org.ID, Pattern: "uber", MatchType: entities.MatchContains, CategoryID: expenseCats[2].ID, Priority: 10, Active: true},
			{OrganizationID: org.ID, Pattern: "farmácia", MatchType: entities.MatchContains, CategoryID: expenseCats[4].ID, Priority: 10, Active: true},
			{OrganizationID: org.ID, Pattern: "assinatura", MatchType: entities.MatchContains, CategoryID: expenseCats[5].ID, Priority: 10, Active: true},
			{OrganizationID: org.ID, Pattern: "streaming", MatchType: entities.MatchContains, CategoryID: expenseCats[5].ID, Priority: 10, Active: true},
		}
		for _, rl := range rules {
			if err := tx.Create(rl).Error; err != nil {
				return err
			}
		}

		// Contacts (clientes/fornecedores) for payables/receivables.
		contacts := []*entities.Contact{
			{OrganizationID: org.ID, Name: "Imobiliária Centro", Type: entities.ContactSupplier, Document: "12.345.678/0001-90", Email: "contato@imobcentro.com.br", Phone: "(41) 3333-1010"},
			{OrganizationID: org.ID, Name: "Mercado Bom Preço", Type: entities.ContactSupplier, Document: "98.765.432/0001-21", Phone: "(41) 3333-2020"},
			{OrganizationID: org.ID, Name: "Cliente Acme Ltda", Type: entities.ContactCustomer, Document: "11.222.333/0001-44", Email: "financeiro@acme.com.br"},
		}
		for _, c := range contacts {
			if err := tx.Create(c).Error; err != nil {
				return err
			}
		}

		// Credit cards.
		cards := []*entities.CreditCard{
			{OrganizationID: org.ID, Name: "Cartão Nubank", Limit: 500000, ClosingDay: 20, DueDay: 27, Color: "#820ad1"},
			{OrganizationID: org.ID, Name: "Cartão Inter", Limit: 300000, ClosingDay: 5, DueDay: 12, Color: "#ff7a00"},
		}
		for _, c := range cards {
			if err := tx.Create(c).Error; err != nil {
				return err
			}
		}

		// Tags (rótulos) so the demo shows the tagging feature.
		tags := []*entities.Tag{
			{OrganizationID: org.ID, Name: "reembolsável", Color: "#3b82f6"},
			{OrganizationID: org.ID, Name: "fixo", Color: "#ef4444"},
			{OrganizationID: org.ID, Name: "pessoal", Color: "#a855f7"},
			{OrganizationID: org.ID, Name: "empresa", Color: "#10b981"},
		}
		for _, t := range tags {
			if err := tx.Create(t).Error; err != nil {
				return err
			}
		}

		// Savings goals.
		deadline := now.AddDate(1, 0, 0)
		goals := []*entities.Goal{
			{OrganizationID: org.ID, Name: "Reserva de emergência", TargetAmount: 2000000, CurrentAmount: 750000, Deadline: &deadline, Color: "#10b981"},
			{OrganizationID: org.ID, Name: "Viagem de férias", TargetAmount: 800000, CurrentAmount: 320000, Color: "#6366f1"},
		}
		for _, g := range goals {
			if err := tx.Create(g).Error; err != nil {
				return err
			}
		}

		// Budgets for the current month (expense categories).
		budgets := []*entities.Budget{
			{OrganizationID: org.ID, CategoryID: expenseCats[0].ID, Amount: 250000, Month: int(now.Month()), Year: now.Year()},
			{OrganizationID: org.ID, CategoryID: expenseCats[1].ID, Amount: 90000, Month: int(now.Month()), Year: now.Year()},
			{OrganizationID: org.ID, CategoryID: expenseCats[2].ID, Amount: 40000, Month: int(now.Month()), Year: now.Year()},
		}
		for _, b := range budgets {
			if err := tx.Create(b).Error; err != nil {
				return err
			}
		}

		// In-app notifications.
		notifications := []*entities.Notification{
			{OrganizationID: org.ID, Type: "vencimento", Title: "Conta a vencer", Message: "A fatura do cartão vence em 3 dias."},
			{OrganizationID: org.ID, Type: "orcamento", Title: "Orçamento de Alimentação", Message: "Você já usou 80% do orçamento de Alimentação."},
			{OrganizationID: org.ID, Type: "info", Title: "Bem-vindo ao finance.sh", Message: "Sua conta demo está pronta para uso.", Read: true},
		}
		for _, n := range notifications {
			if err := tx.Create(n).Error; err != nil {
				return err
			}
		}

		// Recurrence rules (the proper recurring-transaction engine). Two monthly
		// rules anchored to early NEXT month so the demo shows UPCOMING recurrences
		// (NextRunDate in the future) rather than back-dated ones. The worker
		// materialises them once their NextRunDate falls due.
		firstNextMonth := time.Date(now.Year(), now.Month(), 1, 12, 0, 0, 0, time.UTC).AddDate(0, 1, 0)
		rentRunDate := firstNextMonth.AddDate(0, 0, 9)   // ~10th of next month
		salaryRunDate := firstNextMonth.AddDate(0, 0, 4) // ~5th of next month
		recurrences := []*entities.RecurrenceRule{
			{
				OrganizationID: org.ID,
				Type:           entities.TxExpense,
				Amount:         230000, // R$ 2.300,00
				Description:    "Aluguel",
				AccountID:      accounts[0].ID,
				CategoryID:     ptr(expenseCats[0].ID), // Moradia
				ContactID:      ptr(contacts[0].ID),    // Imobiliária Centro
				Paid:           false,
				Frequency:      entities.FreqMonthly,
				Interval:       1,
				StartDate:      rentRunDate,
				MaxOccurrences: 0,
				NextRunDate:    rentRunDate,
				Active:         true,
			},
			{
				OrganizationID: org.ID,
				Type:           entities.TxIncome,
				Amount:         850000, // R$ 8.500,00
				Description:    "Salário",
				AccountID:      accounts[0].ID,
				CategoryID:     ptr(incomeCats[0].ID), // Salário
				Paid:           true,
				Frequency:      entities.FreqMonthly,
				Interval:       1,
				StartDate:      salaryRunDate,
				MaxOccurrences: 0,
				NextRunDate:    salaryRunDate,
				Active:         true,
			},
		}
		for _, rr := range recurrences {
			if err := tx.Create(rr).Error; err != nil {
				return err
			}
		}

		txs := buildSeedTransactions(org.ID, accounts, cards, contacts, incomeCats, expenseCats)
		for _, t := range txs {
			if err := tx.Create(t).Error; err != nil {
				return err
			}
		}

		// Attach a couple of demo tags to some transactions so the join table and
		// the tag filter have data. The first tx is the recurring salary ("fixo");
		// the second is the rent ("fixo" + "pessoal"). Best-effort over the slice.
		if len(txs) > 1 {
			if err := tx.Model(txs[0]).Association("Tags").Append(tags[1]); err != nil { // fixo
				return err
			}
			if err := tx.Model(txs[1]).Association("Tags").Append(tags[1], tags[2]); err != nil { // fixo, pessoal
				return err
			}
		}
		return nil
	})
}

// SeedSuperAdmin creates the single platform super-admin used to access the
// back-office (/admin). It is idempotent: if a user with the super-admin email
// already exists it does nothing. The super-admin is deliberately NOT a member
// of any organization — the platform role is kept separate from per-org RBAC.
//
//	email:    super@finance.sh
//	password: superadmin123
func SeedSuperAdmin(db *gorm.DB) error {
	var count int64
	db.Model(&entities.User{}).Where("email = ?", superAdminEmail).Count(&count)
	if count > 0 {
		return nil
	}

	pw, err := hash.Password("superadmin123")
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	admin := &entities.User{
		Name:             "Super Admin",
		Email:            superAdminEmail,
		PasswordHash:     pw,
		EmailVerified:    true,
		SuperAdmin:       true,
		TwoFactorEnabled: false,
		TermsAcceptedAt:  &now,
		TermsVersion:     "1.0",
	}
	return db.Create(admin).Error
}

// buildSeedTransactions generates ~40 transactions across the last 6 months so
// the dashboard's monthly aggregations and cash-flow chart look alive.
func buildSeedTransactions(orgID uuid.UUID, accts []*entities.Account, cards []*entities.CreditCard, contacts []*entities.Contact, income, expense []*entities.Category) []*entities.Transaction {
	rng := rand.New(rand.NewSource(42)) // deterministic seed for reproducibility
	now := time.Now().UTC()
	checking := accts[0].ID

	var out []*entities.Transaction

	// Recurring income + fixed expenses for each of the last 6 months.
	for m := 5; m >= 0; m-- {
		monthRef := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC).AddDate(0, -m, 0)

		// Salary on the 5th.
		out = append(out, &entities.Transaction{
			OrganizationID: orgID, AccountID: checking,
			CategoryID: ptr(income[0].ID), Type: entities.TxIncome,
			Amount: 850000, Description: "Salário mensal",
			Date: monthRef.AddDate(0, 0, 4), Paid: true, Recurring: true,
		})

		// Rent on the 10th.
		out = append(out, &entities.Transaction{
			OrganizationID: orgID, AccountID: checking,
			CategoryID: ptr(expense[0].ID), Type: entities.TxExpense,
			Amount: 230000, Description: "Aluguel",
			Date: monthRef.AddDate(0, 0, 9), Paid: true, Recurring: true,
		})

		// Streaming subscription on the 15th.
		out = append(out, &entities.Transaction{
			OrganizationID: orgID, AccountID: checking,
			CategoryID: ptr(expense[5].ID), Type: entities.TxExpense,
			Amount: 5990, Description: "Assinatura streaming",
			Date: monthRef.AddDate(0, 0, 14), Paid: true, Recurring: true,
		})

		// A handful of variable expenses.
		variable := []struct {
			cat  *entities.Category
			desc string
			min  int64
			max  int64
		}{
			{expense[1], "Supermercado", 18000, 65000},
			{expense[2], "Combustível", 12000, 30000},
			{expense[3], "Cinema e jantar", 8000, 25000},
			{expense[4], "Farmácia", 3000, 18000},
		}
		for _, vexp := range variable {
			amount := vexp.min + rng.Int63n(vexp.max-vexp.min)
			day := 1 + rng.Intn(27)
			out = append(out, &entities.Transaction{
				OrganizationID: orgID, AccountID: checking,
				CategoryID: ptr(vexp.cat.ID), Type: entities.TxExpense,
				Amount: amount, Description: vexp.desc,
				Date: monthRef.AddDate(0, 0, day), Paid: true,
			})
		}

		// Occasional freelance income.
		if rng.Intn(2) == 0 {
			out = append(out, &entities.Transaction{
				OrganizationID: orgID, AccountID: checking,
				CategoryID: ptr(income[1].ID), Type: entities.TxIncome,
				Amount: 50000 + rng.Int63n(150000), Description: "Projeto freelance",
				Date: monthRef.AddDate(0, 0, 18+rng.Intn(8)), Paid: true,
			})
		}
	}

	// Upcoming unpaid bills (future, expense) so the dashboard widget and the
	// accounts-payable view are filled. Several carry a due_date (vencimento)
	// distinct from the competência date and a fornecedor (contact).
	dueIn := func(days int) *time.Time { d := now.AddDate(0, 0, days); return &d }
	out = append(out,
		&entities.Transaction{
			OrganizationID: orgID, AccountID: checking,
			CategoryID: ptr(expense[0].ID), ContactID: ptr(contacts[0].ID), Type: entities.TxExpense,
			Amount: 230000, Description: "Aluguel (próximo mês)",
			Date: now.AddDate(0, 0, 6), DueDate: dueIn(6), Paid: false, Recurring: true,
		},
		&entities.Transaction{
			OrganizationID: orgID, AccountID: checking,
			CategoryID: ptr(expense[5].ID), Type: entities.TxExpense,
			Amount: 11980, Description: "Fatura do cartão",
			Date: now.AddDate(0, 0, 3), DueDate: dueIn(3), Paid: false,
		},
		&entities.Transaction{
			OrganizationID: orgID, AccountID: checking,
			CategoryID: ptr(expense[2].ID), Type: entities.TxExpense,
			Amount: 18500, Description: "IPVA parcela",
			Date: now.AddDate(0, 0, 12), DueDate: dueIn(12), Paid: false,
		},
		// An overdue payable to a fornecedor (due_date in the past, still unpaid).
		&entities.Transaction{
			OrganizationID: orgID, AccountID: checking,
			CategoryID: ptr(expense[1].ID), ContactID: ptr(contacts[1].ID), Type: entities.TxExpense,
			Amount: 32000, Description: "Compra a prazo - Mercado Bom Preço",
			Date: now.AddDate(0, 0, -20), DueDate: dueIn(-3), Paid: false,
		},
	)

	// Unpaid receivables (income with due_date) so the accounts-receivable view
	// and the forecast inflows are populated.
	out = append(out,
		&entities.Transaction{
			OrganizationID: orgID, AccountID: checking,
			CategoryID: ptr(income[1].ID), ContactID: ptr(contacts[2].ID), Type: entities.TxIncome,
			Amount: 180000, Description: "Nota fiscal Acme - projeto",
			Date: now.AddDate(0, 0, -2), DueDate: dueIn(15), Paid: false,
		},
		&entities.Transaction{
			OrganizationID: orgID, AccountID: checking,
			CategoryID: ptr(income[1].ID), ContactID: ptr(contacts[2].ID), Type: entities.TxIncome,
			Amount: 120000, Description: "Nota fiscal Acme - manutenção",
			Date: now.AddDate(0, 0, 5), DueDate: dueIn(40), Paid: false,
		},
	)

	// Unpaid credit-card purchases (the open invoice) so cards show `used`.
	if len(cards) > 0 {
		out = append(out,
			&entities.Transaction{
				OrganizationID: orgID, AccountID: checking,
				CreditCardID: ptr(cards[0].ID), CategoryID: ptr(expense[1].ID), Type: entities.TxExpense,
				Amount: 45000, Description: "Compra no cartão - mercado",
				Date: now.AddDate(0, 0, -5), Paid: false,
			},
			&entities.Transaction{
				OrganizationID: orgID, AccountID: checking,
				CreditCardID: ptr(cards[0].ID), CategoryID: ptr(expense[3].ID), Type: entities.TxExpense,
				Amount: 18000, Description: "Compra no cartão - streaming anual",
				Date: now.AddDate(0, 0, -2), Paid: false,
			},
		)

		card := cards[0]
		// Card purchases spread across two cycles so the card has a CLOSED invoice
		// (last month, already paid) plus a CURRENT one (this month, unpaid). The
		// closed-cycle purchase is dated ~40 days ago so it falls in a prior fatura.
		closedDate := now.AddDate(0, 0, -40)
		out = append(out,
			&entities.Transaction{
				OrganizationID: orgID, AccountID: checking,
				CreditCardID: ptr(card.ID), CategoryID: ptr(expense[2].ID), Type: entities.TxExpense,
				Amount: 27000, Description: "Combustível no cartão",
				Date: closedDate, Paid: true, PaidAt: &closedDate,
				DueDate: invoiceDuePtr(card, closedDate),
			},
		)

		// A 10x parcelado purchase on the card: "Notebook (i/10)", one parcela per
		// month starting today. Each parcela carries its invoice due date and the
		// shared installment group id. The amount split mirrors the service: the
		// remainder lands on the first parcela so the parcelas sum to the total.
		total := int64(599900) // R$ 5.999,00
		n := 10
		groupID := uuid.New()
		per := total / int64(n)
		remainder := total - per*int64(n)
		for i := 1; i <= n; i++ {
			amount := per
			if i == 1 {
				amount += remainder
			}
			date := now.AddDate(0, i-1, 0)
			out = append(out, &entities.Transaction{
				OrganizationID: orgID, AccountID: checking,
				CreditCardID: ptr(card.ID), CategoryID: ptr(expense[3].ID), Type: entities.TxExpense,
				Amount:      amount,
				Description: "Notebook (" + strconv.Itoa(i) + "/" + strconv.Itoa(n) + ")",
				Date:        date, Paid: false,
				DueDate:            invoiceDuePtr(card, date),
				InstallmentGroupID: &groupID,
				InstallmentNumber:  i,
				InstallmentTotal:   n,
			})
		}
	}

	// A transfer to the investment account.
	out = append(out, &entities.Transaction{
		OrganizationID: orgID, AccountID: checking,
		TransferAccountID: ptr(accts[2].ID), Type: entities.TxTransfer,
		Amount: 100000, Description: "Aporte investimentos",
		Date: now.AddDate(0, 0, -10), Paid: true,
	})

	return out
}

func ptr(id uuid.UUID) *uuid.UUID { return &id }

// invoiceDuePtr returns a pointer to the invoice due date for a card purchase on
// the given date, using the card's billing cycle.
func invoiceDuePtr(card *entities.CreditCard, date time.Time) *time.Time {
	inv := cards.InvoiceFor(card.ClosingDay, card.DueDay, date)
	d := inv.DueDate
	return &d
}
