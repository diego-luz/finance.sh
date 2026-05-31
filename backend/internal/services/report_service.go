package services

import (
	"encoding/csv"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/finance-sh/finance-sh/internal/dto"
	"github.com/finance-sh/finance-sh/internal/entities"
	"github.com/finance-sh/finance-sh/internal/reports"
	"github.com/finance-sh/finance-sh/internal/repositories"
	"github.com/google/uuid"
)

type ReportService struct {
	txs        *repositories.TransactionRepository
	accounts   *repositories.AccountRepository
	categories *repositories.CategoryRepository
	contacts   *repositories.ContactRepository
	users      *repositories.UserRepository
}

func NewReportService(
	txs *repositories.TransactionRepository,
	accounts *repositories.AccountRepository,
	categories *repositories.CategoryRepository,
	contacts *repositories.ContactRepository,
	users *repositories.UserRepository,
) *ReportService {
	return &ReportService{txs: txs, accounts: accounts, categories: categories, contacts: contacts, users: users}
}

// monthRange returns [start, end) for the calendar month of ref (UTC).
func monthRange(ref time.Time) (time.Time, time.Time) {
	start := time.Date(ref.Year(), ref.Month(), 1, 0, 0, 0, 0, time.UTC)
	return start, start.AddDate(0, 1, 0)
}

// orgName looks up the organization display name for report headers; falls back
// to a generic label when it cannot be resolved.
func (s *ReportService) orgName(orgID uuid.UUID) string {
	if s.users == nil {
		return "Organização"
	}
	org, err := s.users.FindOrganization(orgID)
	if err != nil || org == nil {
		return "Organização"
	}
	return org.Name
}

// TransactionsCSV writes the org's filtered transactions as CSV to w. Money is
// rendered in reais with a comma decimal separator (pt-BR convention).
func (s *ReportService) TransactionsCSV(orgID uuid.UUID, f dto.TransactionFilter, w io.Writer) error {
	rows, err := s.txs.ListAllFiltered(orgID, f)
	if err != nil {
		return err
	}

	cw := csv.NewWriter(w)
	if err := cw.Write([]string{"data", "descrição", "tipo", "categoria", "conta", "valor", "pago"}); err != nil {
		return err
	}
	for i := range rows {
		t := &rows[i]
		category := ""
		if t.Category != nil {
			category = t.Category.Name
		}
		account := ""
		if t.Account != nil {
			account = t.Account.Name
		}
		record := []string{
			t.Date.Format("2006-01-02"),
			t.Description,
			translateType(t.Type),
			category,
			account,
			reais(t.Amount),
			boolPT(t.Paid),
		}
		if err := cw.Write(record); err != nil {
			return err
		}
	}
	cw.Flush()
	return cw.Error()
}

// reais converts cents to a "1234,56" string (comma decimal separator).
func reais(cents int64) string {
	neg := cents < 0
	if neg {
		cents = -cents
	}
	s := fmt.Sprintf("%d,%02d", cents/100, cents%100)
	if neg {
		return "-" + s
	}
	return s
}

func translateType(t entities.TransactionType) string {
	switch t {
	case entities.TxIncome:
		return "receita"
	case entities.TxExpense:
		return "despesa"
	case entities.TxTransfer:
		return "transferência"
	default:
		return strings.ToLower(string(t))
	}
}

func boolPT(b bool) string {
	if b {
		return "sim"
	}
	return "não"
}

// ----- Summary (Relatórios page) -----

// Summary returns the org's settled income/expense within [from, to) plus the
// breakdowns by category and by month. When from/to are nil the range defaults
// to the current calendar month.
func (s *ReportService) Summary(orgID uuid.UUID, from, to *time.Time) (*dto.ReportSummary, error) {
	start, end := resolveRange(from, to)

	byMonth, err := s.txs.SummaryByMonth(orgID, start, end)
	if err != nil {
		return nil, err
	}
	byCat, err := s.txs.SummaryByCategory(orgID, start, end)
	if err != nil {
		return nil, err
	}

	var income, expense int64
	for _, m := range byMonth {
		income += m.Income
		expense += m.Expense
	}

	return &dto.ReportSummary{
		Income:     income,
		Expense:    expense,
		Net:        income - expense,
		ByCategory: byCat,
		ByMonth:    byMonth,
	}, nil
}

// resolveRange applies the "current month" default when either bound is missing.
// The end bound is treated as exclusive by the repo queries (date < end), so a
// supplied inclusive end-of-day still works.
func resolveRange(from, to *time.Time) (time.Time, time.Time) {
	if from != nil && to != nil {
		return *from, *to
	}
	now := time.Now().UTC()
	return monthRange(now)
}

// ----- Excel export -----

// ExcelExport gathers the org's data and writes a multi-sheet .xlsx workbook to w.
func (s *ReportService) ExcelExport(orgID uuid.UUID, w io.Writer) error {
	data, err := s.buildExcelData(orgID)
	if err != nil {
		return err
	}
	return reports.WriteExcel(w, data)
}

func (s *ReportService) buildExcelData(orgID uuid.UUID) (reports.ExcelData, error) {
	var data reports.ExcelData

	// Transactions (all, ordered by date asc).
	txRows, err := s.txs.ListAllFiltered(orgID, dto.TransactionFilter{})
	if err != nil {
		return data, err
	}
	for i := range txRows {
		t := &txRows[i]
		cat, acct := "", ""
		if t.Category != nil {
			cat = t.Category.Name
		}
		if t.Account != nil {
			acct = t.Account.Name
		}
		data.Transactions = append(data.Transactions, reports.ExcelTransaction{
			Date:        t.Date,
			Description: t.Description,
			Type:        translateType(t.Type),
			Category:    cat,
			Account:     acct,
			Amount:      t.Amount,
			Paid:        t.Paid,
			DueDate:     t.DueDate,
		})
	}

	// Accounts with computed balances.
	accts, err := s.accounts.List(orgID)
	if err != nil {
		return data, err
	}
	balances, err := s.accounts.Balances(orgID)
	if err != nil {
		return data, err
	}
	for i := range accts {
		a := &accts[i]
		data.Accounts = append(data.Accounts, reports.ExcelAccount{
			Name:           a.Name,
			Type:           translateAccountType(a.Type),
			InitialBalance: a.InitialBalance,
			Balance:        balances[a.ID],
		})
	}

	// Categories.
	cats, err := s.categories.List(orgID)
	if err != nil {
		return data, err
	}
	for i := range cats {
		c := &cats[i]
		data.Categories = append(data.Categories, reports.ExcelCategory{
			Name: c.Name,
			Kind: translateKind(c.Kind),
		})
	}

	// Contacts.
	contacts, err := s.contacts.List(orgID)
	if err != nil {
		return data, err
	}
	for i := range contacts {
		c := &contacts[i]
		data.Contacts = append(data.Contacts, reports.ExcelContact{
			Name:     c.Name,
			Type:     translateContactType(c.Type),
			Document: c.Document,
			Email:    c.Email,
			Phone:    c.Phone,
		})
	}

	// Payables / receivables: all rows of each type, ordered by effective due date.
	now := time.Now().UTC()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	allFilter := dto.PayableFilter{Status: "all", Page: 1, PerPage: 100000}
	data.Payables, err = s.payableRows(orgID, entities.TxExpense, allFilter, today)
	if err != nil {
		return data, err
	}
	data.Receivables, err = s.payableRows(orgID, entities.TxIncome, allFilter, today)
	if err != nil {
		return data, err
	}

	return data, nil
}

func (s *ReportService) payableRows(orgID uuid.UUID, txType entities.TransactionType, f dto.PayableFilter, today time.Time) ([]reports.ExcelPayable, error) {
	rows, _, err := s.txs.Payables(orgID, txType, f, today)
	if err != nil {
		return nil, err
	}
	out := make([]reports.ExcelPayable, 0, len(rows))
	for i := range rows {
		t := &rows[i]
		due := t.Date
		if t.DueDate != nil {
			due = *t.DueDate
		}
		cat, contact := "", ""
		if t.Category != nil {
			cat = t.Category.Name
		}
		if t.Contact != nil {
			contact = t.Contact.Name
		}
		out = append(out, reports.ExcelPayable{
			DueDate:     due,
			Description: t.Description,
			Contact:     contact,
			Category:    cat,
			Amount:      t.Amount,
			Status:      statusPT(t),
		})
	}
	return out, nil
}

// ----- Statement PDF -----

// StatementPDF writes a bank-statement-style PDF for [from, to) to w. When the
// bounds are nil the current calendar month is used.
func (s *ReportService) StatementPDF(orgID uuid.UUID, from, to *time.Time, w io.Writer) error {
	start, end := resolveRange(from, to)

	f := dto.TransactionFilter{From: &start, To: &end}
	rows, err := s.txs.ListAllFiltered(orgID, f)
	if err != nil {
		return err
	}

	data := reports.StatementData{
		OrgName: s.orgName(orgID),
		From:    start,
		To:      end,
	}
	for i := range rows {
		t := &rows[i]
		// The statement reflects movement: skip transfers from the income/expense
		// totals but still list them.
		isExpense := t.Type == entities.TxExpense
		cat, acct := "", ""
		if t.Category != nil {
			cat = t.Category.Name
		}
		if t.Account != nil {
			acct = t.Account.Name
		}
		data.Lines = append(data.Lines, reports.StatementLine{
			Date:        t.Date,
			Description: t.Description,
			Category:    cat,
			Account:     acct,
			Amount:      t.Amount,
			IsExpense:   isExpense,
		})
		switch t.Type {
		case entities.TxIncome:
			data.Income += t.Amount
		case entities.TxExpense:
			data.Expense += t.Amount
		}
	}
	data.Net = data.Income - data.Expense

	return reports.WriteStatementPDF(w, data)
}

// ----- Monthly summary PDF -----

// MonthlyPDF writes a monthly summary PDF for the given month/year to w. When
// month/year are zero the current month/year is used.
func (s *ReportService) MonthlyPDF(orgID uuid.UUID, month, year int, w io.Writer) error {
	now := time.Now().UTC()
	if month < 1 || month > 12 {
		month = int(now.Month())
	}
	if year < 2000 || year > 2100 {
		year = now.Year()
	}
	start := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 1, 0)

	income, expense, err := s.txs.MonthTotals(orgID, start)
	if err != nil {
		return err
	}

	topCats, err := s.txs.TopCategories(orgID, start, end, 5)
	if err != nil {
		return err
	}

	data := reports.MonthlyData{
		OrgName: s.orgName(orgID),
		Month:   month,
		Year:    year,
		Income:  income,
		Expense: expense,
	}
	for _, c := range topCats {
		data.TopExpenses = append(data.TopExpenses, reports.MonthlyCategoryTotal{Name: c.Name, Total: c.Total})
	}

	// Paid bills (settled expenses) within the month.
	paidFilter := dto.TransactionFilter{Type: string(entities.TxExpense), From: &start, To: &end}
	paidRows, err := s.txs.ListAllFiltered(orgID, paidFilter)
	if err != nil {
		return err
	}
	for i := range paidRows {
		t := &paidRows[i]
		if !t.Paid {
			continue
		}
		due := t.Date
		if t.DueDate != nil {
			due = *t.DueDate
		}
		data.PaidBills = append(data.PaidBills, reports.MonthlyBillLine{Date: due, Description: t.Description, Amount: t.Amount})
	}

	// Upcoming (unpaid future expenses), capped.
	upcoming, err := s.txs.UpcomingBills(orgID, 20)
	if err != nil {
		return err
	}
	for i := range upcoming {
		t := &upcoming[i]
		due := t.Date
		if t.DueDate != nil {
			due = *t.DueDate
		}
		data.UpcomingBills = append(data.UpcomingBills, reports.MonthlyBillLine{Date: due, Description: t.Description, Amount: t.Amount})
	}

	return reports.WriteMonthlyPDF(w, data)
}

// ----- pt-BR label helpers -----

func translateKind(k entities.CategoryKind) string {
	if k == entities.CategoryIncome {
		return "receita"
	}
	return "despesa"
}

func translateAccountType(t entities.AccountType) string {
	switch t {
	case entities.AccountBank:
		return "conta corrente"
	case entities.AccountWallet:
		return "carteira"
	case entities.AccountInvestment:
		return "investimento"
	case entities.AccountCreditCard:
		return "cartão de crédito"
	default:
		return string(t)
	}
}

func translateContactType(t entities.ContactType) string {
	switch t {
	case entities.ContactCustomer:
		return "cliente"
	case entities.ContactSupplier:
		return "fornecedor"
	case entities.ContactBoth:
		return "ambos"
	default:
		return string(t)
	}
}

// statusPT renders a payable/receivable lifecycle status in pt-BR.
func statusPT(t *entities.Transaction) string {
	switch transactionStatus(t) {
	case "paid":
		return "pago"
	case "overdue":
		return "vencido"
	default:
		return "em aberto"
	}
}
