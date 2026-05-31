package handlers

import (
	"net/http"

	"github.com/finance-sh/finance-sh/internal/middlewares"
	"github.com/finance-sh/finance-sh/internal/services"
	"github.com/finance-sh/finance-sh/pkg/response"
)

type NotificationHandler struct {
	notifications *services.NotificationService
}

func NewNotificationHandler(notifications *services.NotificationService) *NotificationHandler {
	return &NotificationHandler{notifications: notifications}
}

// List GET /notifications — newest first.
func (h *NotificationHandler) List(w http.ResponseWriter, r *http.Request) {
	orgID := middlewares.OrgID(r.Context())
	items, err := h.notifications.List(orgID)
	if writeServiceError(w, err) {
		return
	}
	response.OK(w, items)
}

// MarkRead POST /notifications/{id}/read
func (h *NotificationHandler) MarkRead(w http.ResponseWriter, r *http.Request) {
	orgID := middlewares.OrgID(r.Context())
	id, ok := urlUUID(w, r, "id")
	if !ok {
		return
	}
	if writeServiceError(w, h.notifications.MarkRead(orgID, id)) {
		return
	}
	response.NoContent(w)
}

// MarkAllRead POST /notifications/read-all
func (h *NotificationHandler) MarkAllRead(w http.ResponseWriter, r *http.Request) {
	orgID := middlewares.OrgID(r.Context())
	if writeServiceError(w, h.notifications.MarkAllRead(orgID)) {
		return
	}
	response.NoContent(w)
}
