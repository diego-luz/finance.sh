package reports

import (
	"fmt"
	"io"

	"github.com/go-pdf/fpdf"
)

// The core PDF fonts (Helvetica/Arial) use cp1252 (Latin-1) encoding, not UTF-8.
// We wrap every string through a UTF-8 -> cp1252 translator so pt-BR accents
// render correctly.

// WriteStatementPDF renders a bank-statement-style report for [From, To]: a
// header (finance.sh + org + período), a transactions table and the period totals.
func WriteStatementPDF(w io.Writer, d StatementData) error {
	pdf := fpdf.New("P", "mm", "A4", "")
	tr := pdf.UnicodeTranslatorFromDescriptor("")
	pdf.SetMargins(12, 12, 12)
	pdf.AddPage()

	// Header.
	pdf.SetFont("Helvetica", "B", 18)
	pdf.CellFormat(0, 10, tr("finance.sh"), "", 1, "L", false, 0, "")
	pdf.SetFont("Helvetica", "", 11)
	pdf.CellFormat(0, 6, tr("Extrato — "+d.OrgName), "", 1, "L", false, 0, "")
	pdf.CellFormat(0, 6, tr(fmt.Sprintf("Período: %s a %s", fmtDate(d.From), fmtDate(d.To))), "", 1, "L", false, 0, "")
	pdf.Ln(3)

	// Table header.
	pdf.SetFont("Helvetica", "B", 9)
	pdf.SetFillColor(31, 41, 55)
	pdf.SetTextColor(255, 255, 255)
	widths := []float64{22, 66, 34, 34, 30}
	headers := []string{"Data", "Descrição", "Categoria", "Conta", "Valor"}
	aligns := []string{"L", "L", "L", "L", "R"}
	for i, h := range headers {
		pdf.CellFormat(widths[i], 7, tr(h), "1", 0, aligns[i], true, 0, "")
	}
	pdf.Ln(-1)

	// Table rows.
	pdf.SetTextColor(20, 20, 20)
	pdf.SetFont("Helvetica", "", 9)
	fill := false
	for _, line := range d.Lines {
		if fill {
			pdf.SetFillColor(243, 244, 246)
		} else {
			pdf.SetFillColor(255, 255, 255)
		}
		cells := []string{
			fmtDate(line.Date),
			truncate(line.Description, 42),
			truncate(line.Category, 22),
			truncate(line.Account, 22),
			reaisSigned(line.Amount, line.IsExpense),
		}
		for i, c := range cells {
			pdf.CellFormat(widths[i], 6, tr(c), "1", 0, aligns[i], true, 0, "")
		}
		pdf.Ln(-1)
		fill = !fill
	}
	pdf.Ln(4)

	// Totals.
	pdf.SetFont("Helvetica", "B", 10)
	totalLine(pdf, tr, "Receitas", reaisBR(d.Income))
	totalLine(pdf, tr, "Despesas", "-"+reaisBR(d.Expense))
	totalLine(pdf, tr, "Saldo do período", reaisBR(d.Net))

	return pdf.Output(w)
}

// WriteMonthlyPDF renders a monthly summary: receitas vs despesas, top expense
// categories, paid bills and upcoming bills for the given month/year.
func WriteMonthlyPDF(w io.Writer, d MonthlyData) error {
	pdf := fpdf.New("P", "mm", "A4", "")
	tr := pdf.UnicodeTranslatorFromDescriptor("")
	pdf.SetMargins(12, 12, 12)
	pdf.AddPage()

	pdf.SetFont("Helvetica", "B", 18)
	pdf.CellFormat(0, 10, tr("finance.sh"), "", 1, "L", false, 0, "")
	pdf.SetFont("Helvetica", "", 11)
	pdf.CellFormat(0, 6, tr("Resumo mensal — "+d.OrgName), "", 1, "L", false, 0, "")
	pdf.CellFormat(0, 6, tr(fmt.Sprintf("Mês de referência: %s de %d", MonthName(d.Month), d.Year)), "", 1, "L", false, 0, "")
	pdf.Ln(3)

	// Receitas vs despesas.
	pdf.SetFont("Helvetica", "B", 12)
	pdf.CellFormat(0, 8, tr("Receitas x Despesas"), "", 1, "L", false, 0, "")
	pdf.SetFont("Helvetica", "", 10)
	totalLine(pdf, tr, "Receitas", reaisBR(d.Income))
	totalLine(pdf, tr, "Despesas", "-"+reaisBR(d.Expense))
	totalLine(pdf, tr, "Saldo", reaisBR(d.Income-d.Expense))
	pdf.Ln(4)

	// Top expense categories.
	pdf.SetFont("Helvetica", "B", 12)
	pdf.CellFormat(0, 8, tr("Maiores despesas por categoria"), "", 1, "L", false, 0, "")
	pdf.SetFont("Helvetica", "", 10)
	if len(d.TopExpenses) == 0 {
		pdf.CellFormat(0, 6, tr("Nenhuma despesa no período."), "", 1, "L", false, 0, "")
	}
	for _, c := range d.TopExpenses {
		pdf.CellFormat(120, 6, tr(truncate(c.Name, 60)), "", 0, "L", false, 0, "")
		pdf.CellFormat(0, 6, tr("R$ "+reaisBR(c.Total)), "", 1, "R", false, 0, "")
	}
	pdf.Ln(4)

	// Paid bills.
	billTable(pdf, tr, "Contas pagas", d.PaidBills)
	pdf.Ln(3)
	// Upcoming bills.
	billTable(pdf, tr, "Contas a vencer", d.UpcomingBills)

	return pdf.Output(w)
}

func billTable(pdf *fpdf.Fpdf, tr func(string) string, title string, bills []MonthlyBillLine) {
	pdf.SetFont("Helvetica", "B", 12)
	pdf.CellFormat(0, 8, tr(title), "", 1, "L", false, 0, "")
	pdf.SetFont("Helvetica", "", 10)
	if len(bills) == 0 {
		pdf.CellFormat(0, 6, tr("Nenhum registro."), "", 1, "L", false, 0, "")
		return
	}
	for _, b := range bills {
		pdf.CellFormat(28, 6, tr(fmtDate(b.Date)), "", 0, "L", false, 0, "")
		pdf.CellFormat(100, 6, tr(truncate(b.Description, 52)), "", 0, "L", false, 0, "")
		pdf.CellFormat(0, 6, tr("R$ "+reaisBR(b.Amount)), "", 1, "R", false, 0, "")
	}
}

func totalLine(pdf *fpdf.Fpdf, tr func(string) string, label, value string) {
	pdf.CellFormat(60, 7, tr(label), "", 0, "L", false, 0, "")
	pdf.CellFormat(0, 7, tr("R$ "+value), "", 1, "R", false, 0, "")
}

func truncate(s string, max int) string {
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	if max <= 1 {
		return string(r[:max])
	}
	return string(r[:max-1]) + "…"
}
