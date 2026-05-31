package database

import (
	"github.com/finance-sh/finance-sh/internal/entities"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SeedDefaults populates a freshly-created organization with sensible default
// categories and accounts so a newly-registered user doesn't land in an empty
// app. It is called ONCE on registration (never on login) and deliberately does
// NOT create any transactions for real tenants — only the demo seed (seed.go)
// fabricates movements.
//
// Colors/icons mirror the strings used by the demo seed so the frontend can
// render the same iconography. Everything is created in a single transaction so
// a partial failure leaves the new org with no defaults rather than a half-set.
func SeedDefaults(db *gorm.DB, orgID uuid.UUID) error {
	return db.Transaction(func(tx *gorm.DB) error {
		// Default categories: income + expense, each with a color (hex) and an
		// icon name matching the demo seed's icon vocabulary.
		categories := []*entities.Category{
			// Income.
			{OrganizationID: orgID, Name: "Salário", Kind: entities.CategoryIncome, Color: "#10b981", Icon: "briefcase"},
			{OrganizationID: orgID, Name: "Outras receitas", Kind: entities.CategoryIncome, Color: "#22c55e", Icon: "trending-up"},
			// Expense.
			{OrganizationID: orgID, Name: "Alimentação", Kind: entities.CategoryExpense, Color: "#f97316", Icon: "utensils"},
			{OrganizationID: orgID, Name: "Moradia", Kind: entities.CategoryExpense, Color: "#ef4444", Icon: "home"},
			{OrganizationID: orgID, Name: "Transporte", Kind: entities.CategoryExpense, Color: "#eab308", Icon: "car"},
			{OrganizationID: orgID, Name: "Lazer", Kind: entities.CategoryExpense, Color: "#a855f7", Icon: "film"},
			{OrganizationID: orgID, Name: "Saúde", Kind: entities.CategoryExpense, Color: "#ec4899", Icon: "heart-pulse"},
			{OrganizationID: orgID, Name: "Educação", Kind: entities.CategoryExpense, Color: "#14b8a6", Icon: "graduation-cap"},
			{OrganizationID: orgID, Name: "Contas/Assinaturas", Kind: entities.CategoryExpense, Color: "#3b82f6", Icon: "repeat"},
			{OrganizationID: orgID, Name: "Outros", Kind: entities.CategoryExpense, Color: "#6366f1", Icon: "tag"},
		}
		byName := make(map[string]uuid.UUID, len(categories))
		for _, c := range categories {
			if err := tx.Create(c).Error; err != nil {
				return err
			}
			byName[c.Name] = c.ID
		}

		// Default categorization rules mapping common pt-BR keywords to the
		// default categories above. All "contains" (case-insensitive). Created
		// only for categories that exist in the seed set.
		type seedRule struct {
			category string
			keywords []string
			priority int
		}
		ruleSeeds := []seedRule{
			{"Transporte", []string{"uber", "99", "taxi", "posto", "combustivel", "combustível", "ipva"}, 10},
			{"Alimentação", []string{"ifood", "restaurante", "mercado", "supermercado", "padaria", "lanchonete"}, 10},
			{"Contas/Assinaturas", []string{"netflix", "spotify", "assinatura", "prime", "disney"}, 10},
			{"Saúde", []string{"farmacia", "farmácia", "drogaria", "hospital", "clinica", "clínica"}, 10},
			{"Moradia", []string{"aluguel", "condominio", "condomínio", "luz", "agua", "água"}, 10},
			{"Salário", []string{"salario", "salário", "pagamento salario"}, 10},
		}
		for _, rs := range ruleSeeds {
			catID, ok := byName[rs.category]
			if !ok {
				continue
			}
			for _, kw := range rs.keywords {
				rule := &entities.CategoryRule{
					OrganizationID: orgID,
					Pattern:        kw,
					MatchType:      entities.MatchContains,
					CategoryID:     catID,
					Priority:       rs.priority,
					Active:         true,
				}
				if err := tx.Create(rule).Error; err != nil {
					return err
				}
			}
		}

		// Default accounts with a zero initial balance.
		accounts := []*entities.Account{
			{OrganizationID: orgID, Name: "Carteira", Type: entities.AccountWallet, InitialBalance: 0, Color: "#f59e0b", Icon: "wallet"},
			{OrganizationID: orgID, Name: "Conta Corrente", Type: entities.AccountBank, InitialBalance: 0, Color: "#10b981", Icon: "bank"},
		}
		for _, a := range accounts {
			if err := tx.Create(a).Error; err != nil {
				return err
			}
		}

		return nil
	})
}
