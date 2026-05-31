package repositories

import (
	"github.com/finance-sh/finance-sh/internal/entities"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// NotificationRepository persists in-app notifications, scoped by organization.
type NotificationRepository struct{ db *gorm.DB }

func NewNotificationRepository(db *gorm.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

func (r *NotificationRepository) Create(n *entities.Notification) error {
	return r.db.Create(n).Error
}

// List returns the org's notifications, newest first.
func (r *NotificationRepository) List(orgID uuid.UUID) ([]entities.Notification, error) {
	var items []entities.Notification
	err := r.db.Where("organization_id = ?", orgID).
		Order("created_at desc").Find(&items).Error
	return items, err
}

// MarkRead flags a single notification as read (scoped by org).
func (r *NotificationRepository) MarkRead(orgID, id uuid.UUID) error {
	res := r.db.Model(&entities.Notification{}).
		Where("organization_id = ? AND id = ?", orgID, id).
		Update("read", true)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// MarkAllRead flags every unread notification of the org as read.
func (r *NotificationRepository) MarkAllRead(orgID uuid.UUID) error {
	return r.db.Model(&entities.Notification{}).
		Where("organization_id = ? AND read = ?", orgID, false).
		Update("read", true).Error
}

// ExistsForType reports whether a notification with the given type already
// references the entity id inside its message (idempotency guard for the
// worker). The worker encodes the transaction/budget id in the message.
func (r *NotificationRepository) ExistsByTypeAndMessage(orgID uuid.UUID, typ, message string) (bool, error) {
	var count int64
	err := r.db.Model(&entities.Notification{}).
		Where("organization_id = ? AND type = ? AND message = ?", orgID, typ, message).
		Count(&count).Error
	return count > 0, err
}
