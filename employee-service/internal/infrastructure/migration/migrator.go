package migration

import (
	"database/sql"
	"embed"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

type Migrator struct {
	db     *sql.DB
	logger *log.Logger
}

func NewMigrator(db *sql.DB, logger *log.Logger) *Migrator {
	if logger == nil {
		logger = log.Default()
	}
	return &Migrator{
		db:     db,
		logger: logger,
	}
}

// RunMigrations выполняет все миграции
func (m *Migrator) RunMigrations() error {
	m.logger.Println("Starting database migrations...")

	// Создаем драйвер для PostgreSQL
	driver, err := postgres.WithInstance(m.db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create postgres driver: %w", err)
	}

	// Создаем источник миграций из embedded файлов
	sourceDriver, err := iofs.New(migrationFiles, "migrations")
	if err != nil {
		return fmt.Errorf("failed to create migration source: %w", err)
	}

	// Создаем мигратор
	migrator, err := migrate.NewWithInstance("iofs", sourceDriver, "postgres", driver)
	if err != nil {
		return fmt.Errorf("failed to create migrator: %w", err)
	}
	defer migrator.Close()

	// Получаем текущую версию
	currentVersion, dirty, err := migrator.Version()
	if err != nil && err != migrate.ErrNilVersion {
		return fmt.Errorf("failed to get current migration version: %w", err)
	}

	if dirty {
		m.logger.Printf("Database is in dirty state at version %d, attempting to fix...", currentVersion)
		if err := migrator.Force(int(currentVersion)); err != nil {
			return fmt.Errorf("failed to force migration version: %w", err)
		}
	}

	// Выполняем миграции
	err = migrator.Up()
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	// Получаем финальную версию
	finalVersion, _, err := migrator.Version()
	if err != nil && err != migrate.ErrNilVersion {
		return fmt.Errorf("failed to get final migration version: %w", err)
	}

	if err == migrate.ErrNoChange {
		m.logger.Printf("Database is up to date at version %d", currentVersion)
	} else {
		m.logger.Printf("Successfully migrated database from version %d to %d", currentVersion, finalVersion)
	}

	return nil
}

// WaitForDatabase ждет доступности базы данных
func (m *Migrator) WaitForDatabase() error {
	m.logger.Println("Waiting for database connection...")
	
	for i := 0; i < 30; i++ {
		if err := m.db.Ping(); err == nil {
			m.logger.Println("Database connection established")
			return nil
		}
		m.logger.Printf("Database not ready, attempt %d/30...", i+1)
	}
	
	return fmt.Errorf("database connection timeout after 30 attempts")
}