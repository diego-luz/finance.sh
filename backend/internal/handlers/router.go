package handlers

import (
	"log/slog"
	"net/http"
	"path/filepath"

	"github.com/finance-sh/finance-sh/internal/config"
	"github.com/finance-sh/finance-sh/internal/entities"
	"github.com/finance-sh/finance-sh/internal/middlewares"
	"github.com/finance-sh/finance-sh/internal/repositories"
	"github.com/finance-sh/finance-sh/internal/services"
	"github.com/finance-sh/finance-sh/internal/web"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"gorm.io/gorm"
)

// Deps bundles everything the router needs, injected from main.
type Deps struct {
	Config         *config.Config
	Logger         *slog.Logger
	DB             *gorm.DB
	Users          *repositories.UserRepository
	Auth           *services.AuthService
	Account        *services.AccountService
	Category       *services.CategoryService
	Categorization *services.CategorizationService
	Contact        *services.ContactService
	Tag            *services.TagService
	Tx             *services.TransactionService
	Search         *services.SearchService
	Dashboard      *services.DashboardService
	Forecast       *services.ForecastService
	CreditCard     *services.CreditCardService
	Goal           *services.GoalService
	Budget         *services.BudgetService
	Member         *services.MemberService
	Organization   *services.OrganizationService
	Notification   *services.NotificationService
	Report         *services.ReportService
	LGPD           *services.LGPDService
	Attachment     *services.AttachmentService
	Import         *services.ImportService
	Recurrence     *services.RecurrenceService
	Audit          *services.AuditService
	Admin          *services.AdminService
	Setup          *services.SetupService
	// DocsDir is the directory containing openapi.yaml (defaults to ./docs).
	DocsDir string
}

// NewRouter wires the global middleware chain and all routes.
func NewRouter(d Deps) *chi.Mux {
	r := chi.NewRouter()

	// Global middleware: request id, real client IP, panic recovery, structured
	// logging, CORS and per-IP rate limiting.
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middlewares.Logger(d.Logger))
	// Security headers (CSP, X-Frame-Options, nosniff, ...) applied to every
	// response, since the binary serves the SPA directly.
	r.Use(middlewares.SecurityHeaders)
	r.Use(middlewares.CORS(d.Config))
	r.Use(middlewares.RateLimit(d.Config.RateLimitRPM))

	health := NewHealthHandler()
	auth := NewAuthHandler(d.Auth)
	me := NewMeHandler(d.Auth, d.LGPD)
	accounts := NewAccountHandler(d.Account)
	categories := NewCategoryHandler(d.Category)
	categorization := NewCategorizationHandler(d.Categorization)
	contacts := NewContactHandler(d.Contact)
	tags := NewTagHandler(d.Tag)
	txs := NewTransactionHandler(d.Tx)
	search := NewSearchHandler(d.Search)
	dash := NewDashboardHandler(d.Dashboard)
	forecast := NewForecastHandler(d.Forecast)
	cards := NewCreditCardHandler(d.CreditCard)
	goals := NewGoalHandler(d.Goal)
	budgets := NewBudgetHandler(d.Budget)
	members := NewMemberHandler(d.Member)
	organizations := NewOrganizationHandler(d.Organization)
	notifications := NewNotificationHandler(d.Notification)
	reports := NewReportHandler(d.Report)
	attachments := NewAttachmentHandler(d.Attachment, d.Config.Storage.MaxAttachmentMB)
	statementImports := NewImportHandler(d.Import, d.Config.Storage.MaxAttachmentMB)
	recurrences := NewRecurrenceHandler(d.Recurrence)
	auditLogs := NewAuditHandler(d.Audit)
	admin := NewAdminHandler(d.Admin)
	setup := NewSetupHandler(d.Setup)

	docsDir := d.DocsDir
	if docsDir == "" {
		docsDir = "docs"
	}

	// Root-level liveness probe (used by the docker healthcheck).
	r.Get("/health", health.Health)

	r.Route("/api/v1", func(r chi.Router) {
		// Public
		r.Get("/health", health.Health)
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", auth.Register)
			r.Post("/login", auth.Login)
			r.Post("/refresh", auth.Refresh)
			r.Post("/logout", auth.Logout)
			r.Post("/forgot-password", auth.ForgotPassword)
			r.Post("/reset-password", auth.ResetPassword)
			// Email verification (soft) and 2FA login completion are public.
			r.Post("/verify-email", auth.VerifyEmail)
			r.Post("/verify-email/resend", auth.ResendVerification)
			r.Post("/2fa/verify", auth.VerifyTwoFactor)
			// Public flag so the SPA can hide the signup UI when self-service
			// registration is disabled.
			r.Get("/registration-open", auth.RegistrationOpen)
		})

		// First-run setup wizard. PUBLIC (no auth, no tenant): the SPA's bootstrap
		// gate calls /setup/status on every page load, and /setup/initialize is
		// the only way to bring up the very first super-admin. Both are guarded
		// server-side by an in-transaction users-count == 0 check.
		r.Route("/setup", func(r chi.Router) {
			r.Get("/status", setup.Status)
			r.Post("/initialize", setup.Initialize)
		})

		// Platform back-office (super-admin). PLATFORM-level, NOT tenant-scoped:
		// it deliberately does NOT use the Tenant middleware. Gated by Auth +
		// RequireSuperAdmin (loads the user and 403s unless User.SuperAdmin).
		r.Route("/admin", func(r chi.Router) {
			r.Use(middlewares.Auth(d.Config))
			r.Use(middlewares.RequireSuperAdmin(d.Users))

			r.Get("/stats", admin.Stats)

			// Read-only views of the instance.
			r.Get("/organizations", admin.ListOrganizations)

			r.Route("/users", func(r chi.Router) {
				r.Get("/", admin.ListUsers)
				r.Post("/{id}/disable", admin.DisableUser)
				r.Post("/{id}/enable", admin.EnableUser)
				r.Post("/{id}/reset-password", admin.ResetUserPassword)
			})
		})

		// Authenticated but tenant-agnostic: accepting an invitation resolves the
		// org from the invitation itself, so it must not require an active tenant
		// (the user may not be a member of any org yet). The self-service account
		// endpoints (2FA, LGPD export/erasure) are also user-scoped, not tenant-
		// scoped, so they live here too.
		r.Group(func(r chi.Router) {
			r.Use(middlewares.Auth(d.Config))
			r.Post("/invitations/accept", members.AcceptInvitation)

			// Two-factor management.
			r.Post("/me/2fa/setup", me.SetupTwoFactor)
			r.Post("/me/2fa/enable", me.EnableTwoFactor)
			r.Post("/me/2fa/disable", me.DisableTwoFactor)

			// LGPD data-subject rights.
			r.Get("/me/export", me.ExportData)
			r.Post("/me/import", me.ImportData)
			r.Delete("/me/account", me.DeleteAccount)

			// Self-service: create an additional organization owned by the caller
			// (tenant-agnostic — e.g. an org "Casa" + an org "Microempresa").
			r.Post("/organizations", organizations.Create)

			// Self-service password change (used by the must-change-password flow
			// after admin-provisioning, and for general password rotation).
			r.Post("/me/change-password", me.ChangePassword)

			// Session management (refresh tokens). Tenant-agnostic: scoped by the
			// authenticated user, not by an active org.
			r.Get("/me/sessions", me.Sessions)
			r.Delete("/me/sessions/{id}", me.RevokeSession)
			r.Post("/me/sessions/revoke-others", me.RevokeOtherSessions)
		})

		// Protected: requires a valid access token and an active tenant.
		r.Group(func(r chi.Router) {
			r.Use(middlewares.Auth(d.Config))
			r.Use(middlewares.Tenant(d.Users))
			r.Use(middlewares.Audit(d.DB))

			r.Get("/me", auth.Me)

			// Active organization config (multi-currency). Read for any member;
			// updating name/currency is owner/admin only. The supported-currency
			// table is a read for any member.
			r.Get("/organization", organizations.Get)
			r.With(middlewares.RequireRole(entities.RoleOwner, entities.RoleAdmin)).
				Put("/organization", organizations.Update)
			r.Get("/currencies", organizations.Currencies)

			// Write-capable financial domains: viewers may read but not mutate.
			r.Group(func(r chi.Router) {
				r.Use(middlewares.RequireWrite)

				r.Route("/accounts", func(r chi.Router) {
					r.Get("/", accounts.List)
					r.Post("/", accounts.Create)
					r.Get("/{id}", accounts.Get)
					r.Put("/{id}", accounts.Update)
					r.Delete("/{id}", accounts.Delete)
				})

				r.Route("/categories", func(r chi.Router) {
					r.Get("/", categories.List)
					r.Post("/", categories.Create)
					r.Get("/{id}", categories.Get)
					r.Put("/{id}", categories.Update)
					r.Delete("/{id}", categories.Delete)
				})

				r.Route("/contacts", func(r chi.Router) {
					r.Get("/", contacts.List)
					r.Post("/", contacts.Create)
					r.Get("/{id}", contacts.Get)
					r.Put("/{id}", contacts.Update)
					r.Delete("/{id}", contacts.Delete)
				})

				// Tags (rótulos). GET is allowed for viewers (RequireWrite lets
				// safe methods through); mutations require write capability.
				r.Route("/tags", func(r chi.Router) {
					r.Get("/", tags.List)
					r.Post("/", tags.Create)
					r.Put("/{id}", tags.Update)
					r.Delete("/{id}", tags.Delete)
				})

				// Automatic categorization — rule mutations and the bulk apply are
				// write operations. The list/suggest reads live in the all-members
				// group below.
				r.Route("/categorization/rules", func(r chi.Router) {
					r.Post("/", categorization.CreateRule)
					r.Put("/{id}", categorization.UpdateRule)
					r.Delete("/{id}", categorization.DeleteRule)
				})
				r.Post("/categorization/apply", categorization.Apply)

				r.Route("/transactions", func(r chi.Router) {
					r.Get("/", txs.List)
					r.Post("/", txs.Create)
					// Bulk actions (org-scoped). Registered before /{id} so the
					// literal paths are matched first.
					r.Post("/bulk-settle", txs.BulkSettle)
					r.Post("/bulk-unsettle", txs.BulkUnsettle)
					r.Post("/bulk-categorize", txs.BulkCategorize)
					r.Post("/bulk-delete", txs.BulkDelete)
					r.Get("/{id}", txs.Get)
					r.Put("/{id}", txs.Update)
					r.Delete("/{id}", txs.Delete)
					// Settlement (baixa) of payables/receivables.
					r.Post("/{id}/settle", txs.Settle)
					r.Post("/{id}/unsettle", txs.Unsettle)
					// Receipt attachments (comprovantes). Upload is a mutation (write);
					// listing is a safe GET, which RequireWrite lets viewers through.
					r.Post("/{id}/attachments", attachments.Upload)
					r.Get("/{id}/attachments", attachments.List)
				})

				r.Route("/credit-cards", func(r chi.Router) {
					r.Get("/", cards.List)
					r.Post("/", cards.Create)
					r.Get("/{id}", cards.Get)
					r.Put("/{id}", cards.Update)
					r.Delete("/{id}", cards.Delete)
					// Invoice settlement (pay/unpay) mutates transactions: write-only.
					r.Post("/{id}/invoices/{reference}/pay", cards.PayInvoice)
					r.Post("/{id}/invoices/{reference}/unpay", cards.UnpayInvoice)
				})

				r.Route("/goals", func(r chi.Router) {
					r.Get("/", goals.List)
					r.Post("/", goals.Create)
					r.Get("/{id}", goals.Get)
					r.Put("/{id}", goals.Update)
					r.Delete("/{id}", goals.Delete)
				})

				r.Route("/budgets", func(r chi.Router) {
					r.Get("/", budgets.List)
					r.Post("/", budgets.Create)
					r.Put("/{id}", budgets.Update)
					r.Delete("/{id}", budgets.Delete)
				})

				// Statement import (OFX/CSV): stateless preview then commit. Both
				// are mutations (preview parses an upload; commit persists rows).
				r.Route("/imports", func(r chi.Router) {
					r.Post("/preview", statementImports.Preview)
					r.Post("/commit", statementImports.Commit)
				})

				// Recurrence rules (recurring-transaction engine). GET is allowed for
				// viewers (RequireWrite lets safe methods through); mutations and the
				// manual run require write capability.
				r.Route("/recurrences", func(r chi.Router) {
					r.Get("/", recurrences.List)
					r.Post("/", recurrences.Create)
					r.Put("/{id}", recurrences.Update)
					r.Delete("/{id}", recurrences.Delete)
					r.Post("/{id}/run", recurrences.Run)
				})
			})

			r.Get("/dashboard", dash.Overview)

			// Global search (read-only; any member of the org).
			r.Get("/search", search.Search)

			// Automatic categorization reads (any member): list the rules and ask
			// for a suggestion for a given description/type.
			r.Get("/categorization/rules", categorization.ListRules)
			r.Get("/categorization/suggest", categorization.Suggest)

			// Attachments addressed by their own id (org-scoped). Download is a
			// read (any member); delete is a mutation (write-capable only).
			r.Get("/attachments/{id}/download", attachments.Download)
			r.With(middlewares.RequireWrite).Delete("/attachments/{id}", attachments.Delete)

			// Credit-card invoices/faturas (read-only; any member). The pay/unpay
			// actions live in the write-capable group above.
			r.Get("/credit-cards/{id}/invoices", cards.Invoices)
			r.Get("/credit-cards/{id}/invoices/{reference}", cards.Invoice)

			// Accounts payable / receivable (read-only; any member).
			r.Get("/payables", txs.Payables)
			r.Get("/receivables", txs.Receivables)

			// Cash-flow forecast (read-only; any member).
			r.Get("/cashflow/forecast", forecast.Forecast)

			// Reports & exports (read-only; any member).
			r.Get("/reports/transactions.csv", reports.TransactionsCSV)
			r.Get("/reports/summary", reports.Summary)
			r.Get("/reports/statement.pdf", reports.StatementPDF)
			r.Get("/reports/monthly.pdf", reports.MonthlyPDF)
			r.Get("/export/data.xlsx", reports.ExcelExport)

			// Notifications: any member of the org.
			r.Route("/notifications", func(r chi.Router) {
				r.Get("/", notifications.List)
				r.Post("/read-all", notifications.MarkAllRead)
				r.Post("/{id}/read", notifications.MarkRead)
			})

			// Members & invitations: owner/admin only.
			r.Group(func(r chi.Router) {
				r.Use(middlewares.RequireRole(entities.RoleOwner, entities.RoleAdmin))

				r.Route("/members", func(r chi.Router) {
					r.Get("/", members.ListMembers)
					r.Put("/{id}", members.UpdateMember)
					r.Delete("/{id}", members.RemoveMember)
				})

				r.Route("/invitations", func(r chi.Router) {
					r.Get("/", members.ListInvitations)
					r.Post("/", members.CreateInvitation)
					r.Delete("/{id}", members.RevokeInvitation)
				})

				// Audit trail (auditoria): read-only, sensitive — owner/admin only.
				r.Get("/audit-logs", auditLogs.List)
			})
		})
	})

	// API docs: serve the raw spec and a Swagger UI page. Gated so they can be
	// disabled in production (SWAGGER_ENABLED=false). The spec/UI are exposed when
	// SWAGGER_ENABLED is true OR the environment is non-production.
	if d.Config.SwaggerEnabled || !d.Config.IsProduction() {
		r.Get("/openapi.yaml", func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Content-Type", "application/yaml")
			http.ServeFile(w, req, filepath.Join(docsDir, "openapi.yaml"))
		})
		r.Get("/swagger", swaggerUI)
		r.Get("/swagger/", swaggerUI)
	}

	// SPA: serve the embedded frontend bundle for every non-API route. Mounted
	// LAST as the root catch-all so the explicit /api/v1, /health, /openapi.yaml
	// and /swagger routes above take priority; everything else (the app shell,
	// hashed assets, and client-side deep links) is handled by the SPA handler.
	// One container serves SPA + API; the operator fronts it with their own
	// reverse proxy for TLS.
	r.Handle("/*", web.Handler())

	return r
}

// swaggerUI serves a minimal Swagger UI page that loads the spec from
// /openapi.yaml using assets from the unpkg CDN.
func swaggerUI(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(swaggerHTML))
}

const swaggerHTML = `<!DOCTYPE html>
<html lang="pt-BR">
<head>
  <meta charset="UTF-8">
  <title>finance.sh API — Swagger UI</title>
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css">
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js" crossorigin></script>
  <script>
    window.onload = function () {
      window.ui = SwaggerUIBundle({
        url: '/openapi.yaml',
        dom_id: '#swagger-ui',
        deepLinking: true,
        presets: [SwaggerUIBundle.presets.apis]
      });
    };
  </script>
</body>
</html>`
