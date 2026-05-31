// Package jobs holds the background jobs run by the in-process scheduler (see
// cmd/api): recurring transaction generation and notification creation. Jobs are
// idempotent so the scheduler can run them on every tick without duplicates.
package jobs

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/finance-sh/finance-sh/internal/entities"
	"github.com/finance-sh/finance-sh/internal/repositories"
	"github.com/finance-sh/finance-sh/internal/services"
	"github.com/finance-sh/finance-sh/pkg/cache"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Runner bundles the repositories the jobs need and a logger.
type Runner struct {
	db            *gorm.DB
	txs           *repositories.TransactionRepository
	budgets       *repositories.BudgetRepository
	notifications *repositories.NotificationRepository
	// recurrence is the proper recurring-transaction engine (RecurrenceRule),
	// constructed here so the worker can drive GenerateDue across all orgs. It is
	// the sole recurrence mechanism; the legacy Transaction.Recurring flag job has
	// been retired.
	recurrence    *services.RecurrenceService
	retentionDays int
	log           *slog.Logger
}

func NewRunner(db *gorm.DB, retentionDays int, log *slog.Logger) *Runner {
	if retentionDays < 1 {
		retentionDays = 90
	}
	// The worker runs in a separate process from the API; it spawns its own
	// in-memory cache so dashboard invalidation calls remain valid (the API
	// refreshes its own cache on read; TTL covers the rest).
	noCache := cache.New()
	recurrenceSvc := services.NewRecurrenceService(
		repositories.NewRecurrenceRuleRepository(db),
		repositories.NewTransactionRepository(db),
		repositories.NewAccountRepository(db),
		repositories.NewCategoryRepository(db),
		repositories.NewContactRepository(db),
		noCache,
		db,
	)
	return &Runner{
		db:            db,
		txs:           repositories.NewTransactionRepository(db),
		budgets:       repositories.NewBudgetRepository(db),
		notifications: repositories.NewNotificationRepository(db),
		recurrence:    recurrenceSvc,
		retentionDays: retentionDays,
		log:           log,
	}
}

// RunAll executes every job for every organization. Errors are logged but never
// abort the run so a single bad org cannot stall the worker.
func (r *Runner) RunAll() {
	orgs, err := r.orgIDs()
	if err != nil {
		r.log.Error("worker: failed to list organizations", "error", err)
		return
	}
	for _, orgID := range orgs {
		if err := r.GenerateBillNotifications(orgID); err != nil {
			r.log.Error("worker: GenerateBillNotifications failed", "org", orgID, "error", err)
		}
	}

	// Recurring-transaction engine (RecurrenceRule). Runs once per cycle: it is
	// global (walks every org's due rules in one query) rather than per-org.
	r.GenerateRecurrenceRules()

	// Data retention: hard-delete rows that have been soft-deleted longer than the
	// retention window. Runs once per cycle (it is global, not per-org).
	if err := r.PurgeSoftDeleted(); err != nil {
		r.log.Error("worker: PurgeSoftDeleted failed", "error", err)
	}

	r.log.Info("worker: cycle complete", "organizations", len(orgs))
}

// GenerateRecurrenceRules drives the proper recurring-transaction engine: it
// generates the transactions every due RecurrenceRule owes up to now, across all
// orgs, and logs how many were created. Idempotent: a rule's NextRunDate is
// advanced past `now` as occurrences are generated, so a re-run produces nothing.
func (r *Runner) GenerateRecurrenceRules() {
	now := time.Now().UTC()
	created, err := r.recurrence.GenerateDue(now)
	if err != nil {
		r.log.Error("worker: GenerateRecurrenceRules failed", "error", err)
		return
	}
	r.log.Info("worker: recurrence rules generated", "transactions", created)
}

// PurgeSoftDeleted permanently removes rows whose deleted_at is older than the
// configured retention window across the financial/PII tables. It uses Unscoped
// so GORM performs a real DELETE rather than another soft-delete. The job is
// idempotent (re-running with nothing to purge is a no-op) and logs its work.
func (r *Runner) PurgeSoftDeleted() error {
	cutoff := time.Now().UTC().AddDate(0, 0, -r.retentionDays)

	// Order matters less here because we hard-delete by timestamp, but we keep a
	// child-before-parent ordering for clarity.
	tables := []struct {
		name  string
		model interface{}
	}{
		{"transactions", &entities.Transaction{}},
		{"budgets", &entities.Budget{}},
		{"goals", &entities.Goal{}},
		{"credit_cards", &entities.CreditCard{}},
		{"categories", &entities.Category{}},
		{"accounts", &entities.Account{}},
		{"notifications", &entities.Notification{}},
		{"invitations", &entities.Invitation{}},
		{"memberships", &entities.Membership{}},
		{"organizations", &entities.Organization{}},
	}

	var total int64
	for _, t := range tables {
		res := r.db.Unscoped().
			Where("deleted_at IS NOT NULL AND deleted_at < ?", cutoff).
			Delete(t.model)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected > 0 {
			total += res.RowsAffected
			r.log.Info("worker: purged soft-deleted rows", "table", t.name, "rows", res.RowsAffected)
		}
	}
	r.log.Info("worker: retention purge complete", "retention_days", r.retentionDays, "total_rows", total)
	return nil
}

// orgIDs returns the id of every organization.
func (r *Runner) orgIDs() ([]uuid.UUID, error) {
	var ids []uuid.UUID
	err := r.db.Model(&entities.Organization{}).Pluck("id", &ids).Error
	return ids, err
}

// GenerateBillNotifications creates "vencimento" notifications for unpaid
// expense bills due within three days (and overdue), plus "orcamento"
// notifications when a budget is overflowed. Both are idempotent via a stable
// message key.
func (r *Runner) GenerateBillNotifications(orgID uuid.UUID) error {
	now := time.Now().UTC()
	until := now.AddDate(0, 0, 3)

	bills, err := r.txs.DueUnpaidBills(orgID, until)
	if err != nil {
		return err
	}
	for i := range bills {
		b := &bills[i]
		message := fmt.Sprintf("Conta \"%s\" vence em %s (ref %s)",
			b.Description, b.Date.Format("02/01/2006"), b.ID.String())
		exists, err := r.notifications.ExistsByTypeAndMessage(orgID, "vencimento", message)
		if err != nil {
			return err
		}
		if exists {
			continue
		}
		if err := r.notifications.Create(&entities.Notification{
			OrganizationID: orgID,
			Type:           "vencimento",
			Title:          "Conta a vencer",
			Message:        message,
		}); err != nil {
			return err
		}
	}

	// Budget overflow notifications.
	budgets, err := r.budgets.ListAll(orgID)
	if err != nil {
		return err
	}
	for i := range budgets {
		bg := &budgets[i]
		spent, err := r.budgets.SpentByCategory(orgID, bg.CategoryID, bg.Month, bg.Year)
		if err != nil {
			return err
		}
		if spent <= bg.Amount {
			continue
		}
		message := fmt.Sprintf("Orçamento %d/%d estourado (ref %s)", bg.Month, bg.Year, bg.ID.String())
		exists, err := r.notifications.ExistsByTypeAndMessage(orgID, "orcamento", message)
		if err != nil {
			return err
		}
		if exists {
			continue
		}
		if err := r.notifications.Create(&entities.Notification{
			OrganizationID: orgID,
			Type:           "orcamento",
			Title:          "Orçamento estourado",
			Message:        message,
		}); err != nil {
			return err
		}
	}
	return nil
}
