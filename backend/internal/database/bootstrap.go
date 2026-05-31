package database

import (
	"crypto/rand"
	"fmt"
	"strings"
	"time"

	"github.com/finance-sh/finance-sh/internal/entities"
	"github.com/finance-sh/finance-sh/pkg/hash"
	"gorm.io/gorm"
)

// BootstrapAdmin creates the first platform super-admin and their organization
// on a FRESH database (no users yet), so the operator can log in immediately on
// the very first boot — the env-driven alternative to the interactive setup
// wizard, the pattern self-hosters expect (Grafana/Nextcloud-style).
//
// It is idempotent and safe to call on every boot: if ANY user already exists it
// does nothing and returns created=false. The whole bootstrap (count guard +
// user/org/membership/subscription) runs in a single transaction so a concurrent
// boot cannot double-create.
//
// The created user has MustChangePassword=true, so the SPA forces a password
// rotation on first login. When password is empty a strong random password is
// generated and returned in genPassword so the caller can surface it once
// (logged to the boot console); when password is non-empty it is used verbatim
// and genPassword is "".
func BootstrapAdmin(db *gorm.DB, email, password, orgName, termsVersion string) (genPassword string, created bool, err error) {
	var count int64
	if err = db.Model(&entities.User{}).Count(&count).Error; err != nil {
		return "", false, fmt.Errorf("bootstrap: count users: %w", err)
	}
	if count > 0 {
		return "", false, nil // platform already has users — nothing to do.
	}

	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" {
		email = "admin@finance.sh"
	}
	if orgName = strings.TrimSpace(orgName); orgName == "" {
		orgName = "Minha Organização"
	}

	// Empty password → generate a strong random one and hand it back so the
	// caller logs it. No known default credential is ever persisted.
	if strings.TrimSpace(password) == "" {
		genPassword, err = generatePassword()
		if err != nil {
			return "", false, fmt.Errorf("bootstrap: generate password: %w", err)
		}
		password = genPassword
	}

	pwHash, err := hash.Password(password)
	if err != nil {
		return "", false, fmt.Errorf("bootstrap: hash password: %w", err)
	}

	now := time.Now().UTC()
	user := &entities.User{
		Name:          "Admin",
		Email:         email,
		PasswordHash:  pwHash,
		EmailVerified: true,
		SuperAdmin:    true,
		// Force rotation on first login. The SPA reads must_change_password from
		// the user DTO and routes to the change-password flow.
		MustChangePassword: true,
		TermsAcceptedAt:    &now,
		TermsVersion:       termsVersion,
	}
	org := &entities.Organization{Name: orgName, Slug: slugifyBootstrap(orgName), Currency: "BRL"}

	err = db.Transaction(func(tx *gorm.DB) error {
		// Re-check inside the tx so two concurrent boots can't both create.
		var n int64
		if err := tx.Model(&entities.User{}).Count(&n).Error; err != nil {
			return fmt.Errorf("bootstrap: count in tx: %w", err)
		}
		if n > 0 {
			return errAlreadyBootstrapped
		}
		if err := tx.Create(user).Error; err != nil {
			return fmt.Errorf("bootstrap: create user: %w", err)
		}
		org.OwnerID = user.ID
		if err := tx.Create(org).Error; err != nil {
			return fmt.Errorf("bootstrap: create org: %w", err)
		}
		if err := tx.Create(&entities.Membership{
			UserID: user.ID, OrganizationID: org.ID, Role: entities.RoleOwner,
		}).Error; err != nil {
			return fmt.Errorf("bootstrap: create membership: %w", err)
		}
		return nil
	})
	if err != nil {
		if err == errAlreadyBootstrapped {
			return "", false, nil // lost the race — another boot did it. Not an error.
		}
		return "", false, err
	}

	// Default categories/accounts for the new tenant. Best-effort outside the tx
	// so a seed failure cannot undo the bootstrap.
	if err := SeedDefaults(db, org.ID); err != nil {
		return genPassword, true, fmt.Errorf("bootstrap: seed defaults: %w", err)
	}
	return genPassword, true, nil
}

// errAlreadyBootstrapped is an internal sentinel: a concurrent boot won the race
// and created the first user. Callers treat it as "not created, no error".
var errAlreadyBootstrapped = fmt.Errorf("already bootstrapped")

// ResetUserPassword sets a fresh random password for the user with the given
// email, forces a change on next login, and revokes the user's refresh tokens
// (kills active sessions). Returns the plaintext password so the operator can
// hand it over. Backs the `-reset-password` CLI flag — the no-SMTP recovery
// path (mirrors Coolify/Miniflux/Nextcloud occ).
func ResetUserPassword(db *gorm.DB, email string) (string, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	var u entities.User
	if err := db.Where("email = ?", email).First(&u).Error; err != nil {
		return "", fmt.Errorf("usuário não encontrado: %s", email)
	}
	pw, err := generatePassword()
	if err != nil {
		return "", err
	}
	h, err := hash.Password(pw)
	if err != nil {
		return "", err
	}
	if err := db.Model(&entities.User{}).Where("id = ?", u.ID).
		Updates(map[string]any{"password_hash": h, "must_change_password": true}).Error; err != nil {
		return "", err
	}
	// Best-effort: revoke refresh tokens so existing sessions die.
	db.Model(&entities.RefreshToken{}).Where("user_id = ?", u.ID).Update("revoked", true)
	return pw, nil
}

// generatePassword builds a readable strong password: 4 groups of 4 chars from
// an unambiguous alphabet (no 0/O/1/l/I), joined by dashes — e.g. 7Kq9-mZ2x-Vp4w-Rt6n.
func generatePassword() (string, error) {
	const alphabet = "ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnpqrstuvwxyz23456789"
	const groups, perGroup = 4, 4
	buf := make([]byte, groups*perGroup)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	out := make([]byte, 0, groups*perGroup+groups-1)
	for i, b := range buf {
		if i > 0 && i%perGroup == 0 {
			out = append(out, '-')
		}
		out = append(out, alphabet[int(b)%len(alphabet)])
	}
	return string(out), nil
}

// slugifyBootstrap is a tiny slug helper for the bootstrap org name. It avoids
// importing the services-layer slugify (which adds a random suffix). Lowercases,
// keeps a-z0-9, collapses the rest to single dashes; falls back to "org".
func slugifyBootstrap(name string) string {
	var b strings.Builder
	lastDash := false
	for _, r := range strings.ToLower(name) {
		switch {
		case (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9'):
			b.WriteRune(r)
			lastDash = false
		default:
			if !lastDash {
				b.WriteByte('-')
				lastDash = true
			}
		}
	}
	s := strings.Trim(b.String(), "-")
	if s == "" {
		return "org"
	}
	return s
}
