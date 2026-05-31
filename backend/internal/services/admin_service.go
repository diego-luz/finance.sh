package services

import (
	"errors"
	"log/slog"
	"math"
	"strings"

	"github.com/finance-sh/finance-sh/internal/dto"
	"github.com/finance-sh/finance-sh/internal/entities"
	"github.com/finance-sh/finance-sh/internal/repositories"
	"github.com/finance-sh/finance-sh/pkg/hash"
	"github.com/google/uuid"
)

// Typed errors specific to the platform back-office. User-facing messages are in
// pt-BR; handlers map them to HTTP status codes.
var (
	// ErrCannotDisableSelf protects a super-admin from locking themselves out.
	ErrCannotDisableSelf = errors.New("você não pode desativar a própria conta")
	// ErrLastSuperAdmin protects the platform from losing its last super-admin.
	ErrLastSuperAdmin = errors.New("não é possível desativar o último super-administrador")
)

// AdminService backs the platform back-office (super-admin) — read-only views of
// the instance plus a couple of user-safety operations (disable, reset password).
// It is NOT tenant-scoped: every operation here is platform-wide.
type AdminService struct {
	admin *repositories.AdminRepository
	users *repositories.UserRepository
}

func NewAdminService(
	admin *repositories.AdminRepository,
	users *repositories.UserRepository,
) *AdminService {
	return &AdminService{admin: admin, users: users}
}

// Stats returns the global platform counters.
func (s *AdminService) Stats() (dto.AdminStats, error) {
	c, err := s.admin.Stats()
	if err != nil {
		return dto.AdminStats{}, err
	}
	return dto.AdminStats{
		Organizations: c.Organizations,
		Users:         c.Users,
		Transactions:  c.Transactions,
	}, nil
}

func clampPage(page, perPage int) (int, int) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	return page, perPage
}

// ListOrganizations returns a paginated, optionally-searched view of every
// organization with its owner and member/transaction counts. Counts and owners
// are batch-resolved (no N+1).
func (s *AdminService) ListOrganizations(search string, page, perPage int) ([]dto.AdminOrgDTO, dto.PageMeta, error) {
	page, perPage = clampPage(page, perPage)
	orgs, total, err := s.admin.ListOrganizations(strings.TrimSpace(search), page, perPage)
	if err != nil {
		return nil, dto.PageMeta{}, err
	}

	ids := make([]uuid.UUID, 0, len(orgs))
	ownerIDs := make([]uuid.UUID, 0, len(orgs))
	for i := range orgs {
		ids = append(ids, orgs[i].ID)
		ownerIDs = append(ownerIDs, orgs[i].OwnerID)
	}

	owners, err := s.admin.OwnersByIDs(ownerIDs)
	if err != nil {
		return nil, dto.PageMeta{}, err
	}
	memberCounts, err := s.admin.MemberCountsByOrg(ids)
	if err != nil {
		return nil, dto.PageMeta{}, err
	}
	txCounts, err := s.admin.TransactionCountsByOrg(ids)
	if err != nil {
		return nil, dto.PageMeta{}, err
	}

	out := make([]dto.AdminOrgDTO, 0, len(orgs))
	for i := range orgs {
		out = append(out, adminOrgDTO(&orgs[i], owners, memberCounts, txCounts))
	}

	pages := int(math.Ceil(float64(total) / float64(perPage)))
	meta := dto.PageMeta{Page: page, PerPage: perPage, Total: total, Pages: pages}
	return out, meta, nil
}

func adminOrgDTO(
	o *entities.Organization,
	owners map[uuid.UUID]entities.User,
	memberCounts, txCounts map[uuid.UUID]int64,
) dto.AdminOrgDTO {
	d := dto.AdminOrgDTO{
		ID:           o.ID.String(),
		Name:         o.Name,
		Slug:         o.Slug,
		Currency:     o.Currency,
		Members:      memberCounts[o.ID],
		Transactions: txCounts[o.ID],
		CreatedAt:    o.CreatedAt,
	}
	if owner, ok := owners[o.OwnerID]; ok {
		d.Owner = &dto.AdminOrgOwnerDTO{
			ID:    owner.ID.String(),
			Name:  owner.Name,
			Email: owner.Email,
		}
	}
	return d
}

// ListUsers returns a paginated, optionally-searched view of every user with
// their org names (batch-resolved, no N+1).
func (s *AdminService) ListUsers(search string, page, perPage int) ([]dto.AdminUserDTO, dto.PageMeta, error) {
	page, perPage = clampPage(page, perPage)
	users, total, err := s.admin.ListUsers(strings.TrimSpace(search), page, perPage)
	if err != nil {
		return nil, dto.PageMeta{}, err
	}

	ids := make([]uuid.UUID, 0, len(users))
	for i := range users {
		ids = append(ids, users[i].ID)
	}
	orgNames, err := s.admin.OrgNamesByUser(ids)
	if err != nil {
		return nil, dto.PageMeta{}, err
	}

	out := make([]dto.AdminUserDTO, 0, len(users))
	for i := range users {
		u := &users[i]
		names := orgNames[u.ID]
		if names == nil {
			names = []string{}
		}
		out = append(out, dto.AdminUserDTO{
			ID:            u.ID.String(),
			Name:          u.Name,
			Email:         u.Email,
			SuperAdmin:    u.SuperAdmin,
			Disabled:      u.Disabled,
			EmailVerified: u.EmailVerified,
			CreatedAt:     u.CreatedAt,
			Organizations: names,
		})
	}

	pages := int(math.Ceil(float64(total) / float64(perPage)))
	meta := dto.PageMeta{Page: page, PerPage: perPage, Total: total, Pages: pages}
	return out, meta, nil
}

// SetUserDisabled toggles a user's disabled flag. callerID is the acting
// super-admin: a super-admin cannot disable themselves, and the platform may
// never lose its last (enabled) super-admin. On disable, the user's refresh
// tokens are revoked so existing sessions die immediately.
func (s *AdminService) SetUserDisabled(callerID, userID uuid.UUID, disabled bool) error {
	user, err := s.users.FindByID(userID)
	if err != nil {
		return err
	}

	if disabled {
		if userID == callerID {
			return ErrCannotDisableSelf
		}
		// Guard: do not disable the last enabled super-admin.
		if user.SuperAdmin && !user.Disabled {
			n, err := s.users.CountSuperAdmins()
			if err != nil {
				return err
			}
			if n <= 1 {
				return ErrLastSuperAdmin
			}
		}
	}

	if err := s.users.SetDisabled(userID, disabled); err != nil {
		return err
	}
	if disabled {
		// Kill active sessions immediately. Best-effort: a revoke failure must not
		// undo the disable (the user can no longer log in regardless).
		if err := s.users.RevokeAllRefreshTokens(userID); err != nil {
			slog.Error("admin disable: failed to revoke refresh tokens", "user_id", userID, "error", err)
		}
	}
	slog.Info("admin set user disabled", "user_id", userID, "disabled", disabled, "by", callerID)
	return nil
}

// ResetUserPassword lets a super-admin set a new password for any user. The
// user is forced to change it again on their next login (must_change_password)
// and all their active sessions are revoked so they must re-authenticate.
func (s *AdminService) ResetUserPassword(callerID, userID uuid.UUID, newPassword string) error {
	user, err := s.users.FindByID(userID)
	if err != nil {
		return err
	}
	if user.Disabled {
		return ErrAccountDisabled
	}
	pwHash, err := hash.Password(newPassword)
	if err != nil {
		return err
	}
	if err := s.users.AdminResetPassword(userID, pwHash); err != nil {
		return err
	}
	// Best-effort: kill existing sessions so the user must log in with the new
	// password (and then hit the forced-change flow).
	if err := s.users.RevokeAllRefreshTokens(userID); err != nil {
		slog.Error("admin reset password: failed to revoke refresh tokens", "user_id", userID, "error", err)
	}
	slog.Info("admin reset user password", "user_id", userID, "by", callerID)
	return nil
}
