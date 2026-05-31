package reports

import (
	"fmt"
	"io"

	"github.com/xuri/excelize/v2"
)

// brlNumFmt is an Excel custom number format for Brazilian reais. Codes 0..163
// are built-in; custom codes are auto-registered by NewStyle via CustomNumFmt.
const brlNumFmt = `#,##0.00`

// WriteExcel renders the org's data as a multi-sheet .xlsx workbook to w. Sheets
// (pt-BR): Transações, Contas, Categorias, Contatos, ContasAPagar, ContasAReceber.
// Money columns are written as numeric reais with a #.##0,00 format so Excel
// treats them as numbers.
func WriteExcel(w io.Writer, d ExcelData) error {
	f := excelize.NewFile()
	defer func() { _ = f.Close() }()

	// Header style: bold with a light fill.
	headerStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Color: "FFFFFF"},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"1F2937"}, Pattern: 1},
	})
	if err != nil {
		return err
	}
	moneyStyle, err := f.NewStyle(&excelize.Style{CustomNumFmt: strPtr(brlNumFmt)})
	if err != nil {
		return err
	}

	// Transações
	if err := f.SetSheetName("Sheet1", "Transações"); err != nil {
		return err
	}
	txHeaders := []interface{}{"Data", "Descrição", "Tipo", "Categoria", "Conta", "Contato", "Valor (R$)", "Vencimento", "Pago"}
	if err := writeHeader(f, "Transações", txHeaders, headerStyle); err != nil {
		return err
	}
	for i, t := range d.Transactions {
		row := i + 2
		due := ""
		if t.DueDate != nil {
			due = fmtDate(*t.DueDate)
		}
		vals := []interface{}{
			fmtDate(t.Date), t.Description, t.Type, t.Category, t.Account, t.Contact,
			reaisFloat(t.Amount), due, boolPT(t.Paid),
		}
		if err := f.SetSheetRow("Transações", cell("A", row), &vals); err != nil {
			return err
		}
		_ = f.SetCellStyle("Transações", cell("G", row), cell("G", row), moneyStyle)
	}
	autoWidth(f, "Transações", txHeaders)

	// Contas
	if err := addSheet(f, "Contas", []interface{}{"Conta", "Tipo", "Saldo Inicial (R$)", "Saldo Atual (R$)"}, headerStyle); err != nil {
		return err
	}
	for i, a := range d.Accounts {
		row := i + 2
		vals := []interface{}{a.Name, a.Type, reaisFloat(a.InitialBalance), reaisFloat(a.Balance)}
		if err := f.SetSheetRow("Contas", cell("A", row), &vals); err != nil {
			return err
		}
		_ = f.SetCellStyle("Contas", cell("C", row), cell("D", row), moneyStyle)
	}

	// Categorias
	if err := addSheet(f, "Categorias", []interface{}{"Categoria", "Tipo"}, headerStyle); err != nil {
		return err
	}
	for i, c := range d.Categories {
		row := i + 2
		vals := []interface{}{c.Name, c.Kind}
		if err := f.SetSheetRow("Categorias", cell("A", row), &vals); err != nil {
			return err
		}
	}

	// Contatos
	if err := addSheet(f, "Contatos", []interface{}{"Nome", "Tipo", "Documento", "E-mail", "Telefone"}, headerStyle); err != nil {
		return err
	}
	for i, c := range d.Contacts {
		row := i + 2
		vals := []interface{}{c.Name, c.Type, c.Document, c.Email, c.Phone}
		if err := f.SetSheetRow("Contatos", cell("A", row), &vals); err != nil {
			return err
		}
	}

	// ContasAPagar / ContasAReceber
	if err := writePayables(f, "ContasAPagar", d.Payables, headerStyle, moneyStyle); err != nil {
		return err
	}
	if err := writePayables(f, "ContasAReceber", d.Receivables, headerStyle, moneyStyle); err != nil {
		return err
	}

	// Make the first sheet active.
	if idx, err := f.GetSheetIndex("Transações"); err == nil {
		f.SetActiveSheet(idx)
	}

	return f.Write(w)
}

func writePayables(f *excelize.File, sheet string, rows []ExcelPayable, headerStyle, moneyStyle int) error {
	headers := []interface{}{"Vencimento", "Descrição", "Contato", "Categoria", "Valor (R$)", "Status"}
	if err := addSheet(f, sheet, headers, headerStyle); err != nil {
		return err
	}
	for i, p := range rows {
		row := i + 2
		vals := []interface{}{fmtDate(p.DueDate), p.Description, p.Contact, p.Category, reaisFloat(p.Amount), p.Status}
		if err := f.SetSheetRow(sheet, cell("A", row), &vals); err != nil {
			return err
		}
		_ = f.SetCellStyle(sheet, cell("E", row), cell("E", row), moneyStyle)
	}
	autoWidth(f, sheet, headers)
	return nil
}

// addSheet creates a new sheet and writes its header row.
func addSheet(f *excelize.File, name string, headers []interface{}, headerStyle int) error {
	if _, err := f.NewSheet(name); err != nil {
		return err
	}
	if err := writeHeader(f, name, headers, headerStyle); err != nil {
		return err
	}
	autoWidth(f, name, headers)
	return nil
}

func writeHeader(f *excelize.File, sheet string, headers []interface{}, headerStyle int) error {
	if err := f.SetSheetRow(sheet, "A1", &headers); err != nil {
		return err
	}
	last := colLetter(len(headers))
	return f.SetCellStyle(sheet, "A1", last+"1", headerStyle)
}

// autoWidth sets a reasonable fixed width per column so the sheet is readable.
func autoWidth(f *excelize.File, sheet string, headers []interface{}) {
	for i := range headers {
		col := colLetter(i + 1)
		_ = f.SetColWidth(sheet, col, col, 22)
	}
}

// reaisFloat converts cents to a float in reais for numeric Excel cells.
func reaisFloat(cents int64) float64 { return float64(cents) / 100.0 }

func boolPT(b bool) string {
	if b {
		return "Sim"
	}
	return "Não"
}

func strPtr(s string) *string { return &s }

// cell builds an A1-style reference from a column letter and a 1-based row.
func cell(col string, row int) string { return fmt.Sprintf("%s%d", col, row) }

// colLetter returns the spreadsheet column letter for a 1-based index
// (1 -> A, 26 -> Z, 27 -> AA).
func colLetter(n int) string {
	name, _ := excelize.ColumnNumberToName(n)
	return name
}
