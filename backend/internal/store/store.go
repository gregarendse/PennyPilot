package store

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pennypilot/pennypilot/backend/internal/config"
	"github.com/pennypilot/pennypilot/backend/migrations"
)

// Store will own database access for accounts, transactions, budgets, and encrypted credentials.
type Store struct {
	db     *sql.DB
	logger *slog.Logger
	cfg    config.Config
}

// New initializes a new Store with the given configuration and logger.
func New(cfg config.Config, logger *slog.Logger) (*Store, error) {
	driver := cfg.DatabaseDriver
	dsn := cfg.DatabaseURL

	db, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	return &Store{
		db:     db,
		logger: logger,
		cfg:    cfg,
	}, nil
}

// Migrate applies pending migrations to the database.
func (s *Store) Migrate() error {
	migrationPath := "."
	if s.cfg.DatabaseDriver == "sqlite3" {
		migrationPath = "sqlite"
	}

	dsn := s.cfg.DatabaseURL
	if s.cfg.DatabaseDriver == "sqlite3" {
		// Ensure it has the sqlite3:// prefix if not present for the migrate tool
		if !strings.HasPrefix(dsn, "sqlite3://") {
			dsn = "sqlite3://" + dsn
		}
	}

	migrationSource, err := iofs.New(migrations.FS, migrationPath)
	if err != nil {
		return fmt.Errorf("failed to load embedded migrations: %w", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", migrationSource, dsn)
	if err != nil {
		return fmt.Errorf("failed to initialize migrations: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	s.logger.Info("migrations applied successfully")
	return nil
}

// Ping checks the database connection.
func (s *Store) Ping(ctx context.Context) error {
	return s.db.PingContext(ctx)
}

// Close closes the database connection.
func (s *Store) Close() error {
	return s.db.Close()
}
