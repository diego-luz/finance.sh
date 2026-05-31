package services

import (
	"context"
	"errors"
	"math"
	"strconv"
	"time"

	"github.com/finance-sh/finance-sh/internal/cards"
	"github.com/finance-sh/finance-sh/internal/dto"
	"github.com/finance-sh/finance-sh/internal/entities"
	"github.com/finance-sh/finance-sh/internal/repositories"
	"github.com/finance-sh/finance-sh/pkg/cache"
	"github.com/finance-sh/finance-sh/pkg/crypto"
	"github.com/google/uuid"
)

// ErrAccountNotInOrg is returned when a transaction references an account that
// does not belong to the caller's organization (tenant guard).
var ErrAccountNotInOrg = errors.New("conta não pertence à organização")

// ErrCreditCardNotInOrg is returned when a transaction references a credit card
// that does not belong to the caller's organization (tenant guard).
var ErrCreditCardNotInOrg = errors.New("cartão de crédito não pertence à organização")

// ErrContactNotInOrg is returned when a transaction references a contact that
// does not belong to the caller's organization (tenant guard).
var ErrContactNotInOrg = errors.New("contato não pertence à organização")

type TransactionService struct {
	txs         *repositories.TransactionRepository
	accounts    *repositories.AccountRepository
	categories  *repositories.CategoryRepository
	creditCards *repositories.CreditCardRepository
	contacts    *repositories.ContactRepository
	tags        *repositories.TagRepository
	attachments *repositories.AttachmentRepository
	cache       *cache.Cache
}

func NewTransactionService(
	txs *repositories.TransactionRepository,
	accounts *repositories.AccountRepository,
	categories *repositories.CategoryRepository,
	creditCards *repositories.CreditCardRepository,
	contacts *repositories.ContactRepository,
	tags *repositories.TagRepository,
	attachments *repositories.AttachmentRepository,
	c *cache.Cache,
) *TransactionService {
	return &TransactionService{txs: txs, accounts: accounts, categories: categories, creditCards: creditCards, contacts: contacts, tags: tags, attachments: attachments, cache: c}
}

// applyTags resolves the request's TagIDs to the org's tags and sets the
// transaction's Tags association to exactly those (Replace also clears on an
// empty list). Foreign ids (not belonging to the org) are silently skipped.
// Invalid (non-UUID) ids are ignored so a malformed entry can't fail the write.
func (s *TransactionService) applyTags(orgID uuid.UUID, t *entities.Transaction, tagIDs []string) error {
	ids := make([]uuid.UUID, 0, len(tagIDs))
	for _, raw := range tagIDs {
		if id, err := parseUUID(raw); err == nil {
			ids = append(ids, id)
		}
	}
	tags, err := s.tags.FindByIDs(orgID, ids)
	if err != nil {
		return err
	}
	return s.txs.ReplaceTags(t, tags)
}

func (s *TransactionService) invalidateDashboard(orgID uuid.UUID) {
	s.cache.Delete(context.Background(), DashboardCacheKey(orgID))
}

// List applies sane pagination defaults then returns a page of transactions
// with pagination metadata.
func (s *TransactionService) List(orgID uuid.UUID, f dto.TransactionFilter) ([]dto.TransactionDTO, dto.PageMeta, error) {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.PerPage < 1 || f.PerPage > 100 {
		f.PerPage = 20
	}

	rows, total, err := s.txs.List(orgID, f)
	if err != nil {
		return nil, dto.PageMeta{}, err
	}

	out := make([]dto.TransactionDTO, 0, len(rows))
	for i := range rows {
		out = append(out, transactionDTO(&rows[i]))
	}
	// Attachment indicator: one grouped query for the whole page (no N+1).
	s.applyAttachmentCounts(orgID, rows, out)

	pages := int(math.Ceil(float64(total) / float64(f.PerPage)))
	meta := dto.PageMeta{Page: f.Page, PerPage: f.PerPage, Total: total, Pages: pages}
	return out, meta, nil
}

// applyAttachmentCounts fills the AttachmentCount field of each DTO using a
// single grouped query for the supplied rows. Best-effort: on error the counts
// stay 0 (the indicator is non-critical and must not fail the listing).
func (s *TransactionService) applyAttachmentCounts(orgID uuid.UUID, rows []entities.Transaction, out []dto.TransactionDTO) {
	if s.attachments == nil || len(rows) == 0 {
		return
	}
	ids := make([]uuid.UUID, 0, len(rows))
	for i := range rows {
		ids = append(ids, rows[i].ID)
	}
	counts, err := s.attachments.CountByTransactions(orgID, ids)
	if err != nil {
		return
	}
	for i := range out {
		out[i].AttachmentCount = counts[rows[i].ID]
	}
}

func (s *TransactionService) Get(orgID, id uuid.UUID) (*dto.TransactionDTO, error) {
	t, err := s.txs.FindByID(orgID, id)
	if err != nil {
		return nil, err
	}
	d := transactionDTO(t)
	// Attachment indicator on the single GET (best-effort; 0 on error).
	if s.attachments != nil {
		if counts, err := s.attachments.CountByTransactions(orgID, []uuid.UUID{t.ID}); err == nil {
			d.AttachmentCount = counts[t.ID]
		}
	}
	return &d, nil
}

func (s *TransactionService) Create(orgID uuid.UUID, req dto.TransactionRequest) (*dto.TransactionDTO, error) {
	// Parcelamento: when Installments > 1, generate a group of N sibling
	// transactions instead of a single row. The DTO returned is the FIRST parcela.
	if req.Installments > 1 {
		first, err := s.createInstallmentGroup(orgID, req)
		if err != nil {
			return nil, err
		}
		s.invalidateDashboard(orgID)
		return s.reload(orgID, first.ID)
	}

	t, err := s.buildEntity(orgID, &entities.Transaction{OrganizationID: orgID}, req)
	if err != nil {
		return nil, err
	}
	if err := s.txs.Create(t); err != nil {
		return nil, err
	}
	// Attach tags (org-scoped) to the freshly-saved transaction.
	if err := s.applyTags(orgID, t, req.TagIDs); err != nil {
		return nil, err
	}
	s.invalidateDashboard(orgID)
	return s.reload(orgID, t.ID)
}

// createInstallmentGroup materialises a parcelado purchase as N transactions
// sharing one InstallmentGroupID, persisted in a single DB transaction.
//
// Amount split: each parcela = amount / N (integer cents); the remainder
// (amount - per*N) is added to the FIRST parcela so the parcelas sum to the
// original amount exactly. Parcela i (1..N) is dated baseDate + (i-1) months and
// described as "<desc> (i/N)". When a credit card is set, each parcela's DueDate
// is the invoice due date for that parcela's month (card cycle).
func (s *TransactionService) createInstallmentGroup(orgID uuid.UUID, req dto.TransactionRequest) (*entities.Transaction, error) {
	// Validate references and map the request onto a template entity. The template
	// carries the resolved account/category/card/contact ids; each parcela clones it.
	tmpl, err := s.buildEntity(orgID, &entities.Transaction{OrganizationID: orgID}, req)
	if err != nil {
		return nil, err
	}

	// When the purchase is on a credit card, resolve the card so we can derive each
	// parcela's invoice due date from the billing cycle.
	var card *entities.CreditCard
	if tmpl.CreditCardID != nil {
		card, err = s.creditCards.FindByID(orgID, *tmpl.CreditCardID)
		if err != nil {
			if errors.Is(err, repositories.ErrNotFound) {
				return nil, ErrCreditCardNotInOrg
			}
			return nil, err
		}
	}

	n := req.Installments
	per := tmpl.Amount / int64(n)
	remainder := tmpl.Amount - per*int64(n)

	groupID := uuid.New()
	parcelas := make([]*entities.Transaction, 0, n)
	for i := 1; i <= n; i++ {
		amount := per
		if i == 1 {
			amount += remainder // keep the sum exact
		}
		date := tmpl.Date.AddDate(0, i-1, 0)

		p := &entities.Transaction{
			OrganizationID:     orgID,
			AccountID:          tmpl.AccountID,
			CategoryID:         tmpl.CategoryID,
			CreditCardID:       tmpl.CreditCardID,
			TransferAccountID:  tmpl.TransferAccountID,
			ContactID:          tmpl.ContactID,
			Type:               tmpl.Type,
			Amount:             amount,
			Description:        installmentDescription(req.Description, i, n),
			Date:               date,
			Paid:               false,
			Recurring:          false,
			Notes:              tmpl.Notes,
			InstallmentGroupID: &groupID,
			InstallmentNumber:  i,
			InstallmentTotal:   n,
		}
		// Due date: when on a card, use the invoice cycle for the parcela's month;
		// otherwise fall back to any explicit DueDate offset by the same months.
		if card != nil {
			inv := cards.InvoiceFor(card.ClosingDay, card.DueDay, date)
			due := inv.DueDate
			p.DueDate = &due
		} else if req.DueDate != nil {
			d := req.DueDate.AddDate(0, i-1, 0)
			p.DueDate = &d
		}
		parcelas = append(parcelas, p)
	}

	if err := s.txs.CreateMany(parcelas); err != nil {
		return nil, err
	}
	// Attach the same tags to every parcela of the group (org-scoped).
	if len(req.TagIDs) > 0 {
		for _, p := range parcelas {
			if err := s.applyTags(orgID, p, req.TagIDs); err != nil {
				return nil, err
			}
		}
	}
	return parcelas[0], nil
}

// installmentDescription appends the " (i/N)" suffix to a parcela description.
func installmentDescription(base string, i, n int) string {
	return base + " (" + itoa(i) + "/" + itoa(n) + ")"
}

func itoa(n int) string { return strconv.Itoa(n) }

func (s *TransactionService) Update(orgID, id uuid.UUID, req dto.TransactionRequest) (*dto.TransactionDTO, error) {
	existing, err := s.txs.FindByID(orgID, id)
	if err != nil {
		return nil, err
	}
	t, err := s.buildEntity(orgID, existing, req)
	if err != nil {
		return nil, err
	}
	if err := s.txs.Update(t); err != nil {
		return nil, err
	}
	// Resync tags (org-scoped). An empty/absent list clears them (Replace).
	if err := s.applyTags(orgID, t, req.TagIDs); err != nil {
		return nil, err
	}
	s.invalidateDashboard(orgID)
	return s.reload(orgID, id)
}

func (s *TransactionService) Delete(orgID, id uuid.UUID) error {
	if err := s.txs.Delete(orgID, id); err != nil {
		return err
	}
	s.invalidateDashboard(orgID)
	return nil
}

// DeleteScoped deletes a transaction. With scope="all" and a transaction that
// belongs to an installment group, the WHOLE group is removed; otherwise (or
// scope="one") only the single transaction is removed. Org-scoped throughout.
func (s *TransactionService) DeleteScoped(orgID, id uuid.UUID, scope string) error {
	if scope == "all" {
		t, err := s.txs.FindByID(orgID, id)
		if err != nil {
			return err
		}
		if t.InstallmentGroupID != nil {
			if _, err := s.txs.DeleteInstallmentGroup(orgID, *t.InstallmentGroupID); err != nil {
				return err
			}
			s.invalidateDashboard(orgID)
			return nil
		}
		// Not part of a group: fall through to single-row delete.
	}
	return s.Delete(orgID, id)
}

// Settle marks a transaction as paid, recording PaidAt (default now) and
// optionally moving it to another account (validated against the org).
func (s *TransactionService) Settle(orgID, id uuid.UUID, req dto.SettleRequest) (*dto.TransactionDTO, error) {
	t, err := s.txs.FindByID(orgID, id)
	if err != nil {
		return nil, err
	}

	if req.AccountID != "" {
		acctID, err := parseUUID(req.AccountID)
		if err != nil {
			return nil, err
		}
		if _, err := s.accounts.FindByID(orgID, acctID); err != nil {
			if errors.Is(err, repositories.ErrNotFound) {
				return nil, ErrAccountNotInOrg
			}
			return nil, err
		}
		t.AccountID = acctID
	}

	paidAt := time.Now().UTC()
	if req.PaidAt != nil {
		paidAt = *req.PaidAt
	}
	t.Paid = true
	t.PaidAt = &paidAt

	if err := s.txs.Update(t); err != nil {
		return nil, err
	}
	s.invalidateDashboard(orgID)
	return s.reload(orgID, id)
}

// Unsettle reverses a settlement: clears Paid and PaidAt.
func (s *TransactionService) Unsettle(orgID, id uuid.UUID) (*dto.TransactionDTO, error) {
	t, err := s.txs.FindByID(orgID, id)
	if err != nil {
		return nil, err
	}
	t.Paid = false
	t.PaidAt = nil
	if err := s.txs.Update(t); err != nil {
		return nil, err
	}
	s.invalidateDashboard(orgID)
	return s.reload(orgID, id)
}

// ----- Bulk actions -----
//
// Each bulk method is org-scoped end-to-end: ids are filtered by
// organization_id in the repo's WHERE clause, so foreign ids are silently
// ignored (they simply don't match). Referenced account/category are validated
// against the org before the write. The dashboard cache is invalidated once.

// BulkSettle marks several transactions paid (paid_at defaults to now), with an
// optional move to AccountID (validated against the org).
func (s *TransactionService) BulkSettle(orgID uuid.UUID, req dto.BulkSettleRequest) (int64, error) {
	ids, err := parseUUIDs(req.IDs)
	if err != nil {
		return 0, err
	}

	var accountID *uuid.UUID
	if req.AccountID != "" {
		acctID, err := parseUUID(req.AccountID)
		if err != nil {
			return 0, err
		}
		if _, err := s.accounts.FindByID(orgID, acctID); err != nil {
			if errors.Is(err, repositories.ErrNotFound) {
				return 0, ErrAccountNotInOrg
			}
			return 0, err
		}
		accountID = &acctID
	}

	paidAt := time.Now().UTC()
	if req.PaidAt != nil {
		paidAt = *req.PaidAt
	}

	n, err := s.txs.BulkSettle(orgID, ids, accountID, paidAt)
	if err != nil {
		return 0, err
	}
	s.invalidateDashboard(orgID)
	return n, nil
}

// BulkUnsettle reverses settlement on several transactions.
func (s *TransactionService) BulkUnsettle(orgID uuid.UUID, req dto.BulkIDsRequest) (int64, error) {
	ids, err := parseUUIDs(req.IDs)
	if err != nil {
		return 0, err
	}
	n, err := s.txs.BulkUnsettle(orgID, ids)
	if err != nil {
		return 0, err
	}
	s.invalidateDashboard(orgID)
	return n, nil
}

// BulkCategorize assigns one category to several transactions. The category must
// belong to the org. Kind compatibility is loose: we do not reject mixing an
// income category onto expense rows (the frontend lets the user re-tag freely).
func (s *TransactionService) BulkCategorize(orgID uuid.UUID, req dto.BulkCategorizeRequest) (int64, error) {
	ids, err := parseUUIDs(req.IDs)
	if err != nil {
		return 0, err
	}
	categoryID, err := parseUUID(req.CategoryID)
	if err != nil {
		return 0, err
	}
	if _, err := s.categories.FindByID(orgID, categoryID); err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			return 0, ErrNotFound
		}
		return 0, err
	}
	n, err := s.txs.BulkCategorize(orgID, ids, categoryID)
	if err != nil {
		return 0, err
	}
	s.invalidateDashboard(orgID)
	return n, nil
}

// BulkDelete soft-deletes several transactions.
func (s *TransactionService) BulkDelete(orgID uuid.UUID, req dto.BulkIDsRequest) (int64, error) {
	ids, err := parseUUIDs(req.IDs)
	if err != nil {
		return 0, err
	}
	n, err := s.txs.BulkDelete(orgID, ids)
	if err != nil {
		return 0, err
	}
	s.invalidateDashboard(orgID)
	return n, nil
}

// Payables lists accounts-payable (expenses) for the org with the AP summary.
func (s *TransactionService) Payables(orgID uuid.UUID, f dto.PayableFilter) ([]dto.TransactionDTO, dto.PageMeta, dto.PayablesSummary, error) {
	return s.listPayables(orgID, entities.TxExpense, f)
}

// Receivables lists accounts-receivable (income) for the org with the AR summary.
func (s *TransactionService) Receivables(orgID uuid.UUID, f dto.PayableFilter) ([]dto.TransactionDTO, dto.PageMeta, dto.PayablesSummary, error) {
	return s.listPayables(orgID, entities.TxIncome, f)
}

func (s *TransactionService) listPayables(orgID uuid.UUID, txType entities.TransactionType, f dto.PayableFilter) ([]dto.TransactionDTO, dto.PageMeta, dto.PayablesSummary, error) {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.PerPage < 1 || f.PerPage > 100 {
		f.PerPage = 20
	}
	switch f.Status {
	case "open", "overdue", "paid", "all", "":
	default:
		f.Status = "open"
	}

	now := time.Now().UTC()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	next7 := today.AddDate(0, 0, 7)

	rows, total, err := s.txs.Payables(orgID, txType, f, today)
	if err != nil {
		return nil, dto.PageMeta{}, dto.PayablesSummary{}, err
	}
	out := make([]dto.TransactionDTO, 0, len(rows))
	for i := range rows {
		out = append(out, transactionDTO(&rows[i]))
	}

	summary, err := s.txs.PayablesSummary(orgID, txType, today, next7)
	if err != nil {
		return nil, dto.PageMeta{}, dto.PayablesSummary{}, err
	}

	pages := int(math.Ceil(float64(total) / float64(f.PerPage)))
	meta := dto.PageMeta{Page: f.Page, PerPage: f.PerPage, Total: total, Pages: pages}
	return out, meta, summary, nil
}

// buildEntity validates ownership of the referenced account, category and
// transfer account, then maps the request onto the target entity.
func (s *TransactionService) buildEntity(orgID uuid.UUID, t *entities.Transaction, req dto.TransactionRequest) (*entities.Transaction, error) {
	accountID, err := parseUUID(req.AccountID)
	if err != nil {
		return nil, err
	}
	if _, err := s.accounts.FindByID(orgID, accountID); err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			return nil, ErrAccountNotInOrg
		}
		return nil, err
	}

	var categoryID *uuid.UUID
	if req.CategoryID != "" {
		cid, err := parseUUID(req.CategoryID)
		if err != nil {
			return nil, err
		}
		if _, err := s.categories.FindByID(orgID, cid); err != nil {
			if errors.Is(err, repositories.ErrNotFound) {
				return nil, ErrNotFound
			}
			return nil, err
		}
		categoryID = &cid
	}

	var creditCardID *uuid.UUID
	if req.CreditCardID != "" {
		ccid, err := parseUUID(req.CreditCardID)
		if err != nil {
			return nil, err
		}
		if _, err := s.creditCards.FindByID(orgID, ccid); err != nil {
			if errors.Is(err, repositories.ErrNotFound) {
				return nil, ErrCreditCardNotInOrg
			}
			return nil, err
		}
		creditCardID = &ccid
	}

	var transferID *uuid.UUID
	if req.TransferAccountID != "" {
		tid, err := parseUUID(req.TransferAccountID)
		if err != nil {
			return nil, err
		}
		if _, err := s.accounts.FindByID(orgID, tid); err != nil {
			if errors.Is(err, repositories.ErrNotFound) {
				return nil, ErrAccountNotInOrg
			}
			return nil, err
		}
		transferID = &tid
	}

	var contactID *uuid.UUID
	if req.ContactID != "" {
		cid, err := parseUUID(req.ContactID)
		if err != nil {
			return nil, err
		}
		if _, err := s.contacts.FindByID(orgID, cid); err != nil {
			if errors.Is(err, repositories.ErrNotFound) {
				return nil, ErrContactNotInOrg
			}
			return nil, err
		}
		contactID = &cid
	}

	t.OrganizationID = orgID
	t.AccountID = accountID
	t.CategoryID = categoryID
	t.CreditCardID = creditCardID
	t.TransferAccountID = transferID
	t.ContactID = contactID
	t.Type = entities.TransactionType(req.Type)
	t.Amount = req.Amount
	t.Description = req.Description
	t.Date = req.Date
	t.DueDate = req.DueDate
	t.Paid = req.Paid
	// Recurring is no longer accepted on the write path (the RecurrenceRule engine
	// supersedes the legacy per-transaction flag). On create the entity defaults to
	// false; on update `t` is the existing row, so its current value is preserved.
	t.Notes = crypto.EncryptedString(req.Notes)
	return t, nil
}

func (s *TransactionService) reload(orgID, id uuid.UUID) (*dto.TransactionDTO, error) {
	t, err := s.txs.FindByID(orgID, id)
	if err != nil {
		return nil, err
	}
	d := transactionDTO(t)
	return &d, nil
}

func transactionDTO(t *entities.Transaction) dto.TransactionDTO {
	d := dto.TransactionDTO{
		ID:          t.ID.String(),
		AccountID:   t.AccountID.String(),
		Type:        string(t.Type),
		Amount:      t.Amount,
		Description: t.Description,
		Date:        t.Date,
		DueDate:     t.DueDate,
		PaidAt:      t.PaidAt,
		Paid:        t.Paid,
		Recurring:   t.Recurring,
		Status:      transactionStatus(t),
		Notes:       t.Notes.String(),
	}
	if t.InstallmentGroupID != nil {
		d.InstallmentGroupID = t.InstallmentGroupID.String()
	}
	// InstallmentTotal==0 means the row is not part of a group; the omitempty tags
	// keep number/total out of the payload in that case.
	if t.InstallmentTotal > 0 {
		d.InstallmentNumber = t.InstallmentNumber
		d.InstallmentTotal = t.InstallmentTotal
	}
	if t.CreditCardID != nil {
		d.CreditCardID = t.CreditCardID.String()
	}
	if t.ContactID != nil {
		d.ContactID = t.ContactID.String()
	}
	if t.Category != nil {
		c := categoryDTO(t.Category)
		d.Category = &c
	}
	if t.Contact != nil {
		d.Contact = &dto.TransactionContactDTO{ID: t.Contact.ID.String(), Name: t.Contact.Name}
	}
	if len(t.Tags) > 0 {
		d.Tags = make([]dto.TagDTO, 0, len(t.Tags))
		for i := range t.Tags {
			d.Tags = append(d.Tags, tagDTO(&t.Tags[i]))
		}
	}
	return d
}

// transactionStatus derives the lifecycle status from the persisted flags:
// "paid" when settled; otherwise "overdue" when the effective due date (DueDate
// when set, else Date) is before today; otherwise "open".
func transactionStatus(t *entities.Transaction) string {
	if t.Paid {
		return "paid"
	}
	due := t.Date
	if t.DueDate != nil {
		due = *t.DueDate
	}
	now := time.Now().UTC()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	if due.Before(today) {
		return "overdue"
	}
	return "open"
}
