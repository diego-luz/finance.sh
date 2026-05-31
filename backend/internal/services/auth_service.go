package services

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"regexp"
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
	"github.com/finance-sh/finance-sh/pkg/lockout"
	"github.com/finance-sh/finance-sh/pkg/logger"
	"github.com/finance-sh/finance-sh/pkg/mailer"
	"github.com/finance-sh/finance-sh/pkg/totp"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// mfaTokenTTL is how long an mfa challenge token stays valid.
const mfaTokenTTL = 5 * time.Minute

// recoveryCodeCount is the number of one-time backup codes generated on enable.
const recoveryCodeCount = 8

// Typed errors the handler layer maps to HTTP status codes. User-facing
// messages are in pt-BR.
var (
	ErrEmailTaken         = errors.New("e-mail já cadastrado")
	ErrInvalidCredentials = errors.New("credenciais inválidas")
	ErrInvalidToken       = errors.New("token inválido ou expirado")
	ErrUserNotFound       = errors.New("usuário não encontrado")
	ErrAccountLocked      = errors.New("conta temporariamente bloqueada por excesso de tentativas")
	ErrInvalidCode        = errors.New("código de verificação inválido")
	Err2FANotPending      = errors.New("configure o 2FA antes de ativá-lo")
	// ErrAccountDisabled is returned at login when a platform super-admin has
	// disabled the user's account. Maps to 403.
	ErrAccountDisabled = errors.New("Conta desativada. Contate o administrador.")
	// ErrRegistrationClosed is returned by Register when self-service signup is
	// turned off (REGISTRATION_OPEN=false) and the request is not an invitation
	// accept. Maps to 403.
	ErrRegistrationClosed = errors.New("Cadastro fechado. Contate o administrador.")
)

// AuthMeta carries request context used for refresh-token auditing.
type AuthMeta struct {
	UserAgent string
	IP        string
}

type AuthService struct {
	users   *repositories.UserRepository
	resets  *repositories.PasswordResetRepository
	cfg     *config.Config
	lockout *lockout.Limiter
	mailer  *mailer.Mailer
	db      *gorm.DB
}

func NewAuthService(
	users *repositories.UserRepository,
	resets *repositories.PasswordResetRepository,
	cfg *config.Config,
	lim *lockout.Limiter,
	mail *mailer.Mailer,
	db *gorm.DB,
) *AuthService {
	return &AuthService{users: users, resets: resets, cfg: cfg, lockout: lim, mailer: mail, db: db}
}

var slugInvalid = regexp.MustCompile(`[^a-z0-9]+`)

// slugify turns an organization name into a URL-safe slug and appends a short
// random suffix to keep the global slug unique without a round-trip check.
func slugify(name string) string {
	s := strings.ToLower(strings.TrimSpace(name))
	s = slugInvalid.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if s == "" {
		s = "org"
	}
	suffix, _ := hash.RandomToken(3) // 6 hex chars
	return fmt.Sprintf("%s-%s", s, suffix)
}

// Register bootstraps a new tenant: a user, their organization, an owner
// membership and a Community-edition subscription (informational, no billing,
// no quota enforcement), records LGPD consent, sends a (soft) email-verification
// link and issues tokens.
func (s *AuthService) Register(req dto.RegisterRequest, meta AuthMeta) (*dto.AuthResponse, error) {
	// Closed-registration gate: public signup creates a brand-new tenant, so it is
	// never an invitation-accept flow (that path lives in MemberService). Block it
	// when self-service registration is turned off. Super-admin provisioning uses
	// the admin service directly and bypasses this entirely.
	if !s.cfg.RegistrationOpen {
		return nil, ErrRegistrationClosed
	}

	email := strings.ToLower(strings.TrimSpace(req.Email))
	if s.users.EmailExists(email) {
		return nil, ErrEmailTaken
	}

	pwHash, err := hash.Password(req.Password)
	if err != nil {
		return nil, err
	}

	// Currency is optional: default to BRL; when provided it must be supported.
	currency := money.DefaultCode
	if req.Currency != "" {
		if !money.IsSupported(req.Currency) {
			return nil, ErrUnsupportedCurrency
		}
		currency = money.Normalize(req.Currency)
	}

	now := time.Now().UTC()
	user := &entities.User{
		Name:            strings.TrimSpace(req.Name),
		Email:           email,
		PasswordHash:    pwHash,
		TermsAcceptedAt: &now,
		TermsVersion:    s.cfg.TermsVersion,
	}
	org := &entities.Organization{
		Name:     strings.TrimSpace(req.OrganizationName),
		Slug:     slugify(req.OrganizationName),
		Currency: currency,
	}

	membership, err := s.users.CreateWithOrganization(user, org)
	if err != nil {
		return nil, err
	}

	// Onboarding: seed default categories + accounts for the new tenant so the
	// freshly-registered user doesn't land in an empty app. Failure here must not
	// block registration — the user can still create their own data. (No
	// transactions are seeded for real tenants.)
	if s.db != nil {
		if err := database.SeedDefaults(s.db, org.ID); err != nil {
			slog.Error("onboarding: default seed failed", "org_id", org.ID, "error", err)
		}
	}

	slog.Info("user registered", "user_id", user.ID, "email", logger.MaskEmail(user.Email))

	// Soft email verification: create a token and "send" the link. Failure here
	// must not block registration (login still works while unverified).
	s.sendVerificationEmail(user)

	return s.issueTokens(user, org, membership.Role, meta)
}

// sendVerificationEmail creates a 24h verification token and dispatches the link.
// Errors are logged but swallowed so they never block the caller.
func (s *AuthService) sendVerificationEmail(user *entities.User) {
	raw, err := hash.RandomToken(24)
	if err != nil {
		slog.Error("verify-email: token generation failed", "user_id", user.ID, "error", err)
		return
	}
	ev := &entities.EmailVerification{
		UserID:    user.ID,
		TokenHash: hash.SHA256(raw),
		ExpiresAt: time.Now().UTC().Add(24 * time.Hour),
	}
	if err := s.users.CreateEmailVerification(ev); err != nil {
		slog.Error("verify-email: persistence failed", "user_id", user.ID, "error", err)
		return
	}
	link := fmt.Sprintf("%s/verify-email?token=%s", strings.TrimRight(s.cfg.FrontendURL, "/"), raw)
	body := fmt.Sprintf("Olá %s,\n\nConfirme seu e-mail no finance.sh acessando o link abaixo:\n%s\n\nO link expira em 24 horas.", user.Name, link)
	_ = s.mailer.Send(user.Email, "Confirme seu e-mail — finance.sh", body)
}

// Login verifies credentials and issues tokens for the user's first org.
//
// Brute-force protection: after cfg.Login.MaxAttempts consecutive failures for
// an email the account is temporarily locked (ErrAccountLocked) for
// cfg.Login.LockoutMin minutes. A successful login resets the counter.
//
// When the user has 2FA enabled, no tokens are issued: a short-lived mfa token
// is returned (LoginResult.MFARequired) to be exchanged at /auth/2fa/verify.
func (s *AuthService) Login(req dto.LoginRequest, meta AuthMeta) (*dto.LoginResult, error) {
	ctx := context.Background()
	email := strings.ToLower(strings.TrimSpace(req.Email))

	if s.lockout.Locked(ctx, email) {
		return nil, ErrAccountLocked
	}

	user, err := s.users.FindByEmail(email)
	if err != nil {
		// Count the failure even for unknown emails to avoid user enumeration.
		s.lockout.RegisterFailure(ctx, email)
		slog.Info("login failed", "email", logger.MaskEmail(email), "reason", "unknown_user")
		return nil, ErrInvalidCredentials
	}
	if !hash.Check(user.PasswordHash, req.Password) {
		locked := s.lockout.RegisterFailure(ctx, email)
		slog.Info("login failed", "email", logger.MaskEmail(email), "reason", "bad_password", "locked", locked)
		if locked {
			return nil, ErrAccountLocked
		}
		return nil, ErrInvalidCredentials
	}

	// Platform-level access gate: a disabled account cannot log in regardless of
	// valid credentials. Reset the failure counter (the password was correct) so
	// the account is not also lockout-throttled on top of being disabled.
	if user.Disabled {
		s.lockout.Reset(ctx, email)
		slog.Info("login blocked", "email", logger.MaskEmail(email), "reason", "disabled")
		return nil, ErrAccountDisabled
	}

	// 2FA gate: issue a challenge token instead of the full session. The lockout
	// counter is NOT reset here — the login is not yet complete. It is reset only
	// after the second factor verifies (see VerifyTwoFactorLogin).
	if user.TwoFactorEnabled {
		mfaToken, err := jwt.GenerateWithPurpose(user.ID.String(), user.Email, "mfa", s.cfg.JWT.AccessSecret, mfaTokenTTL)
		if err != nil {
			return nil, err
		}
		return &dto.LoginResult{MFARequired: true, MFAToken: mfaToken}, nil
	}

	org, role, err := s.primaryOrg(user.ID)
	if err != nil {
		return nil, err
	}
	auth, err := s.issueTokens(user, org, role, meta)
	if err != nil {
		return nil, err
	}

	// Full login succeeded (no 2FA): reset the failure counter.
	s.lockout.Reset(ctx, email)
	return &dto.LoginResult{Auth: auth}, nil
}

// RegistrationOpen reports whether public self-service signup is currently
// allowed (exposed by GET /auth/registration-open so the frontend can hide the
// signup UI).
func (s *AuthService) RegistrationOpen() bool {
	return s.cfg.RegistrationOpen
}

// LockoutMinutes exposes the configured lockout window (for the handler's
// user-facing message).
func (s *AuthService) LockoutMinutes() int {
	if s.lockout == nil {
		return s.cfg.Login.LockoutMin
	}
	return s.lockout.LockoutMinutes()
}

// Refresh validates an opaque refresh token, rotates it (revoke + reissue) and
// returns a fresh pair of tokens.
func (s *AuthService) Refresh(rawToken string, meta AuthMeta) (*dto.AuthResponse, error) {
	tokenHash := hash.SHA256(rawToken)
	stored, err := s.users.FindRefreshToken(tokenHash)
	if err != nil {
		return nil, ErrInvalidToken
	}
	if stored.Revoked || time.Now().After(stored.ExpiresAt) {
		return nil, ErrInvalidToken
	}

	user, err := s.users.FindByID(stored.UserID)
	if err != nil {
		return nil, ErrInvalidToken
	}

	// Rotate: invalidate the presented token before issuing a new one.
	if err := s.users.RevokeRefreshToken(stored.ID); err != nil {
		return nil, err
	}

	org, role, err := s.primaryOrg(user.ID)
	if err != nil {
		return nil, err
	}
	return s.issueTokens(user, org, role, meta)
}

// MailEnabled reports whether SMTP is configured (real e-mail delivery). When
// false, password-reset links are written to the server log instead. Used by the
// forgot-password response so the UI shows the right message.
func (s *AuthService) MailEnabled() bool {
	return strings.TrimSpace(s.cfg.SMTP.Host) != ""
}

// ForgotPassword creates a single-use reset token for the user with the given
// email. To avoid leaking which emails exist, it returns no error when the user
// is absent. When a token is created it is logged (info) and — in development —
// also returned so it can be tested without email delivery. The returned string
// is empty when nothing was created or in production.
func (s *AuthService) ForgotPassword(email string) (string, error) {
	user, err := s.users.FindByEmail(strings.ToLower(strings.TrimSpace(email)))
	if err != nil {
		return "", nil // silent: do not reveal absence
	}

	raw, err := hash.RandomToken(24)
	if err != nil {
		return "", err
	}
	pr := &entities.PasswordReset{
		UserID:    user.ID,
		TokenHash: hash.SHA256(raw),
		ExpiresAt: time.Now().UTC().Add(time.Hour),
	}
	if err := s.resets.Create(pr); err != nil {
		return "", err
	}

	// Dispatch the reset link via the mailer (logs only when SMTP is disabled).
	// PII is masked in the structured log; the raw token is only echoed back in
	// development (see handler) — never logged at info.
	link := fmt.Sprintf("%s/reset-password?token=%s", strings.TrimRight(s.cfg.FrontendURL, "/"), raw)
	body := fmt.Sprintf("Olá %s,\n\nRecebemos um pedido para redefinir sua senha no finance.sh. Acesse:\n%s\n\nO link expira em 1 hora. Se não foi você, ignore este e-mail.", user.Name, link)
	_ = s.mailer.Send(user.Email, "Redefinição de senha — finance.sh", body)
	slog.Info("password reset requested", "user_id", user.ID, "email", logger.MaskEmail(user.Email))

	if s.cfg.IsProduction() {
		return "", nil
	}
	return raw, nil
}

// ChangePassword verifies the user's current password and rotates it to a new
// one. On success: clears must_change_password and revokes every active refresh
// token of the user EXCEPT the one whose raw token is supplied in
// keepRawToken (when empty all are revoked and the client must re-login). The
// three writes run in a single DB transaction so a partial failure can never
// leave the user with the new password but stale sessions still authoritative.
func (s *AuthService) ChangePassword(userID uuid.UUID, currentPassword, newPassword, keepRawToken string) error {
	user, err := s.users.FindByID(userID)
	if err != nil {
		return ErrUserNotFound
	}
	if !hash.Check(user.PasswordHash, currentPassword) {
		return ErrWrongPassword
	}

	newHash, err := hash.Password(newPassword)
	if err != nil {
		return err
	}

	keepHash := ""
	if keepRawToken != "" {
		keepHash = hash.SHA256(strings.TrimSpace(keepRawToken))
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		// Update password + clear the must-change flag in one row write.
		if err := tx.Model(&entities.User{}).
			Where("id = ?", userID).
			Updates(map[string]interface{}{
				"password_hash":        newHash,
				"must_change_password": false,
			}).Error; err != nil {
			return err
		}
		// Revoke every active refresh token of the user, optionally keeping the
		// caller's current session alive.
		q := tx.Model(&entities.RefreshToken{}).
			Where("user_id = ? AND revoked = false", userID)
		if keepHash != "" {
			q = q.Where("token_hash <> ?", keepHash)
		}
		return q.Update("revoked", true).Error
	})
	if err != nil {
		return err
	}
	slog.Info("password changed", "user_id", userID, "email", logger.MaskEmail(user.Email))
	return nil
}

// ResetPassword validates a reset token, sets the new password, marks the token
// used and revokes all of the user's refresh tokens (forces re-login). All four
// steps run in a single DB transaction so a partial failure can never leave the
// token reusable.
func (s *AuthService) ResetPassword(rawToken, newPassword string) error {
	pwHash, err := hash.Password(newPassword)
	if err != nil {
		return err
	}
	tokenHash := hash.SHA256(strings.TrimSpace(rawToken))
	if err := s.resets.ConsumeAndResetPassword(tokenHash, pwHash, time.Now().UTC()); err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			return ErrInvalidToken
		}
		return err
	}
	return nil
}

// Logout revokes the presented refresh token. Unknown tokens are a no-op so the
// client can always treat logout as successful.
func (s *AuthService) Logout(rawToken string) error {
	stored, err := s.users.FindRefreshToken(hash.SHA256(rawToken))
	if err != nil {
		return nil
	}
	return s.users.RevokeRefreshToken(stored.ID)
}

// ----- Session management (refresh tokens) -----

// Sessions lists the user's active (non-revoked, non-expired) refresh tokens as
// session metadata. The token hash is never returned.
func (s *AuthService) Sessions(userID uuid.UUID) ([]dto.SessionDTO, error) {
	tokens, err := s.users.ActiveRefreshTokens(userID)
	if err != nil {
		return nil, err
	}
	out := make([]dto.SessionDTO, 0, len(tokens))
	for i := range tokens {
		t := &tokens[i]
		out = append(out, dto.SessionDTO{
			ID:        t.ID.String(),
			UserAgent: t.UserAgent,
			IP:        t.IP,
			CreatedAt: t.CreatedAt,
			ExpiresAt: t.ExpiresAt,
		})
	}
	return out, nil
}

// RevokeSession revokes a single session (refresh token) that must belong to the
// user. Returns ErrNotFound when the id is unknown or not the user's.
func (s *AuthService) RevokeSession(userID, sessionID uuid.UUID) error {
	return s.users.RevokeRefreshTokenForUser(userID, sessionID)
}

// RevokeOtherSessions revokes all of the user's active sessions except, when a
// raw refresh token is supplied, the one matching its hash. When the token is
// empty every session is revoked (the client must re-login). Returns the count
// revoked.
func (s *AuthService) RevokeOtherSessions(userID uuid.UUID, keepRawToken string) (int64, error) {
	keepHash := ""
	if keepRawToken != "" {
		keepHash = hash.SHA256(strings.TrimSpace(keepRawToken))
	}
	return s.users.RevokeOtherRefreshTokens(userID, keepHash)
}

// Me returns the authenticated user together with all their organizations.
func (s *AuthService) Me(userID uuid.UUID) (*dto.MeResponse, error) {
	user, err := s.users.FindByID(userID)
	if err != nil {
		return nil, ErrUserNotFound
	}
	memberships, err := s.users.Memberships(userID)
	if err != nil {
		return nil, err
	}

	orgs := make([]dto.OrgDTO, 0, len(memberships))
	for _, m := range memberships {
		if m.Organization == nil {
			continue
		}
		orgs = append(orgs, orgDTO(m.Organization, m.Role))
	}

	return &dto.MeResponse{
		User:          userDTO(user),
		Organizations: orgs,
	}, nil
}

// ----- Email verification (soft) -----

// VerifyEmail validates a verification token, marking the user verified and the
// token used. Login is never blocked when unverified; this only flips the flag.
func (s *AuthService) VerifyEmail(rawToken string) error {
	ev, err := s.users.FindEmailVerification(hash.SHA256(strings.TrimSpace(rawToken)))
	if err != nil {
		return ErrInvalidToken
	}
	if ev.Used || time.Now().UTC().After(ev.ExpiresAt) {
		return ErrInvalidToken
	}
	if err := s.users.MarkEmailVerified(ev.UserID); err != nil {
		return err
	}
	return s.users.MarkEmailVerificationUsed(ev.ID)
}

// ResendVerification re-sends a verification link for an unverified user. It is
// silent about whether the email exists / is already verified (generic 200).
func (s *AuthService) ResendVerification(email string) {
	user, err := s.users.FindByEmail(strings.ToLower(strings.TrimSpace(email)))
	if err != nil {
		return
	}
	if user.EmailVerified {
		return
	}
	s.sendVerificationEmail(user)
}

// ----- Two-factor authentication -----

// SetupTwoFactor generates a fresh TOTP secret (issuer "finance.sh", account=email),
// stores it as pending (2FA not yet enabled) and returns the secret + otpauth URL.
func (s *AuthService) SetupTwoFactor(userID uuid.UUID) (*dto.TwoFactorSetupResponse, error) {
	user, err := s.users.FindByID(userID)
	if err != nil {
		return nil, ErrUserNotFound
	}
	secret, url, err := totp.Generate("finance.sh", user.Email)
	if err != nil {
		return nil, err
	}
	if err := s.users.SetTwoFactorSecret(userID, secret); err != nil {
		return nil, err
	}
	// Ensure it stays disabled until the user confirms a valid code.
	if err := s.users.SetTwoFactorEnabled(userID, false); err != nil {
		return nil, err
	}
	return &dto.TwoFactorSetupResponse{Secret: secret, OTPAuthURL: url}, nil
}

// EnableTwoFactor validates the supplied TOTP against the pending secret, enables
// 2FA and returns freshly generated single-use recovery codes (shown once).
func (s *AuthService) EnableTwoFactor(userID uuid.UUID, code string) (*dto.TwoFactorEnableResponse, error) {
	user, err := s.users.FindByID(userID)
	if err != nil {
		return nil, ErrUserNotFound
	}
	secret := user.TwoFactorSecret.String()
	if secret == "" {
		return nil, Err2FANotPending
	}
	if !totp.Validate(strings.TrimSpace(code), secret) {
		return nil, ErrInvalidCode
	}

	codes, err := totp.RecoveryCodes(recoveryCodeCount)
	if err != nil {
		return nil, err
	}
	hashes := make([]string, len(codes))
	for i, c := range codes {
		hashes[i] = hash.SHA256(c)
	}
	if err := s.users.ReplaceRecoveryCodes(userID, hashes); err != nil {
		return nil, err
	}
	if err := s.users.SetTwoFactorEnabled(userID, true); err != nil {
		return nil, err
	}
	slog.Info("2fa enabled", "user_id", userID)
	return &dto.TwoFactorEnableResponse{RecoveryCodes: codes}, nil
}

// DisableTwoFactor validates a current TOTP (or a recovery code) and turns 2FA
// off, clearing the secret and recovery codes.
func (s *AuthService) DisableTwoFactor(userID uuid.UUID, code string) error {
	user, err := s.users.FindByID(userID)
	if err != nil {
		return ErrUserNotFound
	}
	if !user.TwoFactorEnabled {
		return nil // already off; idempotent
	}
	if !s.checkSecondFactor(user, strings.TrimSpace(code)) {
		return ErrInvalidCode
	}
	if err := s.users.DisableTwoFactor(userID); err != nil {
		return err
	}
	slog.Info("2fa disabled", "user_id", userID)
	return nil
}

// VerifyTwoFactorLogin completes a 2FA login: it validates the short-lived mfa
// token, then a TOTP or recovery code, and finally issues the full session.
func (s *AuthService) VerifyTwoFactorLogin(req dto.TwoFactorVerifyRequest, meta AuthMeta) (*dto.AuthResponse, error) {
	ctx := context.Background()
	claims, err := jwt.Parse(req.MFAToken, s.cfg.JWT.AccessSecret)
	if err != nil || claims.Purpose != "mfa" {
		return nil, ErrInvalidToken
	}
	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return nil, ErrInvalidToken
	}
	user, err := s.users.FindByID(userID)
	if err != nil {
		return nil, ErrInvalidToken
	}
	if !user.TwoFactorEnabled {
		return nil, ErrInvalidToken
	}

	// Brute-force protection for the second factor: reuse the login lockout keyed
	// by the user's email so it shares the budget with the password stage.
	email := strings.ToLower(strings.TrimSpace(user.Email))
	if s.lockout.Locked(ctx, email) {
		return nil, ErrAccountLocked
	}
	if !s.checkSecondFactor(user, strings.TrimSpace(req.Code)) {
		if locked := s.lockout.RegisterFailure(ctx, email); locked {
			return nil, ErrAccountLocked
		}
		return nil, ErrInvalidCode
	}

	org, role, err := s.primaryOrg(user.ID)
	if err != nil {
		return nil, err
	}
	auth, err := s.issueTokens(user, org, role, meta)
	if err != nil {
		return nil, err
	}

	// Full login succeeded: reset the failure counter.
	s.lockout.Reset(ctx, email)
	return auth, nil
}

// checkSecondFactor accepts either a valid TOTP for the user's secret or an
// unused recovery code (which it then consumes).
func (s *AuthService) checkSecondFactor(user *entities.User, code string) bool {
	if code == "" {
		return false
	}
	if secret := user.TwoFactorSecret.String(); secret != "" && totp.Validate(code, secret) {
		return true
	}
	rc, err := s.users.FindRecoveryCode(user.ID, hash.SHA256(code))
	if err != nil {
		return false
	}
	_ = s.users.UseRecoveryCode(rc.ID)
	return true
}

// primaryOrg returns the user's first membership organization (used right after
// auth before the client selects a tenant via X-Organization-ID).
func (s *AuthService) primaryOrg(userID uuid.UUID) (*entities.Organization, entities.Role, error) {
	memberships, err := s.users.Memberships(userID)
	if err != nil {
		return nil, "", err
	}
	// No org is valid: a platform super-admin (or an invited user who hasn't
	// accepted yet) has no membership. Return nil org without erroring; the
	// client picks a tenant later via X-Organization-ID.
	if len(memberships) == 0 || memberships[0].Organization == nil {
		return nil, "", nil
	}
	return memberships[0].Organization, memberships[0].Role, nil
}

func (s *AuthService) issueTokens(user *entities.User, org *entities.Organization, role entities.Role, meta AuthMeta) (*dto.AuthResponse, error) {
	access, err := jwt.Generate(user.ID.String(), user.Email, s.cfg.JWT.AccessSecret, s.cfg.JWT.AccessTTL)
	if err != nil {
		return nil, err
	}

	raw, err := hash.RandomToken(32)
	if err != nil {
		return nil, err
	}
	rt := &entities.RefreshToken{
		UserID:    user.ID,
		TokenHash: hash.SHA256(raw),
		ExpiresAt: time.Now().Add(s.cfg.JWT.RefreshTTL),
		UserAgent: meta.UserAgent,
		IP:        meta.IP,
	}
	if err := s.users.SaveRefreshToken(rt); err != nil {
		return nil, err
	}

	return &dto.AuthResponse{
		AccessToken:  access,
		RefreshToken: raw,
		ExpiresIn:    int64(s.cfg.JWT.AccessTTL.Seconds()),
		User:         userDTO(user),
		Organization: orgDTO(org, role),
	}, nil
}

func userDTO(u *entities.User) dto.UserDTO {
	return dto.UserDTO{
		ID:                 u.ID.String(),
		Name:               u.Name,
		Email:              u.Email,
		EmailVerified:      u.EmailVerified,
		TwoFactorEnabled:   u.TwoFactorEnabled,
		AvatarURL:          u.AvatarURL,
		SuperAdmin:         u.SuperAdmin,
		MustChangePassword: u.MustChangePassword,
	}
}

func orgDTO(o *entities.Organization, role entities.Role) dto.OrgDTO {
	// Org-less users (platform super-admin) get an empty org in the auth response.
	if o == nil {
		return dto.OrgDTO{}
	}
	return dto.OrgDTO{
		ID:       o.ID.String(),
		Name:     o.Name,
		Slug:     o.Slug,
		Currency: o.Currency,
		Role:     string(role),
	}
}
