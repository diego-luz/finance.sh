// Package reports builds human-friendly export documents (Excel workbooks and
// PDF statements) from already-fetched, org-scoped data. It holds only
// formatting/layout logic so the service/handler layers stay thin; it never
// touches the database.
package reports

import (
	"fmt"
	"time"
)

// Money in this package is always int64 minor units (centavos).

// reaisBR formats cents as a Brazilian-real string "1.234,56" (dot thousands,
// comma decimals) without the currency symbol. Negative values keep the sign.
func reaisBR(cents int64) string {
	neg := cents < 0
	if neg {
		cents = -cents
	}
	intPart := cents / 100
	dec := cents % 100

	// Group the integer part with dots every three digits.
	s := fmt.Sprintf("%d", intPart)
	var grouped string
	for i, c := range reverse(s) {
		if i > 0 && i%3 == 0 {
			grouped = "." + grouped
		}
		grouped = string(c) + grouped
	}
	out := fmt.Sprintf("%s,%02d", grouped, dec)
	if neg {
		return "-" + out
	}
	return out
}

// reaisSigned renders a value with an explicit +/- sign relative to its kind:
// income is positive, expense negative. Used in the statement "valor" column.
func reaisSigned(cents int64, isExpense bool) string {
	if isExpense {
		return "-" + reaisBR(cents)
	}
	return "+" + reaisBR(cents)
}

func reverse(s string) string {
	r := []rune(s)
	for i, j := 0, len(r)-1; i < j; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	return string(r)
}

// ptMonths are the pt-BR month names indexed 1..12.
var ptMonths = [...]string{
	"", "janeiro", "fevereiro", "março", "abril", "maio", "junho",
	"julho", "agosto", "setembro", "outubro", "novembro", "dezembro",
}

// MonthName returns the pt-BR month name for 1..12 (empty for out-of-range).
func MonthName(m int) string {
	if m < 1 || m > 12 {
		return ""
	}
	return ptMonths[m]
}

// fmtDate renders a date as dd/mm/aaaa.
func fmtDate(t time.Time) string { return t.Format("02/01/2006") }

// ----- Input data structures (decoupled from dto/entities) -----

// StatementLine is one transaction row in the bank-statement PDF.
type StatementLine struct {
	Date        time.Time
	Description string
	Category    string
	Account     string
	Amount      int64
	IsExpense   bool
}

// StatementData is everything the statement PDF needs.
type StatementData struct {
	OrgName string
	From    time.Time
	To      time.Time
	Lines   []StatementLine
	Income  int64
	Expense int64
	Net     int64
}

// MonthlyCategoryTotal is a category total for the monthly summary PDF.
type MonthlyCategoryTotal struct {
	Name  string
	Total int64
}

// MonthlyBillLine is a paid/upcoming bill row in the monthly summary PDF.
type MonthlyBillLine struct {
	Date        time.Time
	Description string
	Amount      int64
}

// MonthlyData is everything the monthly summary PDF needs.
type MonthlyData struct {
	OrgName       string
	Month         int
	Year          int
	Income        int64
	Expense       int64
	TopExpenses   []MonthlyCategoryTotal
	PaidBills     []MonthlyBillLine
	UpcomingBills []MonthlyBillLine
}

// ExcelData is the multi-sheet workbook content.
type ExcelData struct {
	Transactions []ExcelTransaction
	Accounts     []ExcelAccount
	Categories   []ExcelCategory
	Contacts     []ExcelContact
	Payables     []ExcelPayable
	Receivables  []ExcelPayable
}

type ExcelTransaction struct {
	Date        time.Time
	Description string
	Type        string // pt-BR label
	Category    string
	Account     string
	Contact     string
	Amount      int64
	Paid        bool
	DueDate     *time.Time
}

type ExcelAccount struct {
	Name           string
	Type           string
	InitialBalance int64
	Balance        int64
}

type ExcelCategory struct {
	Name string
	Kind string // pt-BR label
}

type ExcelContact struct {
	Name     string
	Type     string
	Document string
	Email    string
	Phone    string
}

// ExcelPayable is a payable/receivable row.
type ExcelPayable struct {
	DueDate     time.Time
	Description string
	Contact     string
	Category    string
	Amount      int64
	Status      string // pt-BR: pago/em aberto/vencido
}
