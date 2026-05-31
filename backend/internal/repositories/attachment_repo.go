package repositories

import (
	"errors"

	"github.com/finance-sh/finance-sh/internal/entities"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AttachmentRepository persists transaction attachments (receipts/comprovantes),
// always scoped by organization_id.
type AttachmentRepository struct{ db *gorm.DB }

func NewAttachmentRepository(db *gorm.DB) *AttachmentRepository {
	return &AttachmentRepository{db: db}
}

func (r *AttachmentRepository) Create(a *entities.Attachment) error {
	return r.db.Create(a).Error
}

func (r *AttachmentRepository) FindByID(orgID, id uuid.UUID) (*entities.Attachment, error) {
	var a entities.Attachment
	err := r.db.Where("organization_id = ? AND id = ?", orgID, id).First(&a).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &a, err
}

// ListByTransaction returns every attachment of a transaction (org-scoped),
// newest first.
func (r *AttachmentRepository) ListByTransaction(orgID, txID uuid.UUID) ([]entities.Attachment, error) {
	var rows []entities.Attachment
	err := r.db.
		Where("organization_id = ? AND transaction_id = ?", orgID, txID).
		Order("created_at desc").
		Find(&rows).Error
	return rows, err
}

func (r *AttachmentRepository) Delete(orgID, id uuid.UUID) error {
	res := r.db.Where("organization_id = ? AND id = ?", orgID, id).Delete(&entities.Attachment{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// CountByTransactions returns a map txID -> attachment count for the given
// transaction ids (org-scoped), computed with a single grouped query so the
// transaction list can show an indicator without N+1 round-trips. Transactions
// with no attachment are simply absent from the map (caller treats them as 0).
func (r *AttachmentRepository) CountByTransactions(orgID uuid.UUID, txIDs []uuid.UUID) (map[uuid.UUID]int, error) {
	out := make(map[uuid.UUID]int, len(txIDs))
	if len(txIDs) == 0 {
		return out, nil
	}
	type row struct {
		TransactionID uuid.UUID
		Cnt           int
	}
	var rows []row
	err := r.db.Model(&entities.Attachment{}).
		Select("transaction_id, COUNT(*) as cnt").
		Where("organization_id = ? AND transaction_id IN ?", orgID, txIDs).
		Group("transaction_id").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	for _, rw := range rows {
		out[rw.TransactionID] = rw.Cnt
	}
	return out, nil
}
