// Package database owns the GORM connection, migrations and seed data.
package database

import (
	"time"

	"github.com/finance-sh/finance-sh/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// Connect opens the PostgreSQL connection, tunes the logger by environment and
// configures the underlying connection pool.
func Connect(cfg *config.Config) (*gorm.DB, error) {
	logLevel := gormlogger.Warn
	if !cfg.IsProduction() {
		logLevel = gormlogger.Info
	}

	db, err := gorm.Open(postgres.Open(cfg.DB.DSN()), &gorm.Config{
		Logger:                                   gormlogger.Default.LogMode(logLevel),
		PrepareStmt:                              true,
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	// Pool sizing keeps connections bounded under load; tune via env in prod.
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(time.Hour)
	sqlDB.SetConnMaxIdleTime(10 * time.Minute)

	return db, nil
}
