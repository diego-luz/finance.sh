// Package middlewares holds HTTP middleware: authentication, multi-tenancy,
// RBAC, CORS and rate limiting. Context helpers expose the resolved identity to
// handlers in a type-safe way.
package middlewares

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/finance-sh/finance-sh/internal/config"
	"github.com/finance-sh/finance-sh/internal/entities"
	"github.com/finance-sh/finance-sh/internal/repositories"
	"github.com/finance-sh/finance-sh/pkg/jwt"
	"github.com/finance-sh/finance-sh/pkg/response"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ctxKey is an unexported type so context values cannot collide with keys set by
// other packages.
type ctxKey int

const (
	ctxUserID ctxKey = iota
	ctxOrgID
	ctxUserRole
)

// UserID returns the authenticated user ID, or uuid.Nil if absent.
func UserID(ctx context.Context) uuid.UUID {
	if v, ok := ctx.Value(ctxUserID).(uuid.UUID); ok {
		return v
	}
	return uuid.Nil
}

// OrgID returns the active tenant organization ID, or uuid.Nil if absent.
func OrgID(ctx context.Context) uuid.UUID {
	if v, ok := ctx.Value(ctxOrgID).(uuid.UUID); ok {
		return v
	}
	return uuid.Nil
}

// UserRole returns the caller's role in the active organization.
func UserRole(ctx context.Context) entities.Role {
	if v, ok := ctx.Value(ctxUserRole).(entities.Role); ok {
		return v
	}
	return ""
}

// Auth validates the Bearer access token and stores the user ID in context.
func Auth(cfg *config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if header == "" || !strings.HasPrefix(header, "Bearer ") {
				response.Error(w, http.StatusUnauthorized, "unauthorized", "Token de acesso ausente")
				return
			}
			raw := strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
			claims, err := jwt.Parse(raw, cfg.JWT.AccessSecret)
			if err != nil {
				response.Error(w, http.StatusUnauthorized, "unauthorized", "Token de acesso inválido")
				return
			}
			// Only plain access tokens (empty or "access" purpose) may authenticate.
			// Reject special-purpose tokens (e.g. the "mfa" 2FA challenge token,
			// signed with the same secret) so they cannot bypass the second factor.
			if claims.Purpose != "" && claims.Purpose != "access" {
				response.Error(w, http.StatusUnauthorized, "unauthorized", "Token de acesso inválido")
				return
			}
			userID, err := uuid.Parse(claims.UserID)
			if err != nil {
				response.Error(w, http.StatusUnauthorized, "unauthorized", "Token de acesso inválido")
				return
			}
			ctx := context.WithValue(r.Context(), ctxUserID, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// Tenant resolves the active organization. It prefers the X-Organization-ID
// header and verifies membership; when the header is absent it falls back to the
// user's first membership. Non-members get 403.
func Tenant(users *repositories.UserRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID := UserID(r.Context())
			if userID == uuid.Nil {
				response.Error(w, http.StatusUnauthorized, "unauthorized", "Não autenticado")
				return
			}

			var role entities.Role
			var orgID uuid.UUID

			if raw := r.Header.Get("X-Organization-ID"); raw != "" {
				id, err := uuid.Parse(raw)
				if err != nil {
					response.Error(w, http.StatusBadRequest, "bad_request", "X-Organization-ID inválido")
					return
				}
				m, err := users.Membership(userID, id)
				if err != nil {
					response.Error(w, http.StatusForbidden, "forbidden", "Acesso negado à organização")
					return
				}
				orgID, role = id, m.Role
			} else {
				memberships, err := users.Memberships(userID)
				if err != nil || len(memberships) == 0 {
					response.Error(w, http.StatusForbidden, "forbidden", "Nenhuma organização disponível")
					return
				}
				orgID, role = memberships[0].OrganizationID, memberships[0].Role
			}

			ctx := context.WithValue(r.Context(), ctxOrgID, orgID)
			ctx = context.WithValue(ctx, ctxUserRole, role)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireRole enforces that the caller holds one of the allowed roles (RBAC).
func RequireRole(roles ...entities.Role) func(http.Handler) http.Handler {
	allowed := make(map[entities.Role]struct{}, len(roles))
	for _, r := range roles {
		allowed[r] = struct{}{}
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if _, ok := allowed[UserRole(r.Context())]; !ok {
				response.Error(w, http.StatusForbidden, "forbidden", "Permissão insuficiente")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RequireSuperAdmin enforces the PLATFORM-level super-admin role for the
// back-office (/admin). It must run AFTER Auth (which sets the user id). It loads
// the user and 403s unless User.SuperAdmin is true. This is intentionally NOT
// tenant-scoped: the super-admin operates platform-wide.
func RequireSuperAdmin(users *repositories.UserRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID := UserID(r.Context())
			if userID == uuid.Nil {
				response.Error(w, http.StatusUnauthorized, "unauthorized", "Não autenticado")
				return
			}
			user, err := users.FindByID(userID)
			if err != nil || !user.SuperAdmin || user.Disabled {
				response.Error(w, http.StatusForbidden, "forbidden", "Acesso restrito ao administrador da plataforma")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RequireWrite blocks viewers (read-only role) from performing mutations. Safe
// methods (GET/HEAD/OPTIONS) are always allowed. Apply it to write-capable
// domains so viewers can read but never modify.
func RequireWrite(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet, http.MethodHead, http.MethodOptions:
			next.ServeHTTP(w, r)
			return
		}
		if UserRole(r.Context()) == entities.RoleViewer {
			response.Error(w, http.StatusForbidden, "forbidden", "Perfil somente leitura não pode realizar esta ação")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// statusRecorder captures the response status so the Audit middleware can decide
// whether a mutation succeeded.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (s *statusRecorder) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}

func (s *statusRecorder) Write(b []byte) (int, error) {
	if s.status == 0 {
		s.status = http.StatusOK
	}
	return s.ResponseWriter.Write(b)
}

// Audit records successful mutations (2xx on POST/PUT/PATCH/DELETE) as AuditLog
// rows. It is best-effort: a failed insert never fails the request. Reads and
// unsuccessful mutations are not logged.
func Audit(db *gorm.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rec := &statusRecorder{ResponseWriter: w}
			next.ServeHTTP(rec, r)

			switch r.Method {
			case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
			default:
				return
			}
			if rec.status < 200 || rec.status >= 300 {
				return
			}

			entity, entityID := auditTarget(r.URL.Path)
			orgID := OrgID(r.Context())
			userID := UserID(r.Context())
			// Metadata is a jsonb column — it MUST be valid JSON (never an empty
			// string, which Postgres rejects). Store the request method + path.
			meta, _ := json.Marshal(map[string]string{"method": r.Method, "path": r.URL.Path})
			row := &entities.AuditLog{
				Action:   r.Method,
				Entity:   entity,
				EntityID: entityID,
				IP:       r.RemoteAddr,
				Metadata: string(meta),
			}
			if orgID != uuid.Nil {
				row.OrganizationID = &orgID
			}
			if userID != uuid.Nil {
				row.UserID = &userID
			}
			// Fire-and-forget; do not block the response on the audit write.
			go func() { _ = db.Create(row).Error }()
		})
	}
}

// auditTarget extracts the resource name (first path segment after /api/v1) and
// a trailing UUID id if present. e.g. /api/v1/accounts/<uuid> -> ("accounts", uuid).
func auditTarget(path string) (entity, id string) {
	trimmed := strings.TrimPrefix(path, "/api/v1/")
	parts := strings.Split(strings.Trim(trimmed, "/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		return "", ""
	}
	entity = parts[0]
	if len(parts) > 1 {
		if _, err := uuid.Parse(parts[1]); err == nil {
			id = parts[1]
		}
	}
	return entity, id
}

// CORS configures cross-origin access for the SPA frontend.
func CORS(cfg *config.Config) func(http.Handler) http.Handler {
	return cors.Handler(cors.Options{
		AllowedOrigins:   cfg.CORSOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type", "X-Organization-ID"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	})
}

// securityCSP is the Content-Security-Policy for the SPA: scripts/connect
// restricted to same-origin, inline styles allowed for Vite + Google Fonts,
// no framing.
const securityCSP = "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; font-src 'self' https://fonts.gstatic.com; img-src 'self' data:; connect-src 'self'; frame-ancestors 'none'; base-uri 'self'; form-action 'self'"

// SecurityHeaders sets the baseline response security headers. The static
// headers are safe on every response (API JSON included); CSP is skipped for the
// Swagger UI, which loads its assets from a CDN and would otherwise be blocked by
// script-src 'self'. HSTS is emitted only when the request arrived over TLS
// (X-Forwarded-Proto=https set by the reverse proxy), since the app itself speaks
// plain HTTP.
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("X-Frame-Options", "DENY")
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		h.Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")

		if !strings.HasPrefix(r.URL.Path, "/swagger") {
			h.Set("Content-Security-Policy", securityCSP)
		}

		// Behind a TLS-terminating reverse proxy, lock the client to HTTPS.
		if r.Header.Get("X-Forwarded-Proto") == "https" {
			h.Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains")
		}

		next.ServeHTTP(w, r)
	})
}

// RateLimit throttles requests per client IP using a per-minute window.
func RateLimit(rpm int) func(http.Handler) http.Handler {
	return httprate.Limit(
		rpm,
		time.Minute,
		httprate.WithKeyByIP(),
		httprate.WithLimitHandler(func(w http.ResponseWriter, _ *http.Request) {
			response.Error(w, http.StatusTooManyRequests, "rate_limited", "Muitas requisições, tente novamente em instantes")
		}),
	)
}

// Logger emits a structured slog line per request with method, path, status and
// latency, tagged with the chi request ID.
func Logger(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			next.ServeHTTP(ww, r)
			log.Info("request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", ww.Status(),
				"bytes", ww.BytesWritten(),
				"duration_ms", time.Since(start).Milliseconds(),
				"request_id", middleware.GetReqID(r.Context()),
				"ip", r.RemoteAddr,
			)
		})
	}
}
