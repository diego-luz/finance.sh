package repositories

import (
	"errors"
	"time"

	"github.com/finance-sh/finance-sh/internal/dto"
	"github.com/finance-sh/finance-sh/internal/entities"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TransactionRepository persists transactions and provides the read-side
// aggregations the dashboard needs. Every query is scoped by organization_id.
type TransactionRepository struct{ db *gorm.DB }

func NewTransactionRepository(db *gorm.DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

func (r *TransactionRepository) Create(t *entities.Transaction) error {
	return r.db.Create(t).Error
}

func (r *TransactionRepository) Update(t *entities.Transaction) error {
	return r.db.Save(t).Error
}

// ReplaceTags sets the transaction's many-to-many Tags association to exactly
// the supplied tags (clearing it when tags is empty). The tags are expected to
// be already validated against the org by the service layer.
func (r *TransactionRepository) ReplaceTags(t *entities.Transaction, tags []entities.Tag) error {
	return r.db.Model(t).Association("Tags").Replace(tags)
}

func (r *TransactionRepository) Delete(orgID, id uuid.UUID) error {
	res := r.db.Where("organization_id = ? AND id = ?", orgID, id).Delete(&entities.Transaction{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *TransactionRepository) FindByID(orgID, id uuid.UUID) (*entities.Transaction, error) {
	var t entities.Transaction
	err := r.db.Preload("Category").Preload("Contact").Preload("Tags").
		Where("organization_id = ? AND id = ?", orgID, id).First(&t).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &t, err
}

// CreateMany inserts several transactions inside a single DB transaction, so a
// parcelado purchase is materialised atomically (all parcelas or none).
func (r *TransactionRepository) CreateMany(txs []*entities.Transaction) error {
	if len(txs) == 0 {
		return nil
	}
	return r.db.Transaction(func(tx *gorm.DB) error {
		for _, t := range txs {
			if err := tx.Create(t).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// DeleteInstallmentGroup soft-deletes every transaction of an installment group
// (org-scoped). Returns the number of rows removed.
func (r *TransactionRepository) DeleteInstallmentGroup(orgID, groupID uuid.UUID) (int64, error) {
	res := r.db.Where("organization_id = ? AND installment_group_id = ?", orgID, groupID).
		Delete(&entities.Transaction{})
	return res.RowsAffected, res.Error
}

// ListByCard returns every transaction booked against the given credit card
// (org-scoped), ordered by date ascending, with Category and Contact preloaded.
// Used to group a card's purchases into invoices.
func (r *TransactionRepository) ListByCard(orgID, cardID uuid.UUID) ([]entities.Transaction, error) {
	var txs []entities.Transaction
	err := r.db.Preload("Category").Preload("Contact").
		Where("organization_id = ? AND credit_card_id = ?", orgID, cardID).
		Order("date asc, created_at asc").
		Find(&txs).Error
	return txs, err
}

// scoped builds a base query already filtered by org and the supplied filter.
func (r *TransactionRepository) scoped(orgID uuid.UUID, f dto.TransactionFilter) *gorm.DB {
	q := r.db.Model(&entities.Transaction{}).Where("organization_id = ?", orgID)
	if f.Type != "" {
		q = q.Where("type = ?", f.Type)
	}
	if f.AccountID != "" {
		q = q.Where("account_id = ?", f.AccountID)
	}
	if f.CategoryID != "" {
		q = q.Where("category_id = ?", f.CategoryID)
	}
	if f.ContactID != "" {
		q = q.Where("contact_id = ?", f.ContactID)
	}
	if f.TagID != "" {
		// JOIN the implicit many2many join table so we can filter by an attached
		// tag. Only one tag_id is filtered, so a single matching join row keeps
		// the count correct without needing DISTINCT.
		q = q.Joins("JOIN transaction_tags tt ON tt.transaction_id = transactions.id").
			Where("tt.tag_id = ?", f.TagID)
	}
	if f.From != nil {
		q = q.Where("date >= ?", *f.From)
	}
	if f.To != nil {
		q = q.Where("date <= ?", *f.To)
	}
	if f.Search != "" {
		q = q.Where("description ILIKE ?", "%"+f.Search+"%")
	}
	return q
}

// List returns a page of transactions ordered by date desc, plus the total
// count for pagination. Category is preloaded for display.
func (r *TransactionRepository) List(orgID uuid.UUID, f dto.TransactionFilter) ([]entities.Transaction, int64, error) {
	var total int64
	if err := r.scoped(orgID, f).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (f.Page - 1) * f.PerPage
	var txs []entities.Transaction
	err := r.scoped(orgID, f).
		Preload("Category").
		Preload("Tags").
		Order("date desc, created_at desc").
		Limit(f.PerPage).
		Offset(offset).
		Find(&txs).Error
	return txs, total, err
}

// SearchByDescription returns the org's transactions whose description matches
// the query (ILIKE), most recent first, capped at limit, with Category and Tags
// preloaded. Used by global search.
func (r *TransactionRepository) SearchByDescription(orgID uuid.UUID, q string, limit int) ([]entities.Transaction, error) {
	var txs []entities.Transaction
	err := r.db.Preload("Category").
		Where("organization_id = ? AND description ILIKE ?", orgID, "%"+q+"%").
		Order("date desc, created_at desc").
		Limit(limit).
		Find(&txs).Error
	return txs, err
}

// ListAllFiltered returns every transaction matching the filter (no pagination),
// with Account and Category preloaded, ordered by date asc. Used for CSV export.
func (r *TransactionRepository) ListAllFiltered(orgID uuid.UUID, f dto.TransactionFilter) ([]entities.Transaction, error) {
	var txs []entities.Transaction
	err := r.scoped(orgID, f).
		Preload("Category").
		Preload("Account").
		Order("date asc").
		Find(&txs).Error
	return txs, err
}

// DueUnpaidBills returns unpaid expense transactions whose due date is on or
// before `until` (overdue + due-soon). Category preloaded. Used by the worker.
func (r *TransactionRepository) DueUnpaidBills(orgID uuid.UUID, until time.Time) ([]entities.Transaction, error) {
	var txs []entities.Transaction
	err := r.db.Preload("Category").
		Where("organization_id = ? AND paid = ? AND type = ? AND date <= ?",
			orgID, false, entities.TxExpense, until).
		Order("date asc").
		Find(&txs).Error
	return txs, err
}

// ExistsExternalIDs reports which of the given external ids already exist for
// the org+account, in a single query. Used by statement import to flag/skip
// duplicates (OFX FITID or CSV hash). Empty ids are ignored. Returns a set
// keyed by the external id (only present/true entries are included).
func (r *TransactionRepository) ExistsExternalIDs(orgID, accountID uuid.UUID, ids []string) (map[string]bool, error) {
	out := make(map[string]bool, len(ids))
	// Filter out empties to keep the IN list tight.
	clean := make([]string, 0, len(ids))
	for _, id := range ids {
		if id != "" {
			clean = append(clean, id)
		}
	}
	if len(clean) == 0 {
		return out, nil
	}
	var found []string
	err := r.db.Model(&entities.Transaction{}).
		Where("organization_id = ? AND account_id = ? AND external_id IN ?", orgID, accountID, clean).
		Pluck("external_id", &found).Error
	if err != nil {
		return nil, err
	}
	for _, id := range found {
		out[id] = true
	}
	return out, nil
}

// ExistsBySignature reports whether a transaction already exists in the
// org+account with the same date (calendar day), absolute amount and
// description. Used as the dedup fallback for imported rows that carry no
// external id.
func (r *TransactionRepository) ExistsBySignature(orgID, accountID uuid.UUID, date time.Time, amount int64, description string) (bool, error) {
	day := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	next := day.AddDate(0, 0, 1)
	var count int64
	err := r.db.Model(&entities.Transaction{}).
		Where("organization_id = ? AND account_id = ? AND amount = ? AND description = ? AND date >= ? AND date < ?",
			orgID, accountID, amount, description, day, next).
		Count(&count).Error
	return count > 0, err
}

// MostCommonCategoryForToken implements the HISTORY fallback for automatic
// categorization: among the org's PAST categorized transactions of the given
// type whose description ILIKE any of the supplied significant tokens, it returns
// the category_id used most often (a single grouped query, limit 1). Returns
// (nil, nil) when there is no history to learn from. txType is "income"|"expense".
func (r *TransactionRepository) MostCommonCategoryForToken(orgID uuid.UUID, txType string, tokens []string) (*uuid.UUID, error) {
	if len(tokens) == 0 {
		return nil, nil
	}
	q := r.db.Model(&entities.Transaction{}).
		Where("organization_id = ? AND type = ? AND category_id IS NOT NULL", orgID, txType)

	// OR together one ILIKE per token. Building the OR group keeps the token
	// match independent of the org/type guards above.
	or := r.db
	for i, tok := range tokens {
		pat := "%" + tok + "%"
		if i == 0 {
			or = r.db.Where("description ILIKE ?", pat)
		} else {
			or = or.Or("description ILIKE ?", pat)
		}
	}
	q = q.Where(or)

	var catID uuid.UUID
	err := q.
		Select("category_id").
		Group("category_id").
		Order("COUNT(*) desc").
		Limit(1).
		Scan(&catID).Error
	if err != nil {
		return nil, err
	}
	if catID == uuid.Nil {
		return nil, nil
	}
	return &catID, nil
}

// UncategorizedIDs returns the ids of every transaction in the org with no
// category assigned (category_id IS NULL), together with its description and
// type so the caller can match them in-memory without a second round-trip.
func (r *TransactionRepository) UncategorizedIDs(orgID uuid.UUID) ([]entities.Transaction, error) {
	var txs []entities.Transaction
	err := r.db.
		Select("id", "description", "type").
		Where("organization_id = ? AND category_id IS NULL", orgID).
		Find(&txs).Error
	return txs, err
}

// MonthTotals returns the paid income and expense totals for the calendar month
// containing the given reference time.
func (r *TransactionRepository) MonthTotals(orgID uuid.UUID, ref time.Time) (income, expense int64, err error) {
	start := time.Date(ref.Year(), ref.Month(), 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 1, 0)

	type sum struct {
		Type  entities.TransactionType
		Total int64
	}
	var rows []sum
	err = r.db.Model(&entities.Transaction{}).
		Select("type, COALESCE(SUM(amount),0) as total").
		Where("organization_id = ? AND paid = ? AND date >= ? AND date < ? AND type IN ?",
			orgID, true, start, end, []entities.TransactionType{entities.TxIncome, entities.TxExpense}).
		Group("type").
		Scan(&rows).Error
	if err != nil {
		return 0, 0, err
	}
	for _, s := range rows {
		switch s.Type {
		case entities.TxIncome:
			income = s.Total
		case entities.TxExpense:
			expense = s.Total
		}
	}
	return income, expense, nil
}

// CashFlow returns income/expense totals grouped by month for the last `months`
// months (oldest first), including months with no movement.
func (r *TransactionRepository) CashFlow(orgID uuid.UUID, months int) ([]dto.CashFlowPoint, error) {
	now := time.Now().UTC()
	currentStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	start := currentStart.AddDate(0, -(months - 1), 0)

	type row struct {
		Month string
		Type  entities.TransactionType
		Total int64
	}
	var rows []row
	err := r.db.Model(&entities.Transaction{}).
		Select("to_char(date, 'YYYY-MM') as month, type, COALESCE(SUM(amount),0) as total").
		Where("organization_id = ? AND paid = ? AND date >= ? AND type IN ?",
			orgID, true, start, []entities.TransactionType{entities.TxIncome, entities.TxExpense}).
		Group("month, type").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	// Pre-seed every month so the chart has no gaps.
	index := map[string]*dto.CashFlowPoint{}
	points := make([]dto.CashFlowPoint, 0, months)
	for i := 0; i < months; i++ {
		key := start.AddDate(0, i, 0).Format("2006-01")
		points = append(points, dto.CashFlowPoint{Month: key})
	}
	for i := range points {
		index[points[i].Month] = &points[i]
	}
	for _, rw := range rows {
		p, ok := index[rw.Month]
		if !ok {
			continue
		}
		switch rw.Type {
		case entities.TxIncome:
			p.Income = rw.Total
		case entities.TxExpense:
			p.Expense = rw.Total
		}
	}
	return points, nil
}

// TopCategories returns the highest-spending expense categories within the
// given window (paid expenses only), capped at `limit`.
func (r *TransactionRepository) TopCategories(orgID uuid.UUID, from, to time.Time, limit int) ([]dto.CategorySpend, error) {
	var out []dto.CategorySpend
	err := r.db.Model(&entities.Transaction{}).
		Select("categories.id as category_id, categories.name as name, categories.color as color, COALESCE(SUM(transactions.amount),0) as total").
		Joins("JOIN categories ON categories.id = transactions.category_id").
		Where("transactions.organization_id = ? AND transactions.paid = ? AND transactions.type = ? AND transactions.date >= ? AND transactions.date < ? AND transactions.category_id IS NOT NULL",
			orgID, true, entities.TxExpense, from, to).
		Group("categories.id, categories.name, categories.color").
		Order("total desc").
		Limit(limit).
		Scan(&out).Error
	return out, err
}

// SummaryByCategory returns the settled (paid) income+expense totals grouped by
// category within [from, to), org-scoped. Uncategorised rows are excluded. The
// category kind is carried so the frontend can split income vs expense.
func (r *TransactionRepository) SummaryByCategory(orgID uuid.UUID, from, to time.Time) ([]dto.ReportCategoryTotal, error) {
	var out []dto.ReportCategoryTotal
	err := r.db.Model(&entities.Transaction{}).
		Select("categories.id as category_id, categories.name as name, categories.color as color, categories.kind as kind, COALESCE(SUM(transactions.amount),0) as total").
		Joins("JOIN categories ON categories.id = transactions.category_id").
		Where("transactions.organization_id = ? AND transactions.paid = ? AND transactions.type IN ? AND transactions.date >= ? AND transactions.date < ? AND transactions.category_id IS NOT NULL",
			orgID, true, []entities.TransactionType{entities.TxIncome, entities.TxExpense}, from, to).
		Group("categories.id, categories.name, categories.color, categories.kind").
		Order("total desc").
		Scan(&out).Error
	return out, err
}

// SummaryByMonth returns settled income/expense grouped by YYYY-MM within
// [from, to), org-scoped, oldest month first.
func (r *TransactionRepository) SummaryByMonth(orgID uuid.UUID, from, to time.Time) ([]dto.ReportMonthTotal, error) {
	type row struct {
		Month string
		Type  entities.TransactionType
		Total int64
	}
	var rows []row
	err := r.db.Model(&entities.Transaction{}).
		Select("to_char(date, 'YYYY-MM') as month, type, COALESCE(SUM(amount),0) as total").
		Where("organization_id = ? AND paid = ? AND type IN ? AND date >= ? AND date < ?",
			orgID, true, []entities.TransactionType{entities.TxIncome, entities.TxExpense}, from, to).
		Group("month, type").
		Order("month asc").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	// Aggregate into a stable map keyed by month, preserving first-seen order
	// (rows already come ordered by month asc from the query).
	agg := map[string]*dto.ReportMonthTotal{}
	order := make([]string, 0, len(rows))
	for _, rw := range rows {
		p, ok := agg[rw.Month]
		if !ok {
			p = &dto.ReportMonthTotal{Month: rw.Month}
			agg[rw.Month] = p
			order = append(order, rw.Month)
		}
		switch rw.Type {
		case entities.TxIncome:
			p.Income = rw.Total
		case entities.TxExpense:
			p.Expense = rw.Total
		}
	}
	out := make([]dto.ReportMonthTotal, 0, len(order))
	for _, m := range order {
		out = append(out, *agg[m])
	}
	return out, nil
}

// UpcomingBills returns unpaid future expenses (bills due soon), ordered by the
// nearest due date, capped at `limit`. Category is preloaded.
func (r *TransactionRepository) UpcomingBills(orgID uuid.UUID, limit int) ([]entities.Transaction, error) {
	var txs []entities.Transaction
	err := r.db.Preload("Category").
		Where("organization_id = ? AND paid = ? AND type = ? AND date >= ?",
			orgID, false, entities.TxExpense, time.Now().UTC()).
		Order("date asc").
		Limit(limit).
		Find(&txs).Error
	return txs, err
}

// CountAll returns the total number of transactions for the org (used for plan
// quota enforcement).
func (r *TransactionRepository) CountAll(orgID uuid.UUID) (int64, error) {
	var n int64
	err := r.db.Model(&entities.Transaction{}).Where("organization_id = ?", orgID).Count(&n).Error
	return n, err
}

// ----- Bulk operations (org-scoped) -----
//
// Every bulk method runs a single UPDATE/DELETE whose WHERE clause is scoped by
// organization_id AND id IN (?), so the operation can never touch rows outside
// the caller's tenant even if foreign ids are supplied. They return the number
// of affected rows.

// BulkSettle marks the given transactions paid, setting paid_at. When accountID
// is non-nil the rows are also moved to that account. Already-validated ids /
// account are expected (the service validates account ownership beforehand).
func (r *TransactionRepository) BulkSettle(orgID uuid.UUID, ids []uuid.UUID, accountID *uuid.UUID, paidAt time.Time) (int64, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	updates := map[string]interface{}{"paid": true, "paid_at": paidAt}
	if accountID != nil {
		updates["account_id"] = *accountID
	}
	res := r.db.Model(&entities.Transaction{}).
		Where("organization_id = ? AND id IN ?", orgID, ids).
		Updates(updates)
	return res.RowsAffected, res.Error
}

// BulkUnsettle clears paid/paid_at on the given transactions.
func (r *TransactionRepository) BulkUnsettle(orgID uuid.UUID, ids []uuid.UUID) (int64, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	res := r.db.Model(&entities.Transaction{}).
		Where("organization_id = ? AND id IN ?", orgID, ids).
		Updates(map[string]interface{}{"paid": false, "paid_at": nil})
	return res.RowsAffected, res.Error
}

// BulkCategorize sets category_id on the given transactions. The category is
// expected to be validated against the org by the service.
func (r *TransactionRepository) BulkCategorize(orgID uuid.UUID, ids []uuid.UUID, categoryID uuid.UUID) (int64, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	res := r.db.Model(&entities.Transaction{}).
		Where("organization_id = ? AND id IN ?", orgID, ids).
		Update("category_id", categoryID)
	return res.RowsAffected, res.Error
}

// BulkDelete soft-deletes the given transactions (GORM honours the DeletedAt on
// Base, so this is a soft delete).
func (r *TransactionRepository) BulkDelete(orgID uuid.UUID, ids []uuid.UUID) (int64, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	res := r.db.Where("organization_id = ? AND id IN ?", orgID, ids).
		Delete(&entities.Transaction{})
	return res.RowsAffected, res.Error
}

// dueExpr is the SQL expression for the effective due date: DueDate when set,
// otherwise the competência Date. Used by AP/AR and the forecast so a bill
// without an explicit vencimento still has a sensible due reference.
const dueExpr = "COALESCE(due_date, date)"

// payablesScoped builds the base query for AP/AR: org + type + status + window.
// The status filter operates on Paid and the effective due date relative to
// `today`: open = !paid & due >= today; overdue = !paid & due < today;
// paid = paid; all/"" = no paid/overdue constraint.
func (r *TransactionRepository) payablesScoped(orgID uuid.UUID, txType entities.TransactionType, f dto.PayableFilter, today time.Time) *gorm.DB {
	q := r.db.Model(&entities.Transaction{}).
		Where("organization_id = ? AND type = ?", orgID, txType)
	switch f.Status {
	case "open":
		q = q.Where("paid = ? AND "+dueExpr+" >= ?", false, today)
	case "overdue":
		q = q.Where("paid = ? AND "+dueExpr+" < ?", false, today)
	case "paid":
		q = q.Where("paid = ?", true)
	}
	if f.ContactID != "" {
		q = q.Where("contact_id = ?", f.ContactID)
	}
	if f.From != nil {
		q = q.Where(dueExpr+" >= ?", *f.From)
	}
	if f.To != nil {
		q = q.Where(dueExpr+" <= ?", *f.To)
	}
	return q
}

// Payables returns a page of AP/AR transactions of the given type matching the
// filter, ordered by effective due date ascending, plus the total count.
// Category and Contact are preloaded for display.
func (r *TransactionRepository) Payables(orgID uuid.UUID, txType entities.TransactionType, f dto.PayableFilter, today time.Time) ([]entities.Transaction, int64, error) {
	var total int64
	if err := r.payablesScoped(orgID, txType, f, today).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (f.Page - 1) * f.PerPage
	var txs []entities.Transaction
	err := r.payablesScoped(orgID, txType, f, today).
		Preload("Category").
		Preload("Contact").
		Order(dueExpr + " asc, created_at asc").
		Limit(f.PerPage).
		Offset(offset).
		Find(&txs).Error
	return txs, total, err
}

// PayablesSummary computes the open/overdue/due-soon aggregates for AP/AR of the
// given type, org-scoped and ignoring the status/window filter (it always
// reflects the full unpaid picture). next7 is the inclusive due-soon horizon.
func (r *TransactionRepository) PayablesSummary(orgID uuid.UUID, txType entities.TransactionType, today, next7 time.Time) (dto.PayablesSummary, error) {
	var s dto.PayablesSummary

	// Each aggregate runs a fresh query (no shared builder) to avoid condition
	// bleed between the three SELECTs.
	base := func() *gorm.DB {
		return r.db.Model(&entities.Transaction{}).
			Where("organization_id = ? AND type = ? AND paid = ?", orgID, txType, false)
	}

	// Total + count of all open (unpaid, due today or later).
	type agg struct {
		Total int64
		Cnt   int64
	}
	var open agg
	if err := base().
		Select("COALESCE(SUM(amount),0) as total, COUNT(*) as cnt").
		Where(dueExpr+" >= ?", today).
		Scan(&open).Error; err != nil {
		return s, err
	}
	s.TotalOpen = open.Total
	s.CountOpen = open.Cnt

	if err := base().
		Select("COALESCE(SUM(amount),0)").
		Where(dueExpr+" < ?", today).
		Scan(&s.TotalOverdue).Error; err != nil {
		return s, err
	}

	if err := base().
		Select("COALESCE(SUM(amount),0)").
		Where(dueExpr+" >= ? AND "+dueExpr+" <= ?", today, next7).
		Scan(&s.DueNext7d).Error; err != nil {
		return s, err
	}
	return s, nil
}

// UnpaidByMonth returns the SUM of unpaid transactions of the given type grouped
// by the YYYY-MM of their effective due date, from `from` (inclusive) onward.
// Returns a map month -> total (cents). Used by the cash-flow forecast.
func (r *TransactionRepository) UnpaidByMonth(orgID uuid.UUID, txType entities.TransactionType, from time.Time) (map[string]int64, error) {
	type row struct {
		Month string
		Total int64
	}
	var rows []row
	err := r.db.Model(&entities.Transaction{}).
		Select("to_char("+dueExpr+", 'YYYY-MM') as month, COALESCE(SUM(amount),0) as total").
		Where("organization_id = ? AND type = ? AND paid = ? AND "+dueExpr+" >= ?",
			orgID, txType, false, from).
		Group("month").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make(map[string]int64, len(rows))
	for _, rw := range rows {
		out[rw.Month] = rw.Total
	}
	return out, nil
}
