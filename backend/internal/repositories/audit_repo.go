package repositories

import (
	"github.com/finance-sh/finance-sh/internal/dto"
	"github.com/finance-sh/finance-sh/internal/entities"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AuditRepository reads the audit_logs table. Every query is scoped by
// organization_id so a tenant only ever sees its own audit trail. The table is
// written by the Audit middleware; this repository is read-only.
type AuditRepository struct{ db *gorm.DB }

func NewAuditRepository(db *gorm.DB) *AuditRepository {
	return &AuditRepository{db: db}
}

// scoped builds a base query already filtered by org and the supplied filter.
func (r *AuditRepository) scoped(orgID uuid.UUID, f dto.AuditFilter) *gorm.DB {
	q := r.db.Model(&entities.AuditLog{}).Where("organization_id = ?", orgID)
	if f.Action != "" {
		q = q.Where("action = ?", f.Action)
	}
	if f.Entity != "" {
		q = q.Where("entity = ?", f.Entity)
	}
	if f.UserID != nil {
		q = q.Where("user_id = ?", *f.UserID)
	}
	if f.From != nil {
		q = q.Where("created_at >= ?", *f.From)
	}
	if f.To != nil {
		q = q.Where("created_at <= ?", *f.To)
	}
	return q
}

// List returns a page of audit logs ordered by created_at desc, plus the total
// count for pagination. User information is not preloaded (UserID is just an id);
// the service resolves names via a batch query.
func (r *AuditRepository) List(orgID uuid.UUID, f dto.AuditFilter) ([]entities.AuditLog, int64, error) {
	var total int64
	if err := r.scoped(orgID, f).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (f.Page - 1) * f.PerPage
	var rows []entities.AuditLog
	err := r.scoped(orgID, f).
		Order("created_at desc").
		Limit(f.PerPage).
		Offset(offset).
		Find(&rows).Error
	return rows, total, err
}
