package services

import (
	"context"
	"errors"
	"sort"
	"time"

	cardcycle "github.com/finance-sh/finance-sh/internal/cards"
	"github.com/finance-sh/finance-sh/internal/dto"
	"github.com/finance-sh/finance-sh/internal/entities"
	"github.com/finance-sh/finance-sh/internal/repositories"
	"github.com/finance-sh/finance-sh/pkg/cache"
	"github.com/google/uuid"
)

type CreditCardService struct {
	cards    *repositories.CreditCardRepository
	txs      *repositories.TransactionRepository
	accounts *repositories.AccountRepository
	cache    *cache.Cache
}

func NewCreditCardService(
	cards *repositories.CreditCardRepository,
	txs *repositories.TransactionRepository,
	accounts *repositories.AccountRepository,
	c *cache.Cache,
) *CreditCardService {
	return &CreditCardService{cards: cards, txs: txs, accounts: accounts, cache: c}
}

func (s *CreditCardService) List(orgID uuid.UUID) ([]dto.CreditCardDTO, error) {
	cards, err := s.cards.List(orgID)
	if err != nil {
		return nil, err
	}
	used, err := s.cards.UsedAmounts(orgID)
	if err != nil {
		return nil, err
	}
	out := make([]dto.CreditCardDTO, 0, len(cards))
	for i := range cards {
		out = append(out, creditCardDTO(&cards[i], used[cards[i].ID]))
	}
	return out, nil
}

func (s *CreditCardService) Get(orgID, id uuid.UUID) (*dto.CreditCardDTO, error) {
	c, err := s.cards.FindByID(orgID, id)
	if err != nil {
		return nil, err
	}
	used, err := s.cards.UsedAmounts(orgID)
	if err != nil {
		return nil, err
	}
	d := creditCardDTO(c, used[c.ID])
	return &d, nil
}

func (s *CreditCardService) Create(orgID uuid.UUID, req dto.CreditCardRequest) (*dto.CreditCardDTO, error) {
	c := &entities.CreditCard{
		OrganizationID: orgID,
		Name:           req.Name,
		Limit:          req.Limit,
		ClosingDay:     req.ClosingDay,
		DueDay:         req.DueDay,
		Color:          defaultStr(req.Color, "#0f1115"),
	}
	if err := s.cards.Create(c); err != nil {
		return nil, err
	}
	d := creditCardDTO(c, 0)
	return &d, nil
}

func (s *CreditCardService) Update(orgID, id uuid.UUID, req dto.CreditCardRequest) (*dto.CreditCardDTO, error) {
	c, err := s.cards.FindByID(orgID, id)
	if err != nil {
		return nil, err
	}
	c.Name = req.Name
	c.Limit = req.Limit
	c.ClosingDay = req.ClosingDay
	c.DueDay = req.DueDay
	if req.Color != "" {
		c.Color = req.Color
	}
	if err := s.cards.Update(c); err != nil {
		return nil, err
	}
	used, err := s.cards.UsedAmounts(orgID)
	if err != nil {
		return nil, err
	}
	d := creditCardDTO(c, used[c.ID])
	return &d, nil
}

func (s *CreditCardService) Delete(orgID, id uuid.UUID) error {
	return s.cards.Delete(orgID, id)
}

// ----- Invoices (faturas) -----
//
// An invoice is not stored: it is computed by grouping a card's transactions by
// the closing year-month their purchase date falls into (see internal/cards).
// This keeps the schema unchanged and the "used"/"available" math consistent.

// invoiceAgg accumulates the totals for one reference while we scan a card's
// transactions. It keeps the matching transaction entities for the detail view.
type invoiceAgg struct {
	inv   cardcycle.Invoice
	total int64
	paid  int64
	rows  []entities.Transaction
}

// Invoices returns the most recent `limit` invoices of a card (newest first),
// built by grouping the card's transactions by computed reference.
func (s *CreditCardService) Invoices(orgID, cardID uuid.UUID, limit int) ([]dto.InvoiceDTO, error) {
	card, err := s.cards.FindByID(orgID, cardID)
	if err != nil {
		return nil, err
	}
	groups, err := s.aggregateInvoices(orgID, card)
	if err != nil {
		return nil, err
	}

	// Sort references newest-first.
	refs := make([]string, 0, len(groups))
	for ref := range groups {
		refs = append(refs, ref)
	}
	sort.Sort(sort.Reverse(sort.StringSlice(refs)))

	if limit < 1 {
		limit = 12
	}
	if limit > len(refs) {
		limit = len(refs)
	}

	today := todayUTC()
	out := make([]dto.InvoiceDTO, 0, limit)
	for _, ref := range refs[:limit] {
		out = append(out, invoiceDTO(groups[ref], today))
	}
	return out, nil
}

// Invoice returns one invoice plus the card transactions in its cycle.
func (s *CreditCardService) Invoice(orgID, cardID uuid.UUID, reference string) (*dto.InvoiceDetailDTO, error) {
	card, err := s.cards.FindByID(orgID, cardID)
	if err != nil {
		return nil, err
	}
	groups, err := s.aggregateInvoices(orgID, card)
	if err != nil {
		return nil, err
	}

	today := todayUTC()
	g, ok := groups[reference]
	if !ok {
		// No transactions for this reference yet: still return a valid (empty)
		// invoice derived from the card cycle so the frontend can render it.
		inv, valid := cardcycle.InvoiceFromReference(card.ClosingDay, card.DueDay, reference)
		if !valid {
			return nil, ErrNotFound
		}
		return &dto.InvoiceDetailDTO{
			Invoice:      invoiceDTO(invoiceAgg{inv: inv}, today),
			Transactions: []dto.TransactionDTO{},
		}, nil
	}

	txDTOs := make([]dto.TransactionDTO, 0, len(g.rows))
	for i := range g.rows {
		txDTOs = append(txDTOs, transactionDTO(&g.rows[i]))
	}
	return &dto.InvoiceDetailDTO{
		Invoice:      invoiceDTO(g, today),
		Transactions: txDTOs,
	}, nil
}

// PayInvoice settles every transaction of an invoice: marks them paid, records
// paid_at and moves them to the paying account (validated against the org), so
// the bank balance reflects the payment and the card's "used" drops. Returns the
// recomputed invoice.
func (s *CreditCardService) PayInvoice(orgID, cardID uuid.UUID, reference string, req dto.InvoicePayRequest) (*dto.InvoiceDTO, error) {
	card, err := s.cards.FindByID(orgID, cardID)
	if err != nil {
		return nil, err
	}

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

	groups, err := s.aggregateInvoices(orgID, card)
	if err != nil {
		return nil, err
	}
	g, ok := groups[reference]
	if !ok || len(g.rows) == 0 {
		return nil, ErrNotFound
	}

	ids := make([]uuid.UUID, 0, len(g.rows))
	for i := range g.rows {
		ids = append(ids, g.rows[i].ID)
	}

	paidAt := time.Now().UTC()
	if req.PaidAt != nil {
		paidAt = *req.PaidAt
	}
	if _, err := s.txs.BulkSettle(orgID, ids, &acctID, paidAt); err != nil {
		return nil, err
	}
	s.invalidateDashboard(orgID)

	return s.invoiceAfterChange(orgID, card, reference)
}

// UnpayInvoice reverses a paid invoice: clears paid/paid_at on its transactions
// (the AccountID move is left in place, as for single unsettle).
func (s *CreditCardService) UnpayInvoice(orgID, cardID uuid.UUID, reference string) (*dto.InvoiceDTO, error) {
	card, err := s.cards.FindByID(orgID, cardID)
	if err != nil {
		return nil, err
	}
	groups, err := s.aggregateInvoices(orgID, card)
	if err != nil {
		return nil, err
	}
	g, ok := groups[reference]
	if !ok || len(g.rows) == 0 {
		return nil, ErrNotFound
	}
	ids := make([]uuid.UUID, 0, len(g.rows))
	for i := range g.rows {
		ids = append(ids, g.rows[i].ID)
	}
	if _, err := s.txs.BulkUnsettle(orgID, ids); err != nil {
		return nil, err
	}
	s.invalidateDashboard(orgID)
	return s.invoiceAfterChange(orgID, card, reference)
}

// invoiceAfterChange recomputes a single invoice after a bulk settle/unsettle.
func (s *CreditCardService) invoiceAfterChange(orgID uuid.UUID, card *entities.CreditCard, reference string) (*dto.InvoiceDTO, error) {
	groups, err := s.aggregateInvoices(orgID, card)
	if err != nil {
		return nil, err
	}
	today := todayUTC()
	if g, ok := groups[reference]; ok {
		d := invoiceDTO(g, today)
		return &d, nil
	}
	inv, valid := cardcycle.InvoiceFromReference(card.ClosingDay, card.DueDay, reference)
	if !valid {
		return nil, ErrNotFound
	}
	d := invoiceDTO(invoiceAgg{inv: inv}, today)
	return &d, nil
}

// aggregateInvoices groups all of a card's transactions by computed reference.
func (s *CreditCardService) aggregateInvoices(orgID uuid.UUID, card *entities.CreditCard) (map[string]invoiceAgg, error) {
	rows, err := s.txs.ListByCard(orgID, card.ID)
	if err != nil {
		return nil, err
	}
	groups := make(map[string]invoiceAgg)
	for i := range rows {
		t := rows[i]
		inv := cardcycle.InvoiceFor(card.ClosingDay, card.DueDay, t.Date)
		g := groups[inv.Reference]
		g.inv = inv
		g.total += t.Amount
		if t.Paid {
			g.paid += t.Amount
		}
		g.rows = append(g.rows, t)
		groups[inv.Reference] = g
	}
	return groups, nil
}

func (s *CreditCardService) invalidateDashboard(orgID uuid.UUID) {
	if s.cache == nil {
		return
	}
	s.cache.Delete(context.Background(), DashboardCacheKey(orgID))
}

func todayUTC() time.Time {
	now := time.Now().UTC()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
}

// invoiceDTO maps an aggregated invoice to its DTO and derives the status.
//
// Status rule:
//   - paid    — there are transactions and all are paid (open == 0, total > 0).
//   - open    — the cycle has not closed yet (today <= period_end), unpaid.
//   - overdue — closed, unpaid, and the due date is before today.
//   - closed  — closed, unpaid, due date not yet reached.
func invoiceDTO(g invoiceAgg, today time.Time) dto.InvoiceDTO {
	open := g.total - g.paid
	status := invoiceStatus(g, open, today)
	return dto.InvoiceDTO{
		Reference:        g.inv.Reference,
		PeriodStart:      g.inv.PeriodStart,
		PeriodEnd:        g.inv.PeriodEnd,
		DueDate:          g.inv.DueDate,
		Total:            g.total,
		PaidTotal:        g.paid,
		OpenTotal:        open,
		Status:           status,
		TransactionCount: len(g.rows),
	}
}

func invoiceStatus(g invoiceAgg, open int64, today time.Time) string {
	if g.total > 0 && open == 0 {
		return "paid"
	}
	// Not yet closed (current cycle): period_end is today or in the future.
	if !g.inv.PeriodEnd.Before(today) {
		return "open"
	}
	// Closed and unpaid.
	if g.inv.DueDate.Before(today) {
		return "overdue"
	}
	return "closed"
}

func creditCardDTO(c *entities.CreditCard, used int64) dto.CreditCardDTO {
	return dto.CreditCardDTO{
		ID:         c.ID.String(),
		Name:       c.Name,
		Limit:      c.Limit,
		ClosingDay: c.ClosingDay,
		DueDay:     c.DueDay,
		Color:      c.Color,
		Used:       used,
		Available:  c.Limit - used,
	}
}
