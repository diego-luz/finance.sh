package services

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/finance-sh/finance-sh/internal/config"
	"github.com/finance-sh/finance-sh/internal/database"
	"github.com/finance-sh/finance-sh/internal/dto"
	"github.com/finance-sh/finance-sh/internal/entities"
	"github.com/finance-sh/finance-sh/internal/money"
	"github.com/finance-sh/finance-sh/internal/repositories"
	"github.com/finance-sh/finance-sh/pkg/hash"
	"github.com/finance-sh/finance-sh/pkg/jwt"
	"github.com/finance-sh/finance-sh/pkg/logger"
	"gorm.io/gorm"
)

// Typed errors for the first-run setup wizard. Handlers map them to HTTP codes.
var (
	// ErrAlreadyInitialized is returned by SetupService.Initialize when at least
	// one user already exists at the start of the transaction. The platform is
	// already bootstrapped and the wizard must not run again.
	ErrAlreadyInitialized = errors.New("plataforma já inicializada")
	// ErrWeakPassword mirrors the auth-layer minimum-length rule for setup. The
	// DTO already enforces it via struct tags; this error guards the service
	// boundary so other callers (tests, future entry points) cannot bypass it.
	ErrWeakPassword = errors.New("senha muito fraca")
)

// minSetupPasswordLen mirrors the validate:"min=8" tag on SetupUser.Password so
// programmatic callers (not the HTTP layer) still get the same guard.
const minSetupPasswordLen = 8

// SetupService backs the first-run wizard. It is invoked at most once per
// instance: the very first call creates the platform super-admin, their first
// organization and the owner membership inside a single DB transaction. Every
// subsequent call must short-circuit with ErrAlreadyInitialized.
//
// It is PLATFORM-level (no tenant) and PUBLIC (no auth) — the only access guard
// is the user-count == 0 invariant re-checked inside the transaction.
type SetupService struct {
	users *repositories.UserRepository
	cfg   *config.Config
	db    *gorm.DB
}

func NewSetupService(
	users *repositories.UserRepository,
	cfg *config.Config,
	db *gorm.DB,
) *SetupService {
	return &SetupService{users: users, cfg: cfg, db: db}
}

// NeedsSetup reports whether the platform still needs first-run initialization
// (i.e. no users exist yet). Cheap COUNT(*) on users; no cache by design.
func (s *SetupService) NeedsSetup(ctx context.Context) (bool, error) {
	n, err := s.users.CountAll()
	if err != nil {
		return false, fmt.Errorf("setup: count users: %w", err)
	}
	return n == 0, nil
}

// Initialize bootstraps the platform with the very first super-admin user and
// their first organization. Everything (count guard, user/org/membership/sub
// creation) runs in a single DB transaction so a concurrent caller cannot
// double-initialize. On success returns the same shape as the auth login
// response so the frontend can reuse its auth store.
func (s *SetupService) Initialize(req dto.SetupInitializeRequest, meta AuthMeta) (*dto.AuthResponse, error) {
	ctx := context.Background()

	// Defensive: the DTO already validates these via struct tags, but a
	// programmatic caller could bypass that. Re-check the bare minimum so the
	// service contract is self-contained.
	if len(strings.TrimSpace(req.User.Password)) < minSetupPasswordLen {
		return nil, ErrWeakPassword
	}
	if !money.IsSupported(req.Organization.Currency) {
		return nil, ErrUnsupportedCurrency
	}

	pwHash, err := hash.Password(req.User.Password)
	if err != nil {
		return nil, fmt.Errorf("setup: hash password: %w", err)
	}

	currency := money.Normalize(req.Organization.Currency)
	now := time.Now().UTC()
	email := strings.ToLower(strings.TrimSpace(req.User.Email))

	user := &entities.User{
		Name:            strings.TrimSpace(req.User.Name),
		Email:           email,
		PasswordHash:    pwHash,
		EmailVerified:   true, // first super-admin is self-vetted
		SuperAdmin:      true,
		TermsAcceptedAt: &now,
		TermsVersion:    s.cfg.TermsVersion,
	}
	org := &entities.Organization{
		Name:     strings.TrimSpace(req.Organization.Name),
		Slug:     slugify(req.Organization.Name),
		Currency: currency,
	}

	var membership entities.Membership
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// Race-condition guard: re-check inside the tx that no users exist.
		// Without this, two concurrent callers could both pass the public
		// NeedsSetup check and both succeed.
		var n int64
		if err := tx.Model(&entities.User{}).Count(&n).Error; err != nil {
			return fmt.Errorf("setup: count in tx: %w", err)
		}
		if n > 0 {
			return ErrAlreadyInitialized
		}

		if err := tx.Create(user).Error; err != nil {
			return fmt.Errorf("setup: create user: %w", err)
		}
		org.OwnerID = user.ID
		if err := tx.Create(org).Error; err != nil {
			return fmt.Errorf("setup: create org: %w", err)
		}
		membership = entities.Membership{
			UserID:         user.ID,
			OrganizationID: org.ID,
			Role:           entities.RoleOwner,
		}
		if err := tx.Create(&membership).Error; err != nil {
			return fmt.Errorf("setup: create membership: %w", err)
		}

		// Audit log entry. Best-effort within the tx so a partial failure here
		// still rolls back the whole bootstrap. Metadata is jsonb — must be
		// valid JSON (never empty string).
		audit := &entities.AuditLog{
			OrganizationID: &org.ID,
			UserID:         &user.ID,
			Action:         "platform.setup_completed",
			Entity:         "user",
			EntityID:       user.ID.String(),
			IP:             meta.IP,
			Metadata:       fmt.Sprintf(`{"organization_id":%q}`, org.ID.String()),
		}
		if err := tx.Create(audit).Error; err != nil {
			return fmt.Errorf("setup: create audit: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	// Onboarding: seed default categories/accounts for the new tenant. Best-
	// effort outside the tx so a seed failure cannot undo the bootstrap.
	if s.db != nil {
		if err := database.SeedDefaults(s.db, org.ID); err != nil {
			slog.ErrorContext(ctx, "setup: default seed failed", "org_id", org.ID, "error", err)
		}
	}

	slog.InfoContext(ctx, "setup initialized",
		"user_id", user.ID, "org_id", org.ID, "email", logger.MaskEmail(user.Email))

	return s.issueSetupTokens(user, org, membership.Role, meta)
}

// issueSetupTokens mirrors AuthService.issueTokens (access + refresh) so the
// returned envelope is the exact same shape the SPA already consumes after
// /auth/login. We do not call into AuthService directly to keep the wiring
// minimal (no circular ownership of the refresh-token write).
func (s *SetupService) issueSetupTokens(user *entities.User, org *entities.Organization, role entities.Role, meta AuthMeta) (*dto.AuthResponse, error) {
	access, err := jwt.Generate(user.ID.String(), user.Email, s.cfg.JWT.AccessSecret, s.cfg.JWT.AccessTTL)
	if err != nil {
		return nil, fmt.Errorf("setup: sign access token: %w", err)
	}
	raw, err := hash.RandomToken(32)
	if err != nil {
		return nil, fmt.Errorf("setup: random refresh: %w", err)
	}
	rt := &entities.RefreshToken{
		UserID:    user.ID,
		TokenHash: hash.SHA256(raw),
		ExpiresAt: time.Now().Add(s.cfg.JWT.RefreshTTL),
		UserAgent: meta.UserAgent,
		IP:        meta.IP,
	}
	if err := s.users.SaveRefreshToken(rt); err != nil {
		return nil, fmt.Errorf("setup: persist refresh: %w", err)
	}
	return &dto.AuthResponse{
		AccessToken:  access,
		RefreshToken: raw,
		ExpiresIn:    int64(s.cfg.JWT.AccessTTL.Seconds()),
		User:         userDTO(user),
		Organization: orgDTO(org, role),
	}, nil
}
