package services

import (
	"math"

	"github.com/finance-sh/finance-sh/internal/dto"
	"github.com/finance-sh/finance-sh/internal/entities"
	"github.com/finance-sh/finance-sh/internal/repositories"
	"github.com/google/uuid"
)

// AuditService reads the org's audit trail. It resolves the (optional) actor of
// each entry into a minimal user shape via a single batch query, avoiding an FK
// association on AuditLog.
type AuditService struct {
	audit *repositories.AuditRepository
	users *repositories.UserRepository
}

func NewAuditService(
	audit *repositories.AuditRepository,
	users *repositories.UserRepository,
) *AuditService {
	return &AuditService{audit: audit, users: users}
}

// List applies sane pagination defaults, fetches a page of audit logs, then
// batch-resolves the distinct non-nil user ids into a lookup map so each entry
// can carry the actor's name/email. Returns the page plus pagination metadata.
func (s *AuditService) List(orgID uuid.UUID, f dto.AuditFilter) ([]dto.AuditLogDTO, dto.PageMeta, error) {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.PerPage < 1 || f.PerPage > 100 {
		f.PerPage = 20
	}

	rows, total, err := s.audit.List(orgID, f)
	if err != nil {
		return nil, dto.PageMeta{}, err
	}

	// Collect the distinct non-nil user ids on this page.
	seen := make(map[uuid.UUID]struct{}, len(rows))
	ids := make([]uuid.UUID, 0, len(rows))
	for i := range rows {
		if rows[i].UserID == nil {
			continue
		}
		uid := *rows[i].UserID
		if _, ok := seen[uid]; ok {
			continue
		}
		seen[uid] = struct{}{}
		ids = append(ids, uid)
	}

	// Batch-load the users into an id -> user map (single query).
	userByID := make(map[uuid.UUID]*entities.User, len(ids))
	if len(ids) > 0 {
		users, err := s.users.FindByIDs(ids)
		if err != nil {
			return nil, dto.PageMeta{}, err
		}
		for i := range users {
			userByID[users[i].ID] = &users[i]
		}
	}

	out := make([]dto.AuditLogDTO, 0, len(rows))
	for i := range rows {
		var user *entities.User
		if rows[i].UserID != nil {
			user = userByID[*rows[i].UserID]
		}
		out = append(out, auditLogDTO(&rows[i], user))
	}

	pages := int(math.Ceil(float64(total) / float64(f.PerPage)))
	meta := dto.PageMeta{Page: f.Page, PerPage: f.PerPage, Total: total, Pages: pages}
	return out, meta, nil
}

// auditLogDTO maps an AuditLog (with its optional resolved user) to the public
// DTO. user is nil when the log has no user, or when the referenced user could
// not be found (e.g. deleted).
func auditLogDTO(l *entities.AuditLog, user *entities.User) dto.AuditLogDTO {
	d := dto.AuditLogDTO{
		ID:        l.ID.String(),
		CreatedAt: l.CreatedAt,
		Action:    l.Action,
		Entity:    l.Entity,
		EntityID:  l.EntityID,
		IP:        l.IP,
		Metadata:  l.Metadata,
	}
	if user != nil {
		d.User = &dto.UserMiniDTO{
			ID:    user.ID.String(),
			Name:  user.Name,
			Email: user.Email,
		}
	}
	return d
}
