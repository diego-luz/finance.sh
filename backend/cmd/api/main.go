// Command api is the finance.sh HTTP server entrypoint. It wires configuration,
// database, repositories, services and handlers, then serves with graceful
// shutdown.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
	"unicode/utf8"

	"github.com/finance-sh/finance-sh/internal/config"
	"github.com/finance-sh/finance-sh/internal/database"
	"github.com/finance-sh/finance-sh/internal/handlers"
	"github.com/finance-sh/finance-sh/internal/jobs"
	"github.com/finance-sh/finance-sh/internal/repositories"
	"github.com/finance-sh/finance-sh/internal/services"
	"github.com/finance-sh/finance-sh/pkg/cache"
	"github.com/finance-sh/finance-sh/pkg/crypto"
	"github.com/finance-sh/finance-sh/pkg/lockout"
	"github.com/finance-sh/finance-sh/pkg/logger"
	"github.com/finance-sh/finance-sh/pkg/mailer"
	"github.com/finance-sh/finance-sh/pkg/storage"
)

func main() {
	// CLI: no-SMTP password recovery. `/app/api -reset-password <email>` resets
	// that user's password to a generated one, prints it, and exits (no server).
	resetEmail := flag.String("reset-password", "", "reset a user's password (no-SMTP recovery) and exit")
	flag.Parse()

	cfg := config.Load()
	log := logger.New(cfg.Env)

	// Production secret guard: refuse to boot with insecure/default secrets. The
	// dev defaults are safe in development but must never reach production.
	if cfg.IsProduction() {
		var insecure []string
		if cfg.JWT.AccessSecret == "dev-access-secret-change-me" {
			insecure = append(insecure, "JWT_ACCESS_SECRET")
		}
		if cfg.JWT.RefreshSecret == "dev-refresh-secret-change-me" {
			insecure = append(insecure, "JWT_REFRESH_SECRET")
		}
		if cfg.EncryptionKey == "" || cfg.EncryptionKey == crypto.DevDefaultKey {
			insecure = append(insecure, "ENCRYPTION_KEY")
		}
		if len(insecure) > 0 {
			log.Error("refusing to start in production with insecure default secrets; set unique values",
				"variables", insecure)
			os.Exit(1)
		}
	}

	// Initialise field-level encryption (AES-256-GCM). A malformed key is fatal;
	// the dev default logs a WARNING from within pkg/crypto.
	if err := crypto.Init(cfg.EncryptionKey); err != nil {
		log.Error("failed to initialise encryption", "error", err)
		os.Exit(1)
	}

	// Database.
	db, err := database.Connect(cfg)
	if err != nil {
		log.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	// Schema: versioned SQL migrations are the default (production source of
	// truth). Set AUTO_MIGRATE=true to use GORM AutoMigrate instead (handy in
	// dev when iterating on entities before writing a migration).
	if os.Getenv("AUTO_MIGRATE") == "true" {
		if err := database.AutoMigrate(db); err != nil {
			log.Error("failed to run automigrate", "error", err)
			os.Exit(1)
		}
		log.Info("schema applied via AutoMigrate")
	} else {
		if err := database.RunMigrations(db); err != nil {
			log.Error("failed to run migrations", "error", err)
			os.Exit(1)
		}
		log.Info("schema applied via versioned migrations")
	}

	// Password recovery mode: reset the password and exit before starting the
	// server (the no-SMTP escape hatch — Coolify/Miniflux-style).
	if *resetEmail != "" {
		pw, err := database.ResetUserPassword(db, *resetEmail)
		if err != nil {
			log.Error("reset-password failed", "error", err)
			os.Exit(1)
		}
		printResetBanner(*resetEmail, pw)
		os.Exit(0)
	}

	if os.Getenv("SEED") == "true" {
		if err := database.Seed(db); err != nil {
			log.Error("failed to seed database", "error", err)
			os.Exit(1)
		}
		log.Info("database seeded", "demo_user", logger.MaskEmail("demo@finance.sh"))
	}

	// First-boot admin: on a fresh DB (no users) create a super-admin + org so
	// the operator can log in immediately, forced to change the password on first
	// login. No-op once any user exists (idempotent). See config.BootstrapAdmin.
	if cfg.BootstrapAdmin {
		genPw, created, err := database.BootstrapAdmin(db, cfg.AdminEmail, cfg.AdminPassword, cfg.AdminOrgName, cfg.TermsVersion)
		if err != nil {
			log.Error("admin bootstrap failed", "error", err)
			os.Exit(1)
		}
		if created {
			if genPw != "" {
				printAdminBanner(cfg.AdminEmail, genPw)
				log.Warn("default admin created with a generated password — change it on first login",
					"email", cfg.AdminEmail)
			} else {
				log.Info("default admin created (password from ADMIN_PASSWORD); change it on first login",
					"email", cfg.AdminEmail)
			}
		}
	}

	// Dashboard cache: in-memory, process-local. Single API replica in
	// Community Edition makes external caching unnecessary.
	dashCache := cache.New()

	// Login brute-force limiter (in-memory, per-process) and the transactional
	// mailer (logs when SMTP is not configured).
	loginLimiter := lockout.New(cfg.Login.MaxAttempts, cfg.Login.LockoutMin)
	mail := mailer.New(cfg.SMTP, log)

	// Attachment storage lives in Postgres BYTEA — same db handle, no external
	// object store.
	store, err := storage.New(db, log)
	if err != nil {
		log.Warn("failed to initialise attachment storage; attachments will be unavailable", "error", err)
		store = nil
	} else {
		log.Info("attachment storage configured (postgres bytea)", "max_mb", cfg.Storage.MaxAttachmentMB)
	}

	// Repositories.
	userRepo := repositories.NewUserRepository(db)
	accountRepo := repositories.NewAccountRepository(db)
	categoryRepo := repositories.NewCategoryRepository(db)
	categoryRuleRepo := repositories.NewCategoryRuleRepository(db)
	contactRepo := repositories.NewContactRepository(db)
	tagRepo := repositories.NewTagRepository(db)
	txRepo := repositories.NewTransactionRepository(db)
	creditCardRepo := repositories.NewCreditCardRepository(db)
	goalRepo := repositories.NewGoalRepository(db)
	budgetRepo := repositories.NewBudgetRepository(db)
	membershipRepo := repositories.NewMembershipRepository(db)
	organizationRepo := repositories.NewOrganizationRepository(db)
	notificationRepo := repositories.NewNotificationRepository(db)
	passwordResetRepo := repositories.NewPasswordResetRepository(db)
	attachmentRepo := repositories.NewAttachmentRepository(db)
	recurrenceRepo := repositories.NewRecurrenceRuleRepository(db)
	auditRepo := repositories.NewAuditRepository(db)
	adminRepo := repositories.NewAdminRepository(db)

	// Services.
	authSvc := services.NewAuthService(userRepo, passwordResetRepo, cfg, loginLimiter, mail, db)
	lgpdSvc := services.NewLGPDService(db, userRepo)
	accountSvc := services.NewAccountService(accountRepo, dashCache)
	categorySvc := services.NewCategoryService(categoryRepo, dashCache)
	contactSvc := services.NewContactService(contactRepo)
	tagSvc := services.NewTagService(tagRepo)
	txSvc := services.NewTransactionService(txRepo, accountRepo, categoryRepo, creditCardRepo, contactRepo, tagRepo, attachmentRepo, dashCache)
	searchSvc := services.NewSearchService(txRepo, contactRepo, categoryRepo, accountRepo, creditCardRepo, goalRepo)
	dashSvc := services.NewDashboardService(txRepo, accountRepo, dashCache)
	forecastSvc := services.NewForecastService(txRepo, accountRepo)
	creditCardSvc := services.NewCreditCardService(creditCardRepo, txRepo, accountRepo, dashCache)
	goalSvc := services.NewGoalService(goalRepo)
	budgetSvc := services.NewBudgetService(budgetRepo, categoryRepo)
	memberSvc := services.NewMemberService(membershipRepo, userRepo)
	organizationSvc := services.NewOrganizationService(organizationRepo, db)
	notificationSvc := services.NewNotificationService(notificationRepo)
	reportSvc := services.NewReportService(txRepo, accountRepo, categoryRepo, contactRepo, userRepo)
	attachmentSvc := services.NewAttachmentService(attachmentRepo, txRepo, store, cfg)
	// Categorization must be constructed before import: import depends on it
	// (one-way; categorization only depends on repositories — no cycle).
	categorizationSvc := services.NewCategorizationService(categoryRuleRepo, categoryRepo, txRepo, dashCache)
	importSvc := services.NewImportService(txRepo, accountRepo, categoryRepo, categorizationSvc, dashCache)
	recurrenceSvc := services.NewRecurrenceService(recurrenceRepo, txRepo, accountRepo, categoryRepo, contactRepo, dashCache, db)
	auditSvc := services.NewAuditService(auditRepo, userRepo)
	// Platform back-office (super-admin): read-only views + user-safety ops.
	// NOT tenant-scoped.
	adminSvc := services.NewAdminService(adminRepo, userRepo)
	// First-run setup wizard (public bootstrap). Issues its own tokens because
	// the new user has no session yet; bypasses the closed-registration gate.
	setupSvc := services.NewSetupService(userRepo, cfg, db)

	// Router.
	router := handlers.NewRouter(handlers.Deps{
		Config:         cfg,
		Logger:         log,
		DB:             db,
		Users:          userRepo,
		Auth:           authSvc,
		Account:        accountSvc,
		Category:       categorySvc,
		Categorization: categorizationSvc,
		Contact:        contactSvc,
		Tag:            tagSvc,
		Tx:             txSvc,
		Search:         searchSvc,
		Dashboard:      dashSvc,
		Forecast:       forecastSvc,
		CreditCard:     creditCardSvc,
		Goal:           goalSvc,
		Budget:         budgetSvc,
		Member:         memberSvc,
		Organization:   organizationSvc,
		Notification:   notificationSvc,
		Report:         reportSvc,
		LGPD:           lgpdSvc,
		Attachment:     attachmentSvc,
		Import:         importSvc,
		Recurrence:     recurrenceSvc,
		Audit:          auditSvc,
		Admin:          adminSvc,
		Setup:          setupSvc,
		DocsDir:        os.Getenv("DOCS_DIR"),
	})

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	// Root context cancelled on SIGINT/SIGTERM. The in-process scheduler honours
	// it so it exits cleanly together with the HTTP server.
	ctx, cancelRoot := context.WithCancel(context.Background())
	defer cancelRoot()

	// Background scheduler (recurring transactions, bill notifications, retention
	// purge) runs in-process by default — the single app binary owns it, no
	// separate worker container. Set JOBS_IN_PROCESS=false to disable it (e.g.
	// running multiple app replicas where only one should drive the scheduler).
	if cfg.JobsInProcess {
		interval := time.Duration(cfg.JobsIntervalSec) * time.Second
		runner := jobs.NewRunner(db, cfg.RetentionDays, log)
		log.Info("in-process scheduler enabled", "interval_sec", int(interval.Seconds()))
		go runScheduler(ctx, runner, interval, log)
	} else {
		log.Info("in-process scheduler disabled (JOBS_IN_PROCESS=false)")
	}

	// Start the server in a goroutine so we can wait for a shutdown signal.
	go func() {
		log.Info("starting server", "env", cfg.Env, "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	// Graceful shutdown on SIGINT/SIGTERM.
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Info("shutting down server")
	// Cancel the root context so the in-process scheduler stops alongside HTTP.
	cancelRoot()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("graceful shutdown failed", "error", err)
	}
	log.Info("server stopped")
}

// printAdminBanner writes the auto-generated first-boot admin credentials to
// stdout in a box that stands out from structured logs. It is shown ONCE, only
// when the password was generated (not when set via ADMIN_PASSWORD), so the
// operator can grab it from the boot output and log in. The account is forced to
// change its password on first login.
func printAdminBanner(email, password string) {
	const w = 60
	line := strings.Repeat("─", w)
	pad := func(s string) string {
		n := utf8.RuneCountInString(s)
		if n > w-2 {
			s = string([]rune(s)[:w-2])
			n = w - 2
		}
		return "│ " + s + strings.Repeat(" ", w-2-n) + " │"
	}
	fmt.Fprintln(os.Stdout, "┌"+line+"┐")
	fmt.Fprintln(os.Stdout, pad("ADMIN CRIADO — troque a senha no 1º login"))
	fmt.Fprintln(os.Stdout, pad(""))
	fmt.Fprintln(os.Stdout, pad("  email:  "+email))
	fmt.Fprintln(os.Stdout, pad("  senha:  "+password+"  (aleatória)"))
	fmt.Fprintln(os.Stdout, pad(""))
	fmt.Fprintln(os.Stdout, pad("Defina ADMIN_PASSWORD no .env para fixar."))
	fmt.Fprintln(os.Stdout, "└"+line+"┘")
}

// printResetBanner shows the new password set by `-reset-password` in a box that
// stands out from structured logs. The user must change it on next login.
func printResetBanner(email, password string) {
	const w = 60
	line := strings.Repeat("─", w)
	pad := func(s string) string {
		n := utf8.RuneCountInString(s)
		if n > w-2 {
			s = string([]rune(s)[:w-2])
			n = w - 2
		}
		return "│ " + s + strings.Repeat(" ", w-2-n) + " │"
	}
	fmt.Fprintln(os.Stdout, "┌"+line+"┐")
	fmt.Fprintln(os.Stdout, pad("SENHA REDEFINIDA — troque no próximo login"))
	fmt.Fprintln(os.Stdout, pad(""))
	fmt.Fprintln(os.Stdout, pad("  email:  "+email))
	fmt.Fprintln(os.Stdout, pad("  senha:  "+password+"  (aleatória)"))
	fmt.Fprintln(os.Stdout, "└"+line+"┘")
}

// runScheduler drives jobs.Runner on a ticker, colocating the scheduler with the
// API in the single app binary. It runs once immediately so a fresh deploy does
// not wait a full interval, then ticks until ctx is cancelled by SIGINT/SIGTERM.
func runScheduler(ctx context.Context, runner *jobs.Runner, interval time.Duration, log *slog.Logger) {
	log.Info("scheduler: initial run")
	runner.RunAll()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			runner.RunAll()
		case <-ctx.Done():
			log.Info("scheduler: shutting down")
			return
		}
	}
}
