package services

import (
	"github.com/finance-sh/finance-sh/internal/dto"
	"github.com/finance-sh/finance-sh/internal/entities"
	"github.com/finance-sh/finance-sh/internal/repositories"
	"github.com/google/uuid"
)

type NotificationService struct {
	notifications *repositories.NotificationRepository
}

func NewNotificationService(notifications *repositories.NotificationRepository) *NotificationService {
	return &NotificationService{notifications: notifications}
}

func (s *NotificationService) List(orgID uuid.UUID) ([]dto.NotificationDTO, error) {
	items, err := s.notifications.List(orgID)
	if err != nil {
		return nil, err
	}
	out := make([]dto.NotificationDTO, 0, len(items))
	for i := range items {
		out = append(out, notificationDTO(&items[i]))
	}
	return out, nil
}

func (s *NotificationService) MarkRead(orgID, id uuid.UUID) error {
	return s.notifications.MarkRead(orgID, id)
}

func (s *NotificationService) MarkAllRead(orgID uuid.UUID) error {
	return s.notifications.MarkAllRead(orgID)
}

func notificationDTO(n *entities.Notification) dto.NotificationDTO {
	return dto.NotificationDTO{
		ID:        n.ID.String(),
		Type:      n.Type,
		Title:     n.Title,
		Message:   n.Message,
		Read:      n.Read,
		CreatedAt: n.CreatedAt,
	}
}
