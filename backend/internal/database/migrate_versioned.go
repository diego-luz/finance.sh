package database

import (
	"embed"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	migratepg "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"gorm.io/gorm"
)

// migrationFiles embeds the versioned SQL migrations. They are the source of
// truth for the schema in production; AutoMigrate remains as a dev convenience.
//
//go:embed migrations/*.sql
var migrationFiles embed.FS

// RunMigrations applies all pending versioned migrations using the existing
// GORM connection (no second pool). It is a no-op when already up to date.
func RunMigrations(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("obter *sql.DB: %w", err)
	}

	driver, err := migratepg.WithInstance(sqlDB, &migratepg.Config{})
	if err != nil {
		return fmt.Errorf("driver de migração: %w", err)
	}

	src, err := iofs.New(migrationFiles, "migrations")
	if err != nil {
		return fmt.Errorf("fonte de migração: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", src, "postgres", driver)
	if err != nil {
		return fmt.Errorf("instância de migração: %w", err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("aplicar migrações: %w", err)
	}
	return nil
}
